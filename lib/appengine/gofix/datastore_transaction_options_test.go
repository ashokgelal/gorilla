// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(datastoreTransactionOptionsTests)
}

var datastoreTransactionOptionsTests = []testCase{
	{
		Name: "datastore_transaction_options.0",
		In: `package foo

import "appengine/datastore"

func f() {
	datastore.RunInTransaction(c, func(c appengine.Context) os.Error {
		return nil
	})
}
`,
		Out: `package foo

import "appengine/datastore"

func f() {
	datastore.RunInTransaction(c, func(c appengine.Context) os.Error {
		return nil
	}, nil)
}
`,
	},
}
