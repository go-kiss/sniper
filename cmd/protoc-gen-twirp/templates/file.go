package templates

const fileTpl = `
package {{ pkg . }}

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ensure the imports are used
var (
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
)

{{ range .Messages }}
	{{ template "msg" . }}
{{ end }}
`
