package rpc

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"strings"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func serverRegistered(gen *dst.FuncDecl, imports *dst.GenDecl) bool {
	has := false
	deleted := map[int]bool{}
	servers := map[string]bool{}
	for i, s := range gen.Body.List {
		bs, ok := s.(*dst.BlockStmt)
		if !ok {
			continue
		}
		// 提取 s := &foo_v1.FooServer{} 的 foo_v1.FooServer
		ue := bs.List[0].(*dst.AssignStmt).Rhs[0].(*dst.UnaryExpr)
		// s.X 保存 foo_v1
		// s.Sel 保存 FooServer
		s := ue.X.(*dst.CompositeLit).Type.(*dst.SelectorExpr)

		if !hasProto(s) {
			deleted[i] = true
		} else {
			servers[s.X.(*dst.Ident).Name] = true
		}

		if !strings.HasSuffix(s.X.(*dst.Ident).Name, server+"_v"+version) {
			continue
		}

		if s.Sel.Name != upper1st(service)+"Server" {
			continue
		}

		has = true
	}

	stmts := []dst.Stmt{}
	for i, s := range gen.Body.List {
		if !deleted[i] {
			stmts = append(stmts, s)
		}
	}
	gen.Body.List = stmts

	specs := []dst.Spec{}
	for _, s := range imports.Specs {
		i := s.(*dst.ImportSpec)
		if strings.HasPrefix(i.Path.Value, `"`+module()+"/rpc") {
			if !servers[i.Name.Name] {
				continue
			}
		}
		specs = append(specs, s)
	}

	imports.Specs = specs

	return has
}

func hasProto(id *dst.SelectorExpr) bool {
	parts := strings.Split(id.X.(*dst.Ident).Name, "_")
	proto := strings.ToLower(id.Sel.Name[:len(id.Sel.Name)-6]) + ".proto"
	proto = "rpc/" + strings.Join(parts, "/") + "/" + proto

	return fileExists(proto)
}

func genServerRoute(initMux *dst.FuncDecl, imports *dst.GenDecl) {
	if serverRegistered(initMux, imports) {
		return
	}

	args := &regSrvTpl{
		Package: module(),
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

	f, err := decorator.Parse(buf)
	if err != nil {
		panic(err)
	}

	for _, d := range f.Decls {
		if fd, ok := d.(*dst.FuncDecl); ok {
			stmt := fd.Body.List[0].(*dst.BlockStmt)
			initMux.Body.List = append(initMux.Body.List, stmt)
			return
		}
	}
}

func registerServer() {
	routeFile := "cmd/http/http.go"
	b, err := os.ReadFile(routeFile)
	if err != nil {
		panic(err)
	}
	routeAst, err := decorator.Parse(b)
	if err != nil {
		panic(err)
	}

	var imports *dst.GenDecl
	for _, d := range routeAst.Decls {
		gd, ok := d.(*dst.GenDecl)
		if ok && gd.Tok == token.IMPORT {
			imports = gd
			break
		}
	}

	// 处理注册路由
	for _, decl := range routeAst.Decls {
		f, ok := decl.(*dst.FuncDecl)
		if ok && f.Name.Name == "initMux" {
			genServerRoute(f, imports)
			break
		}
	}

	f, err := os.OpenFile(routeFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	alias := server + "_v" + version
	path := fmt.Sprintf(`"%s/rpc/%s/v%s"`, module(), server, version)

	addImport(imports, alias, path)

	if err := decorator.Fprint(f, routeAst); err != nil {
		panic(err)
	}
}

func addImport(imports *dst.GenDecl, alias, path string) {
	var n int
	var is dst.ImportSpec
	// 找到倒数第一个 rpc 导入
	for i := len(imports.Specs) - 1; i >= 0; i-- {
		s := imports.Specs[i].(*dst.ImportSpec)
		if strings.HasPrefix(s.Path.Value, "\""+module()) {
			// 确保没有重复导入
			for j := i; j >= 0; j-- {
				s := imports.Specs[j].(*dst.ImportSpec)
				if s.Path.Value == path {
					return
				}
			}
			// 未导入，准备构造 ImportSepc
			is = *s
			n = i
			break
		}
	}

	is.Name = dst.NewIdent(alias)
	is.Path = &dst.BasicLit{Kind: token.STRING, Value: path}

	// 将新的 import 语句插入到 n 指向位置后面
	ss := make([]dst.Spec, 0, len(imports.Specs)+1)
	ss = append(ss, imports.Specs[:n+1]...)
	ss = append(ss, &is)
	ss = append(ss, imports.Specs[n+1:]...)
	imports.Specs = ss
}
