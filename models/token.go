package models

import "time"

type Token struct {
	Token     string `gorm:"primary_key"`
	User      User
	CreatedAt time.Time
	IsActive  bool
	SessionID string `gorm:"type:uuid"`
}

func (db *DB) GetTokenObject(token string) (*Token, error) {
	db.log.Debug("Get token object", token)
	var ret Token
	resp := db.conn.Where(&Token{Token: token, IsActive: true}).First(&ret)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &ret, resp.Error
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

func (db *DB) GetTokenBySessionID(sessionID string) (*Token, error) {
	db.log.Debug("Get token by session id ", sessionID)
	var token Token
	resp := db.conn.Where(Token{SessionID: sessionID}).First(&token)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &token, resp.Error
}

func (db *DB) DeleteToken(token string) error {
	db.log.Debug("Remove token", token)
	return db.conn.Where(Token{Token: token}).Delete(&Token{}).Error
}

func (db *DB) UpdateToken(token *Token) error {
	db.log.Debug("Update token", token.Token)
	return db.conn.Save(token).Error
}
