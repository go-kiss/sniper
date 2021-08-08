package trace

import (
	"context"
	"io"

	"sniper/pkg/conf"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

var closer io.Closer

func init() {
	// job 机器的 trace 太大了，添加一个开关
	if conf.GetBool("NO_JAEGER") {
		return
	}

	// agent 部署在 k8s 的宿主机
	// 宿主机需要使用 HOST 环境变量获取
	host := conf.Get("HOST")
	if host == "" {
		host = conf.Get("JAEGER_AGENT_HOST")
		if host == "" {
			host = "127.0.0.1"
		}
	}

	port := conf.Get("JAEGER_AGENT_PORT")
	if port == "" {
		port = "6831"
	}

	cfg := config.Configuration{
		ServiceName: conf.AppID,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: conf.GetFloat64("JAEGER_SAMPLER_PARAM"),
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: host + ":" + port,
		},
	}

	tracer, c, err := cfg.NewTracer(
		config.Logger(log.NullLogger),
		config.Metrics(metrics.NullFactory),
	)
	if err != nil {
		panic(err)
	}

	closer = c
	opentracing.SetGlobalTracer(tracer)
}

// GetTraceID 查询 trace_id
func GetTraceID(ctx context.Context) (traceID string) {
	traceID = "no-trace-id"

	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	jctx, ok := (span.Context()).(jaeger.SpanContext)
	if !ok {
		return
	}

	traceID = jctx.TraceID().String()

	return
}

// StartFollowSpanFromContext 开起一个 follow 类型 span
// follow 类型用于异步任务，可能在 root span 结束之后才完成。
func StartFollowSpanFromContext(ctx context.Context, operation string) (opentracing.Span, context.Context) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return opentracing.StartSpanFromContext(ctx, operation)
	}

	return opentracing.StartSpanFromContext(ctx, operation, opentracing.FollowsFrom(span.Context()))
}

// Stop 停止 trace 协程
func Stop() {
	closer.Close()
}
