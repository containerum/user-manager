package postgres

import (
	"time"

	"fmt"

	"context"

	"git.containerum.net/ch/user-manager/models"
	chutils "git.containerum.net/ch/utils"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgresql database driver
	"github.com/mattes/migrate"
	migdrv "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file" // needed to load migrations scripts from files
	"github.com/sirupsen/logrus"
)

type pgDB struct {
	conn *sqlx.DB // do not use it in select/exec operations
	log  *logrus.Entry
	qLog sqlx.QueryerContext
	eLog sqlx.ExecerContext
}

// DBConnect initializes connection to postgresql database.
// github.com/jmoiron/sqlx used to to get work with database.
// Function tries to ping database and apply migrations using github.com/mattes/migrate.
// If migrations applying failed database goes to dirty state and requires manual conflict resolution.
func DBConnect(pgConnStr string) (models.DB, error) {
	log := logrus.WithField("component", "db")
	log.Infoln("Connecting to ", pgConnStr)
	conn, err := sqlx.Open("postgres", pgConnStr)
	if err != nil {
		log.WithError(err).Errorln("Postgres connection failed")
		return nil, err
	}
	if pingErr := conn.Ping(); pingErr != nil {
		return nil, err
	}

	ret := &pgDB{
		conn: conn,
		log:  log,
		qLog: chutils.NewSQLXContextQueryLogger(conn, log),
		eLog: chutils.NewSQLXContextExecLogger(conn, log),
	}

	m, err := ret.migrateUp("migrations")
	if err != nil {
		return nil, err
	}
	version, _, _ := m.Version()
	log.WithField("version", version).Infoln("Migrate up")

	return ret, nil
}

func (db *pgDB) migrateUp(path string) (*migrate.Migrate, error) {
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

func (db *pgDB) Transactional(ctx context.Context, f func(ctx context.Context, tx models.DB) error) (err error) {
	start := time.Now().Format(time.ANSIC)
	e := db.log.WithField("transaction_at", start)
	e.Debugln("Begin transaction")
	tx, txErr := db.conn.Beginx()
	if txErr != nil {
		e.WithError(txErr).Errorln("Begin transaction error")
		return models.ErrTransactionBegin
	}

	arg := &pgDB{
		conn: db.conn,
		log:  e,
		eLog: chutils.NewSQLXContextExecLogger(tx, e),
		qLog: chutils.NewSQLXContextQueryLogger(tx, e),
	}

	// needed for recovering panics in transactions.
	defer func(dberr error) {
		// if panic recovered, try to rollback transaction
		if panicErr := recover(); panicErr != nil {
			dberr = fmt.Errorf("panic in transaction: %v", panicErr)
		}

		if dberr != nil {
			e.WithError(dberr).Debugln("Rollback transaction")
			if rerr := tx.Rollback(); rerr != nil {
				e.WithError(rerr).Errorln("Rollback error")
				err = models.ErrTransactionRollback
			}
			err = dberr // forward error with panic description
			return
		}

		e.Debugln("Commit transaction")
		if cerr := tx.Commit(); cerr != nil {
			e.WithError(cerr).Errorln("Commit error")
			err = models.ErrTransactionCommit
		}
	}(f(ctx, arg))

	return
}

func (db *pgDB) Close() error {
	return db.conn.Close()
}
