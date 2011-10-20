// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package appengine_internal

import "http"

// These functions are the dev implementations of the wrapper functions
// in ../appengine/identity.go. See that file for commentary.

const (
	hDefaultHost = "X-AppEngine-Default-Version-Hostname"
	hVersionId   = "X-AppEngine-Inbound-Version-Id"
)

func DefaultVersionHostname(req interface{}) string {
	return req.(http.Header).Get(hDefaultHost)
}

func VersionID(req interface{}) string {
	return req.(http.Header).Get(hVersionId)
}
