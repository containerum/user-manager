package models

import "time"

type Profile struct {
	ID            string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"` // use UUID v4 as primary key (good support in psql)
	User          User
	Referral      string
	Access        string
	Data          string
	CreatedAt     time.Time
	NullBalanceAt time.Time
	DeletedAt     time.Time
}
