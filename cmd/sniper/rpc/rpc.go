package rpc

import (
	"bytes"
	"strings"
	"text/template"
)

const protoTpl = `
syntax = "proto3";

package {{.Server}}.v{{.Version}};

// FIXME 服务必须写注释
service {{.Service}} {
    // FIXME 接口必须写注释
    rpc Echo({{.Service}}EchoReq) returns ({{.Service}}EchoResp);
}

message {{.Service}}EchoReq {
    // FIXME 请求字段必须写注释
    string msg = 1;
}

message {{.Service}}EchoResp {
    // FIXME 响应字段必须写注释
    string msg = 1;
}
`

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
