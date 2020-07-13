package rpc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

const serverTpl = `
package {{.ServerPkg}}

import (
	"context"

	pb "{{.RPCPkg}}"
)

type {{.Service}}Server struct{}
`

const funcTpl = `
func (s *{{.Service}}Server) {{.Name}}(ctx context.Context, req *pb.{{.ReqType}}) (resp *pb.{{.RespType}}, err error) {
	// FIXME 请开始你的表演
	return
}
`

const echoFuncTpl = `
func (s *{{.Service}}Server) Echo(ctx context.Context, req *pb.{{.Service}}EchoReq) (resp *pb.{{.Service}}EchoResp, err error) {
	return &pb.{{.Service}}EchoResp{Msg: req.Msg}, nil
}
`

func genOrUpdateTwirpServer() {
	if !fileExists(serverFile) {
		genServerFile()
	}

	twirp, _ := parseAST(twirpFile)
	for _, decl := range twirp.Decls {
		if tree, ok := decl.(*ast.GenDecl); ok && tree.Tok == token.TYPE {
			// 补充函数
			appendFuncs(tree)
			// 更新注释
			updateRPCComment(tree)
			break // 只处理一个文件
		}
	}
}

func updateRPCComment(twirp *ast.GenDecl) {
	comments := getRPCComments(twirp)

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, serverFile, nil, parser.ParseComments)
	if err != nil {
		return
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *dst.GenDecl: // 处理 server struct 注释
			if d.Tok != token.TYPE {
				continue
			}
			ts, ok := d.Specs[0].(*dst.TypeSpec)
			if !ok || ts.Name.Name != upper1st(service)+"Server" {
				continue
			}

			api := fmt.Sprintf(
				"%sServer 实现 /twirp/%s.v%s.%s 服务",
				upper1st(service),
				server,
				version,
				upper1st(service),
			)
			if comment, ok := comments[upper1st(service)]; ok {
				d.Decs.Start.Replace("// " + api + "\n")
				for _, c := range comment.List {
					d.Decs.Start.Append(c.Text)
				}
			}
		case *dst.FuncDecl: // 函数处理注释
			api := fmt.Sprintf(
				"%s 实现 /twirp/%s.v%s.%s/%s 接口",
				d.Name.Name,
				server,
				version,
				upper1st(service),
				d.Name.Name,
			)

			if comment, ok := comments[d.Name.Name]; ok {
				d.Decs.Start.Replace("// " + api + "\n")
				for _, c := range comment.List {
					d.Decs.Start.Append(c.Text)
				}
			}
		}
	}

	fb, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	decorator.Fprint(fb, f)
}

func getRPCComments(twirp *ast.GenDecl) (comments map[string]*ast.CommentGroup) {
	comments = make(map[string]*ast.CommentGroup)
	// rpc service注释单独添加
	if twirp.Doc != nil {
		comments[upper1st(service)] = twirp.Doc
	}
	for _, s := range twirp.Specs {
		ts, ok := s.(*ast.TypeSpec)
		if !ok {
			continue
		}

		it, ok := ts.Type.(*ast.InterfaceType)
		if !ok {
			continue
		}

		for _, method := range it.Methods.List {
			name := method.Names[0].Name
			if name == "Do" || name == "ServiceDescriptor" || name == "ProtocGenTwirpVersion" {
				continue
			}

			if method.Doc != nil {
				comments[name] = method.Doc
			}
		}
	}

	return
}

func appendFuncs(twirp *ast.GenDecl) {
	buf := &bytes.Buffer{}

	definedFuncs := scanDefinedFuncs(serverFile)

	for _, s := range twirp.Specs {
		ts, ok := s.(*ast.TypeSpec)
		if !ok {
			continue
		}

		it, ok := ts.Type.(*ast.InterfaceType)
		if !ok {
			continue
		}

		for _, method := range it.Methods.List {
			name := method.Names[0].Name

			if name == "Do" || name == "ServiceDescriptor" || name == "ProtocGenTwirpVersion" {
				continue
			}

			if definedFuncs[name] {
				continue
			}

			appendFunc(buf, method)
		}
	}

	if buf.Len() == 0 {
		return
	}

	f, err := os.OpenFile(serverFile, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(buf.Bytes())

	if err != nil {
		panic(err)
	}
}

func appendFunc(buf *bytes.Buffer, method *ast.Field) {
	args := struct {
		Name     string
		ReqType  string
		RespType string
		Service  string
	}{}

	args.Name = method.Names[0].Name

	ft := method.Type.(*ast.FuncType)
	// 写死函数签名
	args.ReqType = ft.Params.List[1].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	args.RespType = ft.Results.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	args.Service = upper1st(service)

	tpl := funcTpl
	if args.Name == "Echo" {
		tpl = echoFuncTpl
	}

	tmpl, err := template.New("server").Parse(tpl)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(buf, args)
	if err != nil {
		panic(err)
	}

}

func scanDefinedFuncs(file string) map[string]bool {
	parseServer, _ := parseAST(file)
	definedFuncs := make(map[string]bool)
	for _, decl := range parseServer.Decls {
		if fundel, ok := decl.(*ast.FuncDecl); ok {
			definedFuncs[fundel.Name.Name] = true
		}
	}

	return definedFuncs
}

// 判断文件是否存在
func fileExists(file string) bool {
	fd, err := os.Open(file)
	defer fd.Close()

	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func genServerFile() {
	fd, err := createDirAndFile(serverFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	serverPkg := filepath.Base(filepath.Dir(serverFile))

	args := struct {
		RPCPkg    string
		ServerPkg string
		Service   string
	}{rpcPkg, serverPkg, upper1st(service)}

	buf := &bytes.Buffer{}

	tmpl, err := template.New("test").Parse(strings.TrimLeft(serverTpl, "\n"))
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
