package pkg

import (
	_ "sniper/pkg/conf"  // init conf
	_ "sniper/pkg/xhttp" // init http

	"sniper/pkg/log"
)

// GatherMetrics 收集一些被动指标
func GatherMetrics() {
}

// Reset all utils
func Reset() {
	log.Reset()
}

// Stop all utils
func Stop() {
}
