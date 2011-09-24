// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"
	"os"
	"reflect"
)

type Validator func(value interface{}, status *Status) os.Error

type ValidatorMap map[string]Validator

// Status stores the result from a validation.
type Status struct {
	Errors      map[string][]os.Error
	parents     []reflect.Value
	parentNames []string
}

func (s *Status) setParent(v reflect.Value) {
	parentName := v.Type().Name()
	if s.parents == nil {
		s.parents = make([]reflect.Value, 0)
	}
	if s.parentNames == nil {
		s.parentNames = make([]string, 0)
	} else if size := len(s.parentNames); size > 0 {
		parentName = fmt.Sprintf("%s.%s", s.parentNames[size-1], parentName)
	}
	s.parents = append(s.parents, v)
	s.parentNames = append(s.parentNames, parentName)
}

func (s *Status) getParentName() string {
	if s.parentNames != nil && len(s.parentNames) > 0 {
		return s.parentNames[len(s.parentNames)-1]
	}
	return ""
}

func (s *Status) SetError(key string, err os.Error) {
	if s.Errors == nil {
		s.Errors = make(map[string][]os.Error)
	}
	if _, ok := s.Errors[key]; !ok {
		s.Errors[key] = make([]os.Error, 0)
	}
	s.Errors[key] = append(s.Errors[key], err)
}

func (s *Status) GetErrors(key string) []os.Error {
	if s.Errors != nil {
		if err, ok := s.Errors[key]; ok {
			return err
		}
	}
	return nil
}

//
//
// It fails if any of the validators wasn't called.
func Validate(value interface{}, validators ValidatorMap) *Status {
	v := reflect.Indirect(reflect.ValueOf(value))
	if v.Kind() != reflect.Struct {
		// Not really. For now we'll just panic.
		panic("Validate must receive a struct or pointer to struct.")
	}
	status := new(Status)
	validate(v, validators, status)
	return status
}

// validate applies a list of validators to a struct and records all errors.
func validate(value reflect.Value, validators ValidatorMap, status *Status) {
	status.setParent(value)

	// Set a closure to call and record each validator called.
	called := make([]string, 0)
	call := func(key string, fieldId string, value interface{}) {
		if validator, ok := validators[key]; ok {
			if err := validator(value, status); err != nil {
				status.SetError(fieldId, err)
			}
			called = append(called, key)
		}
	}

	// Iterate over all fields.
	t := value.Type()
	parentName := status.getParentName()
	for i := 0; i < value.NumField(); i++ {
		field := t.Field(i)
		fieldId := fmt.Sprintf("%s.%s", parentName, field.Name)
		fieldVal := value.Field(i).Interface()
		// Validators set by field name.
		call(field.Name, fieldId, fieldVal)
		// Validators set by tag.
		if tag := field.Tag.Get("schema"); tag != "" {
			call(fmt.Sprintf("tag:%s", tag), fieldId, fieldVal)
		}
	}

	// Ensure that all validators were called.
	for key, _ := range validators {
		ok := false
		for _, ckey := range called {
			if key == ckey {
				ok = true
				break
			}
		}
		if !ok {
			// TODO move error to top level.
			status.SetError(key, os.NewError("Validator wasn't called."))
		}
	}
}
