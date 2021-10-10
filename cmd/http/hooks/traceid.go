package hooks

import (
	"context"

	"github.com/go-kiss/sniper/pkg/trace"
	"github.com/go-kiss/sniper/pkg/twirp"
	"github.com/opentracing/opentracing-go"
)

var TraceID = &twirp.ServerHooks{
	RequestReceived: func(ctx context.Context) (context.Context, error) {
		traceID := trace.GetTraceID(ctx)
		twirp.SetHTTPResponseHeader(ctx, "x-trace-id", traceID)

		return ctx, nil
	},
	RequestRouted: func(ctx context.Context) (context.Context, error) {
		pkg, _ := twirp.PackageName(ctx)
		service, _ := twirp.ServiceName(ctx)
		method, _ := twirp.MethodName(ctx)

		api := "/" + pkg + "." + service + "/" + method

		_, ctx = opentracing.StartSpanFromContext(ctx, api)

		return ctx, nil
	},
	ResponseSent: func(ctx context.Context) {
		if span := opentracing.SpanFromContext(ctx); span != nil {
			span.Finish()
		}
	},
}
