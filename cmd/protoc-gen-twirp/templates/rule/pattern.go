package rule

const patternTpl = `
		var {{ .Field.GoIdent.GoName }}_Pattern = regexp.MustCompile({{ .Value }})

		if !{{ .Field.GoIdent.GoName }}_Pattern.MatchString({{ .Key }}){
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value does not match regex pattern  {{ escape .Value }}",
			}
		}
`
