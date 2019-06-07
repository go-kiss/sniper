package falsework

import (
	"sniper/cmd/falsework/rpc"

	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(rpc.Cmd)
}

// Cmd 脚手架命令
var Cmd = &cobra.Command{
	Use:   "falsework",
	Short: "脚手架",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
