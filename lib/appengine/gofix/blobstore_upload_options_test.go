// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(blobstoreUploadOptionsTests)
}

var blobstoreUploadOptionsTests = []testCase{
	{
		Name: "blobstore_upload_options.0",
		In: `package foo

import "appengine/blobstore"

func f() {
	url, err := blobstore.UploadURL(c, "/party")
}
`,
		Out: `package foo

import "appengine/blobstore"

func f() {
	url, err := blobstore.UploadURL(c, "/party", nil)
}
`,
	},
}
