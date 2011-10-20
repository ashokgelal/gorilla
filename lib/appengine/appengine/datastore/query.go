// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"

	"appengine"
	"goprotobuf.googlecode.com/hg/proto"

	pb "appengine_internal/datastore"
)

type operator int

const (
	lessThan operator = iota
	lessEq
	equal
	greaterEq
	greaterThan
)

var operatorToProto = map[operator]*pb.Query_Filter_Operator{
	lessThan:    pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN),
	lessEq:      pb.NewQuery_Filter_Operator(pb.Query_Filter_LESS_THAN_OR_EQUAL),
	equal:       pb.NewQuery_Filter_Operator(pb.Query_Filter_EQUAL),
	greaterEq:   pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN_OR_EQUAL),
	greaterThan: pb.NewQuery_Filter_Operator(pb.Query_Filter_GREATER_THAN),
}

// filter is a conditional filter on query results.
type filter struct {
	FieldName string
	Op        operator
	Value     interface{}
}

type sortDirection int

const (
	ascending sortDirection = iota
	descending
)

var sortDirectionToProto = map[sortDirection]*pb.Query_Order_Direction{
	ascending:  pb.NewQuery_Order_Direction(pb.Query_Order_ASCENDING),
	descending: pb.NewQuery_Order_Direction(pb.Query_Order_DESCENDING),
}

// order is a sort order on query results.
type order struct {
	FieldName string
	Direction sortDirection
}

// NewQuery creates a new Query for a specific entity kind.
// The kind must be non-empty.
func NewQuery(kind string) *Query {
	q := &Query{kind: kind}
	if kind == "" {
		q.err = os.NewError("datastore: empty kind")
	}
	return q
}

// Query represents a datastore query.
type Query struct {
	kind     string
	ancestor *Key
	filter   []filter
	order    []order

	keysOnly bool
	limit    int32
	offset   int32

	err os.Error
}

// Ancestor sets the ancestor filter for the Query.
// The ancestor should not be nil.
func (q *Query) Ancestor(ancestor *Key) *Query {
	if ancestor == nil {
		q.err = os.NewError("datastore: nil query ancestor")
		return q
	}
	q.ancestor = ancestor
	return q
}

// Filter adds a field-based filter to the Query.
// The filterStr argument must be a field name followed by optional space,
// followed by an operator, one of ">", "<", ">=", "<=", or "=".
// Fields are compared against the provided value using the operator.
// Multiple filters are AND'ed together.
// The Query is updated in place and returned for ease of chaining.
func (q *Query) Filter(filterStr string, value interface{}) *Query {
	filterStr = strings.TrimSpace(filterStr)
	if len(filterStr) < 1 {
		q.err = os.NewError("datastore: invalid filter: " + filterStr)
		return q
	}
	f := filter{
		FieldName: strings.TrimRight(filterStr, " ><="),
		Value:     value,
	}
	switch op := strings.TrimSpace(filterStr[len(f.FieldName):]); op {
	case "<=":
		f.Op = lessEq
	case ">=":
		f.Op = greaterEq
	case "<":
		f.Op = lessThan
	case ">":
		f.Op = greaterThan
	case "=":
		f.Op = equal
	default:
		q.err = fmt.Errorf("datastore: invalid operator %q in filter %q", op, filterStr)
		return q
	}
	q.filter = append(q.filter, f)
	return q
}

// Order adds a field-based sort to the query.
// Orders are applied in the order they are added.
// The default order is ascending; to sort in descending
// order prefix the fieldName with a minus sign (-).
func (q *Query) Order(fieldName string) *Query {
	fieldName = strings.TrimSpace(fieldName)
	o := order{Direction: ascending, FieldName: fieldName}
	if strings.HasPrefix(fieldName, "-") {
		o.Direction = descending
		o.FieldName = strings.TrimSpace(fieldName[1:])
	} else if strings.HasPrefix(fieldName, "+") {
		q.err = fmt.Errorf("datastore: invalid order: %q", fieldName)
		return q
	}
	if len(o.FieldName) == 0 {
		q.err = os.NewError("datastore: empty order")
		return q
	}
	q.order = append(q.order, o)
	return q
}

// KeysOnly configures the query to return just keys,
// instead of keys and entities.
func (q *Query) KeysOnly() *Query {
	q.keysOnly = true
	return q
}

// Limit sets the maximum number of keys/entities to return.
// A zero value means unlimited. A negative value is invalid.
func (q *Query) Limit(limit int) *Query {
	if limit < 0 {
		q.err = os.NewError("datastore: negative query limit")
		return q
	}
	if limit > math.MaxInt32 {
		q.err = os.NewError("datastore: query limit overflow")
		return q
	}
	q.limit = int32(limit)
	return q
}

// Offset sets how many keys to skip over before returning results.
// A negative value is invalid.
func (q *Query) Offset(offset int) *Query {
	if offset < 0 {
		q.err = os.NewError("datastore: negative query offset")
		return q
	}
	if offset > math.MaxInt32 {
		q.err = os.NewError("datastore: query offset overflow")
		return q
	}
	q.offset = int32(offset)
	return q
}

// zeroLimitPolicy defines how to interpret a zero query/cursor limit. In some
// contexts, it means an unlimited query (to follow Go's idiom of a zero value
// being a useful default value). In other contexts, it means a literal zero,
// such as when issuing a query count, no actual entity data is wanted, only
// the number of skipped results.
type zeroLimitPolicy int

const (
	zeroLimitMeansUnlimited zeroLimitPolicy = iota
	zeroLimitMeansZero
)

// toProto converts the query to a protocol buffer.
func (q *Query) toProto(dst *pb.Query, appID string, zlp zeroLimitPolicy) os.Error {
	if q.kind == "" {
		return os.NewError("datastore: empty query kind")
	}
	dst.Reset()
	dst.App = proto.String(appID)
	dst.Kind = proto.String(q.kind)
	if q.ancestor != nil {
		dst.Ancestor = keyToProto(appID, q.ancestor)
	}
	if q.keysOnly {
		dst.KeysOnly = proto.Bool(true)
		dst.RequirePerfectPlan = proto.Bool(true)
	}
	for _, qf := range q.filter {
		if qf.FieldName == "" {
			return os.NewError("datastore: empty query filter field name")
		}
		p, errStr := valueToProto(appID, qf.FieldName, reflect.ValueOf(qf.Value), false)
		if errStr != "" {
			return os.NewError("datastore: bad query filter value type: " + errStr)
		}
		xf := &pb.Query_Filter{
			Op:       operatorToProto[qf.Op],
			Property: []*pb.Property{p},
		}
		if xf.Op == nil {
			return os.NewError("datastore: unknown query filter operator")
		}
		dst.Filter = append(dst.Filter, xf)
	}
	for _, qo := range q.order {
		if qo.FieldName == "" {
			return os.NewError("datastore: empty query order field name")
		}
		xo := &pb.Query_Order{
			Property:  proto.String(qo.FieldName),
			Direction: sortDirectionToProto[qo.Direction],
		}
		if xo.Direction == nil {
			return os.NewError("datastore: unknown query order direction")
		}
		dst.Order = append(dst.Order, xo)
	}
	if q.limit != 0 || zlp == zeroLimitMeansZero {
		dst.Limit = proto.Int32(q.limit)
	}
	if q.offset != 0 {
		dst.Offset = proto.Int32(q.offset)
	}
	return nil
}

// Count returns the number of results for the query.
func (q *Query) Count(c appengine.Context) (int, os.Error) {
	// Check that the query is well-formed.
	if q.err != nil {
		return 0, q.err
	}

	// Run a copy of the query, with keysOnly true, and an adjusted offset.
	// We also set the limit to zero, as we don't want any actual entity data,
	// just the number of skipped results.
	newQ := *q
	newQ.keysOnly = true
	newQ.limit = 0
	if q.limit == 0 {
		// If the original query was unlimited, set the new query's offset to maximum.
		newQ.offset = math.MaxInt32
	} else {
		newQ.offset = q.offset + q.limit
		if newQ.offset < 0 {
			// Do the best we can, in the presence of overflow.
			newQ.offset = math.MaxInt32
		}
	}
	req := &pb.Query{}
	if err := newQ.toProto(req, c.FullyQualifiedAppID(), zeroLimitMeansZero); err != nil {
		return 0, err
	}
	res := &pb.QueryResult{}
	if err := c.Call("datastore_v3", "RunQuery", req, res, nil); err != nil {
		return 0, err
	}

	// n is the count we will return. For example, suppose that our original
	// query had an offset of 4 and a limit of 2008: the count will be 2008,
	// provided that there are at least 2012 matching entities. However, the
	// RPCs will only skip 1000 results at a time. The RPC sequence is:
	//   call RunQuery with (offset, limit) = (2012, 0)  // 2012 == newQ.offset
	//   response has (skippedResults, moreResults) = (1000, true)
	//   n += 1000  // n == 1000
	//   call Next     with (offset, limit) = (1012, 0)  // 1012 == newQ.offset - n
	//   response has (skippedResults, moreResults) = (1000, true)
	//   n += 1000  // n == 2000
	//   call Next     with (offset, limit) = (12, 0)    // 12 == newQ.offset - n
	//   response has (skippedResults, moreResults) = (12, false)
	//   n += 12    // n == 2012
	//   // exit the loop
	//   n -= 4     // n == 2008
	var n int32
	for {
		// The QueryResult should have no actual entity data, just skipped results.
		if len(res.Result) != 0 {
			return 0, os.NewError("datastore: internal error: Count request returned too much data")
		}
		n += proto.GetInt32(res.SkippedResults)
		if !proto.GetBool(res.MoreResults) {
			break
		}
		if err := callNext(c, res, newQ.offset-n, 0, zeroLimitMeansZero); err != nil {
			return 0, err
		}
	}
	n -= q.offset
	if n < 0 {
		// If the offset was greater than the number of matching entities,
		// return 0 instead of negative.
		n = 0
	}
	return int(n), nil
}

// callNext issues a datastore_v3/Next RPC to advance a cursor, such as that
// returned by a query with more results.
func callNext(c appengine.Context, res *pb.QueryResult, offset, limit int32, zlp zeroLimitPolicy) os.Error {
	if res.Cursor == nil {
		return os.NewError("datastore: internal error: server did not return a cursor")
	}
	// TODO: should I eventually call datastore_v3/DeleteCursor on the cursor?
	req := &pb.NextRequest{
		Cursor: res.Cursor,
		Offset: proto.Int32(offset),
	}
	if limit != 0 || zlp == zeroLimitMeansZero {
		req.Count = proto.Int32(limit)
	}
	if res.CompiledCursor != nil {
		req.Compile = proto.Bool(true)
	}
	res.Reset()
	return c.Call("datastore_v3", "Next", req, res, nil)
}

// GetAll runs the query in the given context and returns all keys that match
// that query, as well as appending the values to dst.
// The dst must be a pointer to a slice of structs, struct pointers, or Maps.
// If q is a ``keys-only'' query, GetAll ignores dst and only returns the keys.
func (q *Query) GetAll(c appengine.Context, dst interface{}) ([]*Key, os.Error) {
	var (
		dv       reflect.Value
		et       reflect.Type
		isMap    bool
		isStruct bool
	)
	if !q.keysOnly {
		dv = reflect.ValueOf(dst)
		if dv.Kind() != reflect.Ptr || dv.Elem().Kind() != reflect.Slice {
			return nil, ErrInvalidEntityType
		}
		if dv.IsNil() {
			return nil, ErrInvalidEntityType
		}
		dv = dv.Elem()

		et = dv.Type().Elem()
		switch {
		case et == reflect.TypeOf(Map(nil)):
			isMap = true
		case et.Kind() == reflect.Ptr && et.Elem().Kind() == reflect.Struct:
			et = et.Elem()
		case et.Kind() == reflect.Struct:
			isStruct = true
		default:
			return nil, ErrInvalidEntityType
		}
	}

	var keys []*Key
	for t := q.Run(c); ; {
		k, e, err := t.next()
		if err == Done {
			break
		}
		if err != nil {
			return keys, err
		}
		if !q.keysOnly {
			var ev reflect.Value
			if isMap {
				ev = reflect.ValueOf(make(Map))
			} else {
				ev = reflect.New(et)
			}
			if _, err = loadEntity(ev.Interface(), k, e); err != nil {
				return keys, err
			}
			if isStruct {
				ev = ev.Elem()
			}
			dv.Set(reflect.Append(dv, ev))
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// Run runs the query in the given context.
func (q *Query) Run(c appengine.Context) *Iterator {
	if q.err != nil {
		return &Iterator{err: q.err}
	}
	t := &Iterator{
		c:      c,
		offset: q.offset,
		limit:  q.limit,
	}
	var req pb.Query
	if err := q.toProto(&req, c.FullyQualifiedAppID(), zeroLimitMeansUnlimited); err != nil {
		t.err = err
		return t
	}
	if err := c.Call("datastore_v3", "RunQuery", &req, &t.res, nil); err != nil {
		t.err = err
		return t
	}
	return t
}

// Iterator is the result of running a query.
type Iterator struct {
	c      appengine.Context
	offset int32
	limit  int32
	res    pb.QueryResult
	err    os.Error
}

// Done is returned when a query iteration has completed.
var Done = os.NewError("datastore: query has no more results")

// Next returns the key of the next result. When there are no more results,
// Done is returned as the error.
// If the query is not keys only, it also loads the entity
// stored for that key into the struct pointer or Map dst, with the same
// semantics and possible errors as for the Get function.
// If the query is keys only, it is valid to pass a nil interface{} for dst.
func (t *Iterator) Next(dst interface{}) (*Key, os.Error) {
	k, e, err := t.next()
	if err != nil || e == nil {
		return k, err
	}
	return loadEntity(dst, k, e)
}

func (t *Iterator) next() (*Key, *pb.EntityProto, os.Error) {
	if t.err != nil {
		return nil, nil, t.err
	}

	// Issue datastore_v3/Next RPCs as necessary.
	for len(t.res.Result) == 0 {
		if !proto.GetBool(t.res.MoreResults) {
			t.err = Done
			return nil, nil, t.err
		}
		t.offset -= proto.GetInt32(t.res.SkippedResults)
		if t.offset < 0 {
			t.offset = 0
		}
		if err := callNext(t.c, &t.res, t.offset, t.limit, zeroLimitMeansUnlimited); err != nil {
			t.err = err
			return nil, nil, t.err
		}
		// For an Iterator, a zero limit means unlimited.
		if t.limit == 0 {
			continue
		}
		t.limit -= int32(len(t.res.Result))
		if t.limit > 0 {
			continue
		}
		t.limit = 0
		if proto.GetBool(t.res.MoreResults) {
			t.err = os.NewError("datastore: internal error: limit exhausted but more_results is true")
			return nil, nil, t.err
		}
	}

	// Pop the EntityProto from the front of t.res.Result and
	// extract its key.
	var e *pb.EntityProto
	e, t.res.Result = t.res.Result[0], t.res.Result[1:]
	if e.Key == nil {
		return nil, nil, os.NewError("datastore: internal error: server did not return a key")
	}
	k, err := protoToKey(e.Key)
	if err != nil || k.Incomplete() {
		return nil, nil, os.NewError("datastore: internal error: server returned an invalid key")
	}
	if proto.GetBool(t.res.KeysOnly) {
		return k, nil, nil
	}
	return k, e, nil
}

// loadEntity loads an EntityProto into a Map or struct.
func loadEntity(dst interface{}, k *Key, e *pb.EntityProto) (*Key, os.Error) {
	if m, ok := dst.(Map); ok {
		return k, loadMap(m, k, e)
	}
	sv, err := asStructValue(dst)
	if err != nil {
		return nil, err
	}
	return k, loadStruct(sv, k, e)
}
