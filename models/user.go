package models

import "github.com/jmoiron/sqlx"

type UserRole int

const (
	RoleUser UserRole = iota
	RoleAdmin
)

type User struct {
	ID            string
	Login         string
	PasswordHash  string // base64
	Salt          string // base64
	Role          UserRole
	IsActive      bool
	IsDeleted     bool
	IsInBlacklist bool
}

const userQueryColumns = "(id, login, password_hash, salt, role, is_active, is_deleted, is_in_blacklist)"

func (db *DB) GetUserByLogin(login string) (*User, error) {
	db.log.Debug("Get user by login", login)
	var user User
	rows, err := db.qLog.Queryx("SELECT "+userQueryColumns+" FROM users WHERE login = '$1' AND NOT is_deleted", login)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	err = rows.StructScan(&user)
	return &user, err
}

func (db *DB) GetUserByID(id string) (*User, error) {
	db.log.Debug("Get user by id", id)
	var user User
	rows, err := db.qLog.Queryx("SELECT "+userQueryColumns+" FROM users WHERE id = '$1' AND NOT is_deleted", id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	err = rows.StructScan(&user)
	return &user, err
}

func (db *DB) CreateUser(user *User) error {
	db.log.Debug("Create user", user.Login)
	rows, err := db.qLog.Queryx("INSERT INTO users (login, password_hash, salt, role) "+
		"VALUES ('$1', '$2', '$3', $4) RETURNING id",
		user.Login, user.PasswordHash, user.Salt, user.Role)
	if err != nil {
		return err
	}
	if rows.Next() {
		rows.Scan(&user.ID)
	}
	return rows.Err()
}

func (db *DB) UpdateUser(user *User) error {
	db.log.Debug("Update user", user.Login)
	_, err := db.eLog.Exec("UPDATE users SET "+
		"login = '$2', password_hash = '$3', salt = '$4', role = $5, is_active = $5 WHERE id = '$1'",
		user.ID, user.Login, user.PasswordHash, user.Salt, user.Role, user.IsActive)
	return err
}

func (db *DB) GetBlacklistedUsers() ([]User, error) {
	db.log.Debug("Get blacklisted users")
	var resp []User
	rows, err := db.qLog.Queryx("SELECT " + userQueryColumns + " FROM users WHERE is_in_blacklist")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var user User
		sqlx.StructScan(rows, &user)
		resp = append(resp, user)
	}
	return resp, rows.Err()
}
