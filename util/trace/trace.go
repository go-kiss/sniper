package trace

import (
	"context"
	"io"
	"net/http"

	"sniper/util/conf"

	opentracing "github.com/opentracing/opentracing-go"
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
			Type:  jaeger.SamplerTypeConst,
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

// InjectTrace 注入 OpenTracing 头信息
func InjectTraceHeader(ctx opentracing.SpanContext, req *http.Request) {
	opentracing.GlobalTracer().Inject(
		ctx,
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	jctx, ok := ctx.(jaeger.SpanContext)
	if !ok {
		return
	}

	// 兼容主站老的 trace 逻辑
	req.Header["Bili-Trace-Id"] = req.Header["Uber-Trace-Id"]

	// Envoy 使用 Zipkin 风格头信息
	// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/observability/tracing
	req.Header.Set("x-b3-traceid", jctx.TraceID().String())
	req.Header.Set("x-b3-spanid", jctx.SpanID().String())
	req.Header.Set("x-b3-parentspanid", jctx.ParentID().String())
	if jctx.IsSampled() {
		req.Header.Set("x-b3-sampled", "1")
	}
	if jctx.IsDebug() {
		req.Header.Set("x-b3-flags", "1")
	}
}

// Stop 停止 trace 协程
func Stop() {
	closer.Close()
}
