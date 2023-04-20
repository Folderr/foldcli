/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authFlag string
var dry bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "folderr",
	Short: "A CLI to manage Folderr installations",
	Long: `A CLI to setup and manage your Folderr instance For example:
folderr init /home/folderr/folderr
folderr init
folderr install
folderr setup`,
	Version: "Alpha 0.0.1",
	// Cleanup for dry-run commands
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if dry && strings.Contains(config.directory, os.TempDir()) {
			// Remove the temp dir
			err := os.RemoveAll(config.directory)
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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type Config struct {
	directory  string
	repository string
	canInstall bool
}

var config Config = Config{}

func ReadConfig() (bool, error) {
	viper.SetConfigType("yaml")
	dir, err := os.UserHomeDir()
	if err != nil && dry {
		println("Error accessing user directory:", err)
		println("This is a warning as you are in dry-run mode")
	} else if err != nil {
		return false, err
	}
	viper.AddConfigPath(dir + "/.folderr/cli")
	err = viper.ReadInConfig()
	// If in dry-run mode we don't care if there is a config or not.
	// The config will NEVER be modified.
	// These tests are still ran for warning purposes.
	if err != nil && dry {
		println("Warning: Your config is not usable.")
		println("Notice: No changes as in dry-run mode.")
		println("Here's the error:", err)
	} else if err != nil && !strings.Contains(err.Error(), "Not Found") {
		return false, err
	} else if err != nil {
		err = os.MkdirAll(dir+"/.folderr/cli", 0660)
		if err != nil {
			return false, err
		}

		_, err = os.Create(dir + "/.folderr/cli/config.yaml")
		if err != nil {
			return false, err
		}
		err = viper.WriteConfig()
		if err != nil {
			return false, err
		}
		return false, nil
	}
	if viper.IsSet("repository") {
		config.repository = viper.GetString("repository")
	}
	if viper.IsSet("directory") {
		config.directory = viper.GetString("directory")
	}
	if config.directory != "" && config.repository != "" {
		config.canInstall = true
	}
	if dry {
		println("Setting custom stuff due to dry-run mode")
		viper.Reset()
		dir, err := os.MkdirTemp(os.TempDir(), ".folderr-cli-")
		if err != nil {
			panic(err)
		}
		viper.AddConfigPath(dir)
		dir, err = os.MkdirTemp(os.TempDir(), "Folderr-")
		if err != nil {
			panic(err)
		}
		// We DO NOT care about any config in dry-run mode.
		config.directory = dir
		config.repository = "https://github.com/Folderr/Docs"
	}
	return true, nil
}

func println(a ...any) {
	fmt.Fprintln(rootCmd.OutOrStdout(), a...)
}

func printf(format string, a ...any) {
	fmt.Fprintf(rootCmd.OutOrStdout(), format, a...)
}

func init() {
	// rootCmd.PersistentFlags().BoolVar(&dry, "dry", false, "Runs the command but does not change ANYTHING")
	rootCmd.PersistentFlags().BoolVar(&dry, "dry", false, "Runs the command but does not change anything")
	rootCmd.ParseFlags(os.Args)
	_, err := ReadConfig()
	if err != nil {
		panic(err)
	}
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}