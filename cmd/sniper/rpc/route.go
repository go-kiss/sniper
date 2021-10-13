package rpc

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func serverImported(imports []*dst.ImportSpec) bool {
	rpc := server + "_v" + version
	for _, i := range imports {
		if i.Name != nil && i.Name.Name == rpc {
			return true
		}
	}
	return false
}

func serverRegistered(gen *dst.FuncDecl) bool {
	for _, s := range gen.Body.List {
		bs, ok := s.(*dst.BlockStmt)
		if !ok {
			continue
		}
		ue, ok := bs.List[0].(*dst.AssignStmt).Rhs[0].(*dst.UnaryExpr)
		if !ok {
			continue
		}
		se, ok := ue.X.(*dst.CompositeLit).Type.(*dst.SelectorExpr)
		if !ok {
			continue
		}
		if se.X.(*dst.Ident).Name != server+"_v"+version {
			continue
		}
		if se.Sel.Name != upper1st(service)+"Server" {
			continue
		}
		return true
	}
	return false
}

func genServerRoute(fd *dst.FuncDecl) {
	if serverRegistered(fd) {
		return
	}

	args := &regSrvTpl{
		Server:  server,
		Version: version,
		Service: upper1st(service),
	}
	t, err := template.New("sniper").Parse(args.tpl())
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, args); err != nil {
		panic(err)
	}

	s := token.NewFileSet()
	f, err := decorator.ParseFile(s, "", buf.String(), parser.ParseComments)
	if err != nil {
		panic(err)
	}

	stmt := f.Decls[0].(*dst.FuncDecl).Body.List[0].(*dst.BlockStmt)
	if len(fd.Body.List) > 0 {
		stmt.Decs.Start.Replace("\n")
	}
	fd.Body.List = append(fd.Body.List, stmt)
}

func genImport(file *dst.File) {
	if serverImported(file.Imports) {
		return
	}

	args := impTpl{
		Name: server + "_v" + version,
		Path: fmt.Sprintf(`"%s/rpc/%s/v%s"`, module(), server, version),
	}
	t, err := template.New("sniper").Parse(args.tpl())
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, args); err != nil {
		panic(err)
	}

	s := token.NewFileSet()
	f, err := decorator.ParseFile(s, "", buf.String(), parser.ParseComments)
	if err != nil {
		panic(err)
	}

	spec := f.Decls[0].(*dst.GenDecl).Specs[0].(*dst.ImportSpec)
	for _, decl := range file.Decls {
		pkg, ok := decl.(*dst.GenDecl)
		// http.go 导包分三组
		//
		// "net/http"
		//
		// "sniper/cmd/http/hooks"
		//
		// "github.com/go-kiss/sniper/pkg/twirp"
		//
		// 下面代码 rpc 包导入语句插入到上面第二组中
		if ok && pkg.Tok == token.IMPORT {
			i := 0 // 记录第二组最后一行的位置
			for _, s := range pkg.Specs {
				i++
				is := s.(*dst.ImportSpec)
				// 第二组的包都以项目包名开头
				if !strings.Contains(is.Path.Value, module()+"/") {
					continue
				}
				// 最后一行的 After 为 EmtyLine，表示下面是空行
				if is.Decs.After != dst.EmptyLine {
					continue
				}
				// 注册新路由后需要清理倒数第二行后面的空行
				is.Decs.After = dst.NewLine
				break
			}
			pkg.Specs = append(pkg.Specs[:i+1], pkg.Specs[i:]...)
			pkg.Specs[i] = spec
			return
		}
	}
}

func registerServer() {
	httpFile := "cmd/http/http.go"
	fset := token.NewFileSet()
	httpAST, err := decorator.ParseFile(fset, httpFile, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	genImport(httpAST)

	// 处理注册路由
	for _, decl := range httpAST.Decls {
		f, ok := decl.(*dst.FuncDecl)
		if ok && f.Name.Name == "initMux" {
			genServerRoute(f)
		}
	}

	f, err := os.OpenFile(httpFile, os.O_WRONLY|os.O_CREATE, 0766)
	if err != nil {
		return
	}
	defer f.Close()
	if err := decorator.Fprint(f, httpAST); err != nil {
		panic(err)
	}
}
