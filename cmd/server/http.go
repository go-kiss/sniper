package server

import (
	"net/http"

	"sniper/cmd/server/hook"
	"sniper/util/twirp"

)

var hooks = twirp.ChainHooks(
	hook.NewRequestID(),
	hook.NewLog(),
)

func initMux(mux *http.ServeMux, isInternal bool) {
}

func initInternalMux(mux *http.ServeMux) {
}
