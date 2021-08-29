package sqldb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"

	"sniper/pkg/log"

	"github.com/go-sql-driver/mysql"
	"github.com/ngrok/sqlmw"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"modernc.org/sqlite"
)

func init() {
	sql.Register("db-sqlite", sqlmw.Driver(&sqlite.Driver{}, observer{}))
	sql.Register("db-mysql", sqlmw.Driver(mysql.MySQLDriver{}, observer{}))
}

// 观察所有 sql 执行情况
type observer struct {
	sqlmw.NullInterceptor
}

func (observer) ConnExecContext(ctx context.Context,
	conn driver.ExecerContext,
	query string, args []driver.NamedValue) (driver.Result, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Exec")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))
	span.SetTag(string(ext.DBStatement), query)

	s := time.Now()
	result, err := conn.ExecContext(ctx, query, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] exec: %s, args: %v, cost: %v",
		query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		name(ctx),
		table,
		cmd,
	).Observe(d.Seconds())

	return result, err
}

func (observer) ConnQueryContext(ctx context.Context,
	conn driver.QueryerContext,
	query string, args []driver.NamedValue) (driver.Rows, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Query")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))
	span.SetTag(string(ext.DBStatement), query)

	s := time.Now()
	rows, err := conn.QueryContext(ctx, query, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] query: %s, args: %v, cost: %v",
		query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		name(ctx),
		table,
		cmd,
	).Observe(d.Seconds())

	return rows, err
}

func (observer) ConnPrepareContext(ctx context.Context,
	conn driver.ConnPrepareContext,
	query string) (driver.Stmt, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Prepare")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))
	span.SetTag(string(ext.DBStatement), query)

	s := time.Now()
	stmt, err := conn.PrepareContext(ctx, query)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] prepare: %s, args: %v, cost: %v",
		query, nil, d)

	table, _ := parseSQL(query)
	sqlDurations.WithLabelValues(
		name(ctx),
		table,
		"prepare",
	).Observe(d.Seconds())

	return stmt, err
}

func (observer) StmtExecContext(ctx context.Context,
	stmt driver.StmtExecContext,
	query string, args []driver.NamedValue) (driver.Result, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "PreparedExec")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))
	span.SetTag(string(ext.DBStatement), query)

	s := time.Now()
	result, err := stmt.ExecContext(ctx, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] prepared exec: %s, args: %v, cost: %v",
		query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		name(ctx),
		table,
		cmd+"-prepared",
	).Observe(d.Seconds())

	return result, err
}

func (observer) StmtQueryContext(ctx context.Context,
	stmt driver.StmtQueryContext,
	query string, args []driver.NamedValue) (driver.Rows, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "PreparedQuery")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))
	span.SetTag(string(ext.DBStatement), query)

	s := time.Now()
	rows, err := stmt.QueryContext(ctx, args)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] prepared query: %s, args: %v, cost: %v",
		query, values(args), d)

	table, cmd := parseSQL(query)
	sqlDurations.WithLabelValues(
		name(ctx),
		table,
		cmd+"-prepared",
	).Observe(d.Seconds())

	return rows, err
}

func (observer) ConnBeginTx(ctx context.Context, conn driver.ConnBeginTx,
	txOpts driver.TxOptions) (driver.Tx, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "Begin")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))

	s := time.Now()
	tx, err := conn.BeginTx(ctx, txOpts)
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] begin, cost: %v", d)

	sqlDurations.WithLabelValues(
		name(ctx),
		"",
		"begin",
	).Observe(d.Seconds())

	return tx, err
}

func (observer) TxCommit(ctx context.Context, tx driver.Tx) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Commit")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))

	s := time.Now()
	err := tx.Commit()
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] commit, cost: %v", d)

	sqlDurations.WithLabelValues(
		name(ctx),
		"",
		"commit",
	).Observe(d.Seconds())

	return err
}

func (observer) TxRollback(ctx context.Context, tx driver.Tx) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Rollback")
	defer span.Finish()

	span.SetTag(string(ext.Component), "sqldb")
	span.SetTag(string(ext.DBInstance), name(ctx))

	s := time.Now()
	err := tx.Rollback()
	d := time.Since(s)

	log.Get(ctx).Debugf("[sqldb] rollback, cost: %v", d)

	sqlDurations.WithLabelValues(
		name(ctx),
		"",
		"rollback",
	).Observe(d.Seconds())

	return err
}
