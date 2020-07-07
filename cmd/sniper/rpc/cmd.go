package rpc

// 几乎所有代码由欧阳完成，我只是搬运过来。

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// 服务相关变量
	rootDir, rootPkg, server, service, version string

	twirpFile, serverFile, rpcPkg string

	hooks = "hooks"

	needLogin bool
)

func init() {
	wd, _ := os.Getwd()
	module := getModuleName(wd)

	Cmd.Flags().StringVar(&rootDir, "root", wd, "项目根目录")
	Cmd.Flags().StringVar(&rootPkg, "package", module, "项目总包名")
	Cmd.Flags().StringVar(&server, "server", "", "服务包名")
	Cmd.Flags().StringVar(&service, "service", "", "子服务名")
	Cmd.Flags().StringVar(&version, "version", "1", "服务版本")
	Cmd.Flags().BoolVar(&needLogin, "need-login", false, "是否校验登录态")

	Cmd.MarkFlagRequired("server")
}

func getModuleName(wd string) (module string) {
	f, err := os.Open(wd + "/go.mod")
	if err != nil {
		return
	}
	defer f.Close()

	l, err := bufio.NewReader(f).ReadString('\n')
	if err != nil {
		panic(err)
	}
	fields := strings.Fields(l)

	module = "sniper"
	if len(fields) == 2 {
		module = fields[1]
	}

	return module
}

// Cmd 接口生成工具
var Cmd = &cobra.Command{
	Use:   "rpc",
	Short: "生成 rpc 接口",
	Long: `脚手架功能：
- 生成 rpc/**/*.proto 模版
- 生成 server/**/*.go 代码
- 注册接口到 http server`,
	Run: func(cmd *cobra.Command, args []string) {
		if needLogin {
			hooks = "loginHooks"
		}

		if service == "" {
			service = server
		}

		genRPC()
		genImplements()
		registerServer()
	},
}

func genRPC() {
	proto := fmt.Sprintf("rpc/%s/v%s/%s.proto", server, version, service)
	protoFile := fmt.Sprintf("%s/%s", rootDir, proto)

	if !fileExists(protoFile) {
		genProto(protoFile)
	}

	cmd := exec.Command("protoc", "--twirp_out=.", "--go_out=.", proto)
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func registerServer() {
	httpFile := fmt.Sprintf("%s/cmd/server/http.go", rootDir)
	httpAST, fset := parseAST(httpFile)

	for _, decl := range httpAST.Decls {
		gen, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if gen.Name.Name == "initMux" && version == "0" {
			continue
		}
		if gen.Name.Name == "initInternalMux" && version != "0" {
			continue
		}

		if serverRegistered(gen) {
			return
		}

		if !serverImported(httpAST.Imports) {
			genPKGTemplate()
			appendImportPKGs(httpAST)
		}

		genServerTemplate()
		appendServer(gen)
	}

	f, err := os.OpenFile(httpFile, os.O_WRONLY|os.O_CREATE, 0766)
	if err != nil {
		return
	}
	defer f.Close()
	if err := printer.Fprint(f, fset, httpAST); err != nil {
		panic(err)
	}
}

func genImplements() {
	twirpFile = fmt.Sprintf("%s/rpc/%s/v%s/%s.twirp.go", rootDir, server, version, service)
	serverFile = fmt.Sprintf("%s/server/%sserver%s/%s.go", rootDir, server, version, service)
	rpcPkg = fmt.Sprintf("%s/rpc/%s/v%s", rootPkg, server, version)

	if !fileExists(twirpFile) {
		panic("twirp file does not exist: " + twirpFile)
	}

	genOrUpdateTwirpServer()
}

func createDirAndFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return os.Create(path)
}

func parseAST(file string) (*ast.File, *token.FileSet) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	a, err := parser.ParseFile(fset, "", string(b), parser.ParseComments)
	if err != nil {
		panic(err)
	}

	return a, fset
}

func upper1st(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)

	if r[0] >= 97 && r[0] <= 122 {
		r[0] -= 32 // 大小写字母ASCII值相差32位
	}

	return string(r)
}
