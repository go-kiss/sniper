package sqldb

import (
	"context"
	"sniper/pkg/conf"
	"strings"
	"sync"

	"github.com/dlmiddlecote/sqlstats"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"
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
// ctx, db := sqldb.Get(ctx, "foo")
// db.ExecContext(ctx, "select ...")
func Get(ctx context.Context, name string) (context.Context, *DB) {
	ctx = context.WithValue(ctx, nameKey{}, name)
	rwl.RLock()
	if db, ok := dbs[name]; ok {
		rwl.RUnlock()
		return ctx, db
	}
	rwl.RUnlock()

	v, _, _ := sfg.Do(name, func() (interface{}, error) {
		dsn := conf.Get("SQLDB_DSN_" + name)
		var driver string
		if strings.HasPrefix(dsn, "file:") || dsn == ":memory:" {
			driver = "db-sqlite"
		} else {
			driver = "db-mysql"
		}

		sdb := sqlx.MustOpen(driver, dsn)

		db := &DB{sdb}

		rwl.Lock()
		defer rwl.Unlock()
		dbs[name] = db

		collector := sqlstats.NewStatsCollector(name, db)
		prometheus.MustRegister(collector)

		return db, nil
	})

	return ctx, v.(*DB)
}
