package rule

const gteTpl = `
		if {{ .Key }} < {{ .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must greater than or equal to {{ escape .Value }}",
			}
		}
`
