package init

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Folderr/foldcli/utilities"
)

// Never run parallel. It fucks up Viper
func TestInit(t *testing.T) {
	os.Setenv("test", "true")
	repo := "https://github.com/Folderr/Docs"
	if utilities.GetGitToken() != "" {
		repo = "https://github.com/Folderr/Folderr"
	}
	actual := &bytes.Buffer{}
	dir, err := utilities.GetConfigDir(true)
	if err != nil {
		t.Fatal(err)
	}
	if dir == "" {
		dir = t.TempDir()
	}
	os.Setenv(utilities.Constants.EnvPrefix+"CFG_TEMPDIR", dir)
	os.Setenv(utilities.Constants.EnvPrefix+"FLDRR_TEMPDIR", dir)
	args := []string{dir, repo, "--dry", "-o"}
	command, args, err := initCmd.Find(append([]string{"folderr"}, args...))
	command.Root().SetOut(actual)
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	// init command usage: init folderr [directory] [repository]
	// we'll use github.com/Folderr/Docs here as its a public repository

	command.Root().SetArgs(append([]string{"init", "folderr"}, args...))
	_, err = command.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Errorf(`Command "`+utilities.Constants.RootCmdName+` init folderr %v %v %v %v" failed because of error, %v`, args[0], args[1], args[2], args[3], err)
	}
	suffix := []string{"It looks like your Folderr CLI is initialized!", "No changes were made."}
	if !strings.Contains(actual.String(), suffix[0]) || !strings.Contains(actual.String(), suffix[1]) {
		t.Errorf(`Unexpected output from "`+utilities.Constants.RootCmdName+` init %v %v %v %v"`, args[0], args[1], args[2], args[3])
	}

	t.Cleanup(func() {
		cfgTemp := os.Getenv(utilities.Constants.EnvPrefix + "CFG_TEMPDIR")
		fldrrTemp := os.Getenv(utilities.Constants.EnvPrefix + "FLDRR_TEMPDIR")

		err := os.RemoveAll(cfgTemp)
		if err != nil {
			t.Logf("Ran into error when removing config directories: %v", err.Error())
		}
		err = os.RemoveAll(fldrrTemp)
		if err != nil {
			t.Logf("Ran into error when removing folderr directories: %v", err.Error())
		}
		os.Unsetenv(utilities.Constants.EnvPrefix + "FLDRR_TEMPDIR")
		os.Unsetenv(utilities.Constants.EnvPrefix + "CFG_TEMPDIR")
	})
}
