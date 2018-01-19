package postgres

import (
	"crypto/sha256"
	"strings"
	"time"

	"context"

	"fmt"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	. "git.containerum.net/ch/user-manager/models"
	"github.com/sirupsen/logrus"
)

const linkQueryColumnsWithUser = "links.link, links.type, links.created_at, links.expired_at, links.is_active, links.sent_at, " +
	"users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist"
const linkQueryColumns = "link, type, created_at, expired_at, is_active, sent_at"

func (db *pgDB) CreateLink(ctx context.Context, linkType umtypes.LinkType, lifeTime time.Duration, user *User) (*Link, error) {
	now := time.Now().UTC()

	db.log.WithFields(logrus.Fields{
		"user":          user.Login,
		"creation_time": now.Format(time.ANSIC),
	}).Infoln("Create new link")

	ret := &Link{User: user}

	ret = &Link{
		Link:      strings.ToUpper(fmt.Sprintf("%x", (sha256.Sum256([]byte(user.ID + string(linkType) + lifeTime.String() + now.String()))))),
		User:      user,
		Type:      linkType,
		CreatedAt: now,
		ExpiredAt: now.Add(lifeTime),
		IsActive:  true,
	}
	rows, err := db.qLog.QueryxContext(ctx, "INSERT INTO links (link, type, created_at, expired_at, is_active, user_id) VALUES "+
		"($1, $2, $3, $4, $5, $6) ON CONFLICT (type, user_id) DO UPDATE SET link = $1, is_active = true, created_at = $3, expired_at = $4 RETURNING "+linkQueryColumns, ret.Link, ret.Type, ret.CreatedAt, ret.ExpiredAt, ret.IsActive, ret.User.ID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, rows.Err()
	}

	if err = rows.Scan(&ret.Link, &ret.Type, &ret.CreatedAt, &ret.ExpiredAt, &ret.IsActive, &ret.SentAt); err != nil {
		return nil, err
	}

	return ret, err
}

func (db *pgDB) GetLinkForUser(ctx context.Context, linkType umtypes.LinkType, user *User) (*Link, error) {
	db.log.Infoln("Get link", linkType, "for", user.Login)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+linkQueryColumns+" FROM links "+
		"WHERE user_id = $1 AND type = $2 AND is_active AND expired_at > NOW()", user.ID, linkType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	link := Link{User: user}
	err = rows.Scan(&link.Link, &link.Type, &link.CreatedAt, &link.ExpiredAt, &link.IsActive, &link.SentAt)

	return &link, err
}

func (db *pgDB) GetLinkFromString(ctx context.Context, strLink string) (*Link, error) {
	db.log.Infoln("Get link", strLink)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+linkQueryColumnsWithUser+" FROM links "+
		"JOIN users ON links.user_id = users.id "+
		"WHERE link = $1 AND links.is_active AND links.expired_at > NOW()", strLink)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	defer rows.Close()
	link := Link{User: &User{}}
	err = rows.Scan(&link.Link, &link.Type, &link.CreatedAt, &link.ExpiredAt, &link.IsActive, &link.SentAt,
		&link.User.ID, &link.User.Login, &link.User.PasswordHash, &link.User.Salt, &link.User.Role,
		&link.User.IsActive, &link.User.IsDeleted, &link.User.IsInBlacklist)

	return &link, err
}

func (db *pgDB) UpdateLink(ctx context.Context, link *Link) error {
	db.log.Infof("Update link %#v", link)
	_, err := db.eLog.ExecContext(ctx, "UPDATE links set type = $2, expired_at = $3, is_active = $4, sent_at = $5 "+
		"WHERE link = $1", link.Link, link.Type, link.ExpiredAt, link.IsActive, link.SentAt)
	return err
}

func (db *pgDB) GetUserLinks(ctx context.Context, user *User) ([]Link, error) {
	db.log.Infoln("Get links for", user.Login)
	var ret []Link
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+linkQueryColumns+" FROM links "+
		"WHERE user_id = $1 AND is_active AND expired_at > NOW()", user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		link := Link{User: user}
		err := rows.Scan(&link.Link, &link.Type, &link.CreatedAt, &link.ExpiredAt, &link.IsActive, &link.SentAt)
		if err != nil {
			return nil, err
		}
		ret = append(ret, link)
	}

	return ret, rows.Err()
}
