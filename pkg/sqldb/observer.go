package sqldb

import (
	"context"
	"database/sql/driver"
	"time"

	"sniper/pkg/log"

	"github.com/ngrok/sqlmw"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// 观察所有 sql 执行情况
type observer struct {
	sqlmw.NullInterceptor
	name string
}

func (o observer) ConnExecContext(ctx context.Context,
	conn driver.ExecerContext,
	query string, args []driver.NamedValue) (driver.Result, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Exec")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)
	ext.DBStatement.Set(span, query)

	s := time.Now()
	result, err := conn.ExecContext(ctx, query, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, exec: %s, args: %v, cost: %v",
		o.name, query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		o.name,
		table,
		cmd,
	).Observe(d.Seconds())

	return result, err
}

func (o observer) ConnQueryContext(ctx context.Context,
	conn driver.QueryerContext,
	query string, args []driver.NamedValue) (driver.Rows, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Query")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)
	ext.DBStatement.Set(span, query)

	s := time.Now()
	rows, err := conn.QueryContext(ctx, query, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, query: %s, args: %v, cost: %v",
		o.name, query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		o.name,
		table,
		cmd,
	).Observe(d.Seconds())

	return rows, err
}

func (o observer) ConnPrepareContext(ctx context.Context,
	conn driver.ConnPrepareContext,
	query string) (driver.Stmt, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Prepare")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)
	ext.DBStatement.Set(span, query)

	s := time.Now()
	stmt, err := conn.PrepareContext(ctx, query)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, prepare: %s, args: %v, cost: %v",
		o.name, query, nil, d)

	table, _ := parseSQL(query)
	sqlDurations.WithLabelValues(
		o.name,
		table,
		"prepare",
	).Observe(d.Seconds())

	return stmt, err
}

func (o observer) StmtExecContext(ctx context.Context,
	stmt driver.StmtExecContext,
	query string, args []driver.NamedValue) (driver.Result, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "PreparedExec")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)
	ext.DBStatement.Set(span, query)

	s := time.Now()
	result, err := stmt.ExecContext(ctx, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, prepared exec: %s, args: %v, cost: %v",
		o.name, query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		o.name,
		table,
		cmd+"-prepared",
	).Observe(d.Seconds())

	return result, err
}

func (o observer) StmtQueryContext(ctx context.Context,
	stmt driver.StmtQueryContext,
	query string, args []driver.NamedValue) (driver.Rows, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "PreparedQuery")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)
	ext.DBStatement.Set(span, query)

	s := time.Now()
	rows, err := stmt.QueryContext(ctx, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, prepared query: %s, args: %v, cost: %v",
		o.name, query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		o.name,
		table,
		cmd+"-prepared",
	).Observe(d.Seconds())

	return rows, err
}

func (o observer) ConnBeginTx(ctx context.Context, conn driver.ConnBeginTx,
	txOpts driver.TxOptions) (driver.Tx, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Begin")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)

	s := time.Now()
	tx, err := conn.BeginTx(ctx, txOpts)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, begin, cost: %v", o.name, d)

	sqlDurations.WithLabelValues(
		o.name,
		"",
		"begin",
	).Observe(d.Seconds())

	return tx, err
}

func (o observer) TxCommit(ctx context.Context, tx driver.Tx) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Commit")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)

	s := time.Now()
	err := tx.Commit()
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, commit, cost: %v", o.name, d)

	sqlDurations.WithLabelValues(
		o.name,
		"",
		"commit",
	).Observe(d.Seconds())

	return err
}

func (o observer) TxRollback(ctx context.Context, tx driver.Tx) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Rollback")
	defer span.Finish()

	ext.Component.Set(span, "sqldb")
	ext.DBInstance.Set(span, o.name)

	s := time.Now()
	err := tx.Rollback()
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] name:%s, rollback, cost: %v", o.name, d)

	sqlDurations.WithLabelValues(
		o.name,
		"",
		"rollback",
	).Observe(d.Seconds())

	return err
}
