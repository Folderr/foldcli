package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Folderr/foldcli/utilities"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var testDBNoConn bool

func TestSetupDb(t *testing.T) {
	os.Setenv("test", "true")
	configDir, err := utilities.GetConfigDir(dry)
	if err != nil {
		t.Fatal("Failed due to error: " + err.Error())
	}
	_, config, secrets, err := utilities.ReadConfig(configDir, dry)
	if err != nil {
		t.Fatal("Failed due to error: " + err.Error())
	}
	if os.Getenv(utilities.Constants.EnvPrefix+"FLDRR_TEMPDIR") == "" {
		os.Setenv(utilities.Constants.EnvPrefix+"FLDRR_TEMPDIR", config.Directory)
		//	t.Logf("Set env var %v to %v\n", utilities.Constants.EnvPrefix+"FLDRR_TEMPDIR", config.Directory)
	}
	if os.Getenv(utilities.Constants.EnvPrefix+"CFG_TEMPDIR") == "" {
		os.Setenv(utilities.Constants.EnvPrefix+"CFG_TEMPDIR", configDir)
		//	t.Logf("Set env var %v to %v\n", utilities.Constants.EnvPrefix+"CFG_TEMPDIR", configDir)
	}
	os.Setenv(utilities.Constants.EnvPrefix+"DB_NAME", "foldcli-db-testing")
	// set DB_URI before running this test
	if os.Getenv(utilities.Constants.EnvPrefix+"MONGO_URI") == "" {
		t.Skip("No MONGO_URI environment variable provided. No DB operations available")
	}
	actual := &bytes.Buffer{}
	options := &git.CloneOptions{
		URL: config.Repository,
	}
	if secrets.GitToken != "" {
		options.Auth = &http.BasicAuth{Username: "git", Password: secrets.GitToken}
	}
	_, err = git.PlainClone(config.Directory, false, options)
	if err != nil {
		t.Skipf("Couldn't install repo. Skipping. See error below. %v\n", err)
	} else {
		t.Logf("Installed Repo %v to %v\n", config.Repository, config.Directory)
	}

	args := []string{"--no-cleanup", "-v"} // name of the testing db for this project
	RootCmd.SetOut(actual)
	RootCmd.SetArgs(append([]string{"setup", "db"}, args...))
	_, err = RootCmd.ExecuteC()
	if strings.Contains(actual.String(), "Server Timeout Error") {
		testDBNoConn = true
		t.Skip("Database could not be connected to. Skipping")
	}
	t.Log(actual.String())
	if err != nil {
		t.Fatalf("Command \""+rootCmdName+" setup db %v\" failed because of error, %v", args[1], err)
	}
	if strings.Contains(actual.String(), "Folderr appears to be setup") {
		t.Fatal("Using pre-setup version of Folderr. Test invalid.")
	}
	if !strings.Contains(actual.String(), "Saved public key to database") || !strings.Contains(actual.String(), "The keys were saved in") {
		t.Logf("Command \""+rootCmdName+" setup db %v\" failed because unexpected output", args[1])
		t.FailNow()
	}
}

func TestSetupDbPresetup(t *testing.T) {
	testDBNoConn = false
	os.Setenv(utilities.Constants.EnvPrefix+"DB_NAME", "foldcli-db-testing")
	// set DB_URI before running this test
	if os.Getenv(utilities.Constants.EnvPrefix+"MONGO_URI") == "" {
		t.Skip("No MONGO_URI environment variable provided. No DB operations available")
	}
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"--no-cleanup", "-v"} // name of the testing db for this project
	RootCmd.SetOut(actual)
	RootCmd.SetArgs(append([]string{"setup", "db"}, args...))
	_, err := RootCmd.ExecuteC()
	if strings.Contains(actual.String(), "Server Timeout Error") {
		testDBNoConn = true
		t.Skip("Database could not be connected to. Skipping")
	}
	t.Log(actual.String())
	if err != nil {
		t.Fatalf("Command \""+rootCmdName+" setup db %v\" failed because of error, %v", args[1], err)
	}
	if !strings.Contains(actual.String(), "Folderr appears to be setup") {
		t.Fatal("Could not find pre-setup version.")
	}
}

func TestSetupDbCleanup(t *testing.T) {
	if testDBNoConn {
		t.Skip("Database can't be connected to. Skipping")
	}
	// set DB_URI before running this test
	if os.Getenv(utilities.Constants.EnvPrefix+"MONGO_URI") == "" {
		t.Skip("No MONGO_URI environment variable provided. No DB operations available")
	}
	configDir, err := utilities.GetConfigDir(dry)
	if err != nil {
		t.Fatal("Failed due to error: " + err.Error())
	}
	_, config, _, err := utilities.ReadConfig(configDir, dry)
	if err != nil {
		t.Fatal("Failed due to error: " + err.Error())
	}
	t.Logf("%+v", config.Directory)
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"--no-cleanup", "-v"} // name of the testing db for this project
	RootCmd.SetOut(actual)
	RootCmd.SetArgs(append([]string{"setup", "db"}, args...))
	if strings.Contains(actual.String(), "Server Timeout Error") {
		t.Skip("Database could not be connected to. Skipping")
	}

	cleanupFolderrDbCmd(actual, config, config.Database.DbName, filepath.Join(configDir, "/keys"))
	t.Log(actual.String())

	if !strings.Contains(actual.String(), "Cleaned up") {
		t.Error("Seems cleanup was useless. Why?")
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
