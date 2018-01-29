package models

import (
	"time"

	"io"

	"context"

	"database/sql"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
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
	Type      umtypes.LinkType
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

// Errors which may occur in transactional operations
var (
	ErrTransactionBegin    = errors.New("transaction begin error")
	ErrTransactionRollback = errors.New("transaction rollback error")
	ErrTransactionCommit   = errors.New("transaction commit error")
)

// DB is an interface for persistent data storage (also sometimes called DAO).
type DB interface {
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetAnyUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	CreateUserWebAPI(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	GetBlacklistedUsers(ctx context.Context, limit, offset int) ([]User, error)
	BlacklistUser(ctx context.Context, user *User) error
	UnBlacklistUser(ctx context.Context, user *User) error

	CreateProfile(ctx context.Context, profile *Profile) error
	GetProfileByID(ctx context.Context, id string) (*Profile, error)
	GetProfileByUser(ctx context.Context, user *User) (*Profile, error)
	UpdateProfile(ctx context.Context, profile *Profile) error
	GetAllProfiles(ctx context.Context, perPage, offset int) ([]UserProfileAccounts, error)

	GetUserByBoundAccount(ctx context.Context, service umtypes.OAuthResource, accountID string) (*User, error)
	GetUserBoundAccounts(ctx context.Context, user *User) (*Accounts, error)
	BindAccount(ctx context.Context, user *User, service umtypes.OAuthResource, accountID string) error
	DeleteBoundAccount(ctx context.Context, user *User, service umtypes.OAuthResource) error

	BlacklistDomain(ctx context.Context, domain string, userID string) error
	UnBlacklistDomain(ctx context.Context, domain string) error
	IsDomainBlacklisted(ctx context.Context, domain string) (bool, error)
	GetBlacklistedDomain(ctx context.Context, domain string) (*DomainBlacklistEntry, error)
	GetBlacklistedDomainsList(ctx context.Context) ([]DomainBlacklistEntry, error)

	CreateLink(ctx context.Context, linkType umtypes.LinkType, lifeTime time.Duration, user *User) (*Link, error)
	GetLinkForUser(ctx context.Context, linkType umtypes.LinkType, user *User) (*Link, error)
	GetLinkFromString(ctx context.Context, strLink string) (*Link, error)
	UpdateLink(ctx context.Context, link *Link) error
	GetUserLinks(ctx context.Context, user *User) ([]Link, error)

	GetTokenObject(ctx context.Context, token string) (*Token, error)
	CreateToken(ctx context.Context, user *User, sessionID string) (*Token, error)
	GetTokenBySessionID(ctx context.Context, sessionID string) (*Token, error)
	DeleteToken(ctx context.Context, token string) error
	UpdateToken(ctx context.Context, token *Token) error

	// Perform operations inside transaction
	// Transaction commits if `f` returns nil error, rollbacks and forwards error otherwise
	// May return ErrTransactionBegin if transaction start failed,
	// ErrTransactionCommit if commit failed, ErrTransactionRollback if rollback failed
	Transactional(ctx context.Context, f func(ctx context.Context, tx DB) error) error

	io.Closer
}
