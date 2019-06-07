package metrics

import (
	"sniper/util/conf"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// RPCDurationsSeconds rpc 服务耗时
	RPCDurationsSeconds *prometheus.HistogramVec
	// DBDurationsSeconds mysql 调用耗时
	DBDurationsSeconds *prometheus.HistogramVec
	// MCDurationsSeconds memcache 调用耗时
	MCDurationsSeconds *prometheus.HistogramVec
	// RedisDurationsSeconds redis 调用耗时
	RedisDurationsSeconds *prometheus.HistogramVec
	// HTTPDurationsSeconds http 调用耗时
	HTTPDurationsSeconds *prometheus.HistogramVec
	// MQDurationsSeconds databus 调用耗时
	MQDurationsSeconds *prometheus.HistogramVec

	// LogTotal log 调用数量统计
	LogTotal *prometheus.CounterVec
	// JobTotal 定时任务数量统计
	JobTotal *prometheus.CounterVec

	// NetPoolHits 命中空闲连接数量
	NetPoolHits *prometheus.CounterVec
	// NetPoolMisses 未命中空闲连接数量
	NetPoolMisses *prometheus.CounterVec
	// NetPoolTimeouts 获取连接超时总数
	NetPoolTimeouts *prometheus.CounterVec
	// NetPoolStale 问题连接总数
	NetPoolStale *prometheus.CounterVec
	// NetPoolTotal 连接总数
	NetPoolTotal *prometheus.GaugeVec
	// NetPoolIdle 空闲连接总数
	NetPoolIdle *prometheus.GaugeVec
)

var defBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1}

func init() {
	RPCDurationsSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "sniper",
		Name:        "rpc_durations_seconds",
		Help:        "RPC latency distributions",
		Buckets:     defBuckets,
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"path", "code"})
	prometheus.MustRegister(RPCDurationsSeconds)

	DBDurationsSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "sniper",
		Name:        "db_durations_seconds",
		Help:        "MySQL latency distributions",
		Buckets:     defBuckets,
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "table", "cmd"})
	prometheus.MustRegister(DBDurationsSeconds)

	MCDurationsSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "sniper",
		Name:        "mc_durations_seconds",
		Help:        "MemCache latency distributions",
		Buckets:     defBuckets,
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "cmd"})
	prometheus.MustRegister(MCDurationsSeconds)

	RedisDurationsSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "sniper",
		Name:        "redis_durations_seconds",
		Help:        "Redis latency distributions",
		Buckets:     defBuckets,
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "cmd"})
	prometheus.MustRegister(RedisDurationsSeconds)

	HTTPDurationsSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "sniper",
		Name:        "http_durations_seconds",
		Help:        "HTTP latency distributions",
		Buckets:     defBuckets,
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"url", "status"})
	prometheus.MustRegister(HTTPDurationsSeconds)

	LogTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "sniper",
		Name:        "log_total",
		Help:        "log total",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"code"})
	prometheus.MustRegister(LogTotal)

	JobTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "sniper",
		Name:        "job_total",
		Help:        "job total",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"code"})
	prometheus.MustRegister(JobTotal)

	MQDurationsSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "sniper",
		Name:        "mq_durations_seconds",
		Help:        "Databus latency distributions",
		Buckets:     defBuckets,
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "role"})
	prometheus.MustRegister(MQDurationsSeconds)

	NetPoolHits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "sniper",
		Name:        "net_pool_hits",
		Help:        "net pool hits",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "type"})
	prometheus.MustRegister(NetPoolHits)

	NetPoolMisses = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "sniper",
		Name:        "net_pool_misses",
		Help:        "net pool misses",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "type"})
	prometheus.MustRegister(NetPoolMisses)

	NetPoolTimeouts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "sniper",
		Name:        "net_pool_timeouts",
		Help:        "net pool timeouts",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "type"})
	prometheus.MustRegister(NetPoolTimeouts)

	NetPoolStale = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "sniper",
		Name:        "net_pool_stale",
		Help:        "net pool stale",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "type"})
	prometheus.MustRegister(NetPoolStale)

	NetPoolTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "sniper",
		Name:        "net_pool_total",
		Help:        "net pool total",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "type"})
	prometheus.MustRegister(NetPoolTotal)

	NetPoolIdle = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "sniper",
		Name:        "net_pool_idle",
		Help:        "net pool idle",
		ConstLabels: map[string]string{"app": conf.AppID},
	}, []string{"name", "type"})
	prometheus.MustRegister(NetPoolIdle)
}
