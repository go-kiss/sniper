package http

import (
	"github.com/spf13/cobra"
)

var port int
var isInternal bool
var isManage bool

// Cmd run http server
var Cmd = &cobra.Command{
	Use:   "http",
	Short: "Run http server",
	Long:  `Run http server`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}

func init() {
	Cmd.Flags().IntVar(&port, "port", 8080, "listen port")
	Cmd.Flags().BoolVar(&isInternal, "internal", false, "internal service")
	Cmd.Flags().BoolVar(&isManage, "manage", false, "manage service")
}
