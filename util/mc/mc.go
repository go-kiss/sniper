// Package mc memcache 客户端组件
package mc

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"sniper/util/conf"
	"sniper/util/errors"
	"sniper/util/log"
	"sniper/util/metrics"

	"github.com/bilibili/memcache"
	"github.com/bilibili/net/pool"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	// ErrCacheMiss 未命中缓存
	ErrCacheMiss = memcache.ErrCacheMiss
	// ErrNotStored value未存下, 没有满足条件(i.e. Add or CompareAndSwap)
	ErrNotStored = memcache.ErrNotStored
)

var mcs = make(map[string]*MC, 4)
var lock = sync.RWMutex{}

// Item cache item
type Item memcache.Item

// ParseInt strconv.ParseInt(i.Value, base, bitSize)
func (i *Item) ParseInt(base int, bitSize int) (int64, error) {
	// memcache 的 decr 行为比较奇怪
	// decr 100 得到 99，但使用 get 得到的却是 "99 "
	s := strings.TrimRight(string(i.Value), " ")

	return strconv.ParseInt(s, base, bitSize)
}

// MC memcache 客户端实例
type MC struct {
	client *memcache.Client
	name   string

	// 连接池状态指标好多是 counter 类型
	// 对于 counter 类型，promethes 只提供 Add 方法
	// 所以需要记录上次状态好计算增量
	stats pool.Stats
}

// Get get mc by name
func Get(ctx context.Context, name string) *MC {
	lock.RLock()
	mc := mcs[name]
	lock.RUnlock()

	if mc != nil {
		return mc
	}

	host := conf.Get("MC_" + name + "_HOSTS")
	initConns := conf.GetInt("MC_" + name + "_INIT_CONNS")
	maxIdleConns := conf.GetInt("MC_" + name + "_MAX_IDLE_CONNS")

	client, err := memcache.New(host, initConns, maxIdleConns)
	if err != nil {
		log.Get(ctx).Panic(err)
	}

	mc = &MC{client: client, name: name}
	lock.Lock()
	mcs[name] = mc
	lock.Unlock()

	return mc
}

// Add 只在 key 不存在的时候设置新值
func (mc *MC) Add(ctx context.Context, item *Item) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Add")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", item.Key)

	logger := log.Get(ctx)

	logger.Debugf(
		"[MC] name:%s Add key:%s, value:%s, exptime:%d, flag:%d",
		mc.name,
		item.Key,
		string(item.Value),
		item.Expiration,
		item.Flags,
	)

	start := time.Now()
	err := mc.client.Add(ctx, (*memcache.Item)(item))
	duration := time.Since(start)

	if !memcache.IsResumableErr(err) {
		logger.Errorf(
			"[MC] name:%s Add key:%s, error:%+v",
			mc.name,
			item.Key,
			err,
		)
	}

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"ADD",
	).Observe(duration.Seconds())

	return errors.Wrap(err)
}

// CompareAndSwap cas
func (mc *MC) CompareAndSwap(ctx context.Context, item *Item) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CompareAndSwap")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", item.Key)
	logger := log.Get(ctx)

	logger.Debugf(
		"[MC] name:%s CompareAndSwap key:%s, value:%s, exptime:%d, flag:%d",
		mc.name,
		item.Key,
		string(item.Value),
		item.Expiration,
		item.Flags,
	)

	start := time.Now()
	err := mc.client.CompareAndSwap(ctx, (*memcache.Item)(item))
	duration := time.Since(start)

	if !memcache.IsResumableErr(err) {
		logger.Errorf(
			"[MC] name:%s CompareAndSwap key:%s, error:%+v",
			mc.name,
			item.Key,
			err,
		)
	}

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"CompareAndSwap",
	).Observe(duration.Seconds())

	return errors.Wrap(err)
}

// Decrement 减小 key 对应的值，key 不存在则报错
func (mc *MC) Decrement(ctx context.Context, key string, delta uint64) (uint64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Decrement")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", key)

	logger := log.Get(ctx)

	logger.Debugf("[MC] name:%s Decrement key:%s delta:%d", mc.name, key, delta)

	start := time.Now()
	r, err := mc.client.Decrement(ctx, key, delta)
	duration := time.Since(start)

	if !memcache.IsResumableErr(err) {
		logger.Errorf("[MC] name:%s Decrement key:%s delta:%d error:%+v", mc.name, key, delta, err)
	}

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"Decrement",
	).Observe(duration.Seconds())

	return r, err
}

// Delete del
func (mc *MC) Delete(ctx context.Context, key string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Delete")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", key)

	logger := log.Get(ctx)

	logger.Debugf("[MC] name:%s Delete %s", mc.name, key)

	start := time.Now()
	err := mc.client.Delete(ctx, key)
	duration := time.Since(start)

	// 删除不存在的 key 没有副作用
	if IsCacheMiss(err) {
		err = nil
	}

	if !memcache.IsResumableErr(err) {
		logger.Errorf("[MC] name:%s Delete %s, error:%+v", mc.name, key, err)
	}

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"Delete",
	).Observe(duration.Seconds())

	return errors.Wrap(err)
}

// Get get
func (mc *MC) Get(ctx context.Context, key string) (*Item, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", key)

	logger := log.Get(ctx)

	logger.Debugf("[MC] name:%s Get %s", mc.name, key)

	start := time.Now()
	i, err := mc.client.Get(ctx, key)
	duration := time.Since(start)

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"Get",
	).Observe(duration.Seconds())

	return (*Item)(i), errors.Wrap(err)
}

// GetMulti mget
// FIXME keys 为空的时候会报错
func (mc *MC) GetMulti(ctx context.Context, keys []string) (map[string]*Item, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetMulti")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", strings.Join(keys, ","))

	logger := log.Get(ctx)

	logger.Debugf("[MC] name:%s GetMulti %s", mc.name, keys)

	start := time.Now()
	i, err := mc.client.GetMulti(ctx, keys)
	duration := time.Since(start)

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"GetMulti",
	).Observe(duration.Seconds())

	items := make(map[string]*Item, len(i))
	for k, v := range i {
		items[k] = (*Item)(v)
	}

	return items, errors.Wrap(err)
}

// Increment key 不存在会自动创建，跟原生 mc 客户端不同
func (mc *MC) Increment(ctx context.Context, key string, delta uint64) (uint64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Increment")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", key)

	logger := log.Get(ctx)

	logger.Debugf("[MC] name:%s Increment key:%s delta:%d ", mc.name, key, delta)

	start := time.Now()
	c := mc.client
	r, err := c.Increment(ctx, key, delta)
	if IsCacheMiss(err) {
		// memcache.ErrNotStored 表示有并发请求执行 Add 成功，继续自增
		if err = c.Add(ctx, &memcache.Item{Key: key, Value: []byte("0")}); err == nil || err == memcache.ErrNotStored {
			r, err = c.Increment(ctx, key, delta)
		}
	}
	duration := time.Since(start)

	if !memcache.IsResumableErr(err) {
		logger.Errorf("[MC] name:%s Increment key:%s delta:%d error:%+v", mc.name, key, delta, err)
	}

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"Increment",
	).Observe(duration.Seconds())

	return r, errors.Wrap(err)
}

// Replace 更新已有数据，没有返回错误
func (mc *MC) Replace(ctx context.Context, item *Item) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Replace")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", item.Key)

	logger := log.Get(ctx)

	logger.Debugf(
		"[MC] name:%s Replace key:%s, value:%s, exptime:%d, flag:%d",
		mc.name,
		item.Key,
		string(item.Value),
		item.Expiration,
		item.Flags,
	)

	start := time.Now()
	err := mc.client.Replace(ctx, (*memcache.Item)(item))
	duration := time.Since(start)

	if !memcache.IsResumableErr(err) {
		logger.Errorf(
			"[MC] name:%s Replace key:%s, error:%+v",
			mc.name,
			item.Key,
			err,
		)
	}

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"Replace",
	).Observe(duration.Seconds())

	return errors.Wrap(err)
}

// Set 无条件设置数据
func (mc *MC) Set(ctx context.Context, item *Item) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Set")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", item.Key)

	logger := log.Get(ctx)

	logger.Debugf(
		"[MC] name:%s Set key:%s, value:%s, exptime:%d, flag:%d",
		mc.name,
		item.Key,
		string(item.Value),
		item.Expiration,
		item.Flags,
	)

	start := time.Now()
	err := mc.client.Set(ctx, (*memcache.Item)(item))
	duration := time.Since(start)

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"Set",
	).Observe(duration.Seconds())

	if !memcache.IsResumableErr(err) {
		logger.Errorf(
			"[MC] name:%s Set key:%s, error:%+v",
			mc.name,
			item.Key,
			err,
		)
	}

	return errors.Wrap(err)
}

// Touch 更新过期时间
func (mc *MC) Touch(ctx context.Context, key string, seconds int32) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Touch")
	defer span.Finish()
	span.SetTag(string(ext.Component), "memcached")
	span.SetTag(string(ext.DBInstance), mc.name)
	span.SetTag("mc.key", key)

	logger := log.Get(ctx)

	logger.Debugf("[MC] name:%s Touch %s %d", mc.name, key, seconds)

	start := time.Now()
	err := mc.client.Touch(ctx, key, seconds)
	duration := time.Since(start)

	if !memcache.IsResumableErr(err) {
		logger.Errorf(
			"[MC] name:%s Touch key:%s, error:%+v",
			mc.name,
			key,
			err,
		)
	}

	metrics.MCDurationsSeconds.WithLabelValues(
		mc.name,
		"Touch",
	).Observe(duration.Seconds())

	return errors.Wrap(err)
}

// Reset 关闭所有 MC 连接
// 新调用 Get 方法时会使用最新 MC 配置创建连接
//
// 如果在配置中开启 HOT_LOAD_MC 开关，则每次下发配置都会重置 MC 连接！
func Reset() {
	if !conf.GetBool("HOT_LOAD_MC") {
		return
	}

	lock.Lock()
	oldMCs := mcs
	mcs = make(map[string]*MC, 4)
	lock.Unlock()
	for k, c := range oldMCs {
		c.client.Close()
		delete(mcs, k)
	}
}

// IsCacheMiss 如果没有命中缓存则返回 true
func IsCacheMiss(err error) bool {
	return errors.Cause(err) == memcache.ErrCacheMiss
}

// IsNotStored value是否因为未满足条件而未存储
func IsNotStored(err error) bool {
	return errors.Cause(err) == memcache.ErrNotStored
}

// GatherMetrics 连接池状态指标
func GatherMetrics() {
	lock.RLock()
	defer lock.RUnlock()

	for _, c := range mcs {
		s := c.client.PoolStats()

		if d := s.Hits - c.stats.Hits; d >= 0 {
			metrics.NetPoolHits.WithLabelValues(
				c.name,
				"mc",
			).Add(float64(d))
		}

		if d := s.Misses - c.stats.Misses; d >= 0 {
			metrics.NetPoolMisses.WithLabelValues(
				c.name,
				"mc",
			).Add(float64(d))
		}

		if d := s.Timeouts - c.stats.Timeouts; d >= 0 {
			metrics.NetPoolTimeouts.WithLabelValues(
				c.name,
				"mc",
			).Add(float64(d))
		}

		if d := s.StaleConns - c.stats.StaleConns; d >= 0 {
			metrics.NetPoolStale.WithLabelValues(
				c.name,
				"mc",
			).Add(float64(d))
		}

		metrics.NetPoolTotal.WithLabelValues(
			c.name,
			"mc",
		).Set(float64(s.TotalConns))

		metrics.NetPoolIdle.WithLabelValues(
			c.name,
			"mc",
		).Set(float64(s.IdleConns))

		c.stats = *s
	}
}
