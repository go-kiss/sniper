package rpc

const (
	serverTpl = `
package {{.Server}}_v{{.Version}}

import (
	"context"

	"github.com/go-kiss/sniper/pkg/twirp"
)

type {{.Service}}Server struct{}

// Hooks 返回 server 和 method 对应的 hooks
// 如果设定了 method 的 hooks，则不再执行 server 一级的 hooks
func (s *{{.Service}}Server) Hooks() map[string]*twirp.ServerHooks {
	return map[string]*twirp.ServerHooks {
		// "": nil, // Server 一级 hooks
		// "Echo": nil, // Echo 方法的 hooks
	}
}
`

	funcTpl = `
func (s *{{.Service}}Server) {{.Name}}(ctx context.Context, req *{{.ReqType}}) (resp *{{.RespType}}, err error) {
	// FIXME 请开始你的表演
	return
}
`

	echoFuncTpl = `
func (s *{{.Service}}Server) Echo(ctx context.Context, req *{{.Service}}EchoReq) (resp *{{.Service}}EchoResp, err error) {
	return &{{.Service}}EchoResp{Msg: req.Msg}, nil
}
`

	regServerTpl = `
package main
func main() {
	{
		server := &{{.Server}}_v{{.Version}}.{{.Service}}Server{}
		hooks := twirp.ChainHooks(commonHooks, hooks.ServerHooks(server))
		handler := {{.Server}}_v{{.Version}}.New{{.Service}}Server(server, hooks)
		mux.Handle({{.Server}}_v{{.Version}}.{{.Service}}PathPrefix, handler)
	}
}
`

	importTpl = `
package main
import(
	{{.PKGName}} {{.RPCPath}}
)
`

	protoTpl = `
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
)
