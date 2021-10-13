package rpc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/ast/astutil"
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
	f, fs := parseAST(p, nil)
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}

		appendFuncs(gd, f, fs)
		updateComments(gd)

		return // 只处理第一个服务
	}
}

func updateComments(d *ast.GenDecl) {
	comments := getComments(d)

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, serverFile, nil, parser.ParseComments)
	if err != nil {
		return
	}

	decls := make([]dst.Decl, 0, len(f.Decls))
	for _, decl := range f.Decls {
		decls = append(decls, decl)

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
				"%sServer 实现 /%s.v%s.%s 服务",
				upper1st(service),
				server,
				version,
				upper1st(service),
			)
			if comment := comments[upper1st(service)]; comment != nil {
				d.Decs.Start.Replace("// " + api + "\n")
				for _, c := range comment.List {
					d.Decs.Start.Append(c.Text)
				}
			}
		case *dst.FuncDecl: // 函数处理注释
			api := fmt.Sprintf(
				"%s 实现 /%s.v%s.%s/%s 接口",
				d.Name.Name,
				server,
				version,
				upper1st(service),
				d.Name.Name,
			)

			if comment, ok := comments[d.Name.Name]; comment != nil {
				d.Decs.Start.Replace("// " + api + "\n")
				for _, c := range comment.List {
					d.Decs.Start.Append(c.Text)
				}
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

func getComments(d *ast.GenDecl) map[string]*ast.CommentGroup {
	comments := make(map[string]*ast.CommentGroup)
	// rpc service注释单独添加
	if d.Doc != nil {
		comments[upper1st(service)] = d.Doc
	}

	for _, s := range d.Specs {
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
			if name == "Do" || name == "ServiceDescriptor" ||
				name == "ProtocGenTwirpVersion" {
				continue
			}

			comments[name] = method.Doc
		}
	}

	return comments
}

func addImport(t string, imported []*ast.ImportSpec, f *ast.File, fs *token.FileSet) {
	var name, path string
	for _, i := range imported {
		path, _ = strconv.Unquote(i.Path.Value)
		if i.Name != nil {
			name = i.Name.Name
			if strings.HasPrefix(t, name+".") {
				break
			}
		} else {
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
			if strings.HasPrefix(t, name+".") {
				break
			}
		}
	}
	astutil.AddNamedImport(fs, f, name, path)
}

func appendFuncs(d *ast.GenDecl, f *ast.File, fs *token.FileSet) {
	buf := &bytes.Buffer{}

	definedFuncs := scanDefinedFuncs(serverFile)
	sf, fs := parseAST(serverFile, nil)

	for _, s := range d.Specs {
		ts, ok := s.(*ast.TypeSpec)
		if !ok {
			continue
		}

		it, ok := ts.Type.(*ast.InterfaceType)
		if !ok {
			continue
		}

		for _, m := range it.Methods.List {
			name := m.Names[0].Name
			if name == "Do" || name == "ServiceDescriptor" ||
				name == "ProtocGenTwirpVersion" {
				continue
			}

			if definedFuncs[name] {
				continue
			}

			ft := m.Type.(*ast.FuncType)

			in := ft.Params.List[1].Type.(*ast.StarExpr).X
			reqType, ok := getName(in)
			if ok {
				addImport(reqType, f.Imports, sf, fs)
			}

			out := ft.Results.List[0].Type.(*ast.StarExpr).X
			respType, ok := getName(out)
			if ok {
				addImport(respType, f.Imports, sf, fs)
			}

			fname := m.Names[0].Name
			appendFunc(buf, fname, reqType, respType)
		}
	}

	if buf.Len() == 0 {
		return
	}

	ff, _ := parseAST("", buf.Bytes())
	sf.Decls = append(sf.Decls, ff.Decls...)

	of, err := os.OpenFile(serverFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer of.Close()

	if err := printer.Fprint(of, fs, sf); err != nil {
		panic(err)
	}
}

func getName(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name, false
	case *ast.SelectorExpr:
		return v.X.(*ast.Ident).Name + "." + v.Sel.Name, true
	}
	return "", false
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

func scanDefinedFuncs(file string) map[string]bool {
	r, _ := parseAST(file, nil)
	fs := make(map[string]bool)
	for _, decl := range r.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok {
			fs[f.Name.Name] = true
		}
	}

	return fs
}
