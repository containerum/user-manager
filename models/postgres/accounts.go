package postgres

import (
	. "git.containerum.net/ch/user-manager/models"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func (db *pgDB) GetUserByBoundAccount(service, accountID string) (*User, error) {
	db.log.WithFields(logrus.Fields{
		"service":    service,
		"account_id": accountID,
	}).Infoln("Get bound account")

	var ret User
	err := sqlx.Get(db.qLog, &ret, "SELECT users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist "+
		"FROM accounts JOIN users ON accounts.user_id = users.id WHERE accounts.$1 = $2", service, accountID)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (db *pgDB) GetUserBoundAccounts(user *User) (*Accounts, error) {
	db.log.Infoln("Get bound accounts for user", user.Login)
	rows, err := db.qLog.Queryx("SELECT id, github, facebook, google FROM accounts WHERE user_id = $1", user.ID)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}

	ret := Accounts{User: user}
	err = rows.Scan(&ret.ID, &ret.Github, &ret.Facebook, &ret.Google)

	return &ret, err
}

func (db *pgDB) BindAccount(user *User, service, accountID string) error {
	db.log.Infof("Bind account %s (%s) for user %s", service, accountID, user.Login)
	_, err := db.eLog.Exec("INSERT INTO accounts (user_id, $2) VALUES ($1, $3) ON CONFLICT (user_id) DO UPDATE SET $2 = $3",
		user.ID, service, accountID)
	return err
}
