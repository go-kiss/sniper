package rpc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const serverTpl = `
package {{.ServerPkg}}

import (
	"context"

	pb "{{.RPCPkg}}"
)

// Server {{.Comment}}
type Server struct{}
`

const funcTpl = `
// {{.Name}} {{.Comment}}
func (s *Server) {{.Name}}(ctx context.Context, req *pb.{{.ReqType}}) (resp *pb.{{.RespType}}, err error) {
	// FIXME 请开始你的表演
	return
}
`

const echoFuncTpl = `
// {{.Name}} {{.Comment}}
func (s *Server) Echo(ctx context.Context, req *pb.EchoReq) (resp *pb.EchoResp, err error) {
	return &pb.EchoResp{Msg: req.Msg}, nil
}
`

type serverTplArgs struct {
	RPCPkg    string
	Comment   string
	ServerPkg string
}

type funcTplArgs struct {
	Comment  string
	Name     string
	ReqType  string
	RespType string
}

func genOrUpdateTwirpServer() {
	if !fileExists(serverFile) {
		genServerFile()
	}

	twirp := parseFileByAst(twirpFile)
	for _, decl := range twirp.Decls {
		if tree, ok := decl.(*ast.GenDecl); ok && tree.Tok == token.TYPE {
			appendFuncs(tree)
			break // 只处理一个文件
		}
	}
}

func appendFuncs(twirp *ast.GenDecl) {
	buf := &bytes.Buffer{}

	definedFuncs := scanDefinedFuncs()

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

			if _, ok := definedFuncs[name]; ok {
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
	args := funcTplArgs{Name: method.Names[0].Name}

	ft := method.Type.(*ast.FuncType)

	// 写死函数签名
	args.ReqType = ft.Params.List[1].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	args.RespType = ft.Results.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	args.Comment = fmt.Sprintf(
		"实现 /twirp/%s.v%s.%s/%s 接口",
		server,
		version,
		strFirstToUpper(server),
		args.Name,
	)

	tpl := funcTpl
	if args.Name == "Echo" {
		tpl = echoFuncTpl
	}

	tmpl, err := template.New("test").Parse(tpl)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(buf, args)
	if err != nil {
		panic(err)
	}

}

func scanDefinedFuncs() map[string]bool {
	parseServer := parseFileByAst(serverFile)
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
	fd, err := createFile(serverFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	serverPkg := filepath.Base(filepath.Dir(serverFile))

	comment := fmt.Sprintf(
		"实现 /twirp/%s.v%s.%s 服务",
		server,
		version,
		strFirstToUpper(server),
	)

	args := serverTplArgs{
		RPCPkg:    rpcPkg,
		Comment:   comment,
		ServerPkg: serverPkg,
	}

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

func parseFileByAst(file string) *ast.File {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	parseServer, err := parser.ParseFile(fset, "", string(b), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	return parseServer
}
