package memdb

import (
	"github.com/prometheus/client_golang/prometheus"
)

var defBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1}

var redisDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "sniper",
	Subsystem: "memdb",
	Name:      "redis_durations_seconds",
	Help:      "redis latency distributions",
	Buckets:   defBuckets,
}, []string{"name", "cmd"})

func init() {
	prometheus.MustRegister(redisDurations)
}
