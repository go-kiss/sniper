package http

import (
	"net/http"

	"sniper/pkg/hooks"
	"sniper/pkg/twirp"
)

var commonHooks = twirp.ChainHooks(hooks.TraceID, hooks.Log)

func initMux(mux *http.ServeMux) {
}
