package postgres

import (
	"context"

	"git.containerum.net/ch/user-manager/pkg/db"
	"github.com/jmoiron/sqlx"
)

const userQueryColumns = "id, login, password_hash, salt, role, is_active, is_deleted, is_in_blacklist"

func (pgdb *pgDB) GetUserByLogin(ctx context.Context, login string) (*db.User, error) {
	pgdb.log.Infoln("Get user by login", login)
	var user db.User
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE login = $1 AND NOT is_deleted", login)
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

func (pgdb *pgDB) GetAnyUserByLogin(ctx context.Context, login string) (*db.User, error) {
	pgdb.log.Infoln("Get user by login", login)
	var user db.User

	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE login = $1", login)
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

func (pgdb *pgDB) GetAnyUserByLoginWOContext(login string) (*db.User, error) {
	pgdb.log.Infoln("Get user by login", login)
	var user db.User

	rows, err := pgdb.conn.DB.Query("SELECT id, salt FROM users WHERE login = $1", login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.Scan(&user.ID, &user.Salt)
	return &user, err
}

func (pgdb *pgDB) GetUserByID(ctx context.Context, id string) (*db.User, error) {
	pgdb.log.Infoln("Get user by id", id)
	var user db.User
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE id = $1 AND NOT is_deleted", id)
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

func (pgdb *pgDB) GetAnyUserByID(ctx context.Context, id string) (*db.User, error) {
	pgdb.log.Infoln("Get user by id", id)
	var user db.User
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE id = $1", id)
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

func (pgdb *pgDB) CreateUser(ctx context.Context, user *db.User) error {
	pgdb.log.Infoln("Create user", user.Login)
	rows, err := pgdb.qLog.QueryxContext(ctx, "INSERT INTO users (login, password_hash, salt, role, is_active) "+
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

func (pgdb *pgDB) CreateUserWOContext(user *db.User) error {
	pgdb.log.Infoln("Create user", user.Login)
	rows, err := pgdb.conn.DB.Query("INSERT INTO users (login, password_hash, salt, role, is_active) "+
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

func (pgdb *pgDB) UpdateUser(ctx context.Context, user *db.User) error {
	pgdb.log.Infoln("Update user", user.Login)
	_, err := pgdb.eLog.ExecContext(ctx, "UPDATE users SET "+
		"login = $2, password_hash = $3, salt = $4, role = $5, is_active = $6, is_deleted = $7 WHERE id = $1",
		user.ID, user.Login, user.PasswordHash, user.Salt, user.Role, user.IsActive, user.IsDeleted)
	return err
}

func (pgdb *pgDB) UpdateUserWOContext(user *db.User) error {
	pgdb.log.Infoln("Update user", user.Login)
	_, err := pgdb.conn.DB.Exec("UPDATE users SET "+
		"password_hash = $2 WHERE id = $1",
		user.ID, user.PasswordHash)
	return err
}

func (pgdb *pgDB) GetBlacklistedUsers(ctx context.Context, limit, offset int) ([]db.User, error) {
	pgdb.log.Infoln("Get blacklisted users")
	resp := make([]db.User, 0)
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT "+userQueryColumns+" FROM users WHERE is_in_blacklist ORDER BY users.login LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user db.User
		err := rows.StructScan(&user)
		if err != nil {
			return nil, err
		}
		resp = append(resp, user)
	}
	return resp, rows.Err()
}

func (pgdb *pgDB) BlacklistUser(ctx context.Context, user *db.User) error {
	pgdb.log.Infoln("Blacklisting user", user.Login)
	_, err := pgdb.eLog.ExecContext(ctx, "UPDATE users SET is_in_blacklist = TRUE WHERE id = $1", user.ID)
	if err != nil {
		return err
	}
	_, err = pgdb.eLog.ExecContext(ctx, "UPDATE profiles SET blacklist_at = NOW() WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}
	user.IsInBlacklist = true
	return nil
}

func (pgdb *pgDB) UnBlacklistUser(ctx context.Context, user *db.User) error {
	pgdb.log.Infoln("Unblacklisting user", user.Login)
	_, err := pgdb.eLog.ExecContext(ctx, "UPDATE users SET is_in_blacklist = FALSE WHERE id = $1", user.ID)
	if err != nil {
		return err
	}
	_, err = pgdb.eLog.ExecContext(ctx, "UPDATE profiles SET blacklist_at = NULL WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}
	user.IsInBlacklist = false
	return nil
}

func (pgdb *pgDB) GetUsersLoginID(ctx context.Context, ids []string) ([]db.User, error) {
	pgdb.log.Infoln("Get all users logins")
	users := make([]db.User, 0) // return empty slice instead of nil if no records found
	query, args, err := sqlx.In("SELECT id, login FROM users WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	rows, err := pgdb.qLog.QueryxContext(ctx, pgdb.conn.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		user := db.User{}
		err := rows.Scan(
			&user.ID, &user.Login,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (pgdb *pgDB) CountAdmins(ctx context.Context) (*int, error) {
	pgdb.log.Infoln("Counting admins")
	var count int
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT count(id) FROM users WHERE role='admin' AND is_active AND NOT is_deleted")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.Scan(&count)
	return &count, err
}
