package utils

import (
	"database/sql"

	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type sqlxQueryLogger struct {
	q sqlx.Queryer
	l *logrus.Entry
}

func NewSQLXQueryLogger(queryer sqlx.Queryer, entry *logrus.Entry) sqlx.Queryer {
	return &sqlxQueryLogger{
		q: queryer,
		l: entry,
	}
}

func (q *sqlxQueryLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := q.q.Query(query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxQueryLogger) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	rows, err := q.q.Queryx(query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxQueryLogger) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	row := q.q.QueryRowx(query, args...)
	q.l.WithField("query", query).Debugln(args...)
	return row
}

type sqlxExecLogger struct {
	e sqlx.Execer
	l *logrus.Entry
}

func NewSQLXExecLogger(execer sqlx.Execer, entry *logrus.Entry) sqlx.Execer {
	return &sqlxExecLogger{
		e: execer,
		l: entry,
	}
}

func (e *sqlxExecLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := e.e.Exec(query, args...)
	e.l.WithField("query", query).WithError(err).Debugln(args...)
	return result, err
}

type sqlxContextQueryLogger struct {
	q sqlx.QueryerContext
	l *logrus.Entry
}

func NewSQLXContextQueryLogger(queryer sqlx.QueryerContext, entry *logrus.Entry) sqlx.QueryerContext {
	return &sqlxContextQueryLogger{
		q: queryer,
		l: entry,
	}
}

func (q *sqlxContextQueryLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := q.q.QueryContext(ctx, query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxContextQueryLogger) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	rows, err := q.q.QueryxContext(ctx, query, args...)
	q.l.WithField("query", query).WithError(err).Debugln(args...)
	return rows, err
}

func (q *sqlxContextQueryLogger) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	rows := q.q.QueryRowxContext(ctx, query, args...)
	q.l.WithField("query", query).Debugln(args...)
	return rows
}

type sqlxContextExecLogger struct {
	e sqlx.ExecerContext
	l *logrus.Entry
}

func NewSQLXContextExecLogger(execer sqlx.ExecerContext, entry *logrus.Entry) sqlx.ExecerContext {
	return &sqlxContextExecLogger{
		e: execer,
		l: entry,
	}
}

func (e *sqlxContextExecLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := e.e.ExecContext(ctx, query, args...)
	e.l.WithField("query", query).WithError(err).Debugln(args...)
	return result, err
}
