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

func (db *DB) GetUserByBoundAccount(service, accountID string) (*User, error) {
	db.log.WithFields(logrus.Fields{
		"service":    service,
		"account_id": accountID,
	}).Debug("Get bound account")

	var accounts Accounts
	var accountFieldIndex = -1

	accountsType := reflect.TypeOf(accounts)
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
		return nil, fmt.Errorf("invalid service specified")
	}

	reflect.ValueOf(&accounts).Field(accountFieldIndex).SetString(accountID) // to make query

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
