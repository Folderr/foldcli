package utilities

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Directory  string   `json:"directory"`
	Repository string   `json:"repository"`
	CanInstall bool     `json:"CanInstall"`
	Database   DBConfig `json:"db" mapstructure:"db"`
}

type DBConfig struct {
	DbName string `json:"dbName"`
	Url    string `json:"url"`
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
	if dry && os.Getenv("test") == "true" {
		dir, err = os.MkdirTemp(os.TempDir(), ".foldcli-")
		if err != nil {
			println("Failed to make temp dir for dry-run")
			panic(err)
		}
	}
	if os.Getenv("test") == "true" || os.Getenv("CI") == "true" {
		var tempdir = os.TempDir()
		var runner = os.Getenv("RUNNER_TEMP")
		if len(runner) > 0 {
			tempdir = runner
		}
		dir, err = os.MkdirTemp(tempdir, ".foldcli-")
		if err != nil {
			println("Failed to make temp dir for dry-run")
			panic(err)
		}
	}
	return dir, nil
}

func ReadConfig(directory string, dry bool) (*viper.Viper, Config, SecretConfig, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	// config stuffs
	dir := directory
	var err error
	v.AddConfigPath(dir)
	err = v.ReadInConfig()
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
		return v, config, secrets, err
	} else if err != nil {
		err = os.MkdirAll(dir, 0770)
		if err != nil {
			return v, config, secrets, err
		}

		_, err = os.Create(dir + "/config.yaml")
		if err != nil {
			return v, config, secrets, err
		}
		err = v.WriteConfig()
		if err != nil {
			return v, config, secrets, err
		}
		return v, Config{}, secrets, nil
	}

	err = v.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	if dry {
		if dbUrl := os.Getenv("DB_URI"); dbUrl != "" {
			config.Database.Url = dbUrl
			v.Set("db.url", config.Database.Url)
		}
	}

	if os.Getenv(Constants.EnvPrefix+"DB_NAME") != "" {
		config.Database.DbName = os.Getenv(Constants.EnvPrefix + "DB_NAME")
	}

	// Make a temp dir for tests & dry runs
	if os.Getenv("test") == "true" || os.Getenv("CI") == "true" {
		var tempdir = os.TempDir()
		var runner = os.Getenv("RUNNER_TEMP")
		if len(runner) > 0 {
			tempdir = runner
		}
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
		if dbUrl := os.Getenv(Constants.EnvPrefix + "DB_URI"); dbUrl != "" {
			config.Database.Url = dbUrl
			v.Set("db.url", config.Database.Url)
		}
		v.Set("repository", config.Repository)
		v.Set("directory", config.Directory)
		v.Set("db.dbName", config.Database.DbName)
		err = v.SafeWriteConfig()
		if err != nil {
			fmt.Println("Tried working with temp directories. No luck.")
			panic(err)
		}
	}

	if GetGitToken() != "" {
		secrets.GitToken = GetGitToken()
	}
	if config.Directory != "" && config.Repository != "" {
		config.CanInstall = true
	}
	return v, config, secrets, nil
}
