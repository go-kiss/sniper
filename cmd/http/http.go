package http

import (
	"net/http"

	"sniper/cmd/http/hook"
	"sniper/pkg/twirp"
)

var hooks = twirp.ChainHooks(
	hook.NewRequestID(),
	hook.NewLog(),
)

func initMux(mux *http.ServeMux, isInternal bool) {
}

func initInternalMux(mux *http.ServeMux) {
}
