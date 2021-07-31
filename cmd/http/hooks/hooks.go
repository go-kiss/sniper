package hooks

import (
	"context"

	"sniper/pkg/twirp"
)

type ctxKeyType int

const (
	sendRespKey ctxKeyType = 0
	spanKey     ctxKeyType = 1
)

type ServerHooker interface {
	Hooks() map[string]*twirp.ServerHooks
}

func ServerHooks(server interface{}) *twirp.ServerHooks {
	hooker, ok := server.(ServerHooker)
	if !ok {
		return nil
	}

	hooks := hooker.Hooks()
	if len(hooks) == 0 {
		return nil
	}

	serverHooks := hooks[""]

	return &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			if serverHooks != nil {
				return serverHooks.CallRequestReceived(ctx)
			}
			return ctx, nil
		},
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			method, _ := twirp.MethodName(ctx)
			if hooks, ok := hooks[method]; ok {
				return hooks.CallRequestRouted(ctx)
			} else if serverHooks != nil {
				return serverHooks.CallRequestRouted(ctx)
			}
			return ctx, nil
		},
		ResponsePrepared: func(ctx context.Context) context.Context {
			method, _ := twirp.MethodName(ctx)
			if hooks, ok := hooks[method]; ok {
				return hooks.CallResponsePrepared(ctx)
			} else if serverHooks != nil {
				return serverHooks.CallResponsePrepared(ctx)
			}
			return ctx
		},
		ResponseSent: func(ctx context.Context) {
			method, _ := twirp.MethodName(ctx)
			if hooks, ok := hooks[method]; ok {
				hooks.CallResponseSent(ctx)
			} else if serverHooks != nil {
				serverHooks.CallResponseSent(ctx)
			}
		},
		Error: func(ctx context.Context, twerr twirp.Error) context.Context {
			method, _ := twirp.MethodName(ctx)
			if hooks, ok := hooks[method]; ok {
				return hooks.CallError(ctx, twerr)
			} else if serverHooks != nil {
				return serverHooks.CallError(ctx, twerr)
			}
			return ctx
		},
	}
}
