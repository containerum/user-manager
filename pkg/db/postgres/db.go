package postgres

import (
	"time"

	"fmt"

	"context"

	"git.containerum.net/ch/user-manager/pkg/db"
	"github.com/containerum/utils/sqlxutil"
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
func DBConnect(pgConnStr string, migrationsPath string) (db.DB, error) {
	log := logrus.WithField("component", "db")
	log.Infoln("Connecting to ", pgConnStr)
	conn, err := sqlx.Open("postgres", pgConnStr)
	if err != nil {
		log.WithError(err).Errorln("Postgres connection failed")
		return nil, err
	}
	if pingErr := conn.Ping(); pingErr != nil {
		return nil, pingErr
	}

	ret := &pgDB{
		conn: conn,
		log:  log,
		qLog: sqlxutil.NewSQLXContextQueryLogger(conn, log),
		eLog: sqlxutil.NewSQLXContextExecLogger(conn, log),
	}

	m, err := ret.migrateUp(migrationsPath)
	if err != nil {
		return nil, err
	}
	version, _, _ := m.Version()
	log.WithField("version", version).Infoln("Migrate up")

	return ret, nil
}

func (pgdb *pgDB) migrateUp(path string) (*migrate.Migrate, error) {
	pgdb.log.Infof("Running migrations")
	instance, err := migdrv.WithInstance(pgdb.conn.DB, &migdrv.Config{})
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

func (pgdb *pgDB) Transactional(ctx context.Context, f func(ctx context.Context, tx db.DB) error) (err error) {
	start := time.Now().Format(time.ANSIC)
	e := pgdb.log.WithField("transaction_at", start)
	e.Debugln("Begin transaction")
	tx, txErr := pgdb.conn.Beginx()
	if txErr != nil {
		e.WithError(txErr).Errorln("Begin transaction error")
		return db.ErrTransactionBegin
	}

	arg := &pgDB{
		conn: pgdb.conn,
		log:  e,
		eLog: sqlxutil.NewSQLXContextExecLogger(tx, e),
		qLog: sqlxutil.NewSQLXContextQueryLogger(tx, e),
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
				err = db.ErrTransactionRollback
			}
			err = dberr // forward error with panic description
			return
		}

		e.Debugln("Commit transaction")
		if cerr := tx.Commit(); cerr != nil {
			e.WithError(cerr).Errorln("Commit error")
			err = db.ErrTransactionCommit
		}
	}(f(ctx, arg))

	return
}

func (pgdb *pgDB) Close() error {
	return pgdb.conn.Close()
}
