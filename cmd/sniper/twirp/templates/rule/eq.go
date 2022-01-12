package rule

const eqTpl = `
		if {{ .Key }} != {{ .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must equal {{ escape .Value }}",
			}
		}
`
