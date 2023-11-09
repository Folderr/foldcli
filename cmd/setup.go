/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Base command for setting up various Folderr apps",
	Long: `Base command for setting up various Folderr apps
DRY RUN MODE NOT IMPLEMENTED
DO NOT RUN IN DRY RUN MODE, IT WILL CHANGE THINGS`,
}

func init() {
	rootCmd.AddCommand(setupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
