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

var serverFile string

func genOrUpdateServer() {
	serverFile = fmt.Sprintf("rpc/%s/v%s/%s.go", server, version, service)

	if !fileExists(serverFile) {
		tpl := &srvTpl{
			Server:  server,
			Version: version,
			Service: upper1st(service),
		}

		save(serverFile, tpl)
	}

	p := fmt.Sprintf("rpc/%s/v%s/%s.twirp.go", server, version, service)
	b, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}
	f, err := decorator.Parse(b)
	if err != nil {
		panic(err)
	}

	for _, d := range f.Decls {
		gd, ok := d.(*dst.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		it, ok := gd.Specs[0].(*dst.TypeSpec).Type.(*dst.InterfaceType)
		if !ok {
			continue
		}

		appendFuncs(it, f.Imports)
		updateComments(it)

		return // 只处理第一个服务
	}
}

func updateComments(twirp *dst.InterfaceType) {
	comments := getComments(twirp)

	b, err := os.ReadFile(serverFile)
	if err != nil {
		return
	}
	f, err := decorator.Parse(b)
	if err != nil {
		return
	}

	decls := make([]dst.Decl, 0, len(f.Decls))
	for _, decl := range f.Decls {
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
	f.Decls = decls

	fb, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	decorator.Fprint(fb, f)
}

func getComments(d *dst.InterfaceType) map[string]dst.Decorations {
	comments := make(map[string]dst.Decorations)
	// rpc service注释单独添加
	if d.Decs.Interface != nil {
		comments[upper1st(service)] = d.Decs.Interface
	}

	for _, method := range d.Methods.List {
		name := method.Names[0].Name
		if name == "Do" || name == "ServiceDescriptor" ||
			name == "ProtocGenTwirpVersion" {
			continue
		}

		comments[name] = method.Decs.Start
	}

	return comments
}

func appendFuncs(twirp *dst.InterfaceType, imports []*dst.ImportSpec) {
	local := server + "_v" + version

	d := decorator.NewDecoratorWithImports(nil, local, goast.New())
	b, err := os.ReadFile(serverFile)
	if err != nil {
		panic(err)
	}
	server, err := d.Parse(b)
	if err != nil {
		panic(err)
	}
	definedFuncs := scanDefinedFuncs(server)

	buf := &bytes.Buffer{}
	buf.WriteString("package main\n")

	rr := simple.RestorerResolver{}
	for _, i := range imports {
		name := i.Name.Name
		path, _ := strconv.Unquote(i.Path.Value)
		// twirp 文件导入的包都有别名
		fmt.Fprintf(buf, "import %s \"%s\"\n", name, path)
		rr[path] = name
	}

	for _, m := range twirp.Methods.List {
		name := m.Names[0].Name

		if name == "Do" || name == "ServiceDescriptor" ||
			name == "ProtocGenTwirpVersion" {
			continue
		}

		if definedFuncs[name] {
			continue
		}

		ft := m.Type.(*dst.FuncType)

		in := ft.Params.List[1].Type.(*dst.StarExpr).X
		out := ft.Results.List[0].Type.(*dst.StarExpr).X

		appendFunc(buf, name, getType(in), getType(out))
	}

	if buf.Len() == 0 {
		return
	}

	ff, err := d.Parse(buf.Bytes())
	if err != nil {
		panic(err)
	}

	for _, d := range ff.Decls {
		if v, ok := d.(*dst.FuncDecl); ok {
			server.Decls = append(server.Decls, v)
		}
	}

	f, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := decorator.NewRestorerWithImports(local, rr)
	if err := r.Fprint(f, server); err != nil {
		panic(err)
	}
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

func scanDefinedFuncs(file *dst.File) map[string]bool {
	fs := make(map[string]bool)

	for _, decl := range file.Decls {
		if f, ok := decl.(*dst.FuncDecl); ok {
			fs[f.Name.Name] = true
		}
	}

	return fs
}
