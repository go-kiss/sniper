package rpc

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"strconv"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/goast"
	"github.com/dave/dst/decorator/resolver/simple"
)

func genOrUpdateServer() {
	serverPkg := server + "_v" + version
	serverPath := fmt.Sprintf("rpc/%s/v%s/%s.go", server, version, service)
	twirpPath := fmt.Sprintf("rpc/%s/v%s/%s.twirp.go", server, version, service)

	serverAst := parseServerAst(serverPath, serverPkg)
	twirpAst := parseServerAst(twirpPath, serverPkg)

	for _, d := range twirpAst.Decls {
		if it, ok := isInterfaceType(d); ok {
			imports := twirpAst.Imports

			appendFuncs(serverAst, it, imports)
			updateComments(serverAst, it)

			saveCode(serverAst, imports, serverPath, serverPkg)

			return // 只处理第一个服务
		}
	}
}

func isInterfaceType(d dst.Decl) (*dst.InterfaceType, bool) {
	gd, ok := d.(*dst.GenDecl)
	if !ok || gd.Tok != token.TYPE {
		return nil, false
	}

	it, ok := gd.Specs[0].(*dst.TypeSpec).Type.(*dst.InterfaceType)
	if !ok {
		return nil, false
	}

	return it, true
}

func parseServerAst(path, pkg string) *dst.File {
	d := decorator.NewDecoratorWithImports(nil, pkg, goast.New())
	ast, err := d.Parse(readCode(path))
	if err != nil {
		panic(err)
	}
	return ast
}

func parseTwirpAst(path string) *dst.File {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	ast, err := decorator.Parse(b)
	if err != nil {
		panic(err)
	}

	return ast
}

func readCode(serverFile string) []byte {
	var code []byte
	if fileExists(serverFile) {
		var err error
		code, err = os.ReadFile(serverFile)
		if err != nil {
			panic(err)
		}
	} else {
		t := &srvTpl{
			Server:  server,
			Version: version,
			Service: upper1st(service),
		}

		buf := &bytes.Buffer{}

		tmpl, err := template.New("sniper").Parse(t.tpl())
		if err != nil {
			panic(err)
		}

		if err := tmpl.Execute(buf, t); err != nil {
			panic(err)
		}
		code = buf.Bytes()
	}
	return code
}

func saveCode(ast *dst.File, imports []*dst.ImportSpec, file, pkg string) {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rr := simple.RestorerResolver{}
	for _, i := range imports {
		alias := i.Name.Name
		path, _ := strconv.Unquote(i.Path.Value)
		rr[path] = alias
	}
	r := decorator.NewRestorerWithImports(pkg, rr)
	if err := r.Fprint(f, ast); err != nil {
		panic(err)
	}
}

func updateComments(serverAst *dst.File, twirp *dst.InterfaceType) {
	comments := getComments(twirp)

	decls := make([]dst.Decl, 0, len(serverAst.Decls))
	for _, decl := range serverAst.Decls {
		decls = append(decls, decl)

		switch d := decl.(type) {
		case *dst.GenDecl: // 服务注释
			if d.Tok != token.TYPE {
				continue
			}
			ts, ok := d.Specs[0].(*dst.TypeSpec)
			if !ok || ts.Name.Name != upper1st(service)+"Server" {
				continue
			}

			api := fmt.Sprintf(
				"%sServer 实现 /%s.v%s.%s 服务",
				upper1st(service),
				server,
				version,
				upper1st(service),
			)
			if c := comments[upper1st(service)]; c != nil {
				d.Decs.Start.Replace("// " + api + "\n")
				d.Decs.Start.Append(c...)
			}
		case *dst.FuncDecl: // 函数注释
			api := fmt.Sprintf(
				"%s 实现 /%s.v%s.%s/%s 接口",
				d.Name.Name,
				server,
				version,
				upper1st(service),
				d.Name.Name,
			)

			if c, ok := comments[d.Name.Name]; c != nil {
				d.Decs.Start.Replace("// " + api + "\n")
				d.Decs.Start.Append(c...)
			} else if !ok {
				if d.Recv != nil && d.Name.IsExported() && d.Name.Name != "Hooks" {
					// 删除 proto 中不存在的方法
					st, ok := d.Recv.List[0].Type.(*dst.StarExpr)
					if ok {
						x, ok := st.X.(*dst.Ident)
						if ok && x.Name == upper1st(service)+"Server" {
							decls = decls[:len(decls)-1]
						}
					}
				}
			}
		}
	}
	serverAst.Decls = decls
}

func getComments(d *dst.InterfaceType) map[string]dst.Decorations {
	comments := map[string]dst.Decorations{}
	// rpc service注释单独添加
	comments[upper1st(service)] = d.Decs.Interface

	for _, method := range d.Methods.List {
		name := method.Names[0].Name

		if isTwirpFunc(name) {
			continue
		}

		comments[name] = method.Decs.Start
	}

	return comments
}

func appendFuncs(serverAst *dst.File, twirp *dst.InterfaceType, imports []*dst.ImportSpec) {
	buf := &bytes.Buffer{}
	buf.WriteString("package main\n")

	for _, i := range imports {
		alias := i.Name.Name
		// twirp 文件导入的包都有别名
		fmt.Fprintf(buf, "import %s %s\n", alias, i.Path.Value)
	}

	definedFuncs := scanDefinedFuncs(serverAst)

	for _, m := range twirp.Methods.List {
		name := m.Names[0].Name

		if isTwirpFunc(name) {
			continue
		}

		ft := m.Type.(*dst.FuncType)

		// 接口定义没有指定参数名
		ft.Params.List[0].Names = []*dst.Ident{{Name: "ctx"}}
		ft.Params.List[1].Names = []*dst.Ident{{Name: "req"}}
		ft.Results.List[0].Names = []*dst.Ident{{Name: "resp"}}
		ft.Results.List[1].Names = []*dst.Ident{{Name: "err"}}

		if f, ok := definedFuncs[name]; ok {
			f.Type = ft
			continue
		}

		in := ft.Params.List[1].Type.(*dst.StarExpr).X
		out := ft.Results.List[0].Type.(*dst.StarExpr).X

		appendFunc(buf, name, getType(in), getType(out))
	}

	pkg := server + "_v" + version
	d := decorator.NewDecoratorWithImports(nil, pkg, goast.New())
	f, err := d.Parse(buf.Bytes())
	if err != nil {
		panic(err)
	}

	for _, d := range f.Decls {
		if v, ok := d.(*dst.FuncDecl); ok {
			name := v.Name.Name
			if _, ok := definedFuncs[name]; !ok {
				serverAst.Decls = append(serverAst.Decls, v)
			}
		}
	}

	for _, d := range serverAst.Decls {
		if v, ok := d.(*dst.FuncDecl); ok {
			name := v.Name.Name
			if f, ok := definedFuncs[name]; ok {
				v.Type = f.Type
			}
		}
	}
}

func isTwirpFunc(name string) bool {
	return name == "Do" || name == "ServiceDescriptor" ||
		name == "ProtocGenTwirpVersion"
}

func getType(e dst.Expr) string {
	switch v := e.(type) {
	case *dst.Ident:
		return v.Name
	case *dst.SelectorExpr:
		return v.X.(*dst.Ident).Name + "." + v.Sel.Name
	}
	return ""
}

func appendFunc(buf *bytes.Buffer, name, reqType, respType string) {
	args := &funcTpl{
		Name:     name,
		ReqType:  reqType,
		RespType: respType,
		Service:  upper1st(service),
	}

	t, err := template.New("server").Parse(args.tpl())
	if err != nil {
		panic(err)
	}

	if err := t.Execute(buf, args); err != nil {
		panic(err)
	}
}

func scanDefinedFuncs(file *dst.File) map[string]*dst.FuncDecl {
	fs := make(map[string]*dst.FuncDecl)

	for _, decl := range file.Decls {
		if f, ok := decl.(*dst.FuncDecl); ok {
			fs[f.Name.Name] = f
		}
	}

	return fs
}
