package server

import (
	"net/http"

	"sniper/cmd/server/hook"

	"github.com/bilibili/twirp"
)

var hooks = twirp.ChainHooks(
	hook.NewRequestID(),
	hook.NewLog(),
)

var loginHooks = twirp.ChainHooks(
	hook.NewRequestID(),
	hook.NewCheckLogin(),
	hook.NewLog(),
)

func initMux(mux *http.ServeMux, isInternal bool) {
}

func initInternalMux(mux *http.ServeMux) {
}
