// Package db 提供 mysql 封装
package db

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"sniper/util/conf"
	"sniper/util/errors"
	"sniper/util/log"
	"sniper/util/metrics"

	"github.com/go-sql-driver/mysql"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var dbs = make(map[string]*DB, 4)
var lock = sync.RWMutex{}

// DB 对象，有限开放 sql.DB 功能，支持上报 metrics
type DB struct {
	db   *sql.DB
	name string
	s    sql.DBStats
}

// Query sql 查询对象
type Query struct {
	table   string
	sql     string
	sqlType string
}

// Conn 简单 DB 接口。用于统一非事务和事务业务逻辑
type Conn interface {
	ExecContext(ctx context.Context, query Query, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query Query, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query Query, args ...interface{}) *sql.Row
}

type unionDB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// SQLInsert 构造 insert 查询
func SQLInsert(table string, sql string) Query {
	return Query{
		table:   table,
		sql:     sql,
		sqlType: "INSERT",
	}
}

// SQLDelete 构造 delete 查询
func SQLDelete(table string, sql string) Query {
	return Query{
		table:   table,
		sql:     sql,
		sqlType: "DELETE",
	}
}

// SQLUpdate 构造 update 查询
func SQLUpdate(table string, sql string) Query {
	return Query{
		table:   table,
		sql:     sql,
		sqlType: "UPDATE",
	}
}

// SQLSelect 构造 select 查询
func SQLSelect(table string, sql string) Query {
	return Query{
		table:   table,
		sql:     sql,
		sqlType: "SELECT",
	}
}

// Get 根据配置名字创建并返回 DB 连接池对象
//
// DB 配置名字格式为 DB_{$name}_DSN
// DB 配置内容格式请参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name
// Get 是并发安全的，可以在多协程下使用
func Get(ctx context.Context, name string) *DB {
	lock.RLock()
	db := dbs[name]
	lock.RUnlock()

	if db != nil {
		return db
	}

	dsn := conf.GetString("DB_" + name + "_DSN")

	sqldb, err := sql.Open("mysql", dsn)

	if err != nil {
		log.Get(ctx).Panic(err)
	}

	sqldb.SetMaxOpenConns(10)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxLifetime(60 * time.Second)

	db = &DB{db: sqldb, name: name}
	lock.Lock()
	dbs[name] = db
	lock.Unlock()

	return db
}

// Reset 关闭所有 DB 连接
// 新调用 Get 方法时会使用最新 DB 配置创建连接
func Reset() {
	for k, db := range dbs {
		db.db.Close()
		delete(dbs, k)
	}
}

// ExecContext 执行查询，无返回数据
func (db *DB) ExecContext(ctx context.Context, query Query, args ...interface{}) (sql.Result, error) {
	return execContext(ctx, db.name, db.db, query, args)
}

func execContext(ctx context.Context, name string, db unionDB, query Query, args []interface{}) (sql.Result, error) {
	log.Get(ctx).Debugf("[DB:%s] sql:%s args:%v", name, query.sql, args)

	span, ctx := opentracing.StartSpanFromContext(ctx, "ExecContext")
	defer span.Finish()

	span.SetTag(string(ext.Component), "mysql")
	span.SetTag(string(ext.DBInstance), name)
	span.SetTag(string(ext.DBStatement), query.sql)

	start := time.Now()
	r, err := db.ExecContext(ctx, query.sql, args...)
	duration := time.Since(start)

	metrics.DBDurationsSeconds.WithLabelValues(
		name,
		query.table,
		query.sqlType,
	).Observe(duration.Seconds())

	return r, errors.Wrap(err)
}

// QueryContext 执行查询，返回多行数据
func (db *DB) QueryContext(ctx context.Context, query Query, args ...interface{}) (*sql.Rows, error) {
	return queryContext(ctx, db.name, db.db, query, args)
}

func queryContext(ctx context.Context, name string, db unionDB, query Query, args []interface{}) (*sql.Rows, error) {
	log.Get(ctx).Debugf("[DB:%s] sql:%s args:%v", name, query.sql, args)

	span, ctx := opentracing.StartSpanFromContext(ctx, "QueryContext")
	defer span.Finish()

	span.SetTag(string(ext.Component), "mysql")
	span.SetTag(string(ext.DBInstance), name)
	span.SetTag(string(ext.DBStatement), query.sql)

	start := time.Now()
	r, err := db.QueryContext(ctx, query.sql, args...)
	duration := time.Since(start)

	metrics.DBDurationsSeconds.WithLabelValues(
		name,
		query.table,
		query.sqlType,
	).Observe(duration.Seconds())

	return r, errors.Wrap(err)
}

// QueryRowContext 执行查询，至多返回一行数据
func (db *DB) QueryRowContext(ctx context.Context, query Query, args ...interface{}) *sql.Row {
	return queryRowContext(ctx, db.name, db.db, query, args)
}

func queryRowContext(ctx context.Context, name string, db unionDB, query Query, args []interface{}) *sql.Row {
	log.Get(ctx).Debugf("[DB:%s] %s %v", name, query.sql, args)

	span, ctx := opentracing.StartSpanFromContext(ctx, "QueryRowContext")
	defer span.Finish()

	span.SetTag(string(ext.Component), "mysql")
	span.SetTag(string(ext.DBInstance), name)
	span.SetTag(string(ext.DBStatement), query.sql)

	start := time.Now()
	r := db.QueryRowContext(ctx, query.sql, args...)
	duration := time.Since(start)

	metrics.DBDurationsSeconds.WithLabelValues(
		name,
		query.table,
		query.sqlType,
	).Observe(duration.Seconds())

	return r
}

// Tx 事务对象简单封装
type Tx struct {
	tx    *sql.Tx
	db    string
	start time.Time
	sqls  []string
	args  [][]interface{}
}

func newTx(ctx context.Context, tx *sql.Tx, db string) *Tx {
	return &Tx{
		db:    db,
		tx:    tx,
		start: time.Now(),
	}
}

// msg 目前仅有 commit 和 rollback 两种
func (tx *Tx) log(ctx context.Context, msg string) {
	duration := time.Since(tx.start)

	if msg == "commit" {
		log.Get(ctx).Debugf("commit, total cost:%s", duration)
	} else {
		log.Get(ctx).Warnf("rollback, total cost:%s", duration)
	}

	metrics.DBDurationsSeconds.WithLabelValues(
		tx.db,
		"_tx",
		msg,
	).Observe(duration.Seconds())
}

// ExecContext 执行写查询
// 不鼓励在事务中使用读查询，所以只提供 ExecContext 方法
// 框架会根据返回错误自动提交或者回滚，所以不提供相应方法
func (tx *Tx) ExecContext(ctx context.Context, query Query, args ...interface{}) (sql.Result, error) {
	return execContext(ctx, tx.db, tx.tx, query, args)
}

// QueryContext 在事务中查询多行数据
func (tx *Tx) QueryContext(ctx context.Context, query Query, args ...interface{}) (*sql.Rows, error) {
	return queryContext(ctx, tx.db, tx.tx, query, args)
}

// QueryRowContext 在事务中查询单行数据
func (tx *Tx) QueryRowContext(ctx context.Context, query Query, args ...interface{}) *sql.Row {
	return queryRowContext(ctx, tx.db, tx.tx, query, args)
}

func (tx *Tx) rollback(ctx context.Context) error {
	tx.log(ctx, "rollback")
	if err := tx.tx.Rollback(); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (tx *Tx) commit(ctx context.Context) error {
	tx.log(ctx, "commit")
	if err := tx.tx.Commit(); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

// TxFunc 事务函数，返回 err 会回滚未提交事务，否则自动提交事务
type TxFunc func(ctx context.Context, tx Conn) error

// ExecTx 执行一次事务，回调函数返回 err 或者 panic 或者 ctx 取消都会回滚事务。
// 返回的 err 为 Commit 或者 Rollback 的错误
func (db *DB) ExecTx(ctx context.Context, f TxFunc) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ExecTx")
	defer span.Finish()

	span.SetTag(string(ext.Component), "mysql")

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err)
	}

	logger := log.Get(ctx)

	logger.Info("BeginTx")

	mytx := newTx(ctx, tx, db.name)

	defer func() {
		if p := recover(); p != nil {
			mytx.rollback(ctx)
			panic(p)
		}
	}()

	if err := f(ctx, mytx); err != nil {
		if err := mytx.rollback(ctx); err != nil {
			logger.Error("rollback failed", err)
		}
		return err
	}

	return mytx.commit(ctx)
}

// IsNoRowsErr 判断是否为 ErrNoRows 错误
func IsNoRowsErr(err error) bool {
	if err == nil {
		return false
	}

	return errors.Cause(err) == sql.ErrNoRows
}

// IsDuplicateEntryErr 判断是否为唯一键冲突错误
func IsDuplicateEntryErr(err error) bool {
	if err == nil {
		return false
	}

	// https://stackoverflow.com/a/41666013
	if me, ok := errors.Cause(err).(*mysql.MySQLError); ok {
		return me.Number == 1062
	}

	return false
}

// GatherMetrics 连接池状态指标
func GatherMetrics() {
	lock.RLock()
	defer lock.RUnlock()

	for _, c := range dbs {
		s := c.db.Stats()

		metrics.DBMaxOpenConnections.WithLabelValues(c.name).Set(float64(s.MaxOpenConnections))
		metrics.DBOpenConnections.WithLabelValues(c.name).Set(float64(s.OpenConnections))
		metrics.DBInUseConnections.WithLabelValues(c.name).Set(float64(s.InUse))
		metrics.DBIdleConnections.WithLabelValues(c.name).Set(float64(s.Idle))

		if d := s.WaitCount - c.s.WaitCount; d > 0 {
			metrics.DBWaitCount.WithLabelValues(c.name).Add(float64(d))
		}

		if d := s.MaxIdleClosed - c.s.MaxIdleClosed; d > 0 {
			metrics.DBMaxIdleClosed.WithLabelValues(c.name).Add(float64(d))
		}

		if d := s.MaxLifetimeClosed - c.s.MaxLifetimeClosed; d > 0 {
			metrics.DBMaxLifetimeClosed.WithLabelValues(c.name).Add(float64(d))
		}

		c.s = s
	}
}
