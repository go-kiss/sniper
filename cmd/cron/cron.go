package cron

import (
	"context"
	"fmt"
	"time"
)

func init() {
	manual("foo", func(ctx context.Context) error {
		fmt.Printf("manual run foo with args: %+v\n", onceArgs)
		return nil
	})

	cron("bar", "@every 1m", func(ctx context.Context) error {
		fmt.Printf("run bar @%v\n", time.Now())
		return nil
	})
}
