package models

import (
	"crypto/sha512"
	"encoding/hex"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type LinkType string

const (
	LinkTypeConfirm   LinkType = "confirm"
	LinkTypePwdChange LinkType = "pwd_change"
	LinkTypeDelete    LinkType = "delete"
)

type Link struct {
	Link      string `gorm:"primary_key"`
	User      User
	UserID    string `gorm:"type:uuid;ForeignKey:UserID"`
	Type      LinkType
	CreatedAt time.Time
	ExpiredAt time.Time
	IsActive  bool
	SentAt    time.Time
}

func (db *DB) CreateLink(linkType LinkType, lifeTime time.Duration, user *User) (*Link, error) {
	now := time.Now().UTC()
	ret := &Link{
		Link:      strings.ToUpper(hex.EncodeToString(sha512.New().Sum([]byte(user.ID)))),
		User:      *user,
		Type:      linkType,
		CreatedAt: now,
		ExpiredAt: now.Add(lifeTime),
		IsActive:  true,
	}
	db.log.WithFields(logrus.Fields{
		"user":          user.Login,
		"creation_time": now.Format(time.ANSIC),
	}).Debug("Create activation link")
	return ret, db.conn.Create(ret).Error
}

func (db *DB) GetLinkForUser(linkType LinkType, user *User) (*Link, error) {
	db.log.Debug("Get link", linkType, "for", user.Login)
	var link Link
	resp := db.conn.
		Where("type = ? AND is_active = true AND expires_at > ?", linkType, time.Now().UTC()).
		Model(user).
		Related(&link)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &link, resp.Error
}

func (db *DB) GetLinkFromString(strLink string) (*Link, error) {
	db.log.Debug("Get link", strLink)
	var link Link
	resp := db.conn.Where(Link{Link: strLink}).First(&link)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &link, resp.Error
}

func (db *DB) UpdateLink(link *Link) error {
	db.log.Debugf("Update link %#v", link)
	resp := db.conn.Save(link)
	return resp.Error
}

func (db *DB) GetUserLinks(user *User) ([]*Link, error) {
	db.log.Debug("Get links for", user.Login)
	var ret []*Link
	err := db.conn.Model(Link{}).Find(&ret).Error
	return ret, err
}
