package rule

const inTpl = `
		var {{ .Field.GoIdent.GoName }}_In = map[{{ goType .Field.Desc.Kind }}]struct{}{
			{{ range slice .Value }}
				{{ . }}:{},
			{{ end }}
		}

		if _, ok := {{ .Field.GoIdent.GoName }}_In[{{ .Key }}]; !ok {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "value must be in list {{ escape .Value }}",
			}
		}
`
