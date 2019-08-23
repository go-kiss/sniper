package rpc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"text/template"
)

// 追加的server
var blockStmt *ast.BlockStmt

// 追加的import
var importRPCSpec *ast.ImportSpec
var importServerSpec *ast.ImportSpec

var fset = token.NewFileSet()

const regServerTpl = `
package main
func main(){
	test := "test"
	{
		server := &{{.Server}}server{{.Version}}.Server{}
		handler := {{.Server}}_v{{.Version}}.New{{.UpperServer}}Server(server, {{.Hooks}})
		mux.Handle({{.Server}}_v{{.Version}}.{{.UpperServer}}PathPrefix, handler)
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
	Server      string
	Hooks       string
	Version     string
	UpperServer string
}

type importTplArgs struct {
	PKGName    string
	RPCPath    string
	ServerPath string
}

func parseAndUpdateHTTPFile() {
	addServer := parseFile(httpFile)
	for _, decl := range addServer.Decls {
		gen, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		//在version="0"时只注册内部服务 version!="0"时只注册外部服务
		if gen.Name.Name == "initMux" && version == "0" {
			continue
		}
		if gen.Name.Name == "initInternalMux" && version != "0" {
			continue
		}
		// 判断服务是否已经注册
		if serverRegistered(gen) {
			return
		}
		// 生成头文件模版
		genPKGTemplate()
		// 生成server模版
		genServerTemplate()
		// 追加server
		appendServer(gen)
	}
	// 追加文件头
	appendImportPKGs(addServer)
	f, err := os.OpenFile(httpFile, os.O_WRONLY|os.O_CREATE, 0766)
	if err != nil {
		return
	}
	defer f.Close()
	if err := printer.Fprint(f, fset, addServer); err != nil {
		panic(err)
	}
}

// 构造新加的server的文法
func appendServer(gen *ast.FuncDecl) {
	var serverList []ast.Stmt
	for _, server := range gen.Body.List {
		serverList = append(serverList, server)
	}

	var lastPos token.Pos

	if l := len(gen.Body.List); l > 0 {
		lastBlockPos := gen.Body.List[l-1].(*ast.BlockStmt).Rbrace
		updateBodyPOS(lastBlockPos)
		lastPos = gen.Body.List[l-1].(*ast.BlockStmt).Rbrace
	} else {
		lastPos = gen.Pos()
	}

	updateBodyPOS(lastPos)

	serverList = append(serverList, blockStmt)
	gen.Body.List = serverList
}

// 追加import
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
		if se.X.(*ast.Ident).Name == server+"server"+version {
			return true
		}
	}
	return false
}

func parseFile(file string) *ast.File {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	server, err := parser.ParseFile(fset, "", string(b), parser.ParseComments)
	if err != nil {
		panic(err)
	}
	return server
}

func strFirstToUpper(str string) string {
	var upperStr string
	vv := []rune(str)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				// string的码表相差32位
				vv[i] -= 32
				upperStr += string(vv[i])
			} else {
				fmt.Println("not begins with lowercase letter")
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

func genServerTemplate() {
	args := regServerTplArgs{Server: server, Version: version, UpperServer: strFirstToUpper(server), Hooks: hooks}
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
	for _, decl := range importAst.Decls {
		blockStmt = decl.(*ast.FuncDecl).Body.List[1].(*ast.BlockStmt)
	}
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
