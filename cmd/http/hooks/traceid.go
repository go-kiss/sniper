package hooks

import (
	"context"
	"time"

	"sniper/pkg/ctxkit"
	"sniper/pkg/trace"
	"sniper/pkg/twirp"

	"github.com/opentracing/opentracing-go"
)

var TraceID = &twirp.ServerHooks{
	RequestReceived: func(ctx context.Context) (context.Context, error) {
		ctx = context.WithValue(ctx, ctxkit.StartTimeKey, time.Now())

		traceID := trace.GetTraceID(ctx)
		twirp.SetHTTPResponseHeader(ctx, "x-trace-id", traceID)

		ctx = ctxkit.WithTraceID(ctx, traceID)

		return ctx, nil
	},
	RequestRouted: func(ctx context.Context) (context.Context, error) {
		pkg, _ := twirp.PackageName(ctx)
		service, _ := twirp.ServiceName(ctx)
		method, _ := twirp.MethodName(ctx)

		api := "/" + pkg + "." + service + "/" + method

		span, ctx := opentracing.StartSpanFromContext(ctx, api)
		ctx = context.WithValue(ctx, spanKey, span)

		return ctx, nil
	},
	ResponseSent: func(ctx context.Context) {
		if span, ok := ctx.Value(spanKey).(opentracing.Span); ok {
			span.Finish()
		}
	},
}
