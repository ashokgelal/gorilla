// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"
	"os"
)

// ----------------------------------------------------------------------------
// Node
// ----------------------------------------------------------------------------

// Node represents a node in the schema.
//
// A node stores a name, type, parent and children nodes, optional filter and
// extra attributes.
type Node struct {
	name      string
	typ       NodeType
	filter    NodeFilter
	validator NodeValidator
	parent    *Node
	children  []*Node
	attrs     map[string]string
}

// NewNode returns a new node with the given name and type.
func NewNode(name string, typ NodeType) *Node {
	if name == "" {
		// TODO generate a dummy name?
		panic("Node name is required.")
	}
	return &Node{
		name: name,
		typ:  typ,
	}
}

// Clone clones this node and, if deep is true, all its child nodes.
//
// The returned node doesn't have a parent set.
func (n *Node) Clone(deep bool) *Node {
	node := &Node{
		name:      n.name,
		typ:       n.typ,
		filter:    n.filter,
		validator: n.validator,
	}
	if deep && n.children != nil {
		// Clone all child nodes resetting the parent.
		node.children = make([]*Node, len(n.children))
		var child *Node
		for k, v := range n.children {
			child = v.Clone(deep)
			child.setParent(node)
			node.children[k] = child
		}
	}
	if n.attrs != nil {
		// Clone all attributes.
		node.attrs = make(map[string]string)
		for k, v := range n.attrs {
			node.attrs[k] = v
		}
	}
	return node
}

// Attr returns the value of an extra attribute with the given key, if any.
func (n *Node) Attr(key string) (rv string) {
	if n.attrs != nil {
		if v, ok := n.attrs[key]; ok {
			rv = v
		}
	}
	return
}

// SetAttrs sets extra key/value attributes for this node.
func (n *Node) SetAttrs(pairs ...string) *Node {
	if n.attrs == nil {
		n.attrs = make(map[string]string)
	} else {
		for k, _ := range n.attrs {
			n.attrs[k] = "", false
		}
	}
	length := len(pairs)
	for i := 0; i < length; i += 2 {
		if i+1 < length {
			n.attrs[pairs[i]] = pairs[i+1]
		}
	}
	return n
}

// Children returns the array of child nodes of this node.
func (n *Node) Children() []*Node {
	return n.children
}

// Add appends a child to this node.
func (n *Node) Add(child *Node) *Node {
	if child.Parent() != nil {
		panic("Child node already has a parent.")
	}
	child.setParent(n)
	if n.children == nil {
		n.children = make([]*Node, 0)
	}
	n.children = append(n.children, child)
	return n
}

// Name returns the name of this node.
//
// The returned name is concatenated with parent names in dotted notation.
func (n *Node) Name() string {
	if n.parent != nil {
		return fmt.Sprintf("%s.%s", n.parent.Name(), n.name)
	}
	return n.name
}

// SimpleName returns the name of this node without parent names.
func (n *Node) SimpleName() string {
	return n.name
}

// SetName sets this node's name.
func (n *Node) SetName(name string) *Node {
	n.name = name
	return n
}

// Parent returns the parent node of this node.
func (n *Node) Parent() *Node {
	return n.parent
}

// SetParent sets this node's parent.
func (n *Node) setParent(parent *Node) *Node {
	n.parent = parent
	return n
}

// Type returns this node's type.
func (n *Node) Type() NodeType {
	return n.typ
}

// SetType sets this node's type.
func (n *Node) SetType(typ NodeType) *Node {
	n.typ = typ
	return n
}

// Filter returns the filter of this node.
func (n *Node) Filter() NodeFilter {
	return n.filter
}

// SetFilter sets this node's filter.
func (n *Node) SetFilter(filter NodeFilter) *Node {
	n.filter = filter
	return n
}

// SetFilterFunc sets this node's filter, as a function.
func (n *Node) SetFilterFunc(filter FilterFunc) *Node {
	n.filter = &simpleFilter{f: filter}
	return n
}

// Validator returns the validator of this node.
func (n *Node) Validator() NodeValidator {
	return n.validator
}

// SetValidator sets this node's validator.
func (n *Node) SetValidator(validator NodeValidator) *Node {
	n.validator = validator
	return n
}

// SetValidatorFunc sets this node's validator, as a function.
func (n *Node) SetValidatorFunc(validator ValidatorFunc) *Node {
	n.validator = &simpleValidator{f: validator}
	return n
}

// Serialize
func (n *Node) Serialize(src RawData, dst NodeData) (errors []os.Error) {
	/*
	Problems to solve here:

	- Each type must check if the value is set.
	- If set, prepare value for type conversion -> call filter, if any.
	- Type conversion.
	- Validation -> call validator, if any.

	Side notes:
	- types must be able to access other type values -> solve the 'validate password confimartion' problem.

	*/
	if n.filter != nil {
		v, err := n.filter.Filter(n, "TODO")
		if err != nil {
			// TODO
		}
		if v == nil {
			// TODO
		}
	}

	n.typ.Serialize(n, src, dst)

	if n.validator != nil {
		err := n.validator.Validate(n, "TODO")
		if err != nil {
			// TODO
		}
	}

	return
}

// ----------------------------------------------------------------------------
// Node factories
// ----------------------------------------------------------------------------

// Bool returns a new node with the given name and type bool.
func Bool(name string) *Node {
	return NewNode(name, boolType)
}

// Float32 returns a new node with the given name and type float32.
func Float32(name string) *Node {
	return NewNode(name, float32Type)
}

// Float64 returns a new node with the given name and type float64.
func Float64(name string) *Node {
	return NewNode(name, float64Type)
}

// Int returns a new node with the given name and type int.
func Int(name string) *Node {
	return NewNode(name, intType)
}

// Int8 returns a new node with the given name and type int8.
func Int8(name string) *Node {
	return NewNode(name, int8Type)
}

// Int16 returns a new node with the given name and type int16.
func Int16(name string) *Node {
	return NewNode(name, int16Type)
}

// Int32 returns a new node with the given name and type int32.
func Int32(name string) *Node {
	return NewNode(name, int32Type)
}

// Int64 returns a new node with the given name and type int64.
func Int64(name string) *Node {
	return NewNode(name, int64Type)
}

// Map returns a new node with the given name and type map.
func Map(name string) *Node {
	return NewNode(name, mapType)
}

// String returns a new node with the given name and type string.
func String(name string) *Node {
	return NewNode(name, stringType)
}

// ----------------------------------------------------------------------------
// UrlValues
// ----------------------------------------------------------------------------

// UrlValues wraps data provided using a map. Typically, url.Values.
type UrlValues struct{
	Raw map[string][]string
}

// Get
func (v *UrlValues) Get(key string) []string {
	if values, ok := v.Raw[key]; ok {
		return values
	}
	return nil
}

// ----------------------------------------------------------------------------
// NodeValues
// ----------------------------------------------------------------------------

// NodeValues stores flattened data for a serialized node.
type NodeValues struct {
	values map[string]interface{}
	errors map[string]os.Error
}

// Get returns a value for a flattened key, or nil if it is not set.
func (v *NodeValues) Get(key string) (rv interface{}) {
	if v.values != nil {
		if value, ok := v.values[key]; ok {
			rv = value
		}
	}
	return
}

// Set sets a value for a flattened key.
func (v *NodeValues) Set(key string, val interface{}) {
	if v.values == nil {
		v.values = make(map[string]interface{})
	}
	v.values[key] = val
}

// Values
func (v *NodeValues) Values() interface{} {
	return v.values
}

// Errors
func (v *NodeValues) Errors() interface{} {
	// TODO
	return nil
}

// ----------------------------------------------------------------------------
// Interfaces
// ----------------------------------------------------------------------------

// NodeType
type NodeType interface {
	Serialize(*Node, RawData, NodeData)
}

// NodeFilter
type NodeFilter interface {
	Filter(*Node, interface{}) (interface{}, os.Error)
}

// NodeValidator
type NodeValidator interface {
	Validate(*Node, interface{}) os.Error
}

// RawData is the interface for data providers.
//
// TODO: this interface is not set in stone.
type RawData interface {
	Get(key string) []string
}

// NodeData is the interface to store validated values and validation errors.
//
// TODO: this interface is not set in stone.
type NodeData interface {
	Get(key string) interface{}
	Set(key string, val interface{})
	Values() interface{}
	Errors() interface{}
}

// ----------------------------------------------------------------------------

// FilterFunc is a convenience type to set a filter with a simple function.
type FilterFunc func(*Node, interface{}) (interface{}, os.Error)

// Convenience for FilterFunc.
type simpleFilter struct {
	f FilterFunc
}

func (f *simpleFilter) Filter(n *Node, i interface{}) (interface{}, os.Error) {
	return f.f(n, i)
}

// ----------------------------------------------------------------------------

// ValidatorFunc is a convenience type to set a validator with a simple function.
type ValidatorFunc func(*Node, interface{}) os.Error

// Convenience for ValidatorFunc.
type simpleValidator struct {
	f ValidatorFunc
}

func (v *simpleValidator) Validate(n *Node, i interface{}) os.Error {
	return v.f(n, i)
}
