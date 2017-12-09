package models

type UserRole int

const (
	RoleUser UserRole = iota
	RoleAdmin
)

type User struct {
	ID            string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"` // use UUID v4 as primary key (good support in psql)
	Login         string
	PasswordHash  string // base64
	Salt          string // base64
	Role          UserRole
	IsActive      bool
	IsDeleted     bool
	IsInBlacklist bool
}

func (db *DB) GetUserByLogin(login string) (*User, error) {
	db.log.Debug("Get user by login", login)
	var user User
	res := db.conn.Where(&User{Login: login, IsDeleted: false}).First(&user)
	if res.RecordNotFound() {
		return nil, nil
	}
	return &user, res.Error
}

func (db *DB) GetUserByID(id string) (*User, error) {
	db.log.Debug("Get user by id", id)
	var user User
	res := db.conn.Where(&User{ID: id, IsDeleted: false}).First(&user)
	if res.RecordNotFound() {
		return nil, nil
	}
	return &user, res.Error
}

func (db *DB) CreateUser(user *User) error {
	db.log.Debug("Create user", user.Login)
	return db.conn.Create(user).Error
}

func (db *DB) UpdateUser(user *User) error {
	db.log.Debug("Update user", user.Login)
	return db.conn.Save(user).Error
}
