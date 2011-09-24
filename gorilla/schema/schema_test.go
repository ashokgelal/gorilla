// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func DummyValidator(value interface{}, status *Status) os.Error {
	return os.NewError(reflect.TypeOf(value).String())
}

type Bar struct {
	Field1 string
}

type Foo struct {
	Field1 string      `schema:"extraField1"`
	Field2 int
	Field3 float32
	Field4 *Bar
	Field5 interface{}
}

func TestQueryMatcher(t *testing.T) {
	validators := map[string]Validator{
		"Field1":          DummyValidator,
		"Field2":          DummyValidator,
		"Field3":          DummyValidator,
		"Field4":          DummyValidator,
		"Field5":          DummyValidator,
		"tag:extraField1": DummyValidator,
	}

	foo := Foo{
		Field1: "foo",
		Field2: 42,
		Field3: 4.2,
		Field4: &Bar{Field1: "bar"},
		Field5: []string{"foo", "bar"},
	}
	status := Validate(foo, validators)
	fmt.Printf("errors: %v\n", status.Errors)
}
