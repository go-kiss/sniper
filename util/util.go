package util

import (
	_ "sniper/util/conf" // init conf

	"sniper/util/db"
	"sniper/util/log"
	"sniper/util/mc"
	"sniper/util/redis"
)

// GatherMetrics 收集一些被动指标
func GatherMetrics() {
	mc.GatherMetrics()
	redis.GatherMetrics()
	db.GatherMetrics()
}

// Reset all utils
func Reset() {
	log.Reset()
}

// Stop all utils
func Stop() {
}
