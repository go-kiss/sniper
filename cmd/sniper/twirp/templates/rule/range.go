package rule

const rangeTpl = `
		if {{ rangeRule .Key .Value }} {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must in range {{ escape .Value }}",
			}
		}
`
