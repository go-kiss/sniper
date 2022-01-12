package rule

const minLenTpl = `
		if utf8.RuneCountInString({{ .Key }})  < {{ .Value }}{
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value length must be at least  {{ .Value }}  runes",
			}
		}
`
