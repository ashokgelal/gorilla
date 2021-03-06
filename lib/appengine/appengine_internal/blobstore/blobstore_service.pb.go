// Code generated by protoc-gen-go from "blobstore_service.proto"
// DO NOT EDIT!

package appengine

import proto "goprotobuf.googlecode.com/hg/proto"
import "math"

// Reference proto, math & os imports to suppress error if they are not otherwise used.
var _ = proto.GetString
var _ = math.Inf
var _ error

type BlobstoreServiceError_ErrorCode int32

const (
	BlobstoreServiceError_OK                        BlobstoreServiceError_ErrorCode = 0
	BlobstoreServiceError_INTERNAL_ERROR            BlobstoreServiceError_ErrorCode = 1
	BlobstoreServiceError_URL_TOO_LONG              BlobstoreServiceError_ErrorCode = 2
	BlobstoreServiceError_PERMISSION_DENIED         BlobstoreServiceError_ErrorCode = 3
	BlobstoreServiceError_BLOB_NOT_FOUND            BlobstoreServiceError_ErrorCode = 4
	BlobstoreServiceError_DATA_INDEX_OUT_OF_RANGE   BlobstoreServiceError_ErrorCode = 5
	BlobstoreServiceError_BLOB_FETCH_SIZE_TOO_LARGE BlobstoreServiceError_ErrorCode = 6
	BlobstoreServiceError_ARGUMENT_OUT_OF_RANGE     BlobstoreServiceError_ErrorCode = 8
)

var BlobstoreServiceError_ErrorCode_name = map[int32]string{
	0: "OK",
	1: "INTERNAL_ERROR",
	2: "URL_TOO_LONG",
	3: "PERMISSION_DENIED",
	4: "BLOB_NOT_FOUND",
	5: "DATA_INDEX_OUT_OF_RANGE",
	6: "BLOB_FETCH_SIZE_TOO_LARGE",
	8: "ARGUMENT_OUT_OF_RANGE",
}
var BlobstoreServiceError_ErrorCode_value = map[string]int32{
	"OK":                        0,
	"INTERNAL_ERROR":            1,
	"URL_TOO_LONG":              2,
	"PERMISSION_DENIED":         3,
	"BLOB_NOT_FOUND":            4,
	"DATA_INDEX_OUT_OF_RANGE":   5,
	"BLOB_FETCH_SIZE_TOO_LARGE": 6,
	"ARGUMENT_OUT_OF_RANGE":     8,
}

func NewBlobstoreServiceError_ErrorCode(x BlobstoreServiceError_ErrorCode) *BlobstoreServiceError_ErrorCode {
	e := BlobstoreServiceError_ErrorCode(x)
	return &e
}
func (x BlobstoreServiceError_ErrorCode) String() string {
	return proto.EnumName(BlobstoreServiceError_ErrorCode_name, int32(x))
}

type BlobstoreServiceError struct {
	XXX_unrecognized []byte `json:",omitempty"`
}

func (this *BlobstoreServiceError) Reset()        { *this = BlobstoreServiceError{} }
func (this *BlobstoreServiceError) Error() string { return proto.CompactTextString(this) }

type CreateUploadURLRequest struct {
	SuccessPath               *string `protobuf:"bytes,1,req,name=success_path" json:"success_path,omitempty"`
	MaxUploadSizeBytes        *int64  `protobuf:"varint,2,opt,name=max_upload_size_bytes" json:"max_upload_size_bytes,omitempty"`
	MaxUploadSizePerBlobBytes *int64  `protobuf:"varint,3,opt,name=max_upload_size_per_blob_bytes" json:"max_upload_size_per_blob_bytes,omitempty"`
	XXX_unrecognized          []byte  `json:",omitempty"`
}

func (this *CreateUploadURLRequest) Reset()         { *this = CreateUploadURLRequest{} }
func (this *CreateUploadURLRequest) String() string { return proto.CompactTextString(this) }

type CreateUploadURLResponse struct {
	Url              *string `protobuf:"bytes,1,req,name=url" json:"url,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateUploadURLResponse) Reset()         { *this = CreateUploadURLResponse{} }
func (this *CreateUploadURLResponse) String() string { return proto.CompactTextString(this) }

type DeleteBlobRequest struct {
	BlobKey          []string `protobuf:"bytes,1,rep,name=blob_key" json:"blob_key,omitempty"`
	XXX_unrecognized []byte   `json:",omitempty"`
}

func (this *DeleteBlobRequest) Reset()         { *this = DeleteBlobRequest{} }
func (this *DeleteBlobRequest) String() string { return proto.CompactTextString(this) }

type FetchDataRequest struct {
	BlobKey          *string `protobuf:"bytes,1,req,name=blob_key" json:"blob_key,omitempty"`
	StartIndex       *int64  `protobuf:"varint,2,req,name=start_index" json:"start_index,omitempty"`
	EndIndex         *int64  `protobuf:"varint,3,req,name=end_index" json:"end_index,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *FetchDataRequest) Reset()         { *this = FetchDataRequest{} }
func (this *FetchDataRequest) String() string { return proto.CompactTextString(this) }

type FetchDataResponse struct {
	Data             []byte `protobuf:"bytes,1000,req,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte `json:",omitempty"`
}

func (this *FetchDataResponse) Reset()         { *this = FetchDataResponse{} }
func (this *FetchDataResponse) String() string { return proto.CompactTextString(this) }

type CloneBlobRequest struct {
	BlobKey          []byte `protobuf:"bytes,1,req,name=blob_key" json:"blob_key,omitempty"`
	MimeType         []byte `protobuf:"bytes,2,req,name=mime_type" json:"mime_type,omitempty"`
	TargetAppId      []byte `protobuf:"bytes,3,req,name=target_app_id" json:"target_app_id,omitempty"`
	XXX_unrecognized []byte `json:",omitempty"`
}

func (this *CloneBlobRequest) Reset()         { *this = CloneBlobRequest{} }
func (this *CloneBlobRequest) String() string { return proto.CompactTextString(this) }

type CloneBlobResponse struct {
	BlobKey          []byte `protobuf:"bytes,1,req,name=blob_key" json:"blob_key,omitempty"`
	XXX_unrecognized []byte `json:",omitempty"`
}

func (this *CloneBlobResponse) Reset()         { *this = CloneBlobResponse{} }
func (this *CloneBlobResponse) String() string { return proto.CompactTextString(this) }

type DecodeBlobKeyRequest struct {
	BlobKey          []string `protobuf:"bytes,1,rep,name=blob_key" json:"blob_key,omitempty"`
	XXX_unrecognized []byte   `json:",omitempty"`
}

func (this *DecodeBlobKeyRequest) Reset()         { *this = DecodeBlobKeyRequest{} }
func (this *DecodeBlobKeyRequest) String() string { return proto.CompactTextString(this) }

type DecodeBlobKeyResponse struct {
	Decoded          []string `protobuf:"bytes,1,rep,name=decoded" json:"decoded,omitempty"`
	XXX_unrecognized []byte   `json:",omitempty"`
}

func (this *DecodeBlobKeyResponse) Reset()         { *this = DecodeBlobKeyResponse{} }
func (this *DecodeBlobKeyResponse) String() string { return proto.CompactTextString(this) }

func init() {
	proto.RegisterEnum("appengine.BlobstoreServiceError_ErrorCode", BlobstoreServiceError_ErrorCode_name, BlobstoreServiceError_ErrorCode_value)
}
