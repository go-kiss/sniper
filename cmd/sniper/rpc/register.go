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

// 追加的server
var blockStmt *dst.BlockStmt

// 追加的import
var importRPCSpec *dst.ImportSpec
var importServerSpec *dst.ImportSpec

type regServerTplArgs struct {
	Server  string
	Hooks   string
	Version string
	Service string
}

type importTplArgs struct {
	PKGName    string
	RPCPath    string
	ServerPath string
}

func serverImported(imports []*dst.ImportSpec) bool {
	rpc := server + "_v" + version
	for _, i := range imports {
		if i.Name == nil {
			continue
		}

		if i.Name.Name == rpc {
			return true
		}
	}
	return false
}

func appendServer(gen *dst.FuncDecl) {
	blockStmt.Decs.Start.Replace("\n")
	gen.Body.List = append(gen.Body.List, blockStmt)
}

func appendImportPKGs(file *dst.File) {
	for _, decl := range file.Decls {
		pkg, ok := decl.(*dst.GenDecl)
		if !ok || pkg.Tok != token.IMPORT {
			continue
		}

		pkg.Specs = append(pkg.Specs, importRPCSpec)
	}
}

// 判断服务是否已经注册
func serverRegistered(gen *dst.FuncDecl) bool {
	for _, writeServer := range gen.Body.List {
		bs, ok := writeServer.(*dst.BlockStmt)
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

func genServerTemplate() {
	args := regServerTplArgs{
		Server:  server,
		Version: version,
		Service: upper1st(service),
	}
	tmpl, err := template.New("test").Parse(regServerTpl)
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, args)
	if err != nil {
		panic(err)
	}

	serverSet := token.NewFileSet()
	importAst, err := decorator.ParseFile(serverSet, "", buf.String(), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	blockStmt = importAst.Decls[0].(*dst.FuncDecl).Body.List[0].(*dst.BlockStmt)
}

func genPKGTemplate() {
	importRPC := server + "_v" + version
	importRPCPath := fmt.Sprintf("\"%s/rpc/%s/v%s\"", rootPkg, server, version)
	importServerPath := fmt.Sprintf("\"%s/server/%sserver%s\"", rootPkg, server, version)
	args := importTplArgs{PKGName: importRPC, RPCPath: importRPCPath, ServerPath: importServerPath}
	tmpl, err := template.New("test").Parse(importTpl)
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, args)
	if err != nil {
		panic(err)
	}

	importSet := token.NewFileSet()
	importAst, err := decorator.ParseFile(importSet, "", buf.String(), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	importRPCSpec = importAst.Decls[0].(*dst.GenDecl).Specs[0].(*dst.ImportSpec)
}
