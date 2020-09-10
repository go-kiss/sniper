package rule

const suffixTpl = `
		if !strings.HasSuffix({{ .Key }}, {{ .Value }}) {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value does not have suffix {{ escape .Value }}",
			}
		}
`
