package rule

const minItemsTpl = `
		if len({{ .Key }}) < {{ .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must contain at least {{ .Value }} item(s)",
			}
		}
`
