package server

import (
	"context"

	"io"

	kube_types "github.com/containerum/kube-client/pkg/model"

	"git.containerum.net/ch/auth/proto"
	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
)

// UserManager is an interface for server "business logic"
type UserManager interface {
	BasicLogin(ctx context.Context, request models.LoginRequest) (*authProto.CreateTokenResponse, error)
	OneTimeTokenLogin(ctx context.Context, request models.OneTimeTokenLoginRequest) (*authProto.CreateTokenResponse, error)
	OAuthLogin(ctx context.Context, request models.OAuthLoginRequest) (*authProto.CreateTokenResponse, error)

	ChangePassword(ctx context.Context, request models.PasswordChangeRequest) (*authProto.CreateTokenResponse, error)
	ResetPassword(ctx context.Context, request models.UserLogin) error
	RestorePassword(ctx context.Context, request models.PasswordRestoreRequest) (*authProto.CreateTokenResponse, error)

	Logout(ctx context.Context) error

	// changes DB state
	CreateUser(ctx context.Context, request models.RegisterRequest) (*models.UserLogin, error)
	ActivateUser(ctx context.Context, request models.Link) (*authProto.CreateTokenResponse, error)
	BlacklistUser(ctx context.Context, request models.UserLogin) error
	UnBlacklistUser(ctx context.Context, request models.UserLogin) error
	UpdateUser(ctx context.Context, newData map[string]interface{}) (*models.User, error)
	PartiallyDeleteUser(ctx context.Context) error
	CompletelyDeleteUser(ctx context.Context, userID string) error
	AddBoundAccount(ctx context.Context, request models.OAuthLoginRequest) error
	DeleteBoundAccount(ctx context.Context, request models.BoundAccountDeleteRequest) error

	// admin methods
	AdminCreateUser(ctx context.Context, request models.UserLogin) (*models.UserLogin, error)
	AdminActivateUser(ctx context.Context, request models.UserLogin) error
	AdminDeactivateUser(ctx context.Context, request models.UserLogin) error
	AdminResetPassword(ctx context.Context, request models.UserLogin) (*models.UserLogin, error)
	AdminSetAdmin(ctx context.Context, request models.UserLogin) error
	AdminUnsetAdmin(ctx context.Context, request models.UserLogin) error

	// not changes DB state
	GetUserLinks(ctx context.Context, userID string) (*models.Links, error)
	GetUserInfo(ctx context.Context) (*models.User, error)
	GetUserInfoByID(ctx context.Context, userID string) (*models.User, error)
	GetUserInfoByLogin(ctx context.Context, login string) (*models.User, error)
	GetUsersLoginID(ctx context.Context, ids []string) (*models.LoginID, error)
	GetBlacklistedUsers(ctx context.Context, page int, perPage int) (*models.UserList, error)
	GetUsers(ctx context.Context, page int, perPage int, filters ...string) (*models.UserList, error)
	GetBoundAccounts(ctx context.Context) (models.BoundAccounts, error)

	LinkResend(ctx context.Context, request models.UserLogin) error

	// checks
	CheckAdmin(ctx context.Context) error
	CheckUserExist(ctx context.Context) error

	// Domain blacklist
	AddDomainToBlacklist(ctx context.Context, request models.Domain) error
	RemoveDomainFromBlacklist(ctx context.Context, domain string) error
	GetBlacklistedDomain(ctx context.Context, domain string) (*models.Domain, error)
	GetBlacklistedDomainsList(ctx context.Context) (*models.DomainListResponse, error)

	//User groups
	GetGroupsList(ctx context.Context, userID string) (*kube_types.UserGroups, error)
	GetGroup(ctx context.Context, groupID string) (*kube_types.UserGroup, error)
	CreateGroup(ctx context.Context, request kube_types.UserGroup) (*string, error)
	AddGroupMembers(ctx context.Context, groupID string, request kube_types.UserGroupMembers) error
	DeleteGroupMember(ctx context.Context, group kube_types.UserGroup, username string) error
	UpdateGroupMemberAccess(ctx context.Context, group kube_types.UserGroup, username, access string) error
	DeleteGroup(ctx context.Context, groupID string) error

	CreateFirstAdmin(password string) error
	io.Closer
}

// Services is a collection of resources needed for server functionality.
type Services struct {
	MailClient        clients.MailClient
	DB                db.DB
	AuthClient        clients.AuthClientCloser
	ReCaptchaClient   clients.ReCaptchaClient
	PermissionsClient clients.PermissionsClient
}
