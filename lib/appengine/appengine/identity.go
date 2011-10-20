// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package appengine

import (
	"appengine_internal"
)

// AppID returns the application ID for the current application.
// The string will be a plain application ID (e.g. "appid"), with a
// domain prefix for custom domain deployments (e.g. "example.com:appid").
func AppID(c Context) string { return appengine_internal.AppID(c.FullyQualifiedAppID()) }

// DefaultVersionHostname returns the standard hostname of the default version
// of the current application (e.g. "my-app.appspot.com"). This is suitable for
// use in constructing URLs.
func DefaultVersionHostname(c Context) string {
	return appengine_internal.DefaultVersionHostname(c.Request())
}

// VersionID returns the version ID for the current application.
// It will be of the form "X.Y", where X is specified in app.yaml,
// and Y is a generated number when each version of the app is uploaded.
func VersionID(c Context) string { return appengine_internal.VersionID(c.Request()) }
