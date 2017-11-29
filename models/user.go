package models

type UserRole int

const (
	RoleUser UserRole = iota
	RoleAdmin
)

type User struct {
	ID           string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"` // use UUID v4 as primary key (good support in psql)
	Login        string
	PasswordHash string // base64
	Salt         string // base64
	Role         UserRole
	IsActive     bool
	IsDeleted    bool
}

