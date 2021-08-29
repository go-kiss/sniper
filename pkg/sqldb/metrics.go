package sqldb

import (
	"github.com/prometheus/client_golang/prometheus"
)

var defBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1}

var sqlDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "sniper",
	Subsystem: "sqldb",
	Name:      "sql_durations_seconds",
	Help:      "sql latency distributions",
	Buckets:   defBuckets,
}, []string{"name", "table", "cmd"})

func init() {
	prometheus.MustRegister(sqlDurations)
}
