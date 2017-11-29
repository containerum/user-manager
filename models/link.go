package models

import "time"

type LinkType string

const (
	LinkTypeConfigrm  LinkType = "confirm"
	LinkTypePwdChange LinkType = "pwd_change"
	LinkTypeDelete    LinkType = "delete"
)

type Link struct {
	Link      string `gorm:"primary_key"`
	User      User
	Type      LinkType
	CreatedAt time.Time
	ExpiredAt time.Time
	IsActive  bool
}
