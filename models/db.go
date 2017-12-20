package models

import (
	"errors"
	"time"

	chutils "git.containerum.net/ch/utils"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	migdrv "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
	"github.com/sirupsen/logrus"
)

type DB struct {
	conn *sqlx.DB // do not use it in select/exec operations
	log  *logrus.Entry
	qLog *chutils.SQLXQueryLogger
	eLog *chutils.SQLXExecLogger
}

var (
	ErrTransactionBegin    = errors.New("transaction begin error")
	ErrTransactionRollback = errors.New("transaction rollback error")
	ErrTransactionCommit   = errors.New("transaction commit error")
)

func DBConnect(pgConnStr string) (*DB, error) {
	log := logrus.WithField("component", "db")
	log.Info("Connecting to ", pgConnStr)
	conn, err := sqlx.Open("postgres", pgConnStr)
	if err != nil {
		log.WithError(err).Error("Postgres connection failed")
		return nil, err
	}

	ret := &DB{
		conn: conn,
		log:  log,
		qLog: chutils.NewSQLXQueryLogger(conn, log),
		eLog: chutils.NewSQLXExecLogger(conn, log),
	}

	m, err := ret.migrateUp("migrations")
	if err != nil {
		return nil, err
	}
	version, _, _ := m.Version()
	log.WithField("version", version).Info("Migrate up")

	return ret, nil
}

func (db *DB) migrateUp(path string) (*migrate.Migrate, error) {
	db.log.Infof("Running migrations")
	instance, err := migdrv.WithInstance(db.conn.DB, &migdrv.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+path, "clickhouse", instance)
	if err != nil {
		return nil, err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}
	return m, nil
}

func (db *DB) Transactional(f func(tx *DB) error) error {
	start := time.Now().Format(time.ANSIC)
	e := db.log.WithField("transaction_at", start)
	e.Debug("Begin transaction")
	tx, err := db.conn.Beginx()
	if err != nil {
		e.WithError(err).Error("Begin transaction error")
		return ErrTransactionBegin
	}

	arg := &DB{
		conn: db.conn,
		log:  e,
		eLog: chutils.NewSQLXExecLogger(tx, e),
		qLog: chutils.NewSQLXQueryLogger(tx, e),
	}
	if err := f(arg); err != nil {
		e.WithError(err).Debug("Rollback transaction")
		if rerr := tx.Rollback(); rerr != nil {
			e.WithError(rerr).Error("Rollback error")
			return ErrTransactionRollback
		}
		return err
	}

	e.Debug("Commit transaction")
	if cerr := tx.Commit(); cerr != nil {
		e.WithError(cerr).Error("Commit error")
		return ErrTransactionCommit
	}
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
