package server

import (
	"context"

	"io"

	"git.containerum.net/ch/auth/proto"
	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/db"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
)

// UserManager is an interface for server "business logic"
type UserManager interface {
	BasicLogin(ctx context.Context, request umtypes.LoginRequest) (*authProto.CreateTokenResponse, error)
	OneTimeTokenLogin(ctx context.Context, request umtypes.OneTimeTokenLoginRequest) (*authProto.CreateTokenResponse, error)
	OAuthLogin(ctx context.Context, request umtypes.OAuthLoginRequest) (*authProto.CreateTokenResponse, error)

	ChangePassword(ctx context.Context, request umtypes.PasswordRequest) (*authProto.CreateTokenResponse, error)
	ResetPassword(ctx context.Context, request umtypes.UserLogin) error
	RestorePassword(ctx context.Context, request umtypes.PasswordRequest) (*authProto.CreateTokenResponse, error)

	Logout(ctx context.Context) error

	// changes DB state
	CreateUser(ctx context.Context, request umtypes.RegisterRequest) (*umtypes.User, error)
	ActivateUser(ctx context.Context, request umtypes.Link) (*authProto.CreateTokenResponse, error)
	BlacklistUser(ctx context.Context, request umtypes.UserLogin) error
	UnBlacklistUser(ctx context.Context, request umtypes.UserLogin) error
	UpdateUser(ctx context.Context, newData map[string]interface{}) (*umtypes.User, error)
	PartiallyDeleteUser(ctx context.Context) error
	CompletelyDeleteUser(ctx context.Context, userID string) error
	AddBoundAccount(ctx context.Context, request umtypes.OAuthLoginRequest) error
	DeleteBoundAccount(ctx context.Context, request umtypes.BoundAccountDeleteRequest) error

	// not changes DB state
	GetUserLinks(ctx context.Context, userID string) (*umtypes.Links, error)
	GetUserInfo(ctx context.Context) (*umtypes.User, error)
	GetUserInfoByID(ctx context.Context, userID string) (*umtypes.User, error)
	GetUserInfoByLogin(ctx context.Context, login string) (*umtypes.User, error)
	GetUsersLoginID(ctx context.Context) (*map[string]string, error)
	GetBlacklistedUsers(ctx context.Context, page int, perPage int) (*umtypes.UserList, error)
	GetUsers(ctx context.Context, page int, perPage int, filters ...string) (*umtypes.UserList, error)
	GetBoundAccounts(ctx context.Context) (map[string]string, error)

	LinkResend(ctx context.Context, request umtypes.UserLogin) error

	// checks
	CheckAdmin(ctx context.Context) error
	CheckUserExist(ctx context.Context) error

	// Domain blacklist
	AddDomainToBlacklist(ctx context.Context, request umtypes.Domain) error
	RemoveDomainFromBlacklist(ctx context.Context, domain string) error
	GetBlacklistedDomain(ctx context.Context, domain string) (*umtypes.Domain, error)
	GetBlacklistedDomainsList(ctx context.Context) (*umtypes.DomainListResponse, error)

	io.Closer
}

// Services is a collection of resources needed for server functionality.
type Services struct {
	MailClient            clients.MailClient
	DB                    db.DB
	AuthClient            clients.AuthClientCloser
	ReCaptchaClient       clients.ReCaptchaClient
	ResourceServiceClient clients.ResourceServiceClient
}
