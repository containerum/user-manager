package postgres

import (
	"time"

	"context"

	"git.containerum.net/ch/user-manager/pkg/db"
	chutils "git.containerum.net/ch/user-manager/pkg/utils"
)

const tokenQueryColumnsWithUser = "tokens.token, tokens.created_at, tokens.is_active, tokens.session_id, " +
	"users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist"
const tokenQueryColumns = "token, created_at, is_active, session_id"

func (pgdb *pgDB) GetTokenObject(ctx context.Context, token string) (*db.Token, error) {
	pgdb.log.Infoln("Get token object", token)
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT "+tokenQueryColumnsWithUser+" FROM tokens "+
		"JOIN users ON tokens.user_id = users.id WHERE tokens.token = $1 AND tokens.is_active", token)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	defer rows.Close()
	ret := db.Token{User: &db.User{}}
	err = rows.Scan(&ret.Token, &ret.CreatedAt, &ret.IsActive, &ret.SessionID,
		&ret.User.ID, &ret.User.Login, &ret.User.PasswordHash, &ret.User.Salt, &ret.User.Role,
		&ret.User.IsActive, &ret.User.IsDeleted, &ret.User.IsInBlacklist)
	return &ret, err
}

func (pgdb *pgDB) CreateToken(ctx context.Context, user *db.User, sessionID string) (*db.Token, error) {
	pgdb.log.Infoln("Generate one-time token for", user.Login)
	ret := &db.Token{
		Token:     chutils.GenSalt(user.ID, user.Login),
		User:      user,
		IsActive:  true,
		SessionID: sessionID,
		CreatedAt: time.Now().UTC(),
	}
	_, err := pgdb.eLog.ExecContext(ctx, "INSERT INTO tokens (token, user_id, is_active, session_id, created_at) "+
		"VALUES ($1, $2, $3, $4, $5)", ret.Token, ret.User.ID, ret.IsActive, ret.SessionID, ret.CreatedAt)
	return ret, err
}

func (pgdb *pgDB) GetTokenBySessionID(ctx context.Context, sessionID string) (*db.Token, error) {
	pgdb.log.Infoln("Get token by session id ", sessionID)
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT "+tokenQueryColumnsWithUser+" FROM tokens "+
		"JOIN users ON tokens.user_id = users.id WHERE tokens.session_id = $1 and tokens.is_active", sessionID)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	defer rows.Close()
	ret := db.Token{User: &db.User{}}
	err = rows.Scan(&ret.Token, &ret.CreatedAt, &ret.IsActive, &ret.SessionID,
		&ret.User.ID, &ret.User.Login, &ret.User.PasswordHash, &ret.User.Salt, &ret.User.Role,
		&ret.User.IsActive, &ret.User.IsDeleted, &ret.User.IsInBlacklist)

	return &ret, err
}

func (pgdb *pgDB) DeleteToken(ctx context.Context, token string) error {
	pgdb.log.Infoln("Remove token", token)
	_, err := pgdb.eLog.ExecContext(ctx, "DELETE FROM tokens WHERE token = $1", token)
	return err
}

func (pgdb *pgDB) UpdateToken(ctx context.Context, token *db.Token) error {
	pgdb.log.Infoln("Update token", token.Token)
	_, err := pgdb.eLog.ExecContext(ctx, "UPDATE tokens SET is_active = $2, session_id = $3 WHERE token = $1",
		token.Token, token.IsActive, token.SessionID)
	return err
}
