package rpc

// 几乎所有代码由欧阳完成，我只是搬运过来。

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	// 服务相关变量
	rootDir, rootPkg, server, version string

	twirpFile, serverFile, rpcPkg string

	hooks = "hooks"

	// 要更新的注册文件
	httpFile string

	needLogin bool
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	Cmd.Flags().StringVar(&rootDir, "root", wd, "项目根目录")
	Cmd.Flags().StringVar(&rootPkg, "package", "sniper", "项目总包名")
	Cmd.Flags().StringVar(&server, "service", "", "服务名")
	Cmd.Flags().StringVar(&version, "version", "1", "服务版本")
	Cmd.Flags().BoolVar(&needLogin, "need-login", false, "是否校验登录态")

	Cmd.MarkFlagRequired("service")
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

		genRPC()
		genImplements()
		registerServer()
	},
}

func genRPC() {
	proto := fmt.Sprintf("rpc/%s/v%s/service.proto", server, version)
	protoFile := fmt.Sprintf("%s/%s", rootDir, proto)

	if !fileExists(protoFile) {
		genProto(protoFile)
	}

	// generate twirp
	cmd := exec.Command("protoc", "--twirp_out=.", "--go_out=.", proto)
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func registerServer() {
	httpFile = fmt.Sprintf("%s/cmd/server/http.go", rootDir)
	parseAndUpdateHTTPFile()
}

func genImplements() {
	twirpFile = fmt.Sprintf("%s/rpc/%s/v%s/service.twirp.go", rootDir, server, version)
	serverFile = fmt.Sprintf("%s/server/%sserver%s/server.go", rootDir, server, version)
	rpcPkg = fmt.Sprintf("%s/rpc/%s/v%s", rootPkg, server, version)

	if !fileExists(twirpFile) {
		panic("twirp file does not exist: " + twirpFile)
	}

	genOrUpdateTwirpServer()
}

func createFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return os.Create(path)
}
