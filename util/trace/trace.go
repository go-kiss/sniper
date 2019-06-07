package trace

import (
	"context"
	"net/http"
	"os"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
)

func init() {
	if addr := os.Getenv("DAPPER_ADDR"); addr == "" {
		opentracing.SetGlobalTracer(defaultAmyTracer)
	}
}

// 如果使用 jager 则需改为 uber-trace-id
var traceHeader = "bili-trace-id"

// GetTraceID 查询 trace_id
func GetTraceID(ctx context.Context) string {
	traceID := "no-trace-id"

	if span := opentracing.SpanFromContext(ctx); span != nil {
		headers := http.Header{}

		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(headers),
		)

		trace := headers.Get(traceHeader)
		i := strings.IndexByte(trace, ':')
		if i > 0 {
			traceID = trace[:i]
		}
	}

	return traceID
}
