package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wpengine/lostromos/version"
)

func init() {
	LostromosCmd.AddCommand(commandDefintion)
}

var commandDefintion = &cobra.Command{
	Use:   "version",
	Short: `Show the version number.`,
	Run: func(command *cobra.Command, args []string) {
		version.Print()
	},
}
