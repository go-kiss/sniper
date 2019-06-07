package hook

import (
	"context"
	"time"

	"sniper/util/ctxkit"
	"sniper/util/trace"

	"github.com/bilibili/twirp"
)

// NewRequestID 生成唯一请求标识并记录到 ctx
func NewRequestID() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			ctx = context.WithValue(ctx, ctxkit.StartTimeKey, time.Now())

			traceID := trace.GetTraceID(ctx)
			twirp.SetHTTPResponseHeader(ctx, "x-trace-id", traceID)

			ctx = ctxkit.WithTraceID(ctx, traceID)

			return ctx, nil
		},
	}
}
