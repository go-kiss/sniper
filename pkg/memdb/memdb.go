package memdb

import (
	"context"
	"sniper/pkg/conf"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"
)

var (
	sfg singleflight.Group
	rwl sync.RWMutex

	dbs = map[string]*redis.Client{}
)

type nameKey struct{}

// Get 获取数据库实例
//
// ctx, db := sqldb.Get(ctx, "foo")
// db.ExecContext(ctx, "select ...")
func Get(ctx context.Context, name string) (context.Context, *redis.Client) {
	ctx = context.WithValue(ctx, nameKey{}, name)
	rwl.RLock()
	if db, ok := dbs[name]; ok {
		rwl.RUnlock()
		return ctx, db
	}
	rwl.RUnlock()

	v, _, _ := sfg.Do(name, func() (interface{}, error) {
		opts := &redis.Options{}

		dsn := conf.Get("MEMDB_DSN_" + name)
		setOptions(opts, dsn)

		db := redis.NewClient(opts)

		db.AddHook(observer{})

		collector := NewStatsCollector(name, db)
		prometheus.MustRegister(collector)

		rwl.Lock()
		defer rwl.Unlock()
		dbs[name] = db

		return db, nil
	})

	return ctx, v.(*redis.Client)
}
