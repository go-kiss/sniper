package templates

import (
	"text/template"
)

// Register 注册模版
func Register(tpl *template.Template) {
	template.Must(tpl.New("field").Parse(fieldTpl))
	template.Must(tpl.New("msg").Parse(msgTpl))
	template.Must(tpl.Parse(fileTpl))
}
