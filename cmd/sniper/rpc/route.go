package rpc

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/goast"
	"github.com/dave/dst/decorator/resolver/simple"
)

func serverRegistered(gen *dst.FuncDecl) bool {
	for _, s := range gen.Body.List {
		bs, ok := s.(*dst.BlockStmt)
		if !ok {
			continue
		}
		// s := &foo_v1.FooServer{}
		ue, ok := bs.List[0].(*dst.AssignStmt).Rhs[0].(*dst.UnaryExpr)
		if !ok {
			continue
		}
		i, ok := ue.X.(*dst.CompositeLit).Type.(*dst.Ident)
		if !ok {
			continue
		}
		if !strings.HasSuffix(i.Path, "/"+server+"/v"+version) {
			continue
		}
		if i.Name != upper1st(service)+"Server" {
			continue
		}
		return true
	}
	return false
}

func genServerRoute(initMux *dst.FuncDecl) {
	if serverRegistered(initMux) {
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

	d := decorator.NewDecoratorWithImports(nil, "http", goast.New())
	f, err := d.Parse(buf)
	if err != nil {
		panic(err)
	}

	for _, d := range f.Decls {
		if fd, ok := d.(*dst.FuncDecl); ok {
			stmt := fd.Body.List[0].(*dst.BlockStmt)
			if len(initMux.Body.List) > 0 {
				stmt.Decs.Start.Replace("\n")
			}
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
	d := decorator.NewDecoratorWithImports(nil, "http", goast.New())
	routeAst, err := d.Parse(b)
	if err != nil {
		panic(err)
	}

	// 处理注册路由
	for _, decl := range routeAst.Decls {
		f, ok := decl.(*dst.FuncDecl)
		if ok && f.Name.Name == "initMux" {
			genServerRoute(f)
			break
		}
	}

	f, err := os.OpenFile(routeFile, os.O_WRONLY|os.O_CREATE, 0766)
	if err != nil {
		return
	}
	defer f.Close()

	alias := server + "_v" + version
	path := fmt.Sprintf(`%s/rpc/%s/v%s`, module(), server, version)
	rr := simple.RestorerResolver{path: alias}
	for _, i := range routeAst.Imports {
		alias := ""
		path, _ := strconv.Unquote(i.Path.Value)
		if i.Name != nil {
			alias = i.Name.Name
		} else {
			parts := strings.Split(path, "/")
			alias = parts[len(parts)-1]
		}
		rr[path] = alias
	}
	r := decorator.NewRestorerWithImports("http", rr)
	fr := r.FileRestorer()
	fr.Alias[path] = alias
	if err := fr.Fprint(f, routeAst); err != nil {
		panic(err)
	}
}
