package main

import (
	"github.com/go-kiss/sniper/cmd/sniper/new"
	"github.com/go-kiss/sniper/cmd/sniper/rpc"
	"github.com/spf13/cobra"
)

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
	Cmd.AddCommand(rpc.Cmd)
	Cmd.AddCommand(new.Cmd)
	Cmd.Execute()
}
