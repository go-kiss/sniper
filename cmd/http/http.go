package http

import (
	"net/http"

	"sniper/cmd/http/hooks"

	"github.com/go-kiss/sniper/pkg/twirp"
)

var commonHooks = twirp.ChainHooks(hooks.TraceID, hooks.Log)

func initMux(mux *http.ServeMux) {
}
