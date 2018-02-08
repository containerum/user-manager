package server

import (
	"context"

	"io"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
)

// UserManager is an interface for server "business logic"
type UserManager interface {
	BasicLogin(ctx context.Context, request umtypes.BasicLoginRequest) (*auth.CreateTokenResponse, error)
	OneTimeTokenLogin(ctx context.Context, request umtypes.OneTimeTokenLoginRequest) (*auth.CreateTokenResponse, error)
	OAuthLogin(ctx context.Context, request umtypes.OAuthLoginRequest) (*auth.CreateTokenResponse, error)
	WebAPILogin(ctx context.Context, request umtypes.WebAPILoginRequest) (*umtypes.WebAPILoginResponse, error) // login through old web-api

	ChangePassword(ctx context.Context, request umtypes.PasswordChangeRequest) (*auth.CreateTokenResponse, error)
	ResetPassword(ctx context.Context, request umtypes.PasswordResetRequest) error
	RestorePassword(ctx context.Context, request umtypes.PasswordRestoreRequest) (*auth.CreateTokenResponse, error)

	Logout(ctx context.Context) error

	// changes DB state
	CreateUser(ctx context.Context, request umtypes.UserCreateRequest) (*umtypes.UserCreateResponse, error)
	ActivateUser(ctx context.Context, request umtypes.ActivateRequest) (*auth.CreateTokenResponse, error)
	BlacklistUser(ctx context.Context, request umtypes.UserToBlacklistRequest) error
	UnBlacklistUser(ctx context.Context, request umtypes.UserToBlacklistRequest) error
	UpdateUser(ctx context.Context, newData map[string]interface{}) (*umtypes.UserInfoGetResponse, error)
	PartiallyDeleteUser(ctx context.Context) error
	CompletelyDeleteUser(ctx context.Context, userID string) error
	AddBoundAccount(ctx context.Context, request umtypes.OAuthLoginRequest) error
	DeleteBoundAccount(ctx context.Context, request umtypes.BoundAccountDeleteRequest) error

	// not changes DB state
	GetUserLinks(ctx context.Context, userID string) (*umtypes.LinksGetResponse, error)
	GetUserInfo(ctx context.Context) (*umtypes.UserInfoGetResponse, error)
	GetUserInfoByID(ctx context.Context, userID string) (*umtypes.UserInfoByIDGetResponse, error)
	GetUserInfoByLogin(ctx context.Context, login string) (*umtypes.UserInfoByLoginGetResponse, error)
	GetBlacklistedUsers(ctx context.Context, params umtypes.UserListQuery) (*umtypes.BlacklistGetResponse, error)
	GetUsers(ctx context.Context, params umtypes.UserListQuery, filters ...string) (*umtypes.UserListGetResponse, error)
	GetBoundAccounts(ctx context.Context) (*umtypes.BoundAccountsResponce, error)

	LinkResend(ctx context.Context, request umtypes.ResendLinkRequest) error

	// checks
	CheckAdmin(ctx context.Context) error
	CheckUserExist(ctx context.Context) error

	// Domain blacklist
	AddDomainToBlacklist(ctx context.Context, request umtypes.DomainToBlacklistRequest) error
	RemoveDomainFromBlacklist(ctx context.Context, domain string) error
	GetBlacklistedDomain(ctx context.Context, domain string) (*umtypes.DomainResponse, error)
	GetBlacklistedDomainsList(ctx context.Context) (*umtypes.DomainListResponse, error)

	io.Closer
}

// Services is a collection of resources needed for server functionality.
type Services struct {
	MailClient            clients.MailClient
	DB                    models.DB
	AuthClient            clients.AuthClientCloser
	ReCaptchaClient       clients.ReCaptchaClient
	WebAPIClient          clients.WebAPIClient
	ResourceServiceClient clients.ResourceServiceClient
}

// InternalError describes server errors which should not be exposed to client explicitly.
type InternalError struct {
	Err *errors.Error
}

func (e *InternalError) Error() string {
	return e.Err.Error()
}

// AccessDeniedError describes error if client has no access to resource, method, etc.
type AccessDeniedError struct {
	Err *errors.Error
}

func (e *AccessDeniedError) Error() string {
	return e.Err.Error()
}

// NotFoundError describes error returned if requested resource was not found
type NotFoundError struct {
	Err *errors.Error
}

func (e *NotFoundError) Error() string {
	return e.Err.Error()
}

// BadRequestError describes error returned if request was malformed.
type BadRequestError struct {
	Err *errors.Error
}

func (e *BadRequestError) Error() string {
	return e.Err.Error()
}

// AlreadyExistsError describes error returned if client attempts to create resource or register with username which already exists.
type AlreadyExistsError struct {
	Err *errors.Error
}

func (e *AlreadyExistsError) Error() string {
	return e.Err.Error()
}

// AlreadyExistsError describes error returned if the request has not been applied because it lacks valid authentication credentials for the target resource.
type UnauthorizedError struct {
	Err *errors.Error
}

func (e *UnauthorizedError) Error() string {
	return e.Err.Error()
}

// WebAPIError describes error returned from web-api service.
type WebAPIError struct {
	Err        *errors.Error
	StatusCode int
}

func (e *WebAPIError) Error() string {
	return e.Err.Error()
}
