package memdb

import (
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

var defBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1}

var redisDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "sniper",
	Subsystem: "memdb",
	Name:      "commands_duration_seconds",
	Help:      "commands latency distributions",
	Buckets:   defBuckets,
}, []string{"db_name", "cmd"})

func init() {
	prometheus.MustRegister(redisDurations)
}

type StatsCollector struct {
	db *redis.Client

	// descriptions of exported metrics
	hitDesc     *prometheus.Desc
	missDesc    *prometheus.Desc
	timeoutDesc *prometheus.Desc
	totalDesc   *prometheus.Desc
	idleDesc    *prometheus.Desc
	staleDesc   *prometheus.Desc
}

const (
	namespace = "sniper"
	subsystem = "memdb_connections"
)

// NewStatsCollector creates a new StatsCollector.
func NewStatsCollector(dbName string, db *redis.Client) *StatsCollector {
	labels := prometheus.Labels{"db_name": dbName}
	return &StatsCollector{
		db: db,
		hitDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "hit"),
			"The number number of times free connection was NOT found in the pool.",
			nil,
			labels,
		),
		missDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "miss"),
			"The number of times free connection was found in the pool.",
			nil,
			labels,
		),
		timeoutDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "timeout"),
			"The number of times a wait timeout occurred.",
			nil,
			labels,
		),
		totalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total"),
			"The number of total connections in the pool.",
			nil,
			labels,
		),
		idleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle"),
			"The number of idle connections in the pool.",
			nil,
			labels,
		),
		staleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "stale"),
			"The number of stale connections in the pool.",
			nil,
			labels,
		),
	}
}

// Describe implements the prometheus.Collector interface.
func (c StatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hitDesc
	ch <- c.missDesc
	ch <- c.timeoutDesc
	ch <- c.totalDesc
	ch <- c.idleDesc
	ch <- c.staleDesc
}

// Collect implements the prometheus.Collector interface.
func (c StatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.db.PoolStats()

	ch <- prometheus.MustNewConstMetric(
		c.hitDesc,
		prometheus.CounterValue,
		float64(stats.Hits),
	)
	ch <- prometheus.MustNewConstMetric(
		c.missDesc,
		prometheus.CounterValue,
		float64(stats.Misses),
	)
	ch <- prometheus.MustNewConstMetric(
		c.timeoutDesc,
		prometheus.CounterValue,
		float64(stats.Timeouts),
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalDesc,
		prometheus.GaugeValue,
		float64(stats.TotalConns),
	)
	ch <- prometheus.MustNewConstMetric(
		c.idleDesc,
		prometheus.GaugeValue,
		float64(stats.IdleConns),
	)
	ch <- prometheus.MustNewConstMetric(
		c.staleDesc,
		prometheus.GaugeValue,
		float64(stats.StaleConns),
	)
}
