// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"os"
)

// ----------------------------------------------------------------------------
// Interfaces
// ----------------------------------------------------------------------------

type NodeType interface {
	Validate(node *Node, src interface{}) (interface{}, []os.Error)
}

type NodeFilter interface {
	Filter(node *Node, src interface{}) interface{}
}

type NodeValidator interface {
	Validate(node *Node, src interface{}) os.Error
}

// ----------------------------------------------------------------------------
// NodeValues
// ----------------------------------------------------------------------------

// NodeValues stores flattened data for a validated node.
type NodeValues struct {
	// TODO
}

func (v *NodeValues) String(key string) (rv string) {
	// TODO
	return
}

func (v *NodeValues) Int64(key string) (rv int64) {
	// TODO
	return
}

// ----------------------------------------------------------------------------
// Node
// ----------------------------------------------------------------------------

// Node represents a node in the schema.
//
// A node has a name, type, parent and children nodes, filters and validators.
//
// A node is immutable. When any field value is changed, a clone with the
// updated value is returned.
type Node struct {
	name       string
	ntype      NodeType
	parent     *Node
	children   []*Node
	filters    []NodeFilter
	validators []NodeValidator
	required   bool
}

// NewNode returns a new node with the given name, type and validators.
func NewNode(name string, ntype NodeType, validators ...NodeValidator) *Node {
	return &Node{
		name:       name,
		ntype:      ntype,
		validators: validators,
		required:   true,
	}
}

// Clone clones this node and all child nodes.
func (n *Node) Clone() *Node {
	node := &Node{
		name:       n.name,
		ntype:      n.ntype,
		parent:     n.parent,
		filters:    n.filters,
		validators: n.validators,
	}
	var children []*Node
	if n.children != nil {
		// Clone all child nodes resetting the parent.
		children = make([]*Node, len(n.children))
		for k, v := range n.children {
			children[k] = v.SetParent(node)
		}
	}
	node.children = children
	return node
}

// Add returns a clone of this node with a new child appended.
func (n *Node) Add(child *Node) *Node {
	clone := n.Clone()
	child = child.SetParent(clone)
	if clone.children == nil {
		clone.children = []*Node{child}
	} else {
		clone.children = append(clone.children, child)
	}
	return clone
}

// Children returns the array of child nodes of this node.
func (n *Node) Children() []*Node {
	return n.children
}

// Filters returns the filters of this node.
func (n *Node) Filters() []NodeFilter {
	return n.filters
}

// SetFilters returns a clone of this node with the given filters.
func (n *Node) SetFilters(filters ...NodeFilter) *Node {
	clone := n.Clone()
	clone.filters = filters
	return clone
}

// Name returns the name of this node.
func (n *Node) Name() string {
	return n.name
}

// SetName returns a clone of this node with a new name.
func (n *Node) SetName(name string) *Node {
	clone := n.Clone()
	clone.name = name
	return clone
}

// Parent returns the parent node of this node.
func (n *Node) Parent() *Node {
	return n.parent
}

// SetParent returns a clone of this node with a new parent.
func (n *Node) SetParent(parent *Node) *Node {
	clone := n.Clone()
	clone.parent = parent
	return clone
}

// Required returns the required flag of this node.
func (n *Node) Required() bool {
	return n.required
}

// SetRequired returns a clone of this node with a new required flag.
func (n *Node) SetRequired(required bool) *Node {
	clone := n.Clone()
	clone.required = required
	return clone
}

// Type returns the type of this node.
func (n *Node) Type() NodeType {
	return n.ntype
}

// SetType returns a clone of this node with a new type.
func (n *Node) SetType(ntype NodeType) *Node {
	clone := n.Clone()
	clone.ntype = ntype
	return clone
}

// Validators returns the validators of this node.
func (n *Node) Validators() []NodeValidator {
	return n.validators
}

// SetValidators returns a clone of this node with the given validators.
func (n *Node) SetValidators(validators ...NodeValidator) *Node {
	clone := n.Clone()
	clone.validators = validators
	return clone
}

func (n *Node) Validate(data interface{}) (val *NodeValues, err []os.Error) {
	// TODO: alow data to be struct or map; return error if not one of them.
	//       Convert struct to map before validating?
	val = new(NodeValues)

	// TODO
	return
}

// ----------------------------------------------------------------------------
// Node factories
// ----------------------------------------------------------------------------

// Int64 returns a Int64Type node with the given name and validators.
func Int64(name string, validators ...NodeValidator) *Node {
	return NewNode(name, int64Type, validators...)
}

// Map returns a MapType node with the given name and validators.
func Map(name string, validators ...NodeValidator) *Node {
	return NewNode(name, mapType, validators...)
}

// String returns a StringType node with the given name and validators.
func String(name string, validators ...NodeValidator) *Node {
	return NewNode(name, stringType, validators...)
}

// ----------------------------------------------------------------------------
// Int64Type
// ----------------------------------------------------------------------------

var int64Type = &Int64Type{}

type Int64Type struct {
}

func (t *Int64Type) Validate(node *Node, src interface{}) (dst interface{}, errors []os.Error) {
	// TODO
	return
}

// ----------------------------------------------------------------------------
// MapType
// ----------------------------------------------------------------------------

var mapType = &MapType{}

type MapType struct {
}

func (t *MapType) Validate(node *Node, src interface{}) (dst interface{}, errors []os.Error) {
	// TODO
	return
}

// ----------------------------------------------------------------------------
// StringType
// ----------------------------------------------------------------------------

var stringType = &StringType{}

type StringType struct {
}

func (t *StringType) Validate(node *Node, src interface{}) (dst interface{}, errors []os.Error) {
	// TODO
	return
}
