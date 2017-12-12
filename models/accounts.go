package models

type Accounts struct {
	ID       string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	User     User
	UserID   string `gorm:"type:uuid;ForeignKey:UserID"`
	Github   string
	Facebook string
	Google   string
}
