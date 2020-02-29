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
service {{.UpperServer}} {
    // FIXME 接口必须写注释
    //
    // 这里的行尾注释 sniper:foo 有特殊含义，是可选的
    // 框架会将此处冒号后面的值(foo)注入到 ctx 中，
    // 用户可以使用 twirp.MethodOption(ctx) 查询，并执行不同的逻辑
    // 这个 sniper 前缀可以通过 --twirp_out=option_prefix=sniper:. 自定义
    rpc Echo(EchoReq) returns (EchoResp); // sniper:foo
}

message EchoReq {
    // FIXME 请求字段必须写注释
    string msg = 1;
}

message EchoResp {
    // FIXME 响应字段必须写注释
    string msg = 1;
}
`

type protoTplArgs struct {
	Server      string
	Version     string
	UpperServer string
}

func genProto(protoFile string) {
	fd, err := createFile(protoFile)
	if err != nil {
		panic(err)
	}

	args := protoTplArgs{
		Server:      server,
		Version:     version,
		UpperServer: strFirstToUpper(server),
	}

	buf := &bytes.Buffer{}

	tmpl, err := template.New("test").Parse(strings.TrimLeft(protoTpl, "\n"))
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(buf, args)
	if err != nil {
		panic(err)
	}

	_, err = fd.Write(buf.Bytes())
	if err != nil {
		panic(err)
	}
}
