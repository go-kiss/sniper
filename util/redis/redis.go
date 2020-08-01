package redis

import (
	"context"
	"strings"
	"sync"
	"time"

	"sniper/util/conf"
	"sniper/util/log"
	"sniper/util/metrics"

	"github.com/bilibili/net/pool"
	"github.com/bilibili/redis"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	// FlagNX Not Exist
	FlagNX = redis.FlagNX
)

// Item 值对象
type Item redis.Item

// ZSetValue 有序集合对象
type ZSetValue redis.ZSetValue

// Redis 客户端对象
type Redis struct {
	client redis.Client
	name   string

	// 连接池状态指标好多是 counter 类型
	// 对于 counter 类型，promethes 只提供 Add 方法
	// 所以需要记录上次状态好计算增量
	stats pool.Stats
}

// ErrNonExist 查询的 key 不存在或者榜单中 member 不存在
var ErrNonExist = redis.Nil

var rs = make(map[string]*Redis, 4)
var lock = sync.RWMutex{}

// Get 获取一个 redis 实例
func Get(ctx context.Context, name string) *Redis {
	lock.RLock()
	r := rs[name]
	lock.RUnlock()

	if r != nil {
		return r
	}

	c := redis.New(redis.Options{
		// TODO 支持更多配置
		Address:      conf.Get("REDIS_" + name + "_HOST"),
		PoolSize:     conf.GetInt("REDIS_" + name + "_MAX_CONNS"),
		MinIdleConns: conf.GetInt("REDIS_" + name + "_INIT_CONNS"),
		MaxConnAge:   60 * time.Second, // 连接存活最长时间
		IdleTimeout:  30 * time.Second, // 最大空闲时间
	})
	r = &Redis{client: c, name: name}

	lock.Lock()
	rs[name] = r
	lock.Unlock()

	return r
}

// Get 查询单个缓存
func (r *Redis) Get(ctx context.Context, key string) (item *Item, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] Get %s", r.name, key)

	start := time.Now()
	i, err := r.client.Get(ctx, key)
	item = (*Item)(i)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"Get",
	).Observe(duration.Seconds())

	return
}

// MGet 批量查询缓存
func (r *Redis) MGet(ctx context.Context, keys []string) (items map[string]*Item, err error) {
	keyName := strings.Join(keys, " ")

	span, ctx := opentracing.StartSpanFromContext(ctx, "MGet")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", keyName)

	log.Get(ctx).Debugf("[Redis:%s] MGet %s", r.name, keyName)

	start := time.Now()
	items = make(map[string]*Item, len(keys))
	redisItems, err := r.client.MGet(ctx, keys)
	for k, v := range redisItems {
		items[k] = (*Item)(v)
	}
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"MGet",
	).Observe(duration.Seconds())

	return
}

// Eval 执行lua脚本
func (r *Redis) Eval(ctx context.Context, script string, keys []string, argvs ...interface{}) (*redis.EvalReturn, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Eval")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", "")

	log.Get(ctx).Debugf("[Redis:%s] Eval %s", r.name, script)

	start := time.Now()
	val, err := r.client.Eval(ctx, script, keys, argvs...)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"Eval",
	).Observe(duration.Seconds())

	return val, err
}

// Set 设置单个缓存
func (r *Redis) Set(ctx context.Context, item *Item) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Set")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", item.Key)

	log.Get(ctx).Debugf("[Redis:%s] Set %s", r.name, item.Key)

	start := time.Now()
	err := r.client.Set(ctx, (*redis.Item)(item))
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"Set",
	).Observe(duration.Seconds())

	return err
}

// Del 删除一个或多个缓存
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	keyName := strings.Join(keys, ",")

	span, ctx := opentracing.StartSpanFromContext(ctx, "Del")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", keyName)

	log.Get(ctx).Debugf("[Redis:%s] Del %s", r.name, keyName)

	start := time.Now()
	err := r.client.Del(ctx, keys...)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"Del",
	).Observe(duration.Seconds())

	return err
}

// IncrBy 数值增加
func (r *Redis) IncrBy(ctx context.Context, key string, by int64) (i int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "IncrBy")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] IncrBy %s", r.name, key)

	start := time.Now()
	i, err = r.client.IncrBy(ctx, key, by)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"IncrBy",
	).Observe(duration.Seconds())

	return
}

// DecrBy 数值减去
func (r *Redis) DecrBy(ctx context.Context, key string, by int64) (i int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "DecrBy")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] DecrBy %s", r.name, key)

	start := time.Now()
	i, err = r.client.DecrBy(ctx, key, by)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"DecrBy",
	).Observe(duration.Seconds())

	return
}

// Expire 设置过期时间
// ttl 表示: ttl 秒后 key 失效
func (r *Redis) Expire(ctx context.Context, key string, ttl int32) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Expire")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] Expire %s", r.name, key)

	start := time.Now()
	err := r.client.Expire(ctx, key, ttl)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"Expire",
	).Observe(duration.Seconds())

	return err
}

// TTL 以秒为单位，返回给定 key 的剩余生存时间
func (r *Redis) TTL(ctx context.Context, key string) (ttl int32, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "TTL")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] TTL %s", r.name, key)

	start := time.Now()
	ttl, err = r.client.TTL(ctx, key)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"TTL",
	).Observe(duration.Seconds())

	return
}

// ZAdd 将一个元素及其 score 值加入到有序集 key 当中
func (r *Redis) ZAdd(ctx context.Context, item *Item) (added int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZAdd")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", item.Key)

	log.Get(ctx).Debugf("[Redis:%s] ZAdd %s", r.name, item.Key)

	start := time.Now()
	added, err = r.client.ZAdd(ctx, (*redis.Item)(item))
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZAdd",
	).Observe(duration.Seconds())

	return
}

// ZIncrBy 为有序集 key 的成员 member 的 score 值加上增量 by
func (r *Redis) ZIncrBy(ctx context.Context, key, member string, by float64) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZIncrBy")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZIncrBy %s", r.name, key)

	start := time.Now()
	err := r.client.ZIncrBy(ctx, key, member, by)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZIncrBy",
	).Observe(duration.Seconds())

	return err
}

// ZRangeByScore 返回有序集 key 中，指定分数区间内的成员
// 其中成员的位置按 score 值递增(从小到大)来排序
// 具有相同 score 值的成员按字典序(lexicographical order)来排列
// 仅count > 0时 offset, count 参数有效
func (r *Redis) ZRangeByScore(ctx context.Context, key string, min, max float64, offset, count int64) (values []*ZSetValue, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRangeByScore")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRangeByScore %s", r.name, key)

	now := time.Now()
	res, err := r.client.ZRangeByScore(ctx, key, min, max, offset, count)
	for _, s := range res {
		values = append(values, (*ZSetValue)(s))
	}
	duration := time.Since(now)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRangeByScore",
	).Observe(duration.Seconds())

	return
}

// ZRevRangeByScore 返回有序集 key 中，指定分数区间内的成员
// 其中成员的位置按 score 值递减(从大到小)来排列
// 具有相同 score 值的成员按字典序的逆序(reverse lexicographical order)排列。
// 仅count > 0时 offset, count 参数有效
func (r *Redis) ZRevRangeByScore(ctx context.Context, key string, max, min float64, offset, count int64) (values []*ZSetValue, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRevRangeBySocre")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRevRangeByScore %s", r.name, key)

	now := time.Now()
	res, err := r.client.ZRevRangeByScore(ctx, key, max, min, offset, count)
	for _, s := range res {
		values = append(values, (*ZSetValue)(s))
	}
	duration := time.Since(now)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRevRangeByScore",
	).Observe(duration.Seconds())

	return
}

// ZRange 返回有序集 key 中，指定区间内的成员
// 其中成员的位置按 score 值递增(从小到大)来排序
// 具有相同 score 值的成员按字典序(lexicographical order)来排列
func (r *Redis) ZRange(ctx context.Context, key string, start, stop int64) (values []*ZSetValue, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRange")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRange %s", r.name, key)

	now := time.Now()
	res, err := r.client.ZRange(ctx, key, start, stop)
	for _, s := range res {
		values = append(values, (*ZSetValue)(s))
	}
	duration := time.Since(now)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRange",
	).Observe(duration.Seconds())

	return
}

// ZRevRange 返回有序集 key 中，指定区间内的成员
// 其中成员的位置按 score 值递减(从大到小)来排列
// 具有相同 score 值的成员按字典序的逆序(reverse lexicographical order)排列。
func (r *Redis) ZRevRange(ctx context.Context, key string, start, stop int64) (values []*ZSetValue, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRevRange")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRevRange %s", r.name, key)

	now := time.Now()
	res, err := r.client.ZRevRange(ctx, key, start, stop)
	for _, s := range res {
		values = append(values, (*ZSetValue)(s))
	}
	duration := time.Since(now)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRevRange",
	).Observe(duration.Seconds())

	return
}

// ZRank 返回有序集 key 中成员 member 的排名
// 其中有序集成员按 score 值递增(从小到大)顺序排列
// 排名以 0 为底，也就是说， score 值最小的成员排名为 0
func (r *Redis) ZRank(ctx context.Context, key, member string) (rank int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRank")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRank %s", r.name, key)

	start := time.Now()
	rank, err = r.client.ZRank(ctx, key, member)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRank",
	).Observe(duration.Seconds())

	return
}

// ZRevRank 返回有序集 key 中成员 member 的排名
// 其中有序集成员按 score 值递减(从大到小)排序
func (r *Redis) ZRevRank(ctx context.Context, key, member string) (rank int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRevRank")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRevRank %s", r.name, key)

	start := time.Now()
	rank, err = r.client.ZRevRank(ctx, key, member)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRevRank",
	).Observe(duration.Seconds())

	return
}

// ZScore 返回有序集 key 中，成员 member 的 score 值
func (r *Redis) ZScore(ctx context.Context, key, member string) (score float64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZScore")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZScore %s", r.name, key)

	start := time.Now()
	score, err = r.client.ZScore(ctx, key, member)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZScore",
	).Observe(duration.Seconds())

	return
}

// ZCard 返回有序集 key 的基数
func (r *Redis) ZCard(ctx context.Context, key string) (card int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZCard")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZCard %s", r.name, key)

	start := time.Now()
	card, err = r.client.ZCard(ctx, key)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZCard",
	).Observe(duration.Seconds())

	return
}

// ZCount 返回有序集 key 中， score 值在 min 和 max 之间(默认包括 score 值等于 min 或 max)的成员的数量
func (r *Redis) ZCount(ctx context.Context, key, min, max string) (i int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZCount")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZCount %s", r.name, key)

	start := time.Now()
	i, err = r.client.ZCount(ctx, key, min, max)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZCount",
	).Observe(duration.Seconds())

	return
}

// ZRem 移除有序集 key 中的一个或多个成员
func (r *Redis) ZRem(ctx context.Context, keys ...string) error {
	keyName := strings.Join(keys, ",")

	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRem")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", keyName)

	log.Get(ctx).Debugf("[Redis:%s] ZRem %s", r.name, keyName)

	start := time.Now()
	err := r.client.ZRem(ctx, keys...)
	duration := time.Since(start)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRem",
	).Observe(duration.Seconds())

	return err
}

// ZRemRangeByRank 移除有序集 key 中，指定排名(rank)区间内的所有成员
// 移除排名在 [start, stop] 区间的成员，返回被移除的成员的数量
func (r *Redis) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (i int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRemRangeByRank")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRemRangeByRank %s", r.name, key)

	now := time.Now()
	i, err = r.client.ZRemRangeByRank(ctx, key, start, stop)
	duration := time.Since(now)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRemRangeByRank",
	).Observe(duration.Seconds())

	return
}

// ZRemRangeByScore  移除有序集 key 中，所有 score 值介于 min 和 max 之间(包括等于 min 或 max)的成员
func (r *Redis) ZRemRangeByScore(ctx context.Context, key, min, max string) (i int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ZRemRangeByScore")
	defer span.Finish()
	span.SetTag(string(ext.Component), "redis")
	span.SetTag(string(ext.DBInstance), r.name)
	span.SetTag("redis.key", key)

	log.Get(ctx).Debugf("[Redis:%s] ZRemRangeByScore %s", r.name, key)

	now := time.Now()
	i, err = r.client.ZRemRangeByScore(ctx, key, min, max)
	duration := time.Since(now)

	metrics.RedisDurationsSeconds.WithLabelValues(
		r.name,
		"ZRemRangeByScore",
	).Observe(duration.Seconds())

	return
}

// GatherMetrics 连接池状态指标
func GatherMetrics() {
	lock.RLock()
	defer lock.RUnlock()

	const name = "redis"

	for _, c := range rs {
		s := c.client.PoolStats()

		if d := s.Hits - c.stats.Hits; d >= 0 {
			metrics.NetPoolHits.WithLabelValues(
				c.name,
				name,
			).Add(float64(d))
		}

		if d := s.Misses - c.stats.Misses; d >= 0 {
			metrics.NetPoolMisses.WithLabelValues(
				c.name,
				name,
			).Add(float64(d))
		}

		if d := s.Timeouts - c.stats.Timeouts; d >= 0 {
			metrics.NetPoolTimeouts.WithLabelValues(
				c.name,
				name,
			).Add(float64(d))
		}

		if d := s.StaleConns - c.stats.StaleConns; d >= 0 {
			metrics.NetPoolStale.WithLabelValues(
				c.name,
				name,
			).Add(float64(d))
		}

		metrics.NetPoolTotal.WithLabelValues(
			c.name,
			name,
		).Set(float64(s.TotalConns))

		metrics.NetPoolIdle.WithLabelValues(
			c.name,
			name,
		).Set(float64(s.IdleConns))

		c.stats = *s
	}
}
