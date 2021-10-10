package sqldb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"sync"

	"github.com/go-kiss/sniper/pkg/conf"
	"github.com/dlmiddlecote/sqlstats"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/ngrok/sqlmw"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"
	"modernc.org/sqlite"
)

var (
	sfg singleflight.Group
	rwl sync.RWMutex

	dbs = map[string]*DB{}
)

type nameKey struct{}

// DB 扩展 sqlx.DB
type DB struct {
	*sqlx.DB
}

// Tx 扩展 sqlx.Tx
type Tx struct {
	*sqlx.Tx
}

// Get 获取数据库实例
//
// db := sqldb.Get(ctx, "foo")
// db.ExecContext(ctx, "select ...")
func Get(ctx context.Context, name string) *DB {
	rwl.RLock()
	if db, ok := dbs[name]; ok {
		rwl.RUnlock()
		return db
	}
	rwl.RUnlock()

	v, _, _ := sfg.Do(name, func() (interface{}, error) {
		dsn := conf.Get("SQLDB_DSN_" + name)
		isSqlite := strings.HasPrefix(dsn, "file:") || dsn == ":memory:"
		var driverName string
		var driver driver.Driver
		if isSqlite {
			driverName = "db-sqlite:" + name
			driver = sqlmw.Driver(&sqlite.Driver{}, observer{name: name})
		} else {
			driverName = "db-mysql:" + name
			driver = sqlmw.Driver(mysql.MySQLDriver{}, observer{name: name})
		}

		sql.Register(driverName, driver)
		sdb := sqlx.MustOpen(driverName, dsn)

		db := &DB{sdb}

		rwl.Lock()
		defer rwl.Unlock()
		dbs[name] = db

		collector := sqlstats.NewStatsCollector(name, db)
		prometheus.MustRegister(collector)

		return db, nil
	})

	return v.(*DB)
}
