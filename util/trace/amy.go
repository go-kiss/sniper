package trace

// amyTracer 是一个虚拟 tracer，用来生成 trace-id
// sniper 的 tracer-id 是由 opentracing tracer 生成的
// 但内部使用的 dapper 无法开源，又不能强制大家都接入 jaeger
// 所以默认提供一个虚拟 tracer
//
// 之所以命名为 amy，主要是它排在 dapper 前面，方便使用 init 初始化
// 现在不依赖 init 了

import (
	"math/rand"
	"strconv"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

var defaultAmyTracer = amyTracer{}

type amyTracer struct{}

type amySpan struct {
	ctx amySpanContext
}
type amySpanContext struct {
	id uint64
}

func (a amySpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

func (a amySpan) Context() opentracing.SpanContext {
	return a.ctx
}

func (a amySpan) Finish()                                                {}
func (a amySpan) LogEvent(event string)                                  {}
func (a amySpan) Log(data opentracing.LogData)                           {}
func (a amySpan) LogKV(keyVals ...interface{})                           {}
func (a amySpan) LogFields(fields ...log.Field)                          {}
func (a amySpan) FinishWithOptions(opts opentracing.FinishOptions)       {}
func (a amySpan) LogEventWithPayload(event string, payload interface{})  {}
func (a amySpan) SetTag(key string, value interface{}) opentracing.Span  { return a }
func (a amySpan) SetOperationName(operationName string) opentracing.Span { return a }
func (a amySpan) SetBaggageItem(key, val string) opentracing.Span        { return a }
func (a amySpan) BaggageItem(key string) string                          { return "" }
func (a amySpan) Tracer() opentracing.Tracer                             { return defaultAmyTracer }

// StartSpan belongs to the Tracer interface.
func (a amyTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	startOpts := opentracing.StartSpanOptions{}
	for _, opt := range opts {
		opt.Apply(&startOpts)
	}

	var id uint64

	for _, ref := range startOpts.References {
		if ref.Type == opentracing.ChildOfRef {
			id = ref.ReferencedContext.(amySpanContext).id
			break
		}
		if ref.Type == opentracing.FollowsFromRef {
			id = ref.ReferencedContext.(amySpanContext).id
			break
		}
	}

	if id == 0 {
		id = rand.Uint64()
	}

	return amySpan{ctx: amySpanContext{id: id}}
}

// Inject belongs to the Tracer interface.
func (a amyTracer) Inject(sp opentracing.SpanContext, format interface{}, carrier interface{}) error {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		ctx := sp.(amySpanContext)
		traceID := strconv.FormatUint(ctx.id, 16) + ":0" // 需要一个冒号
		carrier.(opentracing.TextMapWriter).Set(traceHeader, traceID)
	}

	return opentracing.ErrUnsupportedFormat
}

// Extract belongs to the Tracer interface.
func (a amyTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	var id uint64

	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		carrier.(opentracing.TextMapReader).ForeachKey(func(key, val string) (err error) {
			if key == traceHeader {
				if i := strings.IndexByte(val, ':'); i > 0 {
					id, err = strconv.ParseUint(val[:i], 10, 64)
				}
			}

			return
		})
	}

	if id > 0 {
		return amySpanContext{id: id}, nil
	}

	return nil, opentracing.ErrSpanContextNotFound
}
