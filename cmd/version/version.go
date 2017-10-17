package version

import (
	"github.com/spf13/cobra"
	"github.com/wpengine/lostromos/cmd"
)

func init() {
	cmd.RootCmd.AddCommand(commandDefintion)
}

var commandDefintion = &cobra.Command{
	Use:   "version",
	Short: `Show the version number.`,
	Run: func(command *cobra.Command, args []string) {
		cmd.ShowVersion()
	},
}
