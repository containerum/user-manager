package db

import (
	"time"

	"io"

	"context"

	"database/sql"

	"errors"

	"git.containerum.net/ch/user-manager/pkg/models"
	"github.com/lib/pq"
)

//UserProfileAccounts descrobes full information about user
type UserProfileAccounts struct {
	User     *User
	Profile  *Profile
	Accounts *Accounts
}

// User describes user model. It should be used only inside this project.
type User struct {
	ID            string `db:"id"`
	Login         string `db:"login"`
	PasswordHash  string `db:"password_hash"` // base64
	Salt          string `db:"salt"`          // base64
	Role          string `db:"role"`
	IsActive      bool   `db:"is_active"`
	IsDeleted     bool   `db:"is_deleted"`
	IsInBlacklist bool   `db:"is_in_blacklist"`
}

// Profile describes user`s profile model. It should be used only inside this project.
type Profile struct {
	ID          sql.NullString
	Referral    sql.NullString
	Access      sql.NullString
	CreatedAt   pq.NullTime
	BlacklistAt pq.NullTime
	DeletedAt   pq.NullTime
	LastLogin   pq.NullTime

	User *User

	Data map[string]interface{}
}

// Accounts describes user`s bound accounts. It should be used only inside this project.
type Accounts struct {
	ID       sql.NullString
	Github   sql.NullString
	Facebook sql.NullString
	Google   sql.NullString

	User *User
}

// Link describes link (for activation, password change, etc.) model. It should be used only inside this project.
type Link struct {
	Link      string
	Type      models.LinkType
	CreatedAt time.Time
	ExpiredAt time.Time
	IsActive  bool
	SentAt    pq.NullTime

	User *User
}

// Token describes one-time token model (used for login), It should be used only inside this project.
type Token struct {
	Token     string
	CreatedAt time.Time
	IsActive  bool
	SessionID string

	User *User
}

// DomainBlacklistEntry describes one blacklisted email domain.
// Registration with email for this domain must be rejected. It should be used only inside this project.
type DomainBlacklistEntry struct {
	Domain    string         `db:"domain"`
	CreatedAt time.Time      `db:"created_at"`
	AddedBy   sql.NullString `db:"added_by"`
}

// UserGroup describes user group model. It should be used only inside this project.
type UserGroup struct {
	ID         string      `db:"id"`
	Label      string      `db:"label"`
	OwnerID    string      `db:"owner_user_id"`
	OwnerLogin string      `db:"owner_login"`
	CreatedAt  pq.NullTime `db:"created_at"`
}

// UserGroupMember describes user group member model. It should be used only inside this project.
type UserGroupMember struct {
	ID      string      `db:"id"`
	Login   string      `db:"login"`
	GroupID string      `db:"group_id"`
	UserID  string      `db:"user_id"`
	Access  string      `db:"default_access"`
	AddedAt pq.NullTime `db:"added_at"`
}

// Errors which may occur in transactional operations
var (
	ErrTransactionBegin    = errors.New("transaction begin error")
	ErrTransactionRollback = errors.New("transaction rollback error")
	ErrTransactionCommit   = errors.New("transaction commit error")
)

// DB is an interface for persistent data storage (also sometimes called DAO).
type DB interface {
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	GetAnyUserByLogin(ctx context.Context, login string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetAnyUserByID(ctx context.Context, id string) (*User, error)
	GetUsersLoginID(ctx context.Context, ids []string) ([]User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	GetBlacklistedUsers(ctx context.Context, limit, offset int) ([]User, error)
	BlacklistUser(ctx context.Context, user *User) error
	UnBlacklistUser(ctx context.Context, user *User) error

	CreateProfile(ctx context.Context, profile *Profile) error
	GetProfileByID(ctx context.Context, id string) (*Profile, error)
	GetProfileByUser(ctx context.Context, user *User) (*Profile, error)
	UpdateProfile(ctx context.Context, profile *Profile) error
	GetAllProfiles(ctx context.Context, perPage, offset uint) ([]UserProfileAccounts, uint, error)

	GetUserByBoundAccount(ctx context.Context, service models.OAuthResource, accountID string) (*User, error)
	GetUserBoundAccounts(ctx context.Context, user *User) (*Accounts, error)
	BindAccount(ctx context.Context, user *User, service models.OAuthResource, accountID string) error
	DeleteBoundAccount(ctx context.Context, user *User, service models.OAuthResource) error

	BlacklistDomain(ctx context.Context, domain string, userID string) error
	UnBlacklistDomain(ctx context.Context, domain string) error
	IsDomainBlacklisted(ctx context.Context, domain string) (bool, error)
	GetBlacklistedDomain(ctx context.Context, domain string) (*DomainBlacklistEntry, error)
	GetBlacklistedDomainsList(ctx context.Context) ([]DomainBlacklistEntry, error)

	CreateLink(ctx context.Context, linkType models.LinkType, lifeTime time.Duration, user *User) (*Link, error)
	GetLinkForUser(ctx context.Context, linkType models.LinkType, user *User) (*Link, error)
	GetLinkFromString(ctx context.Context, strLink string) (*Link, error)
	UpdateLink(ctx context.Context, link *Link) error
	GetUserLinks(ctx context.Context, user *User) ([]Link, error)

	GetTokenObject(ctx context.Context, token string) (*Token, error)
	CreateToken(ctx context.Context, user *User, sessionID string) (*Token, error)
	GetTokenBySessionID(ctx context.Context, sessionID string) (*Token, error)
	DeleteToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, token *Token) error

	GetGroup(ctx context.Context, groupID string) (*UserGroup, error)
	GetGroupMembers(ctx context.Context, groupID string) ([]UserGroupMember, error)
	GetUserGroupsIDsAccesses(ctx context.Context, userID string, isAdmin bool) (map[string]string, error)
	GetGroupListLabelID(ctx context.Context, ids []string) ([]UserGroup, error)
	GetGroupListByIDs(ctx context.Context, ids []string) ([]UserGroup, error)
	CreateGroup(ctx context.Context, group *UserGroup) error
	DeleteGroup(ctx context.Context, groupID string) error

	AddGroupMembers(ctx context.Context, member *UserGroupMember) error
	DeleteGroupMember(ctx context.Context, userID string, groupID string) error
	DeleteGroupMemberFromAllGroups(ctx context.Context, userID string) error
	UpdateGroupMember(ctx context.Context, userID string, groupID string, access string) error
	CountGroupMembers(ctx context.Context, groupID string) (*uint, error)

	UpdateLastLogin(ctx context.Context, profileID, lastlogin string) error

	CountAdmins(ctx context.Context) (*int, error)

	GetAnyUserByLoginWOContext(login string) (*User, error)
	CreateUserWOContext(user *User) error
	CreateProfileWOContext(profile *Profile) error
	UpdateUserWOContext(user *User) error
	// Perform operations inside transaction
	// Transaction commits if `f` returns nil error, rollbacks and forwards error otherwise
	// May return ErrTransactionBegin if transaction start failed,
	// ErrTransactionCommit if commit failed, ErrTransactionRollback if rollback failed
	Transactional(ctx context.Context, f func(ctx context.Context, tx DB) error) error

	io.Closer
}
