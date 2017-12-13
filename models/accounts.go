package models

import (
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
)

type Accounts struct {
	ID       string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	User     User
	UserID   string `gorm:"type:uuid;ForeignKey:UserID"`
	Github   string `account:"github"`
	Facebook string `account:"facebook"`
	Google   string `account:"google"`
}

func (a *Accounts) setAccountField(service, accountID string) error {
	var accountFieldIndex = -1

	accountsType := reflect.TypeOf(a)
	for i := 0; i < accountsType.NumField(); i++ {
		curField := accountsType.Field(i)
		accountTag, ok := curField.Tag.Lookup("account")
		if !ok {
			continue
		}
		if accountTag == service {
			accountFieldIndex = i
		}
	}
	if accountFieldIndex == -1 {
		return fmt.Errorf("invalid service specified")
	}

	reflect.ValueOf(a).Field(accountFieldIndex).SetString(accountID) // to make query

	return nil
}

func (db *DB) GetUserByBoundAccount(service, accountID string) (*User, error) {
	db.log.WithFields(logrus.Fields{
		"service":    service,
		"account_id": accountID,
	}).Debug("Get bound account")

	var accounts Accounts
	if err := accounts.setAccountField(service, accountID); err != nil {
		return nil, err
	}

	resp := db.conn.Where(accounts).First(&accounts)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &accounts.User, resp.Error
}

func (db *DB) GetUserBoundAccounts(user *User) (*Accounts, error) {
	db.log.Debug("Get bound accounts for user", user.Login)
	var accounts Accounts
	resp := db.conn.Model(user).Related(&accounts)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &accounts, resp.Error
}

func (db *DB) BindAccount(user *User, service, accountID string) error {
	db.log.Debugf("Bind account %s (%s) for user %s", service, accountID, user.Login)

	var accounts Accounts

	resp := db.conn.Model(user).Related(&accounts)
	if resp.RecordNotFound() {
		if err := db.Transactional(func(tx *DB) error {
			return tx.conn.Create(&accounts).Error
		}); err != nil {
			return err
		}
	}

	if err := accounts.setAccountField(service, accountID); err != nil {
		return err
	}

	return db.Transactional(func(tx *DB) error {
		return tx.conn.Save(&accounts).Error
	})
}
