// Code generated by protoc-gen-go from "user_service.proto"
// DO NOT EDIT!

package appengine

import proto "goprotobuf.googlecode.com/hg/proto"
import "math"

// Reference proto, math & os imports to suppress error if they are not otherwise used.
var _ = proto.GetString
var _ = math.Inf
var _ error

type UserServiceError_ErrorCode int32

const (
	UserServiceError_OK                    UserServiceError_ErrorCode = 0
	UserServiceError_REDIRECT_URL_TOO_LONG UserServiceError_ErrorCode = 1
	UserServiceError_NOT_ALLOWED           UserServiceError_ErrorCode = 2
	UserServiceError_OAUTH_INVALID_TOKEN   UserServiceError_ErrorCode = 3
	UserServiceError_OAUTH_INVALID_REQUEST UserServiceError_ErrorCode = 4
	UserServiceError_OAUTH_ERROR           UserServiceError_ErrorCode = 5
)

var UserServiceError_ErrorCode_name = map[int32]string{
	0: "OK",
	1: "REDIRECT_URL_TOO_LONG",
	2: "NOT_ALLOWED",
	3: "OAUTH_INVALID_TOKEN",
	4: "OAUTH_INVALID_REQUEST",
	5: "OAUTH_ERROR",
}
var UserServiceError_ErrorCode_value = map[string]int32{
	"OK": 0,
	"REDIRECT_URL_TOO_LONG": 1,
	"NOT_ALLOWED":           2,
	"OAUTH_INVALID_TOKEN":   3,
	"OAUTH_INVALID_REQUEST": 4,
	"OAUTH_ERROR":           5,
}

func NewUserServiceError_ErrorCode(x UserServiceError_ErrorCode) *UserServiceError_ErrorCode {
	e := UserServiceError_ErrorCode(x)
	return &e
}
func (x UserServiceError_ErrorCode) String() string {
	return proto.EnumName(UserServiceError_ErrorCode_name, int32(x))
}

type UserServiceError struct {
	XXX_unrecognized []byte `json:",omitempty"`
}

func (this *UserServiceError) Reset()        { *this = UserServiceError{} }
func (this *UserServiceError) Error() string { return proto.CompactTextString(this) }

type CreateLoginURLRequest struct {
	DestinationUrl    *string `protobuf:"bytes,1,req,name=destination_url" json:"destination_url,omitempty"`
	AuthDomain        *string `protobuf:"bytes,2,opt,name=auth_domain" json:"auth_domain,omitempty"`
	FederatedIdentity *string `protobuf:"bytes,3,opt,name=federated_identity" json:"federated_identity,omitempty"`
	XXX_unrecognized  []byte  `json:",omitempty"`
}

func (this *CreateLoginURLRequest) Reset()         { *this = CreateLoginURLRequest{} }
func (this *CreateLoginURLRequest) String() string { return proto.CompactTextString(this) }

type CreateLoginURLResponse struct {
	LoginUrl         *string `protobuf:"bytes,1,req,name=login_url" json:"login_url,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateLoginURLResponse) Reset()         { *this = CreateLoginURLResponse{} }
func (this *CreateLoginURLResponse) String() string { return proto.CompactTextString(this) }

type CreateLogoutURLRequest struct {
	DestinationUrl   *string `protobuf:"bytes,1,req,name=destination_url" json:"destination_url,omitempty"`
	AuthDomain       *string `protobuf:"bytes,2,opt,name=auth_domain" json:"auth_domain,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateLogoutURLRequest) Reset()         { *this = CreateLogoutURLRequest{} }
func (this *CreateLogoutURLRequest) String() string { return proto.CompactTextString(this) }

type CreateLogoutURLResponse struct {
	LogoutUrl        *string `protobuf:"bytes,1,req,name=logout_url" json:"logout_url,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateLogoutURLResponse) Reset()         { *this = CreateLogoutURLResponse{} }
func (this *CreateLogoutURLResponse) String() string { return proto.CompactTextString(this) }

type GetOAuthUserRequest struct {
	Scope            *string `protobuf:"bytes,1,opt,name=scope" json:"scope,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *GetOAuthUserRequest) Reset()         { *this = GetOAuthUserRequest{} }
func (this *GetOAuthUserRequest) String() string { return proto.CompactTextString(this) }

type GetOAuthUserResponse struct {
	Email            *string `protobuf:"bytes,1,req,name=email" json:"email,omitempty"`
	UserId           *string `protobuf:"bytes,2,req,name=user_id" json:"user_id,omitempty"`
	AuthDomain       *string `protobuf:"bytes,3,req,name=auth_domain" json:"auth_domain,omitempty"`
	UserOrganization *string `protobuf:"bytes,4,opt,name=user_organization" json:"user_organization,omitempty"`
	IsAdmin          *bool   `protobuf:"varint,5,opt,name=is_admin,def=0" json:"is_admin,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *GetOAuthUserResponse) Reset()         { *this = GetOAuthUserResponse{} }
func (this *GetOAuthUserResponse) String() string { return proto.CompactTextString(this) }

const Default_GetOAuthUserResponse_IsAdmin bool = false

type CheckOAuthSignatureRequest struct {
	XXX_unrecognized []byte `json:",omitempty"`
}

func (this *CheckOAuthSignatureRequest) Reset()         { *this = CheckOAuthSignatureRequest{} }
func (this *CheckOAuthSignatureRequest) String() string { return proto.CompactTextString(this) }

type CheckOAuthSignatureResponse struct {
	OauthConsumerKey *string `protobuf:"bytes,1,req,name=oauth_consumer_key" json:"oauth_consumer_key,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CheckOAuthSignatureResponse) Reset()         { *this = CheckOAuthSignatureResponse{} }
func (this *CheckOAuthSignatureResponse) String() string { return proto.CompactTextString(this) }

type CreateFederatedLoginRequest struct {
	ClaimedId        *string `protobuf:"bytes,1,req,name=claimed_id" json:"claimed_id,omitempty"`
	ContinueUrl      *string `protobuf:"bytes,2,req,name=continue_url" json:"continue_url,omitempty"`
	Authority        *string `protobuf:"bytes,3,opt,name=authority" json:"authority,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateFederatedLoginRequest) Reset()         { *this = CreateFederatedLoginRequest{} }
func (this *CreateFederatedLoginRequest) String() string { return proto.CompactTextString(this) }

type CreateFederatedLoginResponse struct {
	RedirectedUrl    *string `protobuf:"bytes,1,req,name=redirected_url" json:"redirected_url,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateFederatedLoginResponse) Reset()         { *this = CreateFederatedLoginResponse{} }
func (this *CreateFederatedLoginResponse) String() string { return proto.CompactTextString(this) }

type CreateFederatedLogoutRequest struct {
	DestinationUrl   *string `protobuf:"bytes,1,req,name=destination_url" json:"destination_url,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateFederatedLogoutRequest) Reset()         { *this = CreateFederatedLogoutRequest{} }
func (this *CreateFederatedLogoutRequest) String() string { return proto.CompactTextString(this) }

type CreateFederatedLogoutResponse struct {
	LogoutUrl        *string `protobuf:"bytes,1,req,name=logout_url" json:"logout_url,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *CreateFederatedLogoutResponse) Reset()         { *this = CreateFederatedLogoutResponse{} }
func (this *CreateFederatedLogoutResponse) String() string { return proto.CompactTextString(this) }

func init() {
	proto.RegisterEnum("appengine.UserServiceError_ErrorCode", UserServiceError_ErrorCode_name, UserServiceError_ErrorCode_value)
}
