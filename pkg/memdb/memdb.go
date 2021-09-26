package memdb

import (
	"sync"

	"sniper/pkg/conf"

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

// Get 获取缓存实例
//
// db := Get("foo")
// db.Set(ctx, "a", "123", 0)
func Get(name string) *redis.Client {
	rwl.RLock()
	if db, ok := dbs[name]; ok {
		rwl.RUnlock()
		return db
	}
	rwl.RUnlock()

	v, _, _ := sfg.Do(name, func() (interface{}, error) {
		opts := &redis.Options{}

		dsn := conf.Get("MEMDB_DSN_" + name)
		setOptions(opts, dsn)

		db := redis.NewClient(opts)

		db.AddHook(&observer{name: name})

		collector := NewStatsCollector(name, db)
		prometheus.MustRegister(collector)

		rwl.Lock()
		defer rwl.Unlock()
		dbs[name] = db

		return db, nil
	})

	return v.(*redis.Client)
}
