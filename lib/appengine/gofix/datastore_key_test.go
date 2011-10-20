// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(datastoreKeyTests)
}

var datastoreKeyTests = []testCase{
	{
		Name: "datastore_key.0",
		In: `package foo

import "appengine"
import "appengine/datastore"

func f() {
	ctxt := appengine.NewContext(req)
	datastore.NewKey("Gopher", "bruce", 0, dad)
	datastore.NewIncompleteKey("Gopher")
}

func g(ct appengine.Context) {
	datastore.NewKey("Gopher", "", 7, ken)
	datastore.NewIncompleteKey("Gopher")
}

func h() {
	datastore.NewKey("Gopher", "", 2, mum)
	datastore.NewIncompleteKey("Gopher")
}
`,
		Out: `package foo

import "appengine"
import "appengine/datastore"

func f() {
	ctxt := appengine.NewContext(req)
	datastore.NewKey(ctxt, "Gopher", "bruce", 0, dad)
	datastore.NewIncompleteKey(ctxt, "Gopher", nil)
}

func g(ct appengine.Context) {
	datastore.NewKey(ct, "Gopher", "", 7, ken)
	datastore.NewIncompleteKey(ct, "Gopher", nil)
}

func h() {
	datastore.NewKey(c, "Gopher", "", 2, mum)
	datastore.NewIncompleteKey(c, "Gopher", nil)
}
`,
	},
}
