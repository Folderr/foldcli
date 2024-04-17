/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package cmd

import (
	"os"
	"strings"

	"github.com/Folderr/foldcli/utilities"
	"github.com/spf13/cobra"
)

var dry bool

var rootCmdName = utilities.Constants.RootCmdName

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   rootCmdName,
	Short: "A CLI to manage Folderr installations",
	Long: `A CLI to setup and manage your Folderr instance. Get started with:
` + rootCmdName + ` init`,
	Version: "0.0.11",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	// Cleanup for dry-run commands
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		dir, err := utilities.GetConfigDir(dry)
		if err != nil {
			panic(err)
		}
		_, config, _, err := utilities.ReadConfig(dir, dry)
		if err != nil {
			panic(err)
		}
		if dry && strings.Contains(config.Directory, os.TempDir()) {
			// Remove the temp dir
			err := os.RemoveAll(config.Directory)
			if err != nil {
				panic(err)
			}
		}
		if dry && strings.Contains(dir, os.TempDir()) {
			err := os.RemoveAll(dir)
			if err != nil {
				panic(err)
			}
		}
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.PersistentFlags().BoolVar(&dry, "dry", false, "Runs the command but does not change ANYTHING")
	RootCmd.SetVersionTemplate("Folderr CLI (foldcli) version: {{ .Version }}\n")
	RootCmd.PersistentFlags().BoolVar(&dry, "dry", false, "Runs the command but does not change anything")
	RootCmd.ParseFlags(os.Args)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
