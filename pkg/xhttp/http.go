// Package http 提供基础 http 客户端组件
//
// 本包通过替换 http.DefaultTransport 实现以下功能：
// - 日志(logging)
// - 链路追踪(tracing)
// - 指标监控(metrics)
//
// 请务必使用 http.NewRequestWithContext 构造 req 对象，这样才能传递 ctx 信息。
//
// 如果希望使用自定义 Transport，需要将 RoundTrip 的逻辑
// 最终委托给 http.DefaultTransport
//
// 使用示例：
//   req, _ := http.NewRequestWithContext(ctx, method, url, body)
//   c := &http.Client{
//   	Timeout: 1 * time.Second,
//   }
//   resp, err := c.Do(req)
package xhttp

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"sniper/pkg/log"
	"sniper/pkg/metrics"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func init() {
	http.DefaultTransport = &roundTripper{
		r: http.DefaultTransport,
	}
}

type roundTripper struct {
	r http.RoundTripper
}

var digitsRE = regexp.MustCompile(`\b\d+\b`)

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	span, ctx := opentracing.StartSpanFromContext(ctx, "DoHTTP")
	defer span.Finish()

	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	start := time.Now()
	resp, err := r.r.RoundTrip(req)
	duration := time.Since(start)

	url := fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)

	status := http.StatusOK
	if err != nil {
		status = http.StatusInternalServerError
	} else {
		status = resp.StatusCode
	}

	log.Get(ctx).Debugf(
		"[HTTP] method:%s url:%s status:%d query:%s",
		req.Method,
		url,
		status,
		req.URL.RawQuery,
	)

	span.SetTag(string(ext.Component), "http")
	span.SetTag(string(ext.HTTPUrl), url)
	span.SetTag(string(ext.HTTPMethod), req.Method)
	span.SetTag(string(ext.HTTPStatusCode), status)

	// 在 url 附带参数会产生大量 metrics 指标，影响 prometheus 性能。
	// 默认会把 url 中带有的纯数字替换成 %d
	// /v123/4/56/foo => /v123/%d/%d/foo
	url = digitsRE.ReplaceAllString(url, "%d")

	metrics.HTTPDurationsSeconds.WithLabelValues(
		url,
		fmt.Sprint(status),
	).Observe(duration.Seconds())

	return resp, err
}
