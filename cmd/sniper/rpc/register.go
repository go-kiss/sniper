package rpc

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
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
		if ok && pkg.Tok == token.IMPORT {
			pkg.Specs = append(pkg.Specs, spec)
			return
		}
	}
}
