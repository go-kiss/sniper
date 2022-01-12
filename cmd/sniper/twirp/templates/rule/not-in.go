package rule

const notInTpl = `
		var {{ .Field.GoIdent.GoName }}_NotIn = map[{{ goType .Field.Desc.Kind }}]struct{}{
			{{ range slice .Value }}
				{{ . }}:{},
			{{ end }}
		}

		if _, ok := {{ .Field.GoIdent.GoName }}_NotIn[{{ .Key }}]; ok {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must be not in list {{ escape .Value }}",
			}
		}
`
