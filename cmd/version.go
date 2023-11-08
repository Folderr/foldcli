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
	Short: "Print the version of Folderr CLI",
	Long:  "Here are the versions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Folderr CLI version:", rootCmd.Version)
	},
}
