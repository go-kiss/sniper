package rule

const lteTpl = `
		if {{ .Key }} > {{ .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must less than or equal to {{ escape .Value }}",
			}
		}
`
