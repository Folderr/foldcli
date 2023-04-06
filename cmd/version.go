package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(version)
}

var version = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Folderr Manage",
	Long:  "Here are the versions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Folderr Manager version: Alpha 0.0.1")
	},
}
