/*
Copyright © 2023 Folderr <contact@folderr.net>
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

var rootCmdName = "foldcli"
var envPrefix = "FOLDCLI_"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   rootCmdName,
	Short: "A CLI to manage Folderr installations",
	Long: `A CLI to setup and manage your Folderr instance. Get started with:
` + rootCmdName + ` init /home/folderr/folderr https://github.com/Folderr/<repo>
` + rootCmdName + ` init`,
	Version: "Alpha 0.0.3",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	// Cleanup for dry-run commands
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if dry && strings.Contains(config.directory, os.TempDir()) {
			// Remove the temp dir
			err := os.RemoveAll(config.directory)
			if err != nil {
				panic(err)
			}
		}
		if dry && strings.Contains(ConfigDir, os.TempDir()) {
			err := os.RemoveAll(ConfigDir)
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

var ConfigDir string

func getToken() string {
	token := os.Getenv(strings.ToLower(envPrefix) + "token")
	if token == "" {
		token = os.Getenv(envPrefix + "TOKEN")
	}
	return token
}

func getConfigDir() (string, error) {
	dir, err := os.UserHomeDir()
	if dry {
		fmt.Println("Using dry-run mode")
	}
	if err != nil && dry {
		println("Error accessing user directory:", err)
		println("This is a warning as you are in dry-run mode")
	} else if err != nil {
		return "", err
	}
	dir = dir + "/.folderr/cli"
	return dir, nil
}

func ReadConfig() (bool, error) {
	viper.SetConfigType("yaml")
	dir, err := getConfigDir()
	if err != nil {
		return false, err
	}
	// config stuffs
	if dry && os.Getenv("test") == "true" {
		dir, err = os.MkdirTemp(os.TempDir(), ".folderr-cli-")
		if err != nil {
			println("Failed to make temp dir for dry-run")
			panic(err)
		}
	}
	viper.AddConfigPath(dir)
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
		err = os.MkdirAll(dir, 0770)
		if err != nil {
			return false, err
		}

		_, err = os.Create(dir + "/config.yaml")
		if err != nil {
			return false, err
		}
		err = viper.WriteConfig()
		if err != nil {
			return false, err
		}
		return false, nil
	}
	// Make a temp dir for tests & dry runs
	if (err != nil && dry) || os.Getenv("test") == "true" {
		dir, err = os.MkdirTemp(os.TempDir(), ".folderr-cli-")
		var runner = os.Getenv("RUNNER_TEMP")
		if len(runner) > 0 && os.Getenv("CI") == "true" {
			dir, err = os.MkdirTemp(runner, ".folderr-cli-")
		}
		if err != nil {
			println("Failed to make temp dir for dry-run")
			panic(err)
		}
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.AddConfigPath(dir)
		ldir, err := os.MkdirTemp(os.TempDir(), "Folderr-")
		if err != nil {
			panic(err)
		}
		// We DO NOT care about any config in dry-run mode.
		config.directory = ldir
		if getToken() != "" {
			config.repository = "https://github.com/Folderr/Folderr"
		} else {
			config.repository = "https://github.com/Folderr/Docs"
		}
		viper.Set("repository", config.repository)
		viper.Set("directory", config.directory)
		err = viper.SafeWriteConfig()
		if err != nil {
			fmt.Println("Tried working with temp directories. No luck.")
			panic(err)
		}
	}
	if getToken() != "" {
		authFlag = getToken()
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
	ConfigDir = dir
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
	rootCmd.SetVersionTemplate("Folderr CLI Version: {{ .Version }}")
	rootCmd.PersistentFlags().BoolVar(&dry, "dry", false, "Runs the command but does not change anything")
	rootCmd.ParseFlags(os.Args)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
