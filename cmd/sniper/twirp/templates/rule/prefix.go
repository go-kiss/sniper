package rule

const prefixTpl = `
		if !strings.HasPrefix({{ .Key }}, {{ .Value }}) {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value does not have prefix {{ escape .Value }}",
			}
		}
`
