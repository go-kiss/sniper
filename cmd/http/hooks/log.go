package hooks

import (
	"context"

	"github.com/go-kiss/sniper/pkg/conf"
	"github.com/go-kiss/sniper/pkg/log"
	"github.com/go-kiss/sniper/pkg/trace"
	"github.com/go-kiss/sniper/pkg/twirp"
	"github.com/opentracing/opentracing-go"
)

type bizResponse interface {
	GetCode() int32
	GetMsg() string
}

var Log = &twirp.ServerHooks{
	ResponseSent: func(ctx context.Context) {
		var bizCode int32
		var bizMsg string
		resp, _ := twirp.Response(ctx)
		if br, ok := resp.(bizResponse); ok {
			bizCode = br.GetCode()
			bizMsg = br.GetMsg()
		}

		span := opentracing.SpanFromContext(ctx)
		duration := trace.GetDuration(span)

		status, _ := twirp.StatusCode(ctx)
		if _, ok := ctx.Deadline(); ok {
			if ctx.Err() != nil {
				status = "503"
			}
		}

		hreq, _ := twirp.HttpRequest(ctx)
		path := hreq.URL.Path

		// 外部爬接口脚本会请求任意 API
		// 导致 prometheus 无法展示数据
		if status != "404" {
			rpcDurations.WithLabelValues(
				path,
				status,
			).Observe(duration.Seconds())
		}

		form := hreq.Form
		// 新版本采用json/protobuf形式，公共参数需要读取query
		if len(form) == 0 {
			form = hreq.URL.Query()
		}
		// 移除日志中的敏感信息
		if conf.IsProdEnv {
			form.Del("access_key")
			form.Del("appkey")
			form.Del("sign")
		}

		log.Get(ctx).WithFields(log.Fields{
			"ip":       hreq.RemoteAddr,
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
			log.Get(ctx).Errorf("%+v", err)
		} else if c >= 400 {
			log.Get(ctx).Warn(err)
		}

		return ctx
	},
}
