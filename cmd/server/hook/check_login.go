package hook

import (
	"context"

	"sniper/util/ctxkit"
	"sniper/util/errors"

	"github.com/bilibili/twirp"
)

// NewCheckLogin 检查用户登录态，未登录直接报错返回
func NewCheckLogin() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			if ctxkit.GetUserID(ctx) == 0 {
				return ctx, errors.NotLoginError
			}

			return ctx, nil
		},
	}
}
