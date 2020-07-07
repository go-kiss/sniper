package rpc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"text/template"
)

// 追加的server
var blockStmt *ast.BlockStmt

// 追加的import
var importRPCSpec *ast.ImportSpec
var importServerSpec *ast.ImportSpec

const regServerTpl = `
package main
func main() {
	{
		server := &{{.Server}}server{{.Version}}.{{.Service}}Server{}
		handler := {{.Server}}_v{{.Version}}.New{{.Service}}Server(server, {{.Hooks}})
		mux.Handle({{.Server}}_v{{.Version}}.{{.Service}}PathPrefix, handler)
	}
}
`

const importTpl = `
package main
import(
	{{.PKGName}} {{.RPCPath}}
	{{.ServerPath}}
)
`

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

func serverImported(imports []*ast.ImportSpec) bool {
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

func appendServer(gen *ast.FuncDecl) {
	var lastPos token.Pos

	if l := len(gen.Body.List); l > 0 {
		lastBlockPos := gen.Body.List[l-1].(*ast.BlockStmt).Rbrace
		updateBodyPOS(lastBlockPos)
		lastPos = gen.Body.List[l-1].(*ast.BlockStmt).Rbrace
	} else {
		lastPos = gen.Pos()
	}

	updateBodyPOS(lastPos)

	gen.Body.List = append(gen.Body.List, blockStmt)
}

func appendImportPKGs(file *ast.File) {
	for _, decl := range file.Decls {
		imp, ok := decl.(*ast.GenDecl)
		if !ok || imp.Tok != token.IMPORT {
			continue
		}
		appendImportPKG(imp)
	}
}

func appendImportPKG(pkg *ast.GenDecl) {
	var pkgList []ast.Spec
	for _, spec := range pkg.Specs {
		pkgList = append(pkgList, spec)
	}
	insertPos := pkg.Rparen - 1
	updatePKGPOS(insertPos)
	// 添加import包
	pkgList = append(pkgList, importRPCSpec, importServerSpec)
	pkg.Specs = pkgList
}

// 判断服务是否已经注册
func serverRegistered(gen *ast.FuncDecl) bool {
	for _, writeServer := range gen.Body.List {
		bs, ok := writeServer.(*ast.BlockStmt)
		if !ok {
			continue
		}
		ue, ok := bs.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr)
		if !ok {
			continue
		}
		se, ok := ue.X.(*ast.CompositeLit).Type.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if se.X.(*ast.Ident).Name != server+"server"+version {
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
		Hooks:   hooks,
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
	importAst, err := parser.ParseFile(serverSet, "", buf.String(), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	blockStmt = importAst.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.BlockStmt)
}

func updateBodyPOS(lastPos token.Pos) {
	blockStmt.Lbrace = lastPos + 4
	// 0段元素位置确定
	blockStmt.List[0].(*ast.AssignStmt).TokPos = lastPos + 5
	blockStmt.List[0].(*ast.AssignStmt).Lhs[0].(*ast.Ident).NamePos = lastPos + 5
	blockStmt.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr).OpPos = lastPos + 5
	blockStmt.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.NamePos = lastPos + 5
	blockStmt.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).NamePos = lastPos + 5
	blockStmt.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Lbrace = lastPos + 5
	blockStmt.List[0].(*ast.AssignStmt).Rhs[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Rbrace = lastPos + 5

	// 1段元素位置确定
	blockStmt.List[1].(*ast.AssignStmt).TokPos = lastPos + 6
	blockStmt.List[1].(*ast.AssignStmt).Lhs[0].(*ast.Ident).NamePos = lastPos + 6
	blockStmt.List[1].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.NamePos = lastPos + 6
	blockStmt.List[1].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).NamePos = lastPos + 6
	blockStmt.List[1].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Args[0].(*ast.Ident).NamePos = lastPos + 6
	blockStmt.List[1].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Args[1].(*ast.Ident).NamePos = lastPos + 6
	blockStmt.List[1].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Lparen = lastPos + 6
	blockStmt.List[1].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Rparen = lastPos + 6

	// 2段元素位置确定
	blockStmt.List[2].(*ast.ExprStmt).X.(*ast.CallExpr).Lparen = lastPos + 7
	blockStmt.List[2].(*ast.ExprStmt).X.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).NamePos = lastPos + 7
	blockStmt.List[2].(*ast.ExprStmt).X.(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.NamePos = lastPos + 7
	blockStmt.List[2].(*ast.ExprStmt).X.(*ast.CallExpr).Args[0].(*ast.SelectorExpr).X.(*ast.Ident).NamePos = lastPos + 7
	blockStmt.List[2].(*ast.ExprStmt).X.(*ast.CallExpr).Args[0].(*ast.SelectorExpr).Sel.NamePos = lastPos + 7
	blockStmt.List[2].(*ast.ExprStmt).X.(*ast.CallExpr).Args[1].(*ast.Ident).NamePos = lastPos + 7
	blockStmt.List[2].(*ast.ExprStmt).X.(*ast.CallExpr).Rparen = lastPos + 7

	blockStmt.Rbrace = lastPos + 8
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
	importAst, err := parser.ParseFile(importSet, "", buf.String(), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	for _, decl := range importAst.Decls {
		importRPCSpec = decl.(*ast.GenDecl).Specs[0].(*ast.ImportSpec)
		importServerSpec = decl.(*ast.GenDecl).Specs[1].(*ast.ImportSpec)
	}
}

func updatePKGPOS(pos token.Pos) {
	importRPCSpec.Name.NamePos = pos + 4
	importRPCSpec.Path.ValuePos = pos + 4
	importServerSpec.Path.ValuePos = pos + 6
}
