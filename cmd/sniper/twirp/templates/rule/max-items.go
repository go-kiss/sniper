package rule

const maxItemsTpl = `
		if len({{ .Key }}) > {{ .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must contain at most {{ .Value }} item(s)",
			}
		}
`
