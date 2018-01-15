package postgres

import (
	"time"

	"context"

	. "git.containerum.net/ch/user-manager/models"
	chutils "git.containerum.net/ch/user-manager/utils"
)

const tokenQueryColumnsWithUser = "tokens.token, tokens.created_at, tokens.is_active, tokens.session_id, " +
	"users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist"
const tokenQueryColumns = "token, created_at, is_active, session_id"

func (db *pgDB) GetTokenObject(ctx context.Context, token string) (*Token, error) {
	db.log.Infoln("Get token object", token)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+tokenQueryColumnsWithUser+" FROM tokens "+
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

func (db *pgDB) CreateToken(ctx context.Context, user *User, sessionID string) (*Token, error) {
	db.log.Infoln("Generate one-time token for", user.Login)
	ret := &Token{
		Token:     chutils.GenSalt(user.ID, user.Login),
		User:      user,
		IsActive:  true,
		SessionID: sessionID,
		CreatedAt: time.Now().UTC(),
	}
	_, err := db.eLog.ExecContext(ctx, "INSERT INTO tokens (token, user_id, is_active, session_id, created_at) "+
		"VALUES ($1, $2, $3, $4, $5)", ret.Token, ret.User.ID, ret.IsActive, ret.SessionID, ret.CreatedAt)
	return ret, err
}

func (db *pgDB) GetTokenBySessionID(ctx context.Context, sessionID string) (*Token, error) {
	db.log.Infoln("Get token by session id ", sessionID)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+tokenQueryColumnsWithUser+" FROM tokens "+
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

func (db *pgDB) DeleteToken(ctx context.Context, token string) error {
	db.log.Infoln("Remove token", token)
	_, err := db.eLog.ExecContext(ctx, "DELETE FROM tokens WHERE token = $1", token)
	return err
}

func (db *pgDB) UpdateToken(ctx context.Context, token *Token) error {
	db.log.Infoln("Update token", token.Token)
	_, err := db.eLog.ExecContext(ctx, "UPDATE tokens SET is_active = $2, session_id = $3 WHERE token = $1",
		token.Token, token.IsActive, token.SessionID)
	return err
}
