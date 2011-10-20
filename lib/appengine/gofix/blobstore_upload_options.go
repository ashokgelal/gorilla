// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
)

var blobstoreUploadOptionsFix = fix{
	"blobstore_upload_options",
	blobstoreUploadOptions,
	"Add an options argument to blobstore.UploadURL.",
}

func init() {
	register(blobstoreUploadOptionsFix)
}

func blobstoreUploadOptions(f *ast.File) bool {
	if !imports(f, "appengine/blobstore") {
		return false
	}

	fixed := false
	walk(f, func(n interface{}) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		if isPkgDot(call.Fun, "blobstore", "UploadURL") && len(call.Args) == 2 {
			call.Args = append(call.Args, ast.NewIdent("nil"))
			fixed = true
		}
	})
	return fixed
}
