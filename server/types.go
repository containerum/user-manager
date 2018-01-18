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

// "Business logic" here
type UserManager interface {
	BasicLogin(ctx context.Context, request umtypes.BasicLoginRequest) (*auth.CreateTokenResponse, error)
	OneTimeTokenLogin(ctx context.Context, request umtypes.OneTimeTokenLoginRequest) (*auth.CreateTokenResponse, error)
	OAuthLogin(ctx context.Context, request umtypes.OAuthLoginRequest) (*auth.CreateTokenResponse, error)
	WebAPILogin(ctx context.Context, request umtypes.WebAPILoginRequest) (map[string]interface{}, error) // login through old web-api

	ChangePassword(ctx context.Context, request umtypes.PasswordChangeRequest) (*auth.CreateTokenResponse, error)
	ResetPassword(ctx context.Context, request umtypes.PasswordResetRequest) error
	RestorePassword(ctx context.Context, request umtypes.PasswordRestoreRequest) (*auth.CreateTokenResponse, error)

	Logout(ctx context.Context) error

	// changes DB state
	CreateUser(ctx context.Context, request umtypes.UserCreateRequest) (*umtypes.UserCreateResponse, error)
	ActivateUser(ctx context.Context, request umtypes.ActivateRequest) (*auth.CreateTokenResponse, error)
	BlacklistUser(ctx context.Context, request umtypes.UserToBlacklistRequest) error
	UpdateUser(ctx context.Context, newData map[string]interface{}) (*umtypes.UserInfoGetResponse, error)
	PartiallyDeleteUser(ctx context.Context) error
	CompletelyDeleteUser(ctx context.Context, userID string) error

	// not changes DB state
	GetUserLinks(ctx context.Context, userID string) (*umtypes.LinksGetResponse, error)
	GetUserInfo(ctx context.Context) (*umtypes.UserInfoGetResponse, error)
	GetUserInfoByID(ctx context.Context, userID string) (*umtypes.UserInfoByIDGetResponse, error)
	GetBlacklistedUsers(ctx context.Context, params umtypes.UserListQuery) (*umtypes.BlacklistGetResponse, error)
	GetUsers(ctx context.Context, params umtypes.UserListQuery, filters ...string) (*umtypes.UserListGetResponse, error)

	LinkResend(ctx context.Context, request umtypes.ResendLinkRequest) error

	io.Closer
}

type Services struct {
	MailClient            clients.MailClient
	DB                    models.DB
	AuthClient            clients.AuthClientCloser
	ReCaptchaClient       clients.ReCaptchaClient
	WebAPIClient          clients.WebAPIClient
	ResourceServiceClient clients.ResourceServiceClient
}

type InternalError struct {
	Err *errors.Error
}

func (e *InternalError) Error() string {
	return e.Err.Error()
}

type AccessDeniedError struct {
	Err *errors.Error
}

func (e *AccessDeniedError) Error() string {
	return e.Err.Error()
}

type NotFoundError struct {
	Err *errors.Error
}

func (e *NotFoundError) Error() string {
	return e.Err.Error()
}

type BadRequestError struct {
	Err *errors.Error
}

func (e *BadRequestError) Error() string {
	return e.Err.Error()
}

type AlreadyExistsError struct {
	Err *errors.Error
}

func (e *AlreadyExistsError) Error() string {
	return e.Err.Error()
}

type WebAPIError struct {
	Err        *errors.Error
	StatusCode int
}

func (e *WebAPIError) Error() string {
	return e.Err.Error()
}
