package models

import "time"

type Token struct {
	Token     string `gorm:"primary_key"`
	User      User
	CreatedAt time.Time
	IsActive  bool
	SessionID string `gorm:"type:uuid"`
}
