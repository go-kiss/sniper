package main

import (
	_ "net/http/pprof" // 注册 pprof 接口

	"sniper/cmd/job"
	"sniper/cmd/server"

	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs" // 根据容器配额设置 maxprocs
)

func main() {
	root := cobra.Command{Use: "sniper"}

	root.AddCommand(
		server.Cmd,
		job.Cmd,
	)

	root.Execute()
}
