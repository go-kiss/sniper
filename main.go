package main

import (
	_ "net/http/pprof" // 注册 pprof 接口

	"sniper/cmd/cron"
	"sniper/cmd/http"

	"github.com/spf13/cobra"
)

func main() {
	root := cobra.Command{Use: "sniper"}

	root.AddCommand(
		cron.Cmd,
		http.Cmd,
	)

	root.Execute()
}
