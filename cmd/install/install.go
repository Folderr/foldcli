package install

import (
	"github.com/Folderr/foldcli/cmd"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:       "install",
	Short:     "Base command for installing Folderr projects",
	Long:      "Base command for installing Folderr projects",
	ValidArgs: []string{"directory", "repository"},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
