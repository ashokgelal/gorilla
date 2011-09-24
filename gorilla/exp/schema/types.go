// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	//"os"
	"reflect"
	"strconv"
	"strings"
)

// All built-in types
// ==================
// Basic types: Bool, Float32, Float64, Int, Int8, Int16, Int32, Int64, String
//
// Composite types: Map
// TODO: Array, Tuple, Date, Time
var (
	boolType    = new(BoolType)
	float32Type = new(Float32Type)
	float64Type = new(Float64Type)
	intType     = new(IntType)
	int8Type    = new(Int8Type)
	int16Type   = new(Int16Type)
	int32Type   = new(Int32Type)
	int64Type   = new(Int64Type)
	mapType     = new(MapType)
	stringType  = new(StringType)
)

// ----------------------------------------------------------------------------
// Bool
// ----------------------------------------------------------------------------

// BoolType serializes and deserializes a bool type.
type BoolType struct {
}

// Serialize
func (t *BoolType) Serialize(node *Node, src RawData, dst NodeData) {
	// TODO
}

// ----------------------------------------------------------------------------
// Float
// ----------------------------------------------------------------------------

// Float32Type serializes and deserializes a float32 type.
type Float32Type struct {
}

// Serialize
func (t *Float32Type) Serialize(node *Node, src RawData, dst NodeData) {
	serializeFloat(node, src, dst, reflect.Float32)
}

// ----------------------------------------------------------------------------

// Float64Type serializes and deserializes a float64 type.
type Float64Type struct {
}

// Serialize
func (t *Float64Type) Serialize(node *Node, src RawData, dst NodeData) {
	serializeFloat(node, src, dst, reflect.Float64)
}

// ----------------------------------------------------------------------------

// serializeFloat
func serializeFloat(node *Node, src RawData, dst NodeData, kind reflect.Kind) {
	name := node.Name()
	if values := src.Get(name); values != nil && len(values) > 0 {
		value, err := strconv.Atof64(strings.TrimSpace(values[0]))
		if err != nil {
			// TODO add error to the list of errors.
		} else {
			switch kind {
			case reflect.Float32:
				dst.Set(name, float32(value))
			default:
				dst.Set(name, value)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Int
// ----------------------------------------------------------------------------

// IntType serializes and deserializes an int type.
type IntType struct {
}

// Serialize
func (t *IntType) Serialize(node *Node, src RawData, dst NodeData) {
	serializeInt(node, src, dst, reflect.Int)
}

// ----------------------------------------------------------------------------

// Int8Type serializes and deserializes an int8 type.
type Int8Type struct {
}

// Serialize
func (t *Int8Type) Serialize(node *Node, src RawData, dst NodeData) {
	serializeInt(node, src, dst, reflect.Int8)
}

// ----------------------------------------------------------------------------

// Int16Type serializes and deserializes an int16 type.
type Int16Type struct {
}

// Serialize
func (t *Int16Type) Serialize(node *Node, src RawData, dst NodeData) {
	serializeInt(node, src, dst, reflect.Int16)
}

// ----------------------------------------------------------------------------

// Int32Type serializes and deserializes an int32 type.
type Int32Type struct {
}

// Serialize
func (t *Int32Type) Serialize(node *Node, src RawData, dst NodeData) {
	serializeInt(node, src, dst, reflect.Int32)
}

// ----------------------------------------------------------------------------

// Int64Type serializes and deserializes an int64 type.
type Int64Type struct {
}

// Serialize
func (t *Int64Type) Serialize(node *Node, src RawData, dst NodeData) {
	serializeInt(node, src, dst, reflect.Int64)
}

// ----------------------------------------------------------------------------

// serializeInt
func serializeInt(node *Node, src RawData, dst NodeData,	kind reflect.Kind) {
	name := node.Name()
	if values := src.Get(name); values != nil && len(values) > 0 {
		value, err := strconv.Atoi64(strings.TrimSpace(values[0]))
		if err != nil {
			// TODO add error to the list of errors.
		} else {
			// TODO should we do anything if data is lost after conversion?
			switch kind {
			case reflect.Int:
				dst.Set(name, int(value))
			case reflect.Int8:
				dst.Set(name, int8(value))
			case reflect.Int16:
				dst.Set(name, int16(value))
			case reflect.Int32:
				dst.Set(name, int32(value))
			default:
				dst.Set(name, value)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Map
// ----------------------------------------------------------------------------

// MapType serializes and deserializes a map[string]interface{} type.
type MapType struct {
}

// Serialize
func (t *MapType) Serialize(node *Node, src RawData, val NodeData) {
	for _, n := range node.Children() {
		n.Serialize(src, val)
	}
}

// ----------------------------------------------------------------------------
// String
// ----------------------------------------------------------------------------

// StringType serializes and deserializes a string type.
type StringType struct {
}

// Serialize
func (t *StringType) Serialize(node *Node, src RawData, val NodeData) {
	// TODO add error to the list of errors.
	var value string
	name := node.Name()
	if values := src.Get(name); values != nil && len(values) > 0 {
		value = values[0]
	}
	val.Set(name, value)
}
