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
    rpc Echo(EchoReq) returns (EchoResp);
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
