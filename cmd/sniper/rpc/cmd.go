package rpc

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	server, service, version string
)

func init() {
	Cmd.Flags().StringVar(&server, "server", "", "服务")
	Cmd.Flags().StringVar(&service, "service", "", "子服务")
	Cmd.Flags().StringVar(&version, "version", "1", "版本")

	Cmd.MarkFlagRequired("server")
}

// Cmd 接口生成工具
var Cmd = &cobra.Command{
	Use:   "rpc",
	Short: "生成 rpc 接口",
	Long: `脚手架功能：
- 生成 rpc/**/*.proto 模版
- 生成 rpc/**/*.go
- 生成 rpc/**/*.pb.go
- 生成 rpc/**/*.twirp.go
- 注册接口到 http server`,
	Run: func(cmd *cobra.Command, args []string) {
		if isSniperDir() {
			color.Red("只能在 sniper 项目根目录运行!")
			os.Exit(1)
		}

		if service == "" {
			service = server
		}

		genProto()
		genOrUpdateServer()
		registerServer()
	},
}
