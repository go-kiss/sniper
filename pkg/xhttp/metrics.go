package xhttp

import (
	"github.com/prometheus/client_golang/prometheus"
)

var defBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1}

var httpDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "sniper",
	Subsystem: "http",
	Name:      "req_durations_seconds",
	Help:      "HTTP latency distributions",
	Buckets:   defBuckets,
}, []string{"url", "status"})

func init() {
	prometheus.MustRegister(httpDurations)
}
