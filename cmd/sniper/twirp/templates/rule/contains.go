package rule

const containsTpl = `
		if !strings.Contains({{ .Key }}, {{ .Value }}) {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value not contains {{ escape .Value }}",
			}
		}
`
