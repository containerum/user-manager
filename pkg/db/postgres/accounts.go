package postgres

import (
	"errors"

	"context"

	"fmt"

	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
	"github.com/sirupsen/logrus"
)

func (pgdb *pgDB) GetUserByBoundAccount(ctx context.Context, service models.OAuthResource, accountID string) (*db.User, error) {
	pgdb.log.WithFields(logrus.Fields{
		"service":    service,
		"account_id": accountID,
	}).Infoln("Get bound account")

	switch service {
	case models.GitHubOAuth, models.FacebookOAuth, models.GoogleOAuth:
	default:
		return nil, errors.New("unrecognised service " + string(service))
	}

	var ret db.User

	rows, err := pgdb.qLog.QueryxContext(ctx, fmt.Sprintf(`SELECT users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist
	FROM accounts JOIN users ON accounts.user_id = users.id WHERE accounts.%v = '%v'`, service, accountID))

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.StructScan(&ret)
	return &ret, err
}

func (pgdb *pgDB) GetUserBoundAccounts(ctx context.Context, user *db.User) (*db.Accounts, error) {
	pgdb.log.Infoln("Get bound accounts for user", user.Login)
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT id, github, facebook, google FROM accounts WHERE user_id = $1", user.ID)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}

	ret := db.Accounts{User: user}
	err = rows.Scan(&ret.ID, &ret.Github, &ret.Facebook, &ret.Google)

	return &ret, err
}

func (pgdb *pgDB) BindAccount(ctx context.Context, user *db.User, service models.OAuthResource, accountID string) error {
	pgdb.log.Infof("Bind account %s (%s) for user %s", service, accountID, user.Login)
	switch service {
	case models.GitHubOAuth:
		_, err := pgdb.eLog.ExecContext(ctx, `INSERT INTO accounts (user_id, github, facebook, google) 
													VALUES ($1, $2, '', '')
													ON CONFLICT (user_id) DO UPDATE SET github = $2`, user.ID, accountID)
		return err
	case models.FacebookOAuth:
		_, err := pgdb.eLog.ExecContext(ctx, `INSERT INTO accounts (user_id, github, facebook, google) 
													VALUES ($1, '', $2, '')
													ON CONFLICT (user_id) DO UPDATE SET facebook = $2`, user.ID, accountID)
		return err
	case models.GoogleOAuth:
		_, err := pgdb.eLog.ExecContext(ctx, `INSERT INTO accounts (user_id, github, facebook, google) 
													VALUES ($1, '', '', $2)
													ON CONFLICT (user_id) DO UPDATE SET google = $2`, user.ID, accountID)
		return err
	default:
		return errors.New("unrecognised service " + string(service))
	}
	// see migrations/1515872648_accounts_constraint.up.sql
}

func (pgdb *pgDB) DeleteBoundAccount(ctx context.Context, user *db.User, service models.OAuthResource) error {
	pgdb.log.Infof("Deleting account %s for user %s", service, user.Login)
	switch service {
	case models.GitHubOAuth, models.FacebookOAuth, models.GoogleOAuth:
	default:
		return errors.New("unrecognised service " + string(service))
	}

	_, err := pgdb.eLog.ExecContext(ctx, fmt.Sprintf(`INSERT INTO accounts (user_id, github, facebook, google)
															VALUES ('%v', '', '', '')
															ON CONFLICT (user_id) DO UPDATE SET %v = ''`, user.ID, service))
	return err
}
