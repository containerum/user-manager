package postgres

import (
	"time"

	. "git.containerum.net/ch/user-manager/models"
	chutils "git.containerum.net/ch/user-manager/utils"
)

const tokenQueryColumnsWithUser = "tokens.token, tokens.created_at, tokens.is_active, tokens.session_id, " +
	"users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist"
const tokenQueryColumns = "token, created_at, is_active, session_id"

func (db *PgDB) GetTokenObject(token string) (*Token, error) {
	db.log.Infoln("Get token object", token)
	rows, err := db.qLog.Queryx("SELECT "+tokenQueryColumnsWithUser+" FROM tokens "+
		"JOIN users ON tokens.user_id = users.id WHERE tokens.token = $1 AND tokens.is_active", token)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	defer rows.Close()
	ret := Token{User: &User{}}
	err = rows.Scan(&ret.Token, &ret.CreatedAt, &ret.IsActive, &ret.SessionID,
		&ret.User.ID, &ret.User.Login, &ret.User.PasswordHash, &ret.User.Salt, &ret.User.Role,
		&ret.User.IsActive, &ret.User.IsDeleted, &ret.User.IsInBlacklist)
	return &ret, err
}

func (db *PgDB) CreateToken(user *User, sessionID string) (*Token, error) {
	db.log.Infoln("Generate one-time token for", user.Login)
	ret := &Token{
		Token:     chutils.GenSalt(user.ID, user.Login),
		User:      user,
		IsActive:  true,
		SessionID: sessionID,
		CreatedAt: time.Now().UTC(),
	}
	_, err := db.eLog.Exec("INSERT INTO tokens (token, user_id, is_active, session_id, created_at) "+
		"VALUES ($1, $2, $3, $4, $5)", ret.Token, ret.User.ID, ret.IsActive, ret.SessionID, ret.CreatedAt)
	return ret, err
}

func (db *PgDB) GetTokenBySessionID(sessionID string) (*Token, error) {
	db.log.Infoln("Get token by session id ", sessionID)
	rows, err := db.qLog.Queryx("SELECT "+tokenQueryColumnsWithUser+" FROM tokens "+
		"JOIN users ON tokens.user_id = users.id WHERE tokens.session_id = $1 and tokens.is_active", sessionID)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	defer rows.Close()
	ret := Token{User: &User{}}
	err = rows.Scan(&ret.Token, &ret.CreatedAt, &ret.IsActive, &ret.SessionID,
		&ret.User.ID, &ret.User.Login, &ret.User.PasswordHash, &ret.User.Salt, &ret.User.Role,
		&ret.User.IsActive, &ret.User.IsDeleted, &ret.User.IsInBlacklist)

	return &ret, err
}

func (db *PgDB) DeleteToken(token string) error {
	db.log.Infoln("Remove token", token)
	_, err := db.eLog.Exec("DELETE FROM tokens WHERE token = $1", token)
	return err
}

func (db *PgDB) UpdateToken(token *Token) error {
	db.log.Infoln("Update token", token.Token)
	_, err := db.eLog.Exec("UPDATE tokens SET is_active = $2, session_id = $3 WHERE token = $1",
		token.Token, token.IsActive, token.SessionID)
	return err
}
