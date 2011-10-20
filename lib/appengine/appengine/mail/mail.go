// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

/*
Package mail provides the means of sending email from an
App Engine application.

Example:
	msg := &mail.Message{
		Sender:  "romeo@montague.com",
		To:      []string{"Juliet <juliet@capulet.org>"},
		Subject: "See you tonight",
		Body:    "Don't forget our plans. Hark, 'til later.",
	}
	if err := mail.Send(c, msg); err != nil {
		c.Errorf("Alas, my user, the email failed to sendeth: %v", err)
	}
*/
package mail

import (
	"os"

	"appengine"
	"appengine_internal"

	mail_proto "appengine_internal/mail"
)

// A Message represents an email message.
// Addresses may be of any form permitted by RFC 822.
type Message struct {
	// Sender must be set, and must be either an application admin
	// or the currently signed-in user.
	Sender  string
	ReplyTo string // may be empty

	// At least one of these slices must have a non-zero length.
	To, Cc, Bcc []string

	Subject string
	Body    string

	Attachments []Attachment

	// TODO: "HTML" body.
}

// An Attachment represents an email attachment.
type Attachment struct {
	// Name must be set to a valid file name.
	Name string
	Data []byte
}

// Send sends an email message.
func Send(c appengine.Context, msg *Message) os.Error {
	req := &mail_proto.MailMessage{
		Sender:   &msg.Sender,
		To:       msg.To,
		Cc:       msg.Cc,
		Bcc:      msg.Bcc,
		Subject:  &msg.Subject,
		TextBody: &msg.Body,
	}
	if msg.ReplyTo != "" {
		req.ReplyTo = &msg.ReplyTo
	}
	if len(msg.Attachments) > 0 {
		req.Attachment = make([]*mail_proto.MailAttachment, len(msg.Attachments))
		for i, att := range msg.Attachments {
			req.Attachment[i] = &mail_proto.MailAttachment{
				FileName: &att.Name,
				Data:     att.Data,
			}
		}
	}
	res := &struct{}{} // VoidProto
	if err := c.Call("mail", "Send", req, res, nil); err != nil {
		return err
	}
	return nil
}

func init() {
	appengine_internal.RegisterErrorCodeMap("mail", mail_proto.MailServiceError_ErrorCode_name)
}
