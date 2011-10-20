// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package appengine_internal

import (
	"bufio"
	"http"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"appengine_internal/remote_api"
	"goprotobuf.googlecode.com/hg/proto"
)

// IsDevAppServer returns whether the App Engine app is running in the
// development App Server.
func IsDevAppServer() bool {
	return true
}

// serveHTTP serves App Engine HTTP requests.
func serveHTTP(netw, addr string) {
	if netw == "unix" {
		os.Remove(addr)
	}
	l, err := net.Listen(netw, addr)
	if err != nil {
		log.Fatal("appengine: ", err)
	}
	err = http.Serve(l, http.HandlerFunc(handleFilteredHTTP))
	if err != nil {
		log.Fatal("appengine: ", err)
	}
}

func init() {
	RegisterHTTPFunc(serveHTTP)
}

func handleFilteredHTTP(w http.ResponseWriter, r *http.Request) {
	// Patch up RemoteAddr so it looks reasonable.
	const remoteAddrHeader = "X-AppEngine-Remote-Addr"
	if addr := r.Header.Get(remoteAddrHeader); addr != "" {
		r.RemoteAddr = addr
		r.Header.Del(remoteAddrHeader)
	} else {
		// Should not normally reach here, but pick
		// a sensible default anyway.
		r.RemoteAddr = "127.0.0.1"
	}

	http.DefaultServeMux.ServeHTTP(w, r)
}

// read and write speak a custom protocol with the appserver. Specifically, an
// ASCII header followed by an encoded protocol buffer. The header is the
// length of the protocol buffer, in decimal, followed by a new line character.
// For example: "53\n".

// read reads a protocol buffer from the socketAPI socket.
func read(r *bufio.Reader, pb interface{}) os.Error {
	b, err := r.ReadSlice('\n')
	if err != nil {
		return err
	}
	n, err := strconv.Atoi(string(b[:len(b)-1]))
	if err != nil {
		return err
	}
	if n < 0 {
		return os.NewError("appengine: negative message length")
	}
	b = make([]byte, n)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return err
	}
	return proto.Unmarshal(b, pb)
}

// write writes a protocol buffer to the socketAPI socket.
func write(w *bufio.Writer, pb interface{}) os.Error {
	b, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	_, err = w.WriteString(strconv.Itoa(len(b)))
	if err != nil {
		return err
	}
	err = w.WriteByte('\n')
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	return w.Flush()
}

var (
	mu       sync.Mutex
	apiRead  *bufio.Reader
	apiWrite *bufio.Writer
)

// initAPI prepares the app to execute App Engine API calls,
// forwarding them to the Appserver at the given network address.
func initAPI(netw, addr string) {
	c, err := net.Dial(netw, addr)
	if err != nil {
		log.Fatal("appengine: ", err)
	}
	apiRead, apiWrite = bufio.NewReader(c), bufio.NewWriter(c)
}

func call(service, method string, data []byte) ([]byte, os.Error) {
	mu.Lock()
	defer mu.Unlock()

	req := &remote_api.Request{
		ServiceName: &service,
		Method:      &method,
		Request:     data,
	}
	if err := write(apiWrite, req); err != nil {
		return nil, err
	}
	res := &remote_api.Response{}
	if err := read(apiRead, res); err != nil {
		return nil, err
	}
	if ae := res.ApplicationError; ae != nil {
		// All Remote API application errors are API-level failures.
		return nil, &APIError{Service: service, Detail: *ae.Detail, Code: *ae.Code}
	}
	return res.Response, nil
}

// context represents the context of an in-flight HTTP request.
// It implements the appengine.Context interface.
type context struct {
	RequestHeader http.Header
}

func NewContext(req *http.Request) *context {
	return &context{
		RequestHeader: req.Header,
	}
}

func (c *context) Call(service, method string, in, out interface{}, _ *CallOptions) os.Error {
	data, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	res, err := call(service, method, data)
	if err != nil {
		return err
	}
	return proto.Unmarshal(res, out)
}

func (c *context) Request() interface{} {
	return c.RequestHeader
}

func (c *context) logf(level, format string, args ...interface{}) {
	log.Printf(level+": "+format, args...)
}

func (c *context) Debugf(format string, args ...interface{})    { c.logf("DEBUG", format, args...) }
func (c *context) Infof(format string, args ...interface{})     { c.logf("INFO", format, args...) }
func (c *context) Warningf(format string, args ...interface{})  { c.logf("WARNING", format, args...) }
func (c *context) Errorf(format string, args ...interface{})    { c.logf("ERROR", format, args...) }
func (c *context) Criticalf(format string, args ...interface{}) { c.logf("CRITICAL", format, args...) }

// FullyQualifiedAppID returns the fully-qualified application ID.
// This may contain a partition prefix (e.g. "s~" for High Replication apps),
// or a domain prefix (e.g. "example.com:").
func (c *context) FullyQualifiedAppID() string {
	return c.RequestHeader.Get("X-AppEngine-Inbound-AppId")
}

func (c *context) AppID() string {
	return appID(c.FullyQualifiedAppID())
}
