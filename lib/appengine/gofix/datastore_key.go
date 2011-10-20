// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
)

var datastoreKeyFix = fix{
	"datastore_key",
	datastoreKey,
	`Add an appengine.Context argument to datastore.NewKey and
datastore.NewIncompleteKey, and add a parent key argument to
datastore.NewIncompleteKey.`,
}

func init() {
	register(datastoreKeyFix)
}

func datastoreKey(f *ast.File) bool {
	if !imports(f, "appengine/datastore") {
		return false
	}

	// During the walk, we track the last thing seen that looks like
	// an appengine.Context, and reset it once the walk leaves a func.
	var lastContext *ast.Ident

	fixed := false
	walk(f, func(n interface{}) {
		// If this node is an assignment from an appengine.NewContext invocation,
		// remember the identifier on the LHS.
		if as, ok := n.(*ast.AssignStmt); ok {
			if len(as.Lhs) != 1 || len(as.Rhs) != 1 {
				return
			}
			if !isCall(as.Rhs[0], "appengine", "NewContext") {
				return
			}
			if ident, ok := as.Lhs[0].(*ast.Ident); ok {
				lastContext = ast.NewIdent(ident.Name)
			}
			return
		}

		// If this node is a function type with an appengine.Context argument,
		// remember that argument's identifier.
		// (gofix will walk a FuncDecl after the FuncDecl's Body).
		if ft, ok := n.(*ast.FuncType); ok && ft.Params != nil {
			for _, param := range ft.Params.List {
				if isPkgDot(param.Type, "appengine", "Context") && len(param.Names) > 0 {
					lastContext = ast.NewIdent(param.Names[0].Name)
					return
				}
			}
			return
		}

		// If this node is a FuncDecl, we've finished the function, so reset lastContext.
		if _, ok := n.(*ast.FuncDecl); ok {
			lastContext = nil
			return
		}

		// Only interested in function calls beyond here.
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		if isPkgDot(call.Fun, "datastore", "NewKey") {
			if len(call.Args) == 4 {
				insertContext(f, call, lastContext)
				fixed = true
			}
			return
		}
		if isPkgDot(call.Fun, "datastore", "NewIncompleteKey") {
			if len(call.Args) == 1 {
				insertContext(f, call, lastContext)
				call.Args = append(call.Args, ast.NewIdent("nil"))
				fixed = true
			}
			return
		}
	})
	return fixed
}

// c may be nil.
func insertContext(f *ast.File, call *ast.CallExpr, c *ast.Ident) {
	if c == nil {
		// c is unknown, so use a plain "c".
		c = ast.NewIdent("c")
	}

	call.Args = append([]ast.Expr{c}, call.Args...)
}
