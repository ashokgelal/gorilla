// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"appengine"
	"appengine_internal"
	"goprotobuf.googlecode.com/hg/proto"

	pb "appengine_internal/datastore"
)

// Entities with more than this many indexed properties will not be saved.
const maxIndexedProperties = 5000

// []byte fields more than 1 megabyte long will not be loaded or saved.
const maxBlobLen = 1 << 20

// Time is the number of microseconds since the Unix epoch,
// January 1, 1970 00:00:00 UTC.
//
// It is a distinct type so that loading and saving fields of type Time are
// displayed correctly in App Engine tools like the Admin Console.
type Time int64

// SecondsToTime converts an int64 number of seconds since to Unix epoch
// to a Time value.
func SecondsToTime(n int64) Time {
	return Time(n * 1e6)
}

// Time returns a *time.Time from a datastore time.
func (t Time) Time() *time.Time {
	// TODO: once App Engine has release.r60 or later,
	// support subseconds here.  Currently we just drop them.
	return time.SecondsToUTC(int64(t) / 1e6)
}

// Map is a map representation of an entity's fields. It is more flexible than
// but not as strongly typed as a struct representation.
type Map map[string]interface{}

var (
	// ErrInvalidEntityType is returned when an invalid destination entity type
	// is passed to Get, GetAll, GetMulti or Next.
	ErrInvalidEntityType = os.NewError("datastore: invalid entity type")
	// ErrInvalidKey is returned when an invalid key is presented.
	ErrInvalidKey = os.NewError("datastore: invalid key")
	// ErrNoSuchEntity is returned when no entity was found for a given key.
	ErrNoSuchEntity = os.NewError("datastore: no such entity")
)

// ErrFieldMismatch is returned when a field is to be loaded into a different
// type than the one it was stored from, or when a field is missing or
// unexported in the destination struct.
// StructType is the type of the struct pointed to by the destination argument
// passed to Get or to Iterator.Next.
type ErrFieldMismatch struct {
	Key        *Key
	StructType reflect.Type
	FieldName  string
	Reason     string
}

// String returns a string representation of the error.
func (e *ErrFieldMismatch) String() string {
	return fmt.Sprintf("datastore: cannot load field %q from key %q into a %q: %s",
		e.FieldName, e.Key, e.StructType, e.Reason)
}

// ErrMulti indicates that a batch operation failed on at least one element.
type ErrMulti []os.Error

// String returns a string representation of the error.
func (m ErrMulti) String() string {
	s, n := "", 0
	for _, e := range m {
		if e == nil {
			continue
		}
		if n == 0 {
			s = e.String()
		}
		n++
	}
	switch n {
	case 0:
		return "(0 errors)"
	case 1:
		return s
	case 2:
		return s + " (and 1 other error)"
	}
	return fmt.Sprintf("%s (and %d other errors)", s, n-1)
}

// protoToKey converts a Reference proto to a *Key.
func protoToKey(r *pb.Reference) (k *Key, err os.Error) {
	appID := proto.GetString(r.App)
	for _, e := range r.Path.Element {
		k = &Key{
			kind:     proto.GetString(e.Type),
			stringID: proto.GetString(e.Name),
			intID:    proto.GetInt64(e.Id),
			parent:   k,
			appID:    appID,
		}
		if !k.valid() {
			return nil, ErrInvalidKey
		}
	}
	return
}

// keyToProto converts a *Key to a Reference proto.
func keyToProto(defaultAppID string, k *Key) *pb.Reference {
	appID := k.appID
	if appID == "" {
		appID = defaultAppID
	}
	n := 0
	for i := k; i != nil; i = i.parent {
		n++
	}
	e := make([]*pb.Path_Element, n)
	for i := k; i != nil; i = i.parent {
		n--
		e[n] = &pb.Path_Element{
			Type: &i.kind,
		}
		// At most one of {Name,Id} should be set.
		// Neither will be set for incomplete keys.
		if i.stringID != "" {
			e[n].Name = &i.stringID
		} else if i.intID != 0 {
			e[n].Id = &i.intID
		}
	}
	return &pb.Reference{
		App: proto.String(appID),
		Path: &pb.Path{
			Element: e,
		},
	}
}

// multiKeyToProto is a batch version of keyToProto.
func multiKeyToProto(appID string, key []*Key) []*pb.Reference {
	ret := make([]*pb.Reference, len(key))
	for i, k := range key {
		ret[i] = keyToProto(appID, k)
	}
	return ret
}

// multiValid is a batch version of Key.valid. It returns an os.Error, not a
// []bool.
func multiValid(key []*Key) os.Error {
	invalid := false
	for _, k := range key {
		if !k.valid() {
			invalid = true
			break
		}
	}
	if !invalid {
		return nil
	}
	err := make(ErrMulti, len(key))
	for i, k := range key {
		if !k.valid() {
			err[i] = ErrInvalidKey
		}
	}
	return err
}

// It's unfortunate that the two semantically equivalent concepts pb.Reference
// and pb.PropertyValue_ReferenceValue aren't the same type. For example, the
// two have different protobuf field numbers.

// referenceValueToKey is the same as protoToKey except the input is a
// PropertyValue_ReferenceValue instead of a Reference.
func referenceValueToKey(r *pb.PropertyValue_ReferenceValue) (k *Key, err os.Error) {
	appID := proto.GetString(r.App)
	for _, e := range r.Pathelement {
		k = &Key{
			kind:     proto.GetString(e.Type),
			stringID: proto.GetString(e.Name),
			intID:    proto.GetInt64(e.Id),
			parent:   k,
			appID:    appID,
		}
		if !k.valid() {
			return nil, ErrInvalidKey
		}
	}
	return
}

// keyToReferenceValue is the same as keyToProto except the output is a
// PropertyValue_ReferenceValue instead of a Reference.
func keyToReferenceValue(defaultAppID string, k *Key) *pb.PropertyValue_ReferenceValue {
	ref := keyToProto(defaultAppID, k)
	pe := make([]*pb.PropertyValue_ReferenceValue_PathElement, len(ref.Path.Element))
	for i, e := range ref.Path.Element {
		pe[i] = &pb.PropertyValue_ReferenceValue_PathElement{
			Type: e.Type,
			Id:   e.Id,
			Name: e.Name,
		}
	}
	return &pb.PropertyValue_ReferenceValue{
		App:         ref.App,
		Pathelement: pe,
	}
}

// asStructValue converts a pointer-to-struct to a reflect.Value.
func asStructValue(x interface{}) (reflect.Value, os.Error) {
	pv := reflect.ValueOf(x)
	if pv.Kind() != reflect.Ptr || pv.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, ErrInvalidEntityType
	}
	return pv.Elem(), nil
}

// Get loads the entity stored for k into dst, which may be either a struct
// pointer or a Map. If there is no such entity for the key, Get returns
// ErrNoSuchEntity.
//
// The values of dst's unmatched struct fields or Map entries are not modified.
// In particular, it is recommended to pass either a pointer to a zero valued
// struct or an empty Map on each Get call.
//
// ErrFieldMismatch is returned when a field is to be loaded into a different
// type than the one it was stored from, or when a field is missing or
// unexported in the destination struct. ErrFieldMismatch is only returned if
// dst is a struct pointer.
func Get(c appengine.Context, key *Key, dst interface{}) os.Error {
	err := GetMulti(c, []*Key{key}, []interface{}{dst})
	if errMulti, ok := err.(ErrMulti); ok {
		return errMulti[0]
	}
	return err
}

// GetMulti is a batch version of Get.
func GetMulti(c appengine.Context, key []*Key, dst []interface{}) os.Error {
	if len(key) != len(dst) {
		return os.NewError("datastore: key and dst slices have different length")
	}
	if len(key) == 0 {
		return nil
	}
	if err := multiValid(key); err != nil {
		return err
	}
	req := &pb.GetRequest{
		Key: multiKeyToProto(c.FullyQualifiedAppID(), key),
	}
	res := &pb.GetResponse{}
	err := c.Call("datastore_v3", "Get", req, res, nil)
	if err != nil {
		return err
	}
	if len(key) != len(res.Entity) {
		return os.NewError("datastore: internal error: server returned the wrong number of entities")
	}
	errMulti := make(ErrMulti, len(key))
	for i, e := range res.Entity {
		if e.Entity == nil {
			errMulti[i] = ErrNoSuchEntity
			continue
		}
		if m, ok := dst[i].(Map); ok {
			errMulti[i] = loadMap(m, key[i], e.Entity)
			continue
		}
		sv, err := asStructValue(dst[i])
		if err != nil {
			errMulti[i] = err
			continue
		}
		errMulti[i] = loadStruct(sv, key[i], e.Entity)
	}
	for _, e := range errMulti {
		if e != nil {
			return errMulti
		}
	}
	return nil
}

// Put saves the entity src into the datastore with key k. src may be either a
// struct pointer or a Map; if the former then any unexported fields of that
// struct will be skipped.
// If k is an incomplete key, the returned key will be a unique key
// generated by the datastore.
func Put(c appengine.Context, key *Key, src interface{}) (*Key, os.Error) {
	k, err := PutMulti(c, []*Key{key}, []interface{}{src})
	if err != nil {
		if errMulti, ok := err.(ErrMulti); ok {
			return nil, errMulti[0]
		}
		return nil, err
	}
	return k[0], nil
}

// PutMulti is a batch version of Put.
func PutMulti(c appengine.Context, key []*Key, src []interface{}) ([]*Key, os.Error) {
	if len(key) != len(src) {
		return nil, os.NewError("datastore: key and src slices have different length")
	}
	if len(key) == 0 {
		return nil, nil
	}
	appID := c.FullyQualifiedAppID()
	if err := multiValid(key); err != nil {
		return nil, err
	}
	req := &pb.PutRequest{}
	for i, sIface := range src {
		if m, ok := sIface.(Map); ok {
			sProto, err := saveMap(appID, key[i], m)
			if err != nil {
				return nil, err
			}
			req.Entity = append(req.Entity, sProto)
		} else {
			sv, err := asStructValue(sIface)
			if err != nil {
				return nil, err
			}
			sProto, err := saveStruct(appID, key[i], sv)
			if err != nil {
				return nil, err
			}
			req.Entity = append(req.Entity, sProto)
		}
	}
	res := &pb.PutResponse{}
	err := c.Call("datastore_v3", "Put", req, res, nil)
	if err != nil {
		return nil, err
	}
	if len(key) != len(res.Key) {
		return nil, os.NewError("datastore: internal error: server returned the wrong number of keys")
	}
	ret := make([]*Key, len(key))
	for i := range ret {
		ret[i], err = protoToKey(res.Key[i])
		if err != nil || ret[i].Incomplete() {
			return nil, os.NewError("datastore: internal error: server returned an invalid key")
		}
	}
	return ret, nil
}

// Delete deletes the entity for the given key.
func Delete(c appengine.Context, key *Key) os.Error {
	err := DeleteMulti(c, []*Key{key})
	if errMulti, ok := err.(ErrMulti); ok {
		return errMulti[0]
	}
	return err
}

// DeleteMulti is a batch version of Delete.
func DeleteMulti(c appengine.Context, key []*Key) os.Error {
	if len(key) == 0 {
		return nil
	}
	if err := multiValid(key); err != nil {
		return err
	}
	req := &pb.DeleteRequest{
		Key: multiKeyToProto(c.FullyQualifiedAppID(), key),
	}
	res := &pb.DeleteResponse{}
	return c.Call("datastore_v3", "Delete", req, res, nil)
}

func init() {
	appengine_internal.RegisterErrorCodeMap("datastore_v3", pb.Error_ErrorCode_name)
}
