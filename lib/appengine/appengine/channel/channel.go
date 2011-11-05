// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

/*
The channel package implements the server side of App Engine's Channel API.

Create creates a new channel associated with the given clientID,
which must be unique to the client that will use the returned token.

	token, err := channel.Create(c, "player1")
	if err != nil {
		// handle error
	}
	// return token to the client in an HTTP response

Send sends a message to the client over the channel identified by clientID.

	channel.Send(c, "player1", "Game over!")
*/
package channel

import (
	"json"

	"appengine"
	"appengine_internal"
	"goprotobuf.googlecode.com/hg/proto"

	channel_proto "appengine_internal/channel"
)

// Create creates a channel and returns a token for use by the client.
// The clientID is an appication-provided string used to identify the client.
func Create(c appengine.Context, clientID string) (token string, err error) {
	req := &channel_proto.CreateChannelRequest{
		ApplicationKey: &clientID,
	}
	resp := &channel_proto.CreateChannelResponse{}
	err = c.Call(service, "CreateChannel", req, resp, nil)
	token = proto.GetString(resp.ClientId)
	return
}

// Send sends a message on the channel associated with clientID.
func Send(c appengine.Context, clientID, message string) error {
	req := &channel_proto.SendMessageRequest{
		ApplicationKey: &clientID,
		Message:        &message,
	}
	resp := &struct{}{} // VoidProto
	return c.Call(service, "SendChannelMessage", req, resp, nil)
}

// SendJSON is a helper function that sends a JSON-encoded value
// on the channel associated with clientID.
func SendJSON(c appengine.Context, clientID string, value interface{}) error {
	m, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Send(c, clientID, string(m))
}

var service = "xmpp" // prod

func init() {
	if appengine.IsDevAppServer() {
		service = "channel" // dev
	}
	appengine_internal.RegisterErrorCodeMap(service, channel_proto.ChannelServiceError_ErrorCode_name)
}
