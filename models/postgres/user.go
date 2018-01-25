package postgres

import (
	"context"

	"git.containerum.net/ch/user-manager/models"
)

const userQueryColumns = "id, login, password_hash, salt, role, is_active, is_deleted, is_in_blacklist"

func (db *pgDB) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	db.log.Infoln("Get user by login", login)
	var user models.User
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE login = $1 AND NOT is_deleted", login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.StructScan(&user)
	return &user, err
}

func (db *pgDB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	db.log.Infoln("Get user by id", id)
	var user models.User
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE id = $1 AND NOT is_deleted", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.StructScan(&user)
	return &user, err
}

func (db *pgDB) GetDeletedUserByID(ctx context.Context, id string) (*models.User, error) {
	db.log.Infoln("Get user by id", id)
	var user models.User
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE id = $1 AND is_deleted", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.StructScan(&user)
	return &user, err
}

func (db *pgDB) CreateUser(ctx context.Context, user *models.User) error {
	db.log.Infoln("Create user", user.Login)
	rows, err := db.qLog.QueryxContext(ctx, "INSERT INTO users (login, password_hash, salt, role, is_active) "+
		"VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Login, user.PasswordHash, user.Salt, user.Role, user.IsActive)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return rows.Err()
	}
	err = rows.Scan(&user.ID)
	return err
}

func (db *pgDB) CreateUserWebAPI(ctx context.Context, user *models.User) error {
	db.log.Infoln("Create user", user.Login)
	rows, err := db.qLog.QueryxContext(ctx, "INSERT INTO users (login, password_hash, salt, role, is_active, id) "+
		"VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		user.Login, user.PasswordHash, user.Salt, user.Role, user.IsActive, user.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return rows.Err()
	}
	err = rows.Scan(&user.ID)
	return err
}

func (db *pgDB) UpdateUser(ctx context.Context, user *models.User) error {
	db.log.Infoln("Update user", user.Login)
	_, err := db.eLog.ExecContext(ctx, "UPDATE users SET "+
		"login = $2, password_hash = $3, salt = $4, role = $5, is_active = $6, is_deleted = $7 WHERE id = $1",
		user.ID, user.Login, user.PasswordHash, user.Salt, user.Role, user.IsActive, user.IsDeleted)
	return err
}

func (db *pgDB) GetBlacklistedUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	db.log.Infoln("Get blacklisted users")
	resp := make([]models.User, 0)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE is_in_blacklist LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.StructScan(&user)
		if err != nil {
			return nil, err
		}
		resp = append(resp, user)
	}
	return resp, rows.Err()
}

func (db *pgDB) BlacklistUser(ctx context.Context, user *models.User) error {
	db.log.Infoln("Blacklisting user", user.Login)
	_, err := db.eLog.ExecContext(ctx, "UPDATE users SET is_in_blacklist = TRUE WHERE id = $1", user.ID)
	if err != nil {
		return err
	}
	_, err = db.eLog.ExecContext(ctx, "UPDATE profiles SET blacklist_at = NOW() WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}
	user.IsInBlacklist = true
	return nil
}
