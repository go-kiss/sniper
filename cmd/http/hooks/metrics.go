package hooks

import (
	"github.com/prometheus/client_golang/prometheus"
)

var defBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1}
var rpcDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "sniper",
	Subsystem: "rpc",
	Name:      "server_durations_seconds",
	Help:      "RPC latency distributions",
	Buckets:   defBuckets,
}, []string{"path", "code"})

func init() {
	prometheus.MustRegister(rpcDurations)
}
