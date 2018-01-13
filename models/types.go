package models

import (
	"time"

	"io"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/lib/pq"
)

type User struct {
	ID            string           `db:"id"`
	Login         string           `db:"login"`
	PasswordHash  string           `db:"password_hash"` // base64
	Salt          string           `db:"salt"`          // base64
	Role          umtypes.UserRole `db:"role"`
	IsActive      bool             `db:"is_active"`
	IsDeleted     bool             `db:"is_deleted"`
	IsInBlacklist bool             `db:"is_in_blacklist"`
}

type Profile struct {
	ID          string
	Referral    string
	Access      string
	CreatedAt   time.Time
	BlacklistAt pq.NullTime
	DeletedAt   pq.NullTime

	User *User

	Data umtypes.ProfileData
}

type Accounts struct {
	ID       string
	Github   string
	Facebook string
	Google   string

	User *User
}

type Link struct {
	Link      string
	Type      umtypes.LinkType
	CreatedAt time.Time
	ExpiredAt time.Time
	IsActive  bool
	SentAt    pq.NullTime

	User *User
}

type Token struct {
	Token     string
	CreatedAt time.Time
	IsActive  bool
	SessionID string

	User *User
}

type DomainBlacklistEntry struct {
	Domain    string
	CreatedAt time.Time
}

var (
	ErrTransactionBegin    = errors.New("transaction begin error")
	ErrTransactionRollback = errors.New("transaction rollback error")
	ErrTransactionCommit   = errors.New("transaction commit error")
)

type DB interface {
	GetUserByLogin(login string) (*User, error)
	GetUserByID(id string) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	GetBlacklistedUsers(perPage, page int) ([]User, error)
	BlacklistUser(user *User) error

	CreateProfile(profile *Profile) error
	GetProfileByID(id string) (*Profile, error)
	GetProfileByUser(user *User) (*Profile, error)
	UpdateProfile(profile *Profile) error
	GetAllProfiles(perPage, offset int) ([]Profile, error)

	GetUserByBoundAccount(service, accountID string) (*User, error)
	GetUserBoundAccounts(user *User) (*Accounts, error)
	BindAccount(user *User, service, accountID string) error

	BlacklistDomain(domain string) error
	UnBlacklistDomain(domain string) error
	IsDomainBlacklisted(domain string) (bool, error)
	/*GetBlacklistedDomains() ([]DomainBlacklistEntry, error)*/

	CreateLink(linkType umtypes.LinkType, lifeTime time.Duration, user *User) (*Link, error)
	GetLinkForUser(linkType umtypes.LinkType, user *User) (*Link, error)
	GetLinkFromString(strLink string) (*Link, error)
	UpdateLink(link *Link) error
	GetUserLinks(user *User) ([]Link, error)

	GetTokenObject(token string) (*Token, error)
	CreateToken(user *User, sessionID string) (*Token, error)
	GetTokenBySessionID(sessionID string) (*Token, error)
	DeleteToken(token string) error
	UpdateToken(token *Token) error

	// Perform operations inside transaction
	// Transaction commits if `f` returns nil error, rollbacks and forwards error otherwise
	// May return ErrTransactionBegin if transaction start failed,
	// ErrTransactionCommit if commit failed, ErrTransactionRollback if rollback failed
	Transactional(f func(tx DB) error) error

	io.Closer
}
