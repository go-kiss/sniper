package rpc

import (
	"strings"
)

type tpl interface {
	tpl() string
}

type srvTpl struct {
	Server  string // 服务
	Version string // 版本
	Service string // 子服务
}

func (t *srvTpl) tpl() string {
	return strings.TrimLeft(`
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
`, "\n")
}

type funcTpl struct {
	Service  string // 服务名
	Name     string // 函数名
	ReqType  string // 请求消息类型
	RespType string // 返回消息类型
}

func (t *funcTpl) tpl() string {
	return strings.TrimLeft(`
func (s *{{.Service}}Server) {{.Name}}(ctx context.Context, req *{{.ReqType}}) (resp *{{.RespType}}, err error) {
	{{ if eq .Name  "Echo" }}
	return &{{.Service}}EchoResp{Msg: req.Msg}, nil
	{{ else }}
	// FIXME 请开始你的表演
	return
	{{ end }}
}
`, "\n")
}

type regSrvTpl struct {
	Server  string // 服务
	Version string // 版本
	Service string // 子服务
}

func (t *regSrvTpl) tpl() string {
	return strings.TrimLeft(`
package main
func main() {
	{
		s := &{{.Server}}_v{{.Version}}.{{.Service}}Server{}
		hooks := twirp.ChainHooks(commonHooks, hooks.ServerHooks(s))
		handler := {{.Server}}_v{{.Version}}.New{{.Service}}Server(s,hooks)
		mux.Handle({{.Server}}_v{{.Version}}.{{.Service}}PathPrefix, handler)
	}
}
`, "\n")
}

type impTpl struct {
	Name string
	Path string
}

func (t *impTpl) tpl() string {
	return strings.TrimLeft(`
package main
import(
	{{.Name}} {{.Path}}
)
`, "\n")
}

type protoTpl struct {
	Server  string // 服务
	Version string // 版本
	Service string // 子服务
}

func (t *protoTpl) tpl() string {
	return strings.TrimLeft(`
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
`, "\n")
}
