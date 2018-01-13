package server

import (
	"context"

	"io"

	"git.containerum.net/ch/grpc-proto-files/auth"
	umtypes "git.containerum.net/ch/json-types/user-manager"
)

type UserManager interface {
	BasicLogin(ctx context.Context, request umtypes.BasicLoginRequest) (auth.CreateTokenResponse, error)
	OneTimeTokenLogin(ctx context.Context, request umtypes.OneTimeTokenLoginRequest) (auth.CreateTokenResponse, error)
	OAuthLogin(ctx context.Context, request umtypes.OAuthLoginRequest) (auth.CreateTokenResponse, error)
	WebAPILogin(ctx context.Context, request umtypes.WebAPILoginRequest) (map[string]interface{}, error) // login through old web-api

	ChangePassword(ctx context.Context, request umtypes.PasswordChangeRequest) (auth.CreateTokenResponse, error)
	ResetPassword(ctx context.Context, request umtypes.PasswordResetRequest) error
	RestorePassword(ctx context.Context, request umtypes.PasswordRestoreRequest) (auth.CreateTokenResponse, error)

	Logout(ctx context.Context) error

	// changes DB state
	CreateUser(ctx context.Context, request umtypes.UserCreateRequest) (umtypes.UserCreateResponse, error)
	ActivateUser(ctx context.Context, request umtypes.ActivateRequest) (auth.CreateTokenResponse, error)
	BlacklistUser(ctx context.Context, request umtypes.UserToBlacklistRequest) error
	UpdateUser(ctx context.Context, newData umtypes.ProfileData) (umtypes.UserInfoGetResponse, error)
	PartiallyDeleteUser(ctx context.Context) error
	CompletelyDeleteUser(ctx context.Context) error

	// not changes DB state
	GetUserLinks(ctx context.Context) (umtypes.LinksGetResponse, error)
	GetUserInfo(ctx context.Context) (umtypes.UserInfoGetResponse, error)

	GetBlacklistedUsers(ctx context.Context, params umtypes.UserListQuery) (umtypes.BlacklistGetResponse, error)
	GetUsers(ctx context.Context, params umtypes.UserListQuery, filters ...string) (umtypes.UserListGetResponse, error)

	LinkResend(ctx context.Context, request umtypes.ResendLinkRequest) error

	io.Closer
}
