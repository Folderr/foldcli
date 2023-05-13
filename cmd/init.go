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

var mkdir, override bool

func init() {
	initCmd.Flags().BoolVar(&mkdir, "mkdir", false, "Make directories if they don't exist")
	initCmd.Flags().BoolVarP(&override, "override", "o", false, "Override previous settings")
	initCmd.Flags().StringVarP(&authFlag, "authorization", "a", "", "Authorization token for private repositories")
	rootCmd.AddCommand(initCmd)
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
		result = strings.ReplaceAll(result, "/", "\\")
	}
	return result
}

func dirChecks(input string) bool {
	isValid := isValidPath(input)
	if !isValid {
		println("That is NOT a valid directory!")
		os.Exit(1)
	}
	return checkifExists(input)
}

// Interactive setup
// CANNOT TEST
// COBRA DOES NOT SUPPLY io.ReadCloser or io.WriteCloser
// COBRA SUPPLIES io.Reader and io.Writer
func interactiveDirectory() (bool, error) {
	if config.directory != "" && !override {
		return true, nil
	}
	prompt := promptui.Prompt{
		Label:     "It would seem you have not configured a directory for where Folderr is. Would you like to do that",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil && err.Error() != "" {
		println("Error: ", err)
		return false, err
	}
	if len(result) > 0 && err == nil {
		dirPrompt := promptui.Prompt{
			Label: "Where would you like to setup/install Folderr?",
		}
		result, err = dirPrompt.Run()
		if err != nil {
			return false, err
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
				println("Error: ", err)
				return false, err
			}
			if len(shouldMake) > 0 && err == nil {
				println("Creating Directory \"" + result + "\"...")
				err := os.MkdirAll(result, 0660)
				if err != nil {
					println("Failed to create directory", result)
					return false, err
				}
			}
		}
		println("Setting directory to be...", result)
		viper.Set("directory", result)
	}
	if dry {
		return true, nil
	}
	err = viper.WriteConfig()
	if err != nil {
		return false, err
	}
	return true, nil
}

// CANNOT TEST
// COBRA DOES NOT SUPPLY io.ReadCloser or io.WriteCloser
// COBRA SUPPLIES io.Reader and io.Writer
func interactiveRepository() (bool, error) {
	if config.repository != "" && !override {
		return true, nil
	}
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
		return false, err
	}
	parsed, err := url.Parse(inputUrl)
	if err != nil {
		return false, err
	}
	if parsed.Scheme != "https" {
		println("Cannot work with non-HTTPS urls. Limitation issue.")
		os.Exit(1)
	}
	shouldFail := false
	prompt = promptui.Prompt{
		Label:     "Is authentication required/Is this a private repository",
		IsConfirm: true,
	}
	needAuth, err := prompt.Run()
	if err != nil && err.Error() != "" {
		return false, err
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
			return false, err
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
		println("Goodbye!")
		os.Exit(0)
	}
	println("Cloning the repository to see if its valid...")
	repo, err := git.Clone(memory.NewStorage(), nil, gitOptions)
	if err != nil {
		return false, err
	}
	if repo == nil {
		return false, nil
	}
	println("Setting repository to be", inputUrl)
	if dry {
		return true, nil
	}
	viper.Set("repository", inputUrl)
	err = viper.WriteConfig()
	if err != nil {
		return false, err
	}
	return true, nil
}

// Functions based on args instead of interactive
// CAN TEST
func dirStatic(args []string) (bool, error) {
	if config.directory != "" && !override {
		return true, nil
	}
	if len(args) < 1 {
		return false, nil
	}
	ldir := manipulateDir(args[0])
	exists := dirChecks(ldir)
	if !exists && !mkdir {
		println("Cannot use a directory that does not exist")
		return false, nil
	} else if !exists && mkdir {
		println("Creating Directory \"" + ldir + "\"...")
		err := os.MkdirAll(ldir, 0760)
		if err != nil {
			println("Failed to create directory", ldir)
			return false, err
		}
	}
	println("Setting directory to be", ldir)
	viper.Set("directory", ldir)
	if dry {
		return true, nil
	}
	err := viper.WriteConfig()
	if err != nil {
		return false, err
	}
	return true, nil
}

func repositoryStatic(args []string) (bool, error) {
	if config.repository != "" && !override {
		return true, nil
	}
	if len(args) < 2 {
		return false, nil
	}
	url, err := url.Parse(args[1])
	if err != nil {
		println('f')
		return false, err
	}
	if url.Host == "" || url.Scheme == "" {
		println("URL is invalid!")
		os.Exit(1)
	}
	if url.Scheme != "https" {
		println("Can only work with HTTPS!")
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
	println("Cloning the repository to see if its valid...")
	repo, err := git.Clone(memory.NewStorage(), nil, gitOptions)
	if err != nil {
		return false, err
	}
	if repo == nil {
		return false, nil
	}
	viper.Set("repository", args[1])
	println("Setting repository to be", args[1])
	if dry {
		return true, nil
	}
	err = viper.WriteConfig()
	if err != nil {
		println(err)
		return false, nil
	}
	return true, nil
}

var initCmd = &cobra.Command{
	Use:   "init [directory] [repository]",
	Short: "Initalize your Folderr Manager config",
	Long: `Initalize your Folderr Manager config interactively or non-interactively.
If a repository is provided non-interactively, the authorization flag MUST be supplied if it is private or else it will fail.
Interactivity happens when you do not provide the listed arguments (excluding flags)`,
	ValidArgs: []string{"directory", "repository"},
	Run: func(cmd *cobra.Command, args []string) {
		_, err := ReadConfig()
		if err != nil {
			panic(err)
		}
		var fails []string
		noFail := true
		if len(args) == 0 {
			noFail, err = interactiveDirectory()
		}
		if len(args) > 0 {
			noFail, err = dirStatic(args)
		}
		if err != nil {
			panic(err)
		}
		if !noFail {
			fails = append(fails, "directory")
		}
		if len(args) < 2 || strings.Contains(args[1], "--") {
			noFail, err = interactiveRepository()
		}
		if len(args) > 1 && !strings.Contains(args[1], "--") {
			noFail, err = repositoryStatic(args)
		}
		if err != nil {
			panic(err)
		}
		if !noFail {
			fails = append(fails, "repository")
		}

		if len(fails) > 0 {
			println("One or more aspects of setup failed.")
			println("Failures:", strings.Join(fails, ", "))
		} else {
			println("It looks like your Folderr CLI is initialized!")
		}
		if dry {
			println("No changes were made.")
		}
	},
}
