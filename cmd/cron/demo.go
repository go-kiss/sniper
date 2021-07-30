package cron

import (
	"context"
	"fmt"
	"time"
)

// 定时任务示例，开源专用
// 业务相关任务请使用 cron.go

func init() {
	manual("foo", func(ctx context.Context) error {
		fmt.Printf("manual run foo with args: %+v\n", onceArgs)
		return nil
	})

	cron("bar", "@every 1m", func(ctx context.Context) error {
		fmt.Printf("run bar @%v\n", time.Now())
		return nil
	})

	http("baz", "0 18-23 * * *", func(ctx context.Context) error {
		fmt.Printf("run http task @%v\n", time.Now())
		return nil
	})
}
