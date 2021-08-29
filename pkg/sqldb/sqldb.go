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
	dbs map[string]*sqlx.DB
	rwl sync.RWMutex
)

type nameKey struct{}

// Get 获取数据库实例
//
// ctx, db := sqldb.Get(ctx, "foo")
// db.ExecContext(ctx, "select ...")
func Get(ctx context.Context, name string) (context.Context, *sqlx.DB) {
	ctx = context.WithValue(ctx, nameKey{}, name)
	rwl.RLock()
	defer rwl.RUnlock()
	if db, ok := dbs[name]; ok {
		return ctx, db
	}

	v, _, _ := sfg.Do(name, func() (interface{}, error) {
		dsn := conf.Get("SQLDB_DSN_" + name)
		var driver string
		if strings.HasPrefix(dsn, "file://") {
			driver = "db-sqlite"
		} else {
			driver = "db-mysql"
		}

		db := sqlx.MustOpen(driver, dsn)

		rwl.Lock()
		defer rwl.Unlock()
		dbs[name] = db

		collector := sqlstats.NewStatsCollector(name, db)
		prometheus.MustRegister(collector)

		return db, nil
	})

	return ctx, v.(*sqlx.DB)
}
