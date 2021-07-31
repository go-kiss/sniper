package rpc

var serverTpl = `
package {{.Server}}_v{{.Version}}

import (
	"context"
)

type {{.Service}}Server struct{}
`

var funcTpl = `
func (s *{{.Service}}Server) {{.Name}}(ctx context.Context, req *{{.ReqType}}) (resp *{{.RespType}}, err error) {
	// FIXME 请开始你的表演
	return
}
`

var echoFuncTpl = `
func (s *{{.Service}}Server) Echo(ctx context.Context, req *{{.Service}}EchoReq) (resp *{{.Service}}EchoResp, err error) {
	return &{{.Service}}EchoResp{Msg: req.Msg}, nil
}
`

var regServerTpl = `
package main
func main() {
	{
		server := &{{.Server}}_v{{.Version}}.{{.Service}}Server{}
		handler := {{.Server}}_v{{.Version}}.New{{.Service}}Server(server, hooks)
		mux.Handle({{.Server}}_v{{.Version}}.{{.Service}}PathPrefix, handler)
	}
}
`

var importTpl = `
package main
import(
	{{.PKGName}} {{.RPCPath}}
)
`
