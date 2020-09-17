package templates

const fieldTpl = `
	{{ range validate . }}
		{{ . }}
	{{ end }}

	{{ message . }}
`
