package main

import (
	"sniper/cmd/falsework"
	"sniper/cmd/job"
	"sniper/cmd/server"

	"github.com/spf13/cobra"
)

func main() {
	root := cobra.Command{Use: "sniper"}

	root.AddCommand(
		falsework.Cmd,
		server.Cmd,
		job.Cmd,
	)

	root.Execute()
}
