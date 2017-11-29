package models

import (
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

func DBConnect(pgConnStr string) (*DB, error) {
	log := logrus.WithField("component", "db").Logger
	log.Info("Connecting to", pgConnStr)
	conn, err := gorm.Open("postgres", pgConnStr)
	if err != nil {
		log.WithError(err).Error("Postgres connection failed")
		return nil, err
	}
	log.Info("Run migrations")
	if err := conn.AutoMigrate(migrateModels...).Error; err != nil {
		log.WithError(err).Error("Run migrations failed")
		return nil, err
	}
	return &DB{
		conn: conn,
		log:  log,
	}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
