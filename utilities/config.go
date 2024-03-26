package utilities

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Directory  string
	Repository string
	CanInstall bool
}

type SecretConfig struct {
	GitToken string
}

func GetGitToken() string {
	token := os.Getenv(strings.ToLower(Constants.EnvPrefix) + "git_token")
	if token == "" {
		token = os.Getenv(Constants.EnvPrefix + "GIT_TOKEN")
	}
	return token
}

func GetConfigDir(dry bool) (string, error) {
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

func ReadConfig(directory string, dry bool) (bool, Config, SecretConfig, error) {
	viper.SetConfigType("yaml")
	// config stuffs
	dir := directory
	var err error
	if dry && os.Getenv("test") == "true" {
		dir, err = os.MkdirTemp(os.TempDir(), ".foldcli-")
		if err != nil {
			println("Failed to make temp dir for dry-run")
			panic(err)
		}
	}
	viper.AddConfigPath(dir)
	err = viper.ReadInConfig()
	config := Config{}
	secrets := SecretConfig{}
	// If in dry-run mode we don't care if there is a config or not.
	// The config will NEVER be modified.
	// These tests are still ran for warning purposes.
	if err != nil && dry {
		println("Warning: Your config is not usable.")
		println("Notice: No changes as in dry-run mode.")
		println("Here's the error:", err)
	} else if err != nil && !strings.Contains(err.Error(), "Not Found") {
		return false, config, secrets, err
	} else if err != nil {
		err = os.MkdirAll(dir, 0770)
		if err != nil {
			return false, config, secrets, err
		}

		_, err = os.Create(dir + "/config.yaml")
		if err != nil {
			return false, config, secrets, err
		}
		err = viper.WriteConfig()
		if err != nil {
			return false, config, secrets, err
		}
		return false, Config{}, secrets, nil
	}

	// Make a temp dir for tests & dry runs
	if (err != nil && dry) || os.Getenv("test") == "true" || os.Getenv("CI") == "true" {
		var tempdir = os.TempDir()
		var runner = os.Getenv("RUNNER_TEMP")
		println(runner)
		if len(runner) > 0 {
			tempdir = runner
		}
		dir, err = os.MkdirTemp(tempdir, ".foldcli-")
		if err != nil {
			println("Failed to make temp dir for dry-run")
			panic(err)
		}
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.AddConfigPath(dir)
		ldir, err := os.MkdirTemp(tempdir, "Folderr-")
		if err != nil {
			panic(err)
		}
		// We DO NOT care about any config in dry-run mode.
		config.Directory = ldir
		if GetGitToken() != "" {
			config.Repository = "https://github.com/Folderr/Folderr"
		} else {
			config.Repository = "https://github.com/Folderr/Docs"
		}
		viper.Set("repository", config.Repository)
		viper.Set("directory", config.Directory)
		err = viper.SafeWriteConfig()
		if err != nil {
			fmt.Println("Tried working with temp directories. No luck.")
			panic(err)
		}
	}
	if GetGitToken() != "" {
		secrets.GitToken = GetGitToken()
	}
	if viper.IsSet("repository") {
		config.Repository = viper.GetString("repository")
	}
	if viper.IsSet("directory") {
		config.Directory = viper.GetString("directory")
	}
	if config.Directory != "" && config.Repository != "" {
		config.CanInstall = true
	}
	return true, config, secrets, nil
}
