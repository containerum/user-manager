package token

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"git.containerum.net/ch/auth/utils"
	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
)

// Kind is a token kind (see OAuth 2 standards). According to standard it can be only KindAccess and KindRefresh.
type Kind int

const (
	// KindAccess represents access token
	KindAccess Kind = iota

	// KindRefresh represents refresh token
	KindRefresh
)

// ExtensionFields is an advanced fields included to JWT
type ExtensionFields struct {
	UserIDHash  string `json:"user_id,omitempty"`
	Role        string `json:"role,omitempty"`
	PartTokenID string `json:"part_token_id,omitempty"`
}

// IssuedToken describes a token
type IssuedToken struct {
	Value    string
	ID       *common.UUID
	LifeTime time.Duration
}

// Issuer is interface for creating access and refresh tokens.
type Issuer interface {
	IssueTokens(extensionFields ExtensionFields) (accessToken, refreshToken *IssuedToken, err error)
}

// ValidationResult describes token validation result.
type ValidationResult struct {
	Valid bool
	ID    *common.UUID
	Kind  Kind
}

// Validator is interface for validating tokens
type Validator interface {
	ValidateToken(token string) (result *ValidationResult, err error)
}

// IssuerValidator is an interface for token factory
type IssuerValidator interface {
	Issuer
	Validator
}

// EncodeAccessObjects encodes resource access objects to store in database
func EncodeAccessObjects(req []*auth.AccessObject) string {
	ret, _ := json.Marshal(req)
	return base64.StdEncoding.EncodeToString(ret)
}

// DecodeAccessObjects decodes resource access object from database record
func DecodeAccessObjects(value string) (ret []*auth.AccessObject) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return make([]*auth.AccessObject, 0)
	}
	err = json.Unmarshal(decoded, &ret)
	if err != nil {
		return make([]*auth.AccessObject, 0)
	}
	return
}

// RequestToRecord prepares a value to store in database
func RequestToRecord(req *auth.CreateTokenRequest, token *IssuedToken) *auth.StoredToken {
	return &auth.StoredToken{
		TokenId:       token.ID,
		UserAgent:     req.UserAgent,
		Platform:      utils.ShortUserAgent(req.UserAgent),
		Fingerprint:   req.Fingerprint,
		UserId:        req.UserId,
		UserRole:      req.UserRole,
		UserNamespace: EncodeAccessObjects(req.Access.Namespace),
		UserVolume:    EncodeAccessObjects(req.Access.Volume),
		RwAccess:      req.RwAccess,
		UserIp:        req.UserIp,
		PartTokenId:   req.PartTokenId,
	}
}
