/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	transport "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func findCommandV(command string, hasPrefix bool, prefix string) (string, error) {
	execCmd, err := findCommand(command, []string{"-v"})
	if err != nil {
		return "", err
	}
	output, err := execCmd.CombinedOutput()
	if err != nil {
		printf("It would seem I encountered an error. Is %v installed?\n", command)
		println("Here's the error:", err)
		return "", err
	}
	printf("Printing output from executing \"%v -v\"\n", command)
	// transform output into integers.
	out := string(output)
	out = strings.TrimSpace(out)
	println(out)
	if !hasPrefix {
		return out, nil
	}
	if strings.HasPrefix(out, prefix) {
		out, _ = strings.CutPrefix(out, prefix)
	} else {
		printf("Got unexpected output from running \"%v -v\". Contact developers.\n", command)
		println("Contact developers at https://github.com/Folderr/Folderr-CLI/issues")
		println("Output:", out)
		return "", nil
	}
	return out, nil
}

func findCommand(command string, args []string) (*exec.Cmd, error) {
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		printf("I can't find %v. Is %v installed?\n", command, command)
		println("Error for debug purposes:", err)
		return nil, err
	}
	return exec.Command(cmdPath, args...), nil
}

func cloneFolderr(options *git.CloneOptions) (*git.Repository, error) {
	if dry {
		printf("Cloning in directory %v for dry-run mode\n", config.directory)
		repo, err := git.PlainClone(config.directory, false, options)
		return repo, err
	}
	repo, err := git.PlainClone(config.directory, false, options)
	return repo, err
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Folderr into the setup directory",
	Long:  `Checks for Folderrs dependencies and installs Folderr`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := ReadConfig()
		if err != nil {
			panic(err)
		}
		if !config.canInstall {
			println("Folderr CLI is not initialized. Run \"" + rootCmdName + " init\" to fix this issue.")
			return
		}
		println("Checking if NodeJS is installed")
		out, err := findCommandV("node", true, "v")
		if err != nil {
			panic(err)
		}
		if out == "" {
			println("NodeJS not installed. Aborting.")
			println("Install Node before running this command!")
			return
		}
		println("NodeJS appears to be installed!")
		// ensure NPM is installed
		// we don't care about the actual version tbh.
		println("Checking if NPM is installed")
		npm, err := findCommandV("npm", false, "")
		if err != nil {
			panic(err)
		}
		if npm == "" {
			println("NPM not installed. Aborting.")
			println("Install NPM before running this command!")
			return
		}
		println("NPM appears to be installed")
		println("Checking for TypeScript installation")
		tsc, err := findCommandV("tsc", true, "Version ")
		if err != nil {
			panic(err)
		}
		if tsc == "" {
			println("TypeScript not installed. Aborting.")
			println("Install TypeScript before running this command!")
			return
		}
		println("TypeScript appears to be installed")

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
		println("Checking Node version for support & compatibility")
		if 16 >= versions[0] && versions[0] <= 18 {
			println("Supported")
		} else if versions[0] <= 16 {
			println("Your Node Version is too old!")
			println("Update your Node version before running this command!")
			return
		} else if versions[0] >= 20 {
			println("We're not sure Folderr will work with this new of a version of Node")
		} else if versions[0] >= 18 {
			println("Your Node version is not tested but should work")
		}

		// Check install folder for Folderr repository
		repo, err := git.PlainOpen(config.directory)
		if err != nil && !strings.Contains(err.Error(), "repository does not exist") {
			println("An error occurred while checking if the repository already exists")
			println("Error:", err)
			panic(err)
		}
		// If the repo exists, consider Folderr installed.
		// If in dry-run mode we can ignore this, as no changes occur.
		if repo != nil && !dry {
			println("Found repository, Folderr is installed.")
			os.Exit(1)
		}

		// Clone Folderr.
		gitOptions := &git.CloneOptions{
			URL: config.repository,
		}
		if authFlag != "" {
			gitOptions.Auth = &transport.BasicAuth{
				Username: "git",
				Password: authFlag,
			}
		}

		println("Cloning repository...")
		repo, err = cloneFolderr(gitOptions)
		if err != nil {
			if errors.Is(err, git.ErrRepositoryNotExists) {
				println("That repository doesn't exist")
				os.Exit(1)
			} else if strings.Contains(err.Error(), "authorization") || strings.Contains(err.Error(), "authentication") {
				println("Authentication required. Please pass either the authorization flag or set the " + envPrefix + "TOKEN environment variable.\n" +
					"See \"" + rootCmdName + " install --help\" for more info")
				os.Exit(1)
			} else {
				println("An Error Occurred while cloning the repository. Error:", err)
				panic(err)
			}
		}
		if repo == nil {
			println("Cannot find the repository for some reason (Not Found).")
			os.Exit(1)
		}

		// Get tags to checkout
		tags, err := repo.Tags()
		if err != nil {
			println("An error occurred while fetching the repository. Error:", err)
		}
		highestVer, highest, err := newDetermineHighestVersion(tags)
		// Determine if tag or commit based releases should be used.
		if err != nil {
			println("An error occurred while determining the highest tag:", err)
		}
		var releaseType string
		if highestVer == nil {
			println("Not using Tags for updating...")
			println("Reason: Latest tag is too old. (Pre V2)")
			viper.Set("releaseType", "commit")
		} else {
			releaseType = "tag"
			viper.Set("releaseType", "tag")
			viper.Set("release", highest.Name().Short())
		}
		if !dry {
			err = viper.WriteConfig()
			if err != nil {
				println("Error Occurred while writing config:", err)
				panic(err)
			}
		}
		println("Clone successful")

		// Get the work tree
		tree, err := repo.Worktree()
		if err != nil {
			println("error Occurred while getting Work Tree:", err)
			panic(err)
		}
		// Check out the CORRECT release type
		if releaseType == "tag" {
			println("Checking out tag", highest.Name().Short())
			err = tree.Checkout(&git.CheckoutOptions{
				Hash: highest.Hash(),
			})
			if err != nil {
				println("Failed to check out tag", highest.Name().Short(), "with error:", err)
				panic(err)
			}
			println("Checked out tag", highest.Name().Short())
		} else {
			if err != nil {
				println("Error while reading the work tree:", err)
				panic(err)
			}
			branches, err := repo.Branches()
			if err != nil {
				println("Error while loading branches", err)
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
				println("FATAL: Suitable Branch Not Found")
				os.Exit(1)
			}
			err = tree.Checkout(&git.CheckoutOptions{Branch: branch.Name()})
			if err != nil {
				println("Error while checking out branch", branch.Name().Short()+",", "error:", err)
				panic(err)
			}
		}
		println("Checkout successful")

		// REMOVE AFTER TESTING
		// IMPL: install dependencies...
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		err = os.Chdir(config.directory)
		if errors.Is(err, fs.ErrPermission) {
			println("Cannot access directory \"" + config.directory + "\".\nDo I have permission to access that?\nIs that a directory?\nIs it executable (linux only)?")
		} else if err != nil {
			panic(err)
		}

		args = []string{"install", "--omit=dev"}
		if dry {
			// After Folderr:frontend is merged with folderr:dev we can remove
			// "--ignore-scripts"
			args = append(args, "--dry-run")
		}
		if config.repository == "https://github.com/Folderr/Folderr" && os.Getenv("test") != "true" {
			args = append(args, "--ignore scripts")
		}
		npmCmd, err := findCommand("npm", args)
		if err != nil {
			panic(err)
		}
		output, err := npmCmd.CombinedOutput()
		if err != nil {
			if len(output) > 0 {
				fmt.Println("NPM install failed, here's the output")
				fmt.Println(string(output))
			}
			panic(err)
		}
		// remove after dev
		println("Output from npm", strings.Join(args, " "))
		println(string(output))
		os.Chdir(cwd)
		println("Install seems to have gone correctly.")
		printf(`To build Folderr go to "%v" and type "npm run build:production"`, config.directory)

		if dry {
			return
		}
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
	installCmd.Flags().StringVarP(&authFlag, "authorization", "a", "", "Authorization token for private repositories")
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
