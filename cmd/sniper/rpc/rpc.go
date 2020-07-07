package rpc

import (
	"bytes"
	"strings"
	"text/template"
)

const protoTpl = `
syntax = "proto3";

package {{.Server}}.v{{.Version}};

option go_package="{{.Server}}_v{{.Version}}";

// FIXME 服务必须写注释
service {{.Service}} {
    // FIXME 接口必须写注释
    //
    // 这里的行尾注释 sniper:foo 有特殊含义，是可选的
    // 框架会将此处冒号后面的值(foo)注入到 ctx 中，
    // 用户可以使用 twirp.MethodOption(ctx) 查询，并执行不同的逻辑
    // 这个 sniper 前缀可以通过 --twirp_out=option_prefix=sniper:. 自定义
    rpc Echo({{.Service}}EchoReq) returns ({{.Service}}EchoResp); // sniper:foo
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
