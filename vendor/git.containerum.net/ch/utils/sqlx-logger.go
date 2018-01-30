package utils

import (
	"database/sql"

	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// sqlx query logger

type sqlxQueryLogger struct {
	sqlx.Queryer
	l *logrus.Entry
}

func NewSQLXQueryLogger(queryer sqlx.Queryer, entry *logrus.Entry) sqlx.Queryer {
	return &sqlxQueryLogger{
		Queryer: queryer,
		l:       entry,
	}
}

func (q *sqlxQueryLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := q.Queryer.Query(query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxQueryLogger) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	rows, err := q.Queryer.Queryx(query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxQueryLogger) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	row := q.Queryer.QueryRowx(query, args...)
	q.l.WithField("query", query).Debugln(args...)
	return row
}

// sqlx exec logger

type sqlxExecLogger struct {
	sqlx.Execer
	l *logrus.Entry
}

func NewSQLXExecLogger(execer sqlx.Execer, entry *logrus.Entry) sqlx.Execer {
	return &sqlxExecLogger{
		Execer: execer,
		l:      entry,
	}
}

func (e *sqlxExecLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := e.Execer.Exec(query, args...)
	e.l.WithField("query", query).WithError(err).Debugln(args...)
	return result, err
}

// sqlx context query logger

type sqlxContextQueryLogger struct {
	sqlx.QueryerContext
	l *logrus.Entry
}

func NewSQLXContextQueryLogger(queryer sqlx.QueryerContext, entry *logrus.Entry) sqlx.QueryerContext {
	return &sqlxContextQueryLogger{
		QueryerContext: queryer,
		l:              entry,
	}
}

func (q *sqlxContextQueryLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := q.QueryerContext.QueryContext(ctx, query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxContextQueryLogger) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	rows, err := q.QueryerContext.QueryxContext(ctx, query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxContextQueryLogger) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	rows := q.QueryerContext.QueryRowxContext(ctx, query, args...)
	q.l.WithField("query", query).Debugln(args...)
	return rows
}

// sqlx context exec logger

type sqlxContextExecLogger struct {
	sqlx.ExecerContext
	l *logrus.Entry
}

func NewSQLXContextExecLogger(execer sqlx.ExecerContext, entry *logrus.Entry) sqlx.ExecerContext {
	return &sqlxContextExecLogger{
		ExecerContext: execer,
		l:             entry,
	}
}

func (e *sqlxContextExecLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := e.ExecerContext.ExecContext(ctx, query, args...)
	e.l.WithField("query", query).WithError(err).Debugln(args...)
	return result, err
}

// sqlx ext logger

type sqlxExtLogger struct {
	sqlx.Ext

	ql sqlxQueryLogger
	el sqlxExecLogger
}

func NewSQLXExtLogger(ext sqlx.Ext, entry *logrus.Entry) sqlx.Ext {
	return &sqlxExtLogger{
		ql: sqlxQueryLogger{Queryer: ext, l: entry},
		el: sqlxExecLogger{Execer: ext, l: entry},
	}
}

func (e *sqlxExtLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return e.ql.Query(query, args...)
}

func (e *sqlxExtLogger) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return e.ql.Queryx(query, args...)
}

func (e *sqlxExtLogger) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return e.ql.QueryRowx(query, args...)
}

func (e *sqlxExtLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
	return e.el.Exec(query, args...)
}

// sqlx ext context logger

type sqlxExtContextLogger struct {
	sqlx.ExtContext

	ql sqlxContextQueryLogger
	el sqlxContextExecLogger
}

func NewSQLXExtContextLogger(ext sqlx.ExtContext, entry *logrus.Entry) sqlx.ExtContext {
	return &sqlxExtContextLogger{
		ql: sqlxContextQueryLogger{QueryerContext: ext, l: entry},
		el: sqlxContextExecLogger{ExecerContext: ext, l: entry},
	}
}

func (e *sqlxExtContextLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return e.ql.QueryContext(ctx, query, args...)
}

func (e *sqlxExtContextLogger) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return e.ql.QueryxContext(ctx, query, args...)
}

func (e *sqlxExtContextLogger) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return e.ql.QueryRowxContext(ctx, query, args...)
}

func (e *sqlxExtContextLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return e.el.ExecContext(ctx, query, args...)
}
