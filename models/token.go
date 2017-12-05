package models

import "time"

type Token struct {
	Token     string `gorm:"primary_key"`
	User      User
	CreatedAt time.Time
	IsActive  bool
	SessionID string `gorm:"type:uuid"`
}

func (db *DB) GetUserByToken(token string) (*User, error) {
	db.log.Debug("Get user by token", token)
	var user User
	resp := db.conn.Where(&Token{Token: token, IsActive: true}).First(&user)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &user, resp.Error
}

func (db *DB) CreateToken(user *User, sessionID string) (*Token, error) {
	db.log.Debug("Generate one-time token for", user.Login)
	ret := &Token{
		Token:     "token", // TODO: token generation here
		User:      *user,
		CreatedAt: time.Now().UTC(),
		IsActive:  true,
		SessionID: sessionID,
	}
	return ret, db.conn.Create(ret).Error
}
