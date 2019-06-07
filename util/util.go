package util

import (
	_ "sniper/util/conf" // init conf

	"sniper/util/log"
	"sniper/util/mc"
)

// GatherMetrics 收集一些被动指标
func GatherMetrics() {
	mc.GatherMetrics()
}

// Reset all utils
func Reset() {
	log.Reset()
}

// Stop all utils
func Stop() {
}
