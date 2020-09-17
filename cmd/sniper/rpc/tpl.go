package rpc

var serverTpl = `
package {{.ServerPkg}}

import (
	"context"

	pb "{{.RPCPkg}}"
)

type {{.Service}}Server struct{}
`

var funcTpl = `
func (s *{{.Service}}Server) {{.Name}}(ctx context.Context, req *pb.{{.ReqType}}) (resp *pb.{{.RespType}}, err error) {
	// FIXME 请开始你的表演
	return
}
`

var echoFuncTpl = `
func (s *{{.Service}}Server) Echo(ctx context.Context, req *pb.{{.Service}}EchoReq) (resp *pb.{{.Service}}EchoResp, err error) {
	return &pb.{{.Service}}EchoResp{Msg: req.Msg}, nil
}
`

var regServerTpl = `
package main
func main() {
	{
		server := &{{.Server}}server{{.Version}}.{{.Service}}Server{}
		handler := {{.Server}}_v{{.Version}}.New{{.Service}}Server(server, hooks)
		mux.Handle({{.Server}}_v{{.Version}}.{{.Service}}PathPrefix, handler)
	}
}
`

var importTpl = `
package main
import(
	{{.PKGName}} {{.RPCPath}}
	{{.ServerPath}}
)
`

func initNewTpl() {
	serverTpl = `
package {{.Server}}_v{{.Version}}

import (
	"context"
)

type {{.Service}}Server struct{}
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
}
