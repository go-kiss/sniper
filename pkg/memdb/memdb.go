package memdb

import (
	"sync"

	"github.com/go-kiss/sniper/pkg/conf"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"
)

var (
	sfg singleflight.Group
	rwl sync.RWMutex

	dbs = map[string]*Client{}
)

type nameKey struct{}

// Client redis 客户端
type Client struct {
	redis.UniversalClient
}

// Get 获取缓存实例
//
// db := Get("foo")
// db.Set(ctx, "a", "123", 0)
func Get(name string) *Client {
	rwl.RLock()
	if db, ok := dbs[name]; ok {
		rwl.RUnlock()
		return db
	}
	rwl.RUnlock()

	v, _, _ := sfg.Do(name, func() (interface{}, error) {
		opts := &redis.UniversalOptions{}

		dsn := conf.Get("MEMDB_DSN_" + name)
		setOptions(opts, dsn)

		rdb := redis.NewUniversalClient(opts)

		rdb.AddHook(observer{name: name})

		collector := NewStatsCollector(name, rdb)
		prometheus.MustRegister(collector)

		db := &Client{rdb}

		rwl.Lock()
		defer rwl.Unlock()
		dbs[name] = db

		return db, nil
	})

	return v.(*Client)
}
