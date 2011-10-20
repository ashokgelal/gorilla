// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
)

var datastoreTransactionOptionsFix = fix{
	"datastore_transaction_options",
	datastoreTransactionOptions,
	"Add an options argument to datastore.RunInTransaction.",
}

func init() {
	register(datastoreTransactionOptionsFix)
}

func datastoreTransactionOptions(f *ast.File) bool {
	if !imports(f, "appengine/datastore") {
		return false
	}

	fixed := false
	walk(f, func(n interface{}) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		if isPkgDot(call.Fun, "datastore", "RunInTransaction") && len(call.Args) == 2 {
			call.Args = append(call.Args, ast.NewIdent("nil"))
			fixed = true
		}
	})
	return fixed
}
