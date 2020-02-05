package hook

import (
	"context"
	"time"

	"sniper/util/conf"
	"sniper/util/ctxkit"
	"sniper/util/log"
	"sniper/util/metrics"

	"github.com/bilibili/twirp"
	"github.com/opentracing/opentracing-go"
)

type bizResponse interface {
	GetCode() int32
	GetMsg() string
}

type ctxKeyType int

const (
	sendRespKey ctxKeyType = iota
)

// NewLog 统一记录请求日志
func NewLog() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		ResponsePrepared: func(ctx context.Context) context.Context {
			span, ctx := opentracing.StartSpanFromContext(ctx, "SendResp")
			ctx = context.WithValue(ctx, sendRespKey, span)
			return ctx
		},
		ResponseSent: func(ctx context.Context) {
			if span, ok := ctx.Value(sendRespKey).(opentracing.Span); ok {
				defer span.Finish()
			}

			span, ctx := opentracing.StartSpanFromContext(ctx, "LogReq")
			defer span.Finish()

			status, _ := twirp.StatusCode(ctx)
			req, _ := twirp.Request(ctx)
			resp, _ := twirp.Response(ctx)

			var bizCode int32
			var bizMsg string
			if br, ok := resp.(bizResponse); ok {
				bizCode = br.GetCode()
				bizMsg = br.GetMsg()
			}

			start := ctx.Value(ctxkit.StartTimeKey).(time.Time)
			duration := time.Since(start)

			if _, ok := ctx.Deadline(); ok {
				if ctx.Err() != nil {
					status = "503"
				}
			}

			path := req.URL.Path

			// 外部爬接口脚本会请求任意 API
			// 导致 prometheus 无法展示数据
			if status != "404" {
				metrics.RPCDurationsSeconds.WithLabelValues(
					path,
					status,
				).Observe(duration.Seconds())
			}

			form := req.Form
			// 移除日志中的敏感信息
			if conf.IsProdEnv {
				form.Del("access_key")
				form.Del("appkey")
				form.Del("sign")
			}

			log.Get(ctx).WithFields(log.Fields{
				"path":     path,
				"status":   status,
				"params":   form.Encode(),
				"cost":     duration.Seconds(),
				"biz_code": bizCode,
				"biz_msg":  bizMsg,
			}).Info("new rpc")
		},
		Error: func(ctx context.Context, err twirp.Error) context.Context {
			c := twirp.ServerHTTPStatusFromErrorCode(err.Code())

			if c >= 500 {
				log.Get(ctx).Errorf("%+v", cause(err))
			} else if c >= 400 {
				log.Get(ctx).Warn(err)
			}

			return ctx
		},
	}
}

func cause(err twirp.Error) error {
	// https://github.com/pkg/errors#retrieving-the-cause-of-an-error
	type causer interface {
		Cause() error
	}
	if c, ok := err.(causer); ok {
		return c.Cause()
	}

	return err
}
