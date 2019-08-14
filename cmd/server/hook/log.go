package hook

import (
	"context"
	"time"

	"sniper/util/ctxkit"
	"sniper/util/log"
	"sniper/util/metrics"

	"github.com/bilibili/twirp"
)

type bizResponse interface {
	GetCode() int32
	GetMsg() string
}

// NewLog 统一记录请求日志
func NewLog() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		ResponseSent: func(ctx context.Context) {
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

			metrics.RPCDurationsSeconds.WithLabelValues(
				path,
				status,
			).Observe(duration.Seconds())

			log.Get(ctx).WithFields(log.Fields{
				"path":     path,
				"status":   status,
				"params":   req.Form.Encode(),
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
