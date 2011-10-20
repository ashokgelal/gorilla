// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

/*
Package capability exposes information about outages and scheduled downtime
for specific API capabilities.

Example:
	if !capability.Enabled(c, "datastore_v3", "write") {
		// show user a different page
	}
*/
package capability

import (
	"appengine"

	capability_proto "appengine_internal/capability"
)

// Enabled returns whether an API's capabilities are enabled.
// The wildcard "*" capability matches every capability of an API.
// If the underlying RPC fails (if the package is unknown, for example),
// false is returned and information is written to the application log.
func Enabled(c appengine.Context, api, capability string) bool {
	req := &capability_proto.IsEnabledRequest{
		Package:    &api,
		Capability: []string{capability},
	}
	res := &capability_proto.IsEnabledResponse{}
	if err := c.Call("capability_service", "IsEnabled", req, res, nil); err != nil {
		c.Warningf("capability.Enabled: RPC failed: %v", err)
		return false
	}
	switch *res.SummaryStatus {
	case capability_proto.IsEnabledResponse_ENABLED,
		capability_proto.IsEnabledResponse_SCHEDULED_FUTURE,
		capability_proto.IsEnabledResponse_SCHEDULED_NOW:
		return true
	case capability_proto.IsEnabledResponse_UNKNOWN:
		c.Errorf("capability.Enabled: unknown API capability %s/%s", api, capability)
		return false
	default:
		return false
	}
	panic("unreachable")
}
