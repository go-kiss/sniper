package main

import (
	start "github.com/go-kiss/sniper/cmd/sniper/init"
	"github.com/go-kiss/sniper/cmd/sniper/rename"
	"github.com/go-kiss/sniper/cmd/sniper/rpc"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(rpc.Cmd)
	Cmd.AddCommand(rename.Cmd)
	Cmd.AddCommand(start.Cmd)
}

// Cmd 脚手架命令
var Cmd = &cobra.Command{
	Use:   "sniper",
	Short: "sniper 脚手架",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	Cmd.Execute()
}
