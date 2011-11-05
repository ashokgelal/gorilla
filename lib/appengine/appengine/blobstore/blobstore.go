// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package blobstore provides a client for App Engine's persistent blob
// storage service.
package blobstore

import (
	"bufio"
	"errors"
	"fmt"
	"http"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"url"

	"appengine"
	"appengine/datastore"
	"appengine_internal"
	"appengine_internal/files"
	"goprotobuf.googlecode.com/hg/proto"

	pb "appengine_internal/blobstore"
)

const (
	blobInfoKind        = "__BlobInfo__"
	blobFileIndexKind   = "__BlobFileIndex__"
	blobKeyPropertyName = "blob_key"
	zeroKey             = appengine.BlobKey("")
)

// BlobInfo is the blob metadata that is stored in the datastore.
type BlobInfo struct {
	BlobKey      appengine.BlobKey
	ContentType  string
	CreationTime *time.Time
	Filename     string // if provided
	Size         int64
}

// Stat returns the BlobInfo for a provided blobKey. If no blob was found for
// that key, Stat returns datastore.ErrNoSuchEntity.
func Stat(c appengine.Context, blobKey appengine.BlobKey) (*BlobInfo, error) {
	dskey := datastore.NewKey(c, blobInfoKind, string(blobKey), 0, nil)
	m := make(datastore.Map)
	if err := datastore.Get(c, dskey, m); err != nil {
		return nil, err
	}
	contentType, ok0 := m["content_type"].(string)
	filename, ok1 := m["filename"].(string)
	size, ok2 := m["size"].(int64)
	creation, ok3 := m["creation"].(datastore.Time)
	if !ok0 || !ok1 || !ok2 || !ok3 {
		return nil, errors.New("blobstore: invalid blob info")
	}
	bi := &BlobInfo{
		BlobKey:      blobKey,
		ContentType:  contentType,
		Filename:     filename,
		Size:         size,
		CreationTime: creation.Time(),
	}
	return bi, nil
}

// Send sets the headers on response to instruct App Engine to send a blob as
// the response body. This is more efficient than reading and writing it out
// manually and isn't subject to normal response size limits.
func Send(response http.ResponseWriter, blobKey appengine.BlobKey) {
	hdr := response.Header()
	hdr.Set("X-AppEngine-BlobKey", string(blobKey))

	if hdr.Get("Content-Type") == "" {
		// This value is known to dev_appserver to mean automatic.
		// In production this is remapped to the empty value which
		// means automatic.
		hdr.Set("Content-Type", "application/vnd.google.appengine.auto")
	}
}

// UploadURL creates an upload URL for the form that the user will
// fill out, passing the application path to load when the POST of the
// form is completed. These URLs expire and should not be reused. The
// opts parameter may be nil.
func UploadURL(c appengine.Context, successPath string, opts *UploadURLOptions) (*url.URL, error) {
	req := &pb.CreateUploadURLRequest{
		SuccessPath: proto.String(successPath),
	}
	if opts != nil {
		if opts.MaxUploadBytes != 0 {
			req.MaxUploadSizeBytes = proto.Int64(opts.MaxUploadBytes)
		}
		if opts.MaxUploadBytesPerBlob != 0 {
			req.MaxUploadSizePerBlobBytes = proto.Int64(opts.MaxUploadBytesPerBlob)
		}
	}
	res := &pb.CreateUploadURLResponse{}
	if err := c.Call("blobstore", "CreateUploadURL", req, res, nil); err != nil {
		return nil, err
	}
	return url.Parse(*res.Url)
}

// UploadURLOptions are the options to create an upload URL.
type UploadURLOptions struct {
	MaxUploadBytes        int64 // optional
	MaxUploadBytesPerBlob int64 // optional
}

// Delete deletes a blob.
func Delete(c appengine.Context, blobKey appengine.BlobKey) error {
	return DeleteMulti(c, []appengine.BlobKey{blobKey})
}

// DeleteMulti deletes multiple blobs.
func DeleteMulti(c appengine.Context, blobKey []appengine.BlobKey) error {
	s := make([]string, len(blobKey))
	for i, b := range blobKey {
		s[i] = string(b)
	}
	req := &pb.DeleteBlobRequest{
		BlobKey: s,
	}
	res := &struct{}{} // unused, a base.VoidProto
	if err := c.Call("blobstore", "DeleteBlob", req, res, nil); err != nil {
		return err
	}
	return nil
}

func errorf(format string, args ...interface{}) error {
	return fmt.Errorf("blobstore: "+format, args...)
}

// ParseUpload parses the synthetic POST request that your app gets from
// App Engine after a user's successful upload of blobs. Given the request,
// ParseUpload returns a map of the blobs received (keyed by HTML form
// element name) and other non-blob POST parameters.
func ParseUpload(req *http.Request) (blobs map[string][]*BlobInfo, other map[string][]string, err error) {
	_, params := mime.ParseMediaType(req.Header.Get("Content-Type"))
	boundary := params["boundary"]
	if boundary == "" {
		return nil, nil, errorf("did not find MIME multipart boundary")
	}

	blobs = make(map[string][]*BlobInfo)
	other = make(map[string][]string)

	mreader := multipart.NewReader(io.MultiReader(req.Body, strings.NewReader("\r\n\r\n")), boundary)
	for {
		part, perr := mreader.NextPart()
		if perr == io.EOF {
			break
		}
		if perr != nil {
			return nil, nil, errorf("error reading next mime part with boundary %q (len=%d): %v",
				boundary, len(boundary), perr)
		}

		bi := &BlobInfo{}
		ctype, params := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		bi.Filename = params["filename"]
		formKey := params["name"]

		ctype, params = mime.ParseMediaType(part.Header.Get("Content-Type"))
		bi.BlobKey = appengine.BlobKey(params["blob-key"])
		if ctype != "message/external-body" || bi.BlobKey == "" {
			if formKey != "" {
				slurp, serr := ioutil.ReadAll(part)
				if serr != nil {
					return nil, nil, errorf("error reading %q MIME part", formKey)
				}
				other[formKey] = append(other[formKey], string(slurp))
			}
			continue
		}

		// App Engine sends a MIME header as the body of each MIME part.
		tp := textproto.NewReader(bufio.NewReader(part))
		header, mimeerr := tp.ReadMIMEHeader()
		if mimeerr != nil {
			return nil, nil, mimeerr
		}
		bi.Size, err = strconv.Atoi64(header.Get("Content-Length"))
		if err != nil {
			return nil, nil, err
		}
		bi.ContentType = header.Get("Content-Type")

		// Parse the time from the MIME header like:
		// X-AppEngine-Upload-Creation: 2011-03-15 21:38:34.712136
		const timeFormat = "2006-01-02 15:04:05"
		createDate := header.Get("X-AppEngine-Upload-Creation")
		if len(createDate) >= len(timeFormat) {
			// Strip off the sub-second precision
			// because time.Parse can't handle it.
			bi.CreationTime, err = time.Parse(timeFormat, createDate[:len(timeFormat)])
			if err != nil {
				return nil, nil, errorf("error parsing X-AppEngine-Upload-Creation: %s", err)
			}
		} else {
			return nil, nil, errorf("expected to find an X-AppEngine-Upload-Creation header")
		}

		blobs[formKey] = append(blobs[formKey], bi)
	}
	return
}

// Reader is a blob reader.
type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// NewReader returns a reader for a blob. It always succeeds; if the blob does
// not exist then an error will be reported upon first read.
func NewReader(c appengine.Context, blobKey appengine.BlobKey) Reader {
	return &reader{
		c:       c,
		blobKey: blobKey,
	}
}

const readBufferSize = 256 * 1024

// reader is a blob reader. It implements the Reader interface.
type reader struct {
	c       appengine.Context
	blobKey appengine.BlobKey
	// buf is the read buffer. r is how much of buf has been read.
	// off is the offset of buf[0] relative to the start of the blob.
	// An invariant is 0 <= r && r <= len(buf).
	// Reads that don't require an RPC call will increment r but not off.
	// Seeks may modify r without discarding the buffer, but only if the
	// invariant can be maintained.
	mu  sync.Mutex
	buf []byte
	r   int
	off int64
}

func (r *reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.r == len(r.buf) {
		if err := r.fetch(r.off + int64(r.r)); err != nil {
			return 0, err
		}
	}
	n := copy(p, r.buf[r.r:])
	r.r += n
	return n, nil
}

func (r *reader) ReadAt(p []byte, off int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	// Convert relative offsets to absolute offsets.
	ab0 := r.off + int64(r.r)
	ab1 := r.off + int64(len(r.buf))
	ap0 := off
	ap1 := off + int64(len(p))
	// Check if we can satisfy the read entirely out of the existing buffer.
	if r.off <= ap0 && ap1 <= ab1 {
		// Convert off from an absolute offset to a relative offset.
		rp0 := int(ap0 - r.off)
		return copy(p, r.buf[rp0:]), nil
	}
	// Restore the original Read/Seek offset after ReadAt completes.
	defer r.seek(ab0)
	// Repeatedly fetch and copy until we have filled p.
	n := 0
	for len(p) > 0 {
		if err := r.fetch(off + int64(n)); err != nil {
			return n, err
		}
		r.r = copy(p, r.buf)
		n += r.r
		p = p[r.r:]
	}
	return n, nil
}

func (r *reader) Seek(offset int64, whence int) (ret int64, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch whence {
	case os.SEEK_SET:
		ret = offset
	case os.SEEK_CUR:
		ret = r.off + int64(r.r) + offset
	case os.SEEK_END:
		return 0, errorf("seeking relative to the end of a blob isn't supported")
	default:
		return 0, errorf("invalid Seek whence value: %d", whence)
	}
	if ret < 0 {
		return 0, errorf("negative Seek offset")
	}
	return r.seek(ret)
}

// fetch fetches readBufferSize bytes starting at the given offset. On success,
// the data is saved as r.buf.
func (r *reader) fetch(off int64) error {
	req := &pb.FetchDataRequest{
		BlobKey:    proto.String(string(r.blobKey)),
		StartIndex: proto.Int64(off),
		EndIndex:   proto.Int64(off + readBufferSize - 1), // EndIndex is inclusive.
	}
	res := &pb.FetchDataResponse{}
	if err := r.c.Call("blobstore", "FetchData", req, res, nil); err != nil {
		return err
	}
	if len(res.Data) == 0 {
		return io.EOF
	}
	r.buf, r.r, r.off = res.Data, 0, off
	return nil
}

// seek seeks to the given offset with an effective whence equal to SEEK_SET.
// It discards the read buffer if the invariant cannot be maintained.
func (r *reader) seek(off int64) (int64, error) {
	delta := off - r.off
	if delta >= 0 && delta < int64(len(r.buf)) {
		r.r = int(delta)
		return off, nil
	}
	r.buf, r.r, r.off = nil, 0, off
	return off, nil
}

const writeBufferSize = 256 * 1024

// Writer is used for writing blobs. Blobs aren't fully written until
// Close is called, at which point the key can be retrieved by calling
// the Key method.
type Writer struct {
	c        appengine.Context
	filename string

	buf      []byte
	writeErr error // set in flush

	// set on Close:
	closed   bool
	closeErr error

	// set on first Key:
	blobKey appengine.BlobKey
}

// Verify that Writer implements the io.WriteCloser interface.
var _ io.WriteCloser = (*Writer)(nil)

// Prefix for all blobstore-based files.
const blobstoreFileDirectory = "/blobstore/"

// Prefix (after blobstoreFileDirectory) on all writable blob filenames.
// The part that follows when this is present is the "creation handle",
// which must then be looked up in datastore to find the blob once it's
// been finalized.
const creationHandlePrefix = "writable:"

// Create begins creating a new blob. The provided mimeType if non-empty
// is stored in the blob's BlobInfo in datastore, else defaults to
// application/octet-stream. The returned Writer should be written to,
// then closed, and then its Key method can be called to retrieve the
// newly-created blob key if there were no errors.
func Create(c appengine.Context, mimeType string) (*Writer, error) {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	req := &files.CreateRequest{
		Filesystem:  proto.String("blobstore"),
		ContentType: files.NewFileContentType_ContentType(files.FileContentType_RAW),
		Parameters: []*files.CreateRequest_Parameter{
			&files.CreateRequest_Parameter{
				Name:  proto.String("content_type"),
				Value: proto.String(mimeType),
			}},
	}
	res := &files.CreateResponse{}
	if err := c.Call("file", "Create", req, res, nil); err != nil {
		return nil, err
	}

	w := &Writer{
		c:        c,
		filename: *res.Filename,
	}
	if !strings.HasPrefix(w.filename, blobstoreFileDirectory) {
		return nil, errorf("unexpected filename from files service")
	}

	oreq := &files.OpenRequest{
		Filename:      res.Filename,
		ContentType:   files.NewFileContentType_ContentType(files.FileContentType_RAW),
		OpenMode:      files.NewOpenRequest_OpenMode(files.OpenRequest_APPEND),
		ExclusiveLock: proto.Bool(true),
	}
	ores := &files.OpenResponse{}
	if err := c.Call("file", "Open", oreq, ores, nil); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, errorf("Writer is already closed")
	}
	w.buf = append(w.buf, p...)
	if len(w.buf) >= writeBufferSize {
		w.flush()
		if w.writeErr != nil {
			return 0, w.writeErr
		}
	}
	return len(p), nil
}

func (w *Writer) flush() {
	if len(w.buf) == 0 {
		return
	}
	req := &files.AppendRequest{
		Filename: proto.String(w.filename),
		Data:     w.buf,
	}
	res := &files.AppendResponse{}
	if err := w.c.Call("file", "Append", req, res, nil); err != nil {
		w.writeErr = err
	}
	w.buf = nil
}

// Close flushes outstanding buffered writes and finalizes the blob. After
// calling Close the key can be retrieved by calling Key.
func (w *Writer) Close() (closeErr error) {
	defer func() {
		// Save the error for Key
		w.closeErr = closeErr
	}()
	if w.closed {
		return errorf("Writer is already closed")
	}
	w.closed = true
	w.flush()
	if w.writeErr != nil {
		return w.writeErr
	}
	req := &files.CloseRequest{
		Filename: proto.String(w.filename),
		Finalize: proto.Bool(true),
	}
	res := &files.CloseResponse{}
	return w.c.Call("file", "Close", req, res, nil)
}

// Key returns the created blob key. It must be called after Close.
// An error is returned if Close wasn't called or returned an error.
func (w *Writer) Key() (appengine.BlobKey, error) {
	if !w.closed {
		return "", errorf("cannot call Key before Close")
	}

	if w.blobKey != "" {
		return w.blobKey, w.closeErr
	}

	handle := w.filename[len(blobstoreFileDirectory):]
	if !strings.HasPrefix(handle, creationHandlePrefix) {
		w.blobKey = appengine.BlobKey(handle)
		return w.blobKey, w.closeErr
	}

	k, err := w.keyNewWay(handle)
	if err == nil {
		w.blobKey = k
		return k, nil
	}

	k, err = w.keyOldWay(handle)
	if err == nil {
		w.blobKey = k
	}

	return k, err
}

func (w *Writer) keyNewWay(handle string) (appengine.BlobKey, error) {
	key := datastore.NewKey(w.c, blobFileIndexKind, handle, 0, nil)
	m := make(datastore.Map)
	err := datastore.Get(w.c, key, m)
	if err != nil {
		return zeroKey, err
	}
	blobkeyStr, ok := m[blobKeyPropertyName].(string)
	if !ok {
		return zeroKey, os.ENOENT
	}

	// Double-check that the BlobInfo actually exists.
	// (Consistent with Python.)
	key = datastore.NewKey(w.c, blobInfoKind, blobkeyStr, 0, nil)
	err = datastore.Get(w.c, key, m)
	if err != nil {
		return zeroKey, err
	}
	return appengine.BlobKey(blobkeyStr), nil
}

// keyOldWay looks up a blobkey from its creation_handle the old way:
// by doing an query against __BlobInfo__ entities.  This is now
// deprecated (corollary: the other way doesn't work yet), so we try
// this only after the new way fails, like Python does.
func (w *Writer) keyOldWay(handle string) (appengine.BlobKey, error) {
	query := datastore.NewQuery(blobInfoKind).
		Filter("creation_handle =", handle).
		KeysOnly().
		Limit(1)
	key, err := query.Run(w.c).Next(nil)
	if err != nil {
		if err != datastore.Done {
			return "", errorf("error looking up __BlobInfo__ entity for creation_handle %q: %v", handle, key)
		}
		return "", errorf("didn't find __BlobInfo__ entity for creation_handle %q", handle)
	}
	return appengine.BlobKey(key.StringID()), w.closeErr
}

func init() {
	appengine_internal.RegisterErrorCodeMap("blobstore", pb.BlobstoreServiceError_ErrorCode_name)
}
