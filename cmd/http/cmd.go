package http

import (
	"github.com/spf13/cobra"
)

var port int

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
}
