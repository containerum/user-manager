package models

import (
	"time"

	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type DB struct {
	conn *gorm.DB
	log  *logrus.Logger
}

// models to automatically migrate at connection
var migrateModels = []interface{}{
	&Accounts{},
	&Link{},
	&Profile{},
	&Token{},
	&User{},
}

var (
	ErrTransactionRollback = errors.New("transaction rollback error")
	ErrTransactionCommit   = errors.New("transaction commit error")
)

func DBConnect(pgConnStr string) (*DB, error) {
	log := logrus.WithField("component", "db").Logger
	log.Info("Connecting to", pgConnStr)
	conn, err := gorm.Open("postgres", pgConnStr)
	if err != nil {
		log.WithError(err).Error("Postgres connection failed")
		return nil, err
	}
	log.Info("Run migrations")
	if err := conn.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.WithError(err).Error("UUID extension create failed")
		return nil, err
	}
	if err := conn.AutoMigrate(migrateModels...).Error; err != nil {
		log.WithError(err).Error("Run migrations failed")
		return nil, err
	}
	return &DB{
		conn: conn,
		log:  log,
	}, nil
}

func (db *DB) Transactional(f func(tx *DB) error) error {
	start := time.Now().Format(time.ANSIC)
	e := db.log.WithField("transaction_at", start)
	e.Debug("Begin transaction")
	tx := db.conn.Begin()
	if err := f(&DB{
		conn: tx,
		log:  e.Logger,
	}); err != nil {
		e.WithError(err).Debug("Rollback transaction")
		if rerr := tx.Rollback().Error; rerr != nil {
			e.WithError(rerr).Error("Rollback error")
			return ErrTransactionRollback
		}
		return err
	}
	e.Debug("Commit transaction")
	if cerr := tx.Commit().Error; cerr != nil {
		e.WithError(cerr).Error("Commit error")
		return ErrTransactionCommit
	}
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
