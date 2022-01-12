package rule

const notContainsTpl = `
		if strings.Contains({{ .Key }}, {{ .Value }}) {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value contains {{ escape .Value }}",
			}
		}
`
