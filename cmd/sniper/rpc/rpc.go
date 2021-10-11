package rpc

import (
	"bytes"
	"strings"
	"text/template"
)

func genProto(protoFile string) {
	tpl := strings.TrimLeft(protoTpl, "\n")
	tmpl, err := template.New("proto").Parse(tpl)
	if err != nil {
		panic(err)
	}

	args := struct {
		Server  string
		Version string
		Service string
	}{server, version, upper1st(service)}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, args); err != nil {
		panic(err)
	}

	fd, err := createDirAndFile(protoFile)
	if err != nil {
		panic(err)
	}

	if _, err := fd.Write(buf.Bytes()); err != nil {
		panic(err)
	}
}
