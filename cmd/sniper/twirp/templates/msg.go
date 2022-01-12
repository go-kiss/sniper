package templates

const msgTpl = `
func (m *{{ msgTyp . }}) validate() error {
	if m == nil { return nil }
	
	{{ range .Fields }}
		{{ template "field" . }}
	{{ end }}

	return nil
}

type {{ errname . }} struct {
	field  string
	reason string
}

// Error satisfies the builtin error interface
func (e {{ errname . }}) Error() string {
	return fmt.Sprintf(
		"invalid {{ (msgTyp .) }}.%s: %s",
		e.field,
		e.reason)
}
`
