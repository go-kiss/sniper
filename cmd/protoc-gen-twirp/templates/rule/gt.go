package rule

const gtTpl = `
		if {{ .Key }} <= {{ .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must greater than {{ escape .Value }}",
			}
		}
`
