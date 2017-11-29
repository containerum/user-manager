package models

type Accounts struct {
	ID       string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	User     User
	Github   string
	Facebook string
}
