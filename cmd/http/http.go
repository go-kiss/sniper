package http

import (
	"net/http"

	"sniper/cmd/http/hooks"
	foo_v1 "sniper/rpc/foo/v1"

	"github.com/go-kiss/sniper/pkg/twirp"
)

var commonHooks = twirp.ChainHooks(hooks.TraceID, hooks.Log)

func initMux(mux *http.ServeMux) {
	{
		s := &foo_v1.FooServer{}
		hooks := twirp.ChainHooks(commonHooks, hooks.ServerHooks(s))
		handler := foo_v1.NewFooServer(s, hooks)
		mux.Handle(foo_v1.FooPathPrefix, handler)
	}
}
