package init

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Folderr/foldcli/utilities"
	"github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mkdir, override, dry bool
var authFlag string

func init() {
	folderrCmd.Flags().BoolVar(&mkdir, "mkdir", false, "Make directories if they don't exist")
	folderrCmd.Flags().BoolVarP(&override, "override", "o", false, "Override previous settings")
	folderrCmd.Flags().StringVarP(&authFlag, "authorization", "a", "", "Authorization token for private repositories")
	folderrCmd.Flags().BoolVar(&dry, "dry", false, "Whether or not to run the command in dry-run mode")
	initCmd.AddCommand(folderrCmd)
}

func interactiveDirectory(vip *viper.Viper, config utilities.Config) (bool, error) {
	if config.Directory != "" && !override {
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
		result = utilities.ManipulateDir(result)
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
				err := os.MkdirAll(result, 0770)
				if err != nil {
					println("Failed to create directory", result)
					return false, err
				}
			}
		}
		println("Setting directory to be...", result)
		vip.Set("directory", result)
	}
	if dry {
		return true, nil
	}
	err = vip.WriteConfig()
	if err != nil {
		return false, err
	}
	return true, nil
}

// CANNOT TEST
// COBRA DOES NOT SUPPLY io.ReadCloser or io.WriteCloser
// COBRA SUPPLIES io.Reader and io.Writer
func interactiveRepository(vip *viper.Viper, config utilities.Config) (bool, error) {
	if config.Repository != "" && !override {
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
	vip.Set("repository", inputUrl)
	err = vip.WriteConfig()
	if err != nil {
		return false, err
	}
	return true, nil
}

// Functions based on args instead of interactive
// CAN TEST
func dirStatic(vip *viper.Viper, config utilities.Config, args []string) (bool, error) {
	if config.Directory != "" && !override {
		return true, nil
	}
	if len(args) < 1 {
		return false, nil
	}
	ldir := utilities.ManipulateDir(args[0])
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
	vip.Set("directory", ldir)
	if dry {
		return true, nil
	}
	err := vip.WriteConfig()
	if err != nil {
		return false, err
	}
	return true, nil
}

func repositoryStatic(vip *viper.Viper, config utilities.Config, args []string) (bool, error) {
	if config.Repository != "" && !override {
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
	vip.Set("repository", args[1])
	println("Setting repository to be", args[1])
	if dry {
		return true, nil
	}
	err = vip.WriteConfig()
	if err != nil {
		println(err)
		return false, nil
	}
	return true, nil
}

var folderrCmd = &cobra.Command{
	Use:   "folderr [directory] [repository]",
	Short: "Initalize config for foldcli Folderr commands",
	Long: `Initalize your Folderr CLI config interactively or non-interactively.
If a repository is provided non-interactively, the authorization flag MUST be supplied if it is private or else it will fail.
Interactivity happens when you do not provide the listed arguments (excluding flags)`,
	ValidArgs: []string{"directory", "repository"},
	RunE: func(command *cobra.Command, args []string) error {
		dir, err := utilities.GetConfigDir(dry)
		if err != nil {
			return err
		}
		vip, config, _, err := utilities.ReadConfig(dir, dry)
		if err != nil {
			panic(err)
		}
		var fails []string
		noFail := true
		if len(args) == 0 {
			noFail, err = interactiveDirectory(vip, config)
		}
		if len(args) > 0 {
			noFail, err = dirStatic(vip, config, args)
		}
		if err != nil {
			panic(err)
		}
		if !noFail {
			fails = append(fails, "directory")
		}
		if len(args) < 2 || strings.Contains(args[1], "--") {
			noFail, err = interactiveRepository(vip, config)
		}
		if len(args) > 1 && !strings.Contains(args[1], "--") {
			noFail, err = repositoryStatic(vip, config, args)
		}
		if err != nil {
			panic(err)
		}
		if !noFail {
			fails = append(fails, "repository")
		}

		if len(fails) > 0 {
			command.Println("One or more aspects of setup failed.")
			command.Println("Failures:", strings.Join(fails, ", "))
		} else {
			command.Println("It looks like your Folderr CLI is initialized!")
		}
		if dry {
			command.Println("No changes were made.")
		}
		return nil
	},
}
