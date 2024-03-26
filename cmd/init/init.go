package init

import (
	"os"

	"github.com/Folderr/foldcli/cmd"
	"github.com/Folderr/foldcli/utilities"
	"github.com/spf13/cobra"
)

func init() {
	cmd.RootCmd.AddCommand(initCmd)
}

func dirChecks(input string) bool {
	isValid := utilities.IsValidPath(input)
	if !isValid {
		println("That is NOT a valid directory!")
		os.Exit(1)
	}
	return utilities.CheckIfDirExists(input)
}

var initCmd = &cobra.Command{
	Use:       "init",
	Short:     "Dummy command for initializing different CLI modules",
	Long:      "Dummy command for initializing different CLI modules",
	ValidArgs: []string{"directory", "repository"},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
