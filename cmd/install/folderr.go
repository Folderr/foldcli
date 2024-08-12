/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package install

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/Folderr/foldcli/utilities"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	transport "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func cloneFolderr(w io.Writer, config utilities.Config, options *git.CloneOptions, dry bool) (*git.Repository, error) {
	if dry {
		fmt.Fprintf(w, "Cloning in directory %v for dry-run mode\n", config.Directory)
		repo, err := git.PlainClone(config.Directory, false, options)
		return repo, err
	}
	repo, err := git.PlainClone(config.Directory, false, options)
	return repo, err
}

var dry bool
var authFlag string
var sharedConfig utilities.Config

// installCmd represents the install command
var installFolderr = &cobra.Command{
	Use:   "folderr",
	Short: "Install Folderr into the directory setup with \"foldcli init folderr\"",
	Long:  `Checks for Folderrs dependencies and installs Folderr`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var config utilities.Config
		if sharedConfig.Directory != "" {
			cmd.Println("Shared config directory not found")
			config = sharedConfig
		} else {
			dir, err := utilities.GetConfigDir(dry)
			if err != nil {
				return err
			}
			_, config, _, err = utilities.ReadConfig(dir, dry)
			if err != nil {
				panic(err)
			}
		}
		if !config.CanInstall {
			cmd.Println("Folderr CLI is not initialized. Run \"" + utilities.Constants.RootCmdName + " init\" to fix this issue.")
			return nil
		}
		cmd.Println("Checking if NodeJS is installed")
		out, err := utilities.FindSystemCommandVersion(cmd.OutOrStdout(), "node", true, "v")
		if err != nil {
			return err
		}
		if out == "" {
			cmd.Println("NodeJS not installed. Aborting.")
			cmd.Println("Install Node before running this command!")
			return nil
		}
		cmd.Println("NodeJS appears to be installed!")
		// ensure NPM is installed
		// we don't care about the actual version tbh.
		cmd.Println("Checking if NPM is installed")
		npm, err := utilities.FindSystemCommandVersion(cmd.OutOrStdout(), "npm", false, "")
		if err != nil {
			panic(err)
		}
		if npm == "" {
			cmd.Println("NPM not installed. Aborting.")
			cmd.Println("Install NPM before running this command!")
			return nil
		}
		cmd.Println("NPM appears to be installed")
		cmd.Println("Checking for TypeScript installation")
		tsc, err := utilities.FindSystemCommandVersion(cmd.OutOrStdout(), "tsc", true, "Version ")
		if err != nil {
			return err
		}
		swc, err := utilities.FindSystemCommandVersion(cmd.OutOrStdout(), "swc", true, "@swc/cli: ")
		if err != nil {
			return err
		}
		if tsc == "" && swc == "" {
			cmd.Println("Neither TypeScript nor SWC not installed. Aborting.")
			cmd.Println("Install TypeScript or SWC before running this command!")
			return nil
		}
		cmd.Println("TypeScript or SWC appears to be installed")

		// Turn Node version into a int!
		versions := []int{}
		for _, i := range strings.Split(out, ".") {
			j, err := strconv.Atoi(i)
			if err != nil {
				panic(err)
			}
			versions = append(versions, j)
		}
		// Check node version compatibility
		// Future versions should use a Matrix included with the repository.
		// like say the engines field in the package.json
		cmd.Println("Checking Node version for support & compatibility")
		if 20 >= versions[0] && versions[0] <= 22 {
			cmd.Println("Supported")
		} else if versions[0] <= 20 {
			cmd.Println("Your Node Version is too old!")
			cmd.Println("Update your Node version before running this command!")
			return nil
		} else if versions[0] >= 20 {
			cmd.Println("We're not sure Folderr will work with this new of a version of Node")
		}

		// Check install folder for Folderr repository
		repo, err := git.PlainOpen(config.Directory)
		if err != nil && !strings.Contains(err.Error(), "repository does not exist") {
			cmd.Println("An error occurred while checking if the repository already exists")
			cmd.Println("Error:", err)
			panic(err)
		}
		// If the repo exists, consider Folderr installed.
		// If in dry-run mode we can ignore this, as no changes occur.
		if repo != nil && !dry {
			cmd.Println("Found repository, Folderr is installed.")
			os.Exit(1)
		}

		// Clone Folderr.
		gitOptions := &git.CloneOptions{
			URL: config.Repository,
		}
		if authFlag != "" {
			gitOptions.Auth = &transport.BasicAuth{
				Username: "git",
				Password: authFlag,
			}
		}

		cmd.Println("Cloning repository...")
		repo, err = cloneFolderr(cmd.OutOrStdout(), config, gitOptions, dry)
		if err != nil {
			if errors.Is(err, git.ErrRepositoryNotExists) {
				cmd.Println("That repository doesn't exist")
				os.Exit(1)
			} else if strings.Contains(err.Error(), "authorization") || strings.Contains(err.Error(), "authentication") {
				cmd.Println("Authentication required. Please pass either the authorization flag or set the " + utilities.Constants.EnvPrefix + "TOKEN environment variable.\n" +
					"See \"" + utilities.Constants.RootCmdName + " install --help\" for more info")
				os.Exit(1)
			} else {
				cmd.Println("An Error Occurred while cloning the repository. Error:", err)
				panic(err)
			}
		}
		if repo == nil {
			cmd.Println("Cannot find the repository for some reason (Not Found).")
			os.Exit(1)
		}

		// Get tags to checkout
		tags, err := repo.Tags()
		if err != nil {
			cmd.Println("An error occurred while fetching the repository. Error:", err)
		}
		highestVer, highest, err := newDetermineHighestVersion(tags)
		// Determine if tag or commit based releases should be used.
		if err != nil {
			cmd.Println("An error occurred while determining the highest tag:", err)
		}
		var releaseType string
		if highestVer == nil {
			cmd.Println("Not using Tags for updating...")
			cmd.Println("Reason: Latest tag is too old. (Pre V2)")
			viper.Set("releaseType", "commit")
		} else {
			releaseType = "tag"
			viper.Set("releaseType", "tag")
			viper.Set("release", highest.Name().Short())
		}
		if !dry {
			err = viper.WriteConfig()
			if err != nil {
				cmd.Println("Error Occurred while writing config:", err)
				panic(err)
			}
		}
		cmd.Println("Clone successful")

		// Get the work tree
		tree, err := repo.Worktree()
		if err != nil {
			cmd.Println("error Occurred while getting Work Tree:", err)
			panic(err)
		}
		// Check out the CORRECT release type
		if releaseType == "tag" {
			cmd.Println("Checking out tag", highest.Name().Short())
			err = tree.Checkout(&git.CheckoutOptions{
				Hash: highest.Hash(),
			})
			if err != nil {
				cmd.Println("Failed to check out tag", highest.Name().Short(), "with error:", err)
				panic(err)
			}
			cmd.Println("Checked out tag", highest.Name().Short())
		} else {
			if err != nil {
				cmd.Println("Error while reading the work tree:", err)
				panic(err)
			}
			branches, err := repo.Branches()
			if err != nil {
				cmd.Println("Error while loading branches", err)
			}
			var branch *plumbing.Reference
			branches.ForEach(func(r *plumbing.Reference) error {
				if r.Name().Short() == "master" {
					branch = r
				} else if r.Name().Short() == "main" && branch == nil {
					branch = r
				} else if r.Name().Short() == "dev" && branch == nil {
					branch = r
				}
				return nil
			})
			if branch == nil {
				cmd.Println("FATAL: Suitable Branch Not Found")
				os.Exit(1)
			}
			err = tree.Checkout(&git.CheckoutOptions{Branch: branch.Name()})
			if err != nil {
				cmd.Println("Error while checking out branch", branch.Name().Short()+",", "error:", err)
				panic(err)
			}
		}
		cmd.Println("Checkout successful")

		// REMOVE AFTER TESTING
		// IMPL: install dependencies...
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		err = os.Chdir(config.Directory)
		if errors.Is(err, fs.ErrPermission) {
			cmd.Println("Cannot access directory \"" + config.Directory + "\".\nDo I have permission to access that?\nIs that a directory?\nIs it executable (linux only)?")
		} else if err != nil {
			panic(err)
		}

		args = []string{"install", "--omit=dev"}
		if dry {
			// After Folderr:frontend is merged with folderr:dev we can remove
			// "--ignore-scripts"
			args = append(args, "--dry-run")
		}
		npmCmd, err := utilities.FindSystemCommand(cmd.OutOrStdout(), "npm", args)
		if err != nil {
			panic(err)
		}
		output, err := npmCmd.CombinedOutput()
		if err != nil {
			if len(output) > 0 {
				fmt.Println("NPM install failed, here's the output")
				fmt.Println(string(output))
			}
			return err
		}
		// remove after dev
		cmd.Println("Output from npm", strings.Join(args, " "))
		cmd.Println(string(output))
		os.Chdir(cwd)
		cmd.Println("Install seems to have gone correctly.")
		buildCmd := "npm run build"
		if swc == "" {
			buildCmd = "npm run build:tsc"
		}
		cmd.Printf(`To build Folderr go to "%v" and type "%v"`, config.Directory, buildCmd)

		return nil
	},
}

func newDetermineHighestVersion(tags storer.ReferenceIter) (*semver.Version, *plumbing.Reference, error) {
	c, err := semver.NewConstraint(">= 2.0.0-0")
	if err != nil {
		return nil, nil, err
	}

	var highest *plumbing.Reference

	tags.ForEach(func(r *plumbing.Reference) error {
		if !strings.HasPrefix(r.Name().Short(), "v") && !strings.HasPrefix(r.Name().Short(), "V") {
			return nil
		}

		var highestVersion *semver.Version
		if highest != nil {
			// I don't really care if this errors, if it errors then we'll just set highest
			localVersion, _ := semver.NewVersion(highest.Name().Short())
			highestVersion = localVersion
		}
		version := r.Name().Short()
		v, err := semver.NewVersion(version)
		if err != nil {
			return err
		}

		if c.Check(v) {
			if highestVersion == nil {
				highest = r
			} else if v.GreaterThan(highestVersion) {
				highest = r
			}
		}

		return nil
	})

	if highest != nil {
		v, _ := semver.NewVersion(highest.Name().Short())
		return v, highest, nil
	}
	return nil, nil, nil
}

func init() {
	installFolderr.Flags().StringVarP(&authFlag, "authorization", "a", "", "Authorization token for private repositories")
	installFolderr.Flags().BoolVar(&dry, "dry", false, "Runs the command but does not change anything")
	installCmd.AddCommand(installFolderr)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
