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
	has := false
	for i, s := range gen.Body.List {
		bs, ok := s.(*dst.BlockStmt)
		if !ok {
			continue
		}
		// 提取 s := &foo_v1.FooServer{} 的 foo_v1.FooServer
		// 保存到 id 变量
		ue, ok := bs.List[0].(*dst.AssignStmt).Rhs[0].(*dst.UnaryExpr)
		if !ok {
			continue
		}
		// id.Name 保存 FooServer
		// id.Path 保存 sniper/rpc/bar/v1
		id, ok := ue.X.(*dst.CompositeLit).Type.(*dst.Ident)
		if !ok {
			continue
		}

		if !hasProto(id) {
			// 设成 EmptyStmt 表示删除当前 block
			// 也可以直接操作 gen.Body.List
			// 但不如直接赋值方便
			gen.Body.List[i] = &dst.EmptyStmt{Implicit: true}
		}

		if !strings.HasSuffix(id.Path, "/"+server+"/v"+version) {
			continue
		}

		if id.Name != upper1st(service)+"Server" {
			continue
		}

		has = true
	}
	return has
}

func hasProto(id *dst.Ident) bool {
	parts := strings.Split(id.Path, "/")
	proto := strings.ToLower(id.Name[:len(id.Name)-6]) + ".proto"
	proto = strings.Join(parts[1:], "/") + "/" + proto

	return fileExists(proto)
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

	f, err := os.OpenFile(routeFile, os.O_WRONLY|os.O_TRUNC, 0644)
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
