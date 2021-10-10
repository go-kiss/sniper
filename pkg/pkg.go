package pkg

import (
	_ "github.com/go-kiss/sniper/pkg/conf" // init conf
	_ "github.com/go-kiss/sniper/pkg/http" // init http

	"github.com/go-kiss/sniper/pkg/log"
)

// Reset all utils
func Reset() {
	log.Reset()
}

// Stop all utils
func Stop() {
}
