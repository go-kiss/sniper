package rule

const typeTpl = `
		{{ if eq  .Value  "url" }}
			if _, err := url.Parse({{ .Key }}); err != nil {
				return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
					field:  "{{ .Field.GoName }}",
					reason: "value must be a valid URL",
				}
			}
		{{ else if eq .Value "ip" }}
			if ip := net.ParseIP({{ .Key }}); ip == nil {
				return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
					field:  "{{ .Field.GoName }}",
					reason: "value must be a valid IP address",
				}
			}
		{{ else if eq .Value "phone" }}
			var {{ .Field.GoIdent.GoName }}_Pattern = regexp.MustCompile("1[3-9]\\d{9}")

			if !{{ .Field.GoIdent.GoName }}_Pattern.MatchString({{ .Key }}){
				return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
					field:  "{{ .Field.GoName }}",
					reason: "value does not match regex pattern  {{ escape .Value }}",
				}
			}
		{{ else if eq .Value "email" }}
			var {{ .Field.GoIdent.GoName }}_Pattern = regexp.MustCompile("[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+")

			if !{{ .Field.GoIdent.GoName }}_Pattern.MatchString({{ .Key }}){
				return {{ .Field.Parent.GoIdent.GoName }}ValidationError {
					field:  "{{ .Field.GoName }}",
					reason: "value does not match regex pattern  {{ escape .Value }}",
				}
			}
		{{ else }}
			// undefined type
		{{ end }}
`
