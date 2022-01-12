package rule

const uniqueTpl = `
	{{ .Field.GoIdent.GoName }}_Unique := make(map[{{ goType .Field.Desc.Kind }}]struct{}, len({{ .Key }}))

	for idx, item := range {{ .Key }} {
		_, _ = idx, item
		if _, exists :={{ .Field.GoIdent.GoName }}_Unique[item]; exists {
			return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
				field:  "{{ .Field.GoName }}",
				reason: "repeated value must contain unique items",
			}
		}
		{{ .Field.GoIdent.GoName }}_Unique[item] = struct{}{}
	}
`
