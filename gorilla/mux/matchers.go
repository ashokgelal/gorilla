// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"http"
)

// ----------------------------------------------------------------------------
// Matchers
// ----------------------------------------------------------------------------

// MatcherFunc is the type used by custom matchers.
type MatcherFunc func(*http.Request) bool

// routeMatcher is the interface used by the router, route and route matchers.
//
// Only Router and Route actually return a route; it indicates a final match.
// Route matchers return nil and the result from the individual match.
type routeMatcher interface {
	Match(*http.Request) (*Route, bool)
}

// customMatcher matches the request using a custom matcher function.
type customMatcher struct {
	matcherFunc MatcherFunc
}

func (m *customMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, m.matcherFunc(request)
}

// headerMatcher matches the request against header values.
type headerMatcher struct {
	headers map[string]string
}

func (m *headerMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchMap(m.headers, request.Header, true)
}

// methodMatcher matches the request against HTTP methods.
type methodMatcher struct {
	methods []string
}

func (m *methodMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchInArray(m.methods, request.Method)
}

// queryMatcher matches the request against URL queries.
type queryMatcher struct {
	queries map[string]string
}

func (m *queryMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchMap(m.queries, request.URL.Query(), false)
}

// schemeMatcher matches the request against URL schemes.
type schemeMatcher struct {
	schemes []string
}

func (m *schemeMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchInArray(m.schemes, request.URL.Scheme)
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// matchInArray returns true if the given string value is in the array.
func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// matchMap returns true if the given key/value pairs exist in a given map.
func matchMap(toCheck map[string]string, toMatch map[string][]string,
			  canonicalKey bool) bool {
	for k, v := range toCheck {
		// Check if key exists.
		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		if values, keyExists := toMatch[k]; !keyExists {
			return false
		} else if v != "" {
			// If value was defined as an empty string we only check that the
			// key exists. Otherwise we also check if the value exists.
			valueExists := false
			for _, value := range values {
				if v == value {
					valueExists = true
					break
				}
			}
			if !valueExists {
				return false
			}
		}
	}
	return true
}
