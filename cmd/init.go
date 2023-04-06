package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func readConfig() {
	viper.SetConfigType("yaml")
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	viper.AddConfigPath(dir + "/.folderr/cli")
	err = viper.ReadInConfig()
	if err != nil && !strings.Contains(err.Error(), "Not Found") {
		panic(err)
	} else if err != nil {
		err = os.MkdirAll(dir+"/.folderr/cli", 0660)
		if err != nil {
			panic(err)
		}

		_, err = os.Create(dir + "/.folderr/cli/config.yaml")
		if err != nil {
			panic(err)
		}
		err = viper.WriteConfig()
		if err != nil {
			panic(err)
		}
	}
}

var mkdir, override bool
var authFlag string

func init() {
	initCmd.Flags().BoolVar(&mkdir, "mkdir", false, "Make directories if they don't exist")
	initCmd.Flags().BoolVarP(&override, "override", "o", false, "Override previous settings")
	initCmd.Flags().StringVarP(&authFlag, "authorization", "a", "", "Authorization token for private repositories")
	rootCmd.AddCommand(initCmd)
	readConfig()
}

func checkifExists(input string) bool {
	if _, err := os.Stat(input); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return false
	}
}

func isValidPath(input string) bool {
	if runtime.GOOS == "windows" { // Having to handle this because the NT kernel.
		if len(strings.SplitN(input, ":", 2)) < 1 {
			return false
		}
		if !strings.Contains(input, "/") && !strings.Contains(input, "\\\\") && !strings.Contains(input, "\\") {
			return false
		}
		return true
	}
	if strings.Contains(input, "/") {
		return true
	}
	return false
}

func manipulateDir(input string) string {
	result := input
	if runtime.GOOS == "windows" && !strings.Contains(input, "\\\\") { // Because windows. Fuck you Windows.
		result = strings.ReplaceAll(input, "\\", "\\\\")
		result = strings.ReplaceAll(result, "/", "\\\\")
	}
	return result
}

func dirChecks(input string) bool {
	isValid := isValidPath(input)
	if !isValid {
		fmt.Println("That is NOT a valid directory!")
		os.Exit(1)
	}
	return checkifExists(input)
}

// Interactive setup
func interactiveDirectory() {
	prompt := promptui.Prompt{
		Label:     "It would seem you have not configured a directory for where Folderr is. Would you like to do that",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil && err.Error() != "" {
		fmt.Println("Error: ", err)
		panic(err)
	}
	if len(result) > 0 && err == nil {
		dirPrompt := promptui.Prompt{
			Label: "Where would you like to setup/install Folderr?",
		}
		result, err = dirPrompt.Run()
		if err != nil {
			panic(err)
		}
		result = manipulateDir(result)
		exists := dirChecks(result)
		if !exists {
			mkdirPrompt := promptui.Prompt{
				Label:     "Would you like to make the directory \"" + result + "\"",
				IsConfirm: true,
			}
			shouldMake, err := mkdirPrompt.Run()
			if err != nil && err.Error() != "" {
				fmt.Println("Error: ", err)
				panic(err)
			}
			if len(shouldMake) > 0 && err == nil {
				fmt.Println("Creating Directory \"" + result + "\"...")
				err := os.MkdirAll(result, 0660)
				if err != nil {
					fmt.Println("Failed to create directory", result)
					panic(err)
				}
			}
		}
		fmt.Println("Setting directory to be...", result)
		viper.Set("directory", result)
	}
	err = viper.WriteConfig()
	if err != nil {
		panic(err)
	}
}

func interactiveRepository() {
	prompt := promptui.Prompt{
		Label: "What URL is the repository you're using for Folderr",
		Validate: func(input string) error {
			inputUrl, err := url.Parse(input)
			if err != nil {
				return err
			}
			if inputUrl.Scheme == "" || inputUrl.Host == "" {
				return fmt.Errorf("URL Not Found")
			}
			return nil
		},
	}
	inputUrl, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	parsed, err := url.Parse(inputUrl)
	if err != nil {
		panic(err)
	}
	if parsed.Scheme != "https" {
		fmt.Println("Cannot work with non-HTTPS urls. Limitation issue.")
		os.Exit(1)
	}
	shouldFail := false
	prompt = promptui.Prompt{
		Label:     "Is authentication required/Is this a private repository",
		IsConfirm: true,
	}
	needAuth, err := prompt.Run()
	if err != nil && err.Error() != "" {
		panic(err)
	}
	gitOptions := &git.CloneOptions{
		URL: inputUrl,
	}
	if len(needAuth) > 0 && err == nil {
		prompt = promptui.Prompt{
			Label: "We need an authentication token. If you do not have a token, say no. What is it",
		}
		resp, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		if strings.ToLower(resp) == "no" {
			shouldFail = true
		} else {
			gitOptions.Auth = &transport.BasicAuth{
				Username: "git",
				Password: resp,
			}
		}
	}
	if shouldFail {
		fmt.Println("Goodbye!")
		os.Exit(0)
	}
	fmt.Println("Cloning the repository to see if its valid...")
	repo, err := git.Clone(memory.NewStorage(), nil, gitOptions)
	if err != nil {
		panic(err)
	}
	if repo != nil {
		fmt.Println("Setting repository to be", inputUrl)
		viper.Set("repository", inputUrl)
		viper.WriteConfig()
	}
}

// Functions based on args instead of interactive
func dirStatic(args []string) {
	if len(args) < 1 {
		return
	}
	dir := manipulateDir(args[0])
	exists := dirChecks(dir)
	if !exists && !mkdir {
		fmt.Println("Cannot use a directory that does not exist")
		return
	} else if !exists && mkdir {
		fmt.Println("Creating Directory \"" + dir + "\"...")
		err := os.MkdirAll(dir, 0760)
		if err != nil {
			fmt.Println("Failed to create directory", dir)
			panic(err)
		}
	}
	fmt.Println("Setting directory to be", dir)
	viper.Set("directory", dir)
	err := viper.WriteConfig()
	if err != nil {
		panic(err)
	}
}

func repositoryStatic(args []string) {
	if len(args) < 2 {
		return
	}
	url, err := url.Parse(args[1])
	if err != nil {
		panic(err)
	}
	if url.Host == "" || url.Scheme == "" {
		fmt.Println("URL is invalid!")
		os.Exit(1)
	}
	if url.Scheme != "https" {
		fmt.Println("Can only work with HTTPS!")
		os.Exit(1)
	}
	gitOptions := &git.CloneOptions{
		URL: args[1],
	}
	if authFlag != "" {
		gitOptions.Auth = &transport.BasicAuth{
			Username: "git",
			Password: authFlag,
		}
	}
	fmt.Println("Cloning the repository to see if its valid...")
	repo, err := git.Clone(memory.NewStorage(), nil, gitOptions)
	if err != nil {
		panic(err)
	}
	if repo != nil {
		fmt.Println("Setting repository to be", args[1])
		viper.Set("repository", args[1])
		viper.WriteConfig()
	}
}

var initCmd = &cobra.Command{
	Use:   "init [directory] [repository]",
	Short: "Initalize your Folderr Manager config",
	Long: `Initalize your Folderr Manager config interactively or non-interactively.
If a repository is provided non-interactively, the authorization flag MUST be supplied if it is private or else it will fail.
Interactivity happens when you do not provide the listed arguments (excluding flags)`,
	ValidArgs: []string{"directory", "repository"},
	Run: func(cmd *cobra.Command, args []string) {
		isSet := viper.IsSet("directory")
		if override {
			isSet = false
		}
		if !isSet && len(args) == 0 {
			interactiveDirectory()
		}
		if !isSet && len(args) > 0 {
			dirStatic(args)
		}
		isSet = viper.IsSet("repository")
		if override {
			isSet = false
		}
		if !isSet && (len(args) < 2 || strings.Contains(args[1], "--")) {
			interactiveRepository()
		}
		if !isSet && len(args) > 1 && !strings.Contains(args[1], "--") {
			repositoryStatic(args)
		}
		fmt.Println("It looks like your Folderr CLI is initialized!")
	},
}
