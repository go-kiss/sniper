package rule

const ltTpl = `
		if {{ .Key }} >= {{ .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must less than {{ escape .Value }}",
			}
		}
`
