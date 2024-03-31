package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Folderr/foldcli/utilities"
)

func TestSetupDb(t *testing.T) {
	// set DB_URI before running this test
	os.Setenv(utilities.Constants.EnvPrefix+"DB_NAME", "foldcli-db-testing")
	if os.Getenv("DB_URI") == "" {
		t.Skip("No DB_URI environment variable provided. No DB operations available")
	}
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"--no-cleanup", "-v"} // name of the testing db for this project
	RootCmd.SetOut(actual)
	RootCmd.SetArgs(append([]string{"setup", "db"}, args...))
	_, err := RootCmd.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Fatalf("Command \""+rootCmdName+" setup db\" %v failed because of error, %v", args[0], err)
	}
	if strings.Contains(actual.String(), "Folderr appears to be setup") {
		t.Fatal("Using pre-setup version of Folderr. Test invalid.")
	}
	if !strings.Contains(actual.String(), "If this is not the location of your Folderr installation, please follow the directions below.") {
		t.Logf("Command \""+rootCmdName+" setup db\" %v failed because unexpected output", args[0])
		t.FailNow()
	}
}

func TestSetupDbPresetup(t *testing.T) {
	// set DB_URI before running this test
	os.Setenv(utilities.Constants.EnvPrefix+"DB_NAME", "foldcli-db-testing")
	if os.Getenv("DB_URI") == "" {
		t.Skip("No DB_URI environment variable provided. No DB operations available")
	}
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"--no-cleanup", "-v"} // name of the testing db for this project
	RootCmd.SetOut(actual)
	RootCmd.SetArgs(append([]string{"setup", "db"}, args...))
	_, err := RootCmd.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Fatalf("Command \""+rootCmdName+" setup db\" %v failed because of error, %v", args[0], err)
	}
	t.Log(strings.Contains(actual.String(), "Folderr appears to be setup"))
	if !strings.Contains(actual.String(), "Folderr appears to be setup") {
		t.Fatal("Could not find pre-setup version.")
	}
}

func TestSetupDbCleanup(t *testing.T) {
	// set DB_URI before running this test
	os.Setenv(utilities.Constants.EnvPrefix+"DB_NAME", "foldcli-db-testing")
	if os.Getenv("DB_URI") == "" {
		t.Skip("No DB_URI environment variable provided. No DB operations available")
	}
	configDir, err := utilities.GetConfigDir(dry)
	if err != nil {
		t.Fatal("Failed due to error: " + err.Error())
	}
	_, config, _, err := utilities.ReadConfig(configDir, dry)
	if err != nil {
		t.Fatal("Failed due to error: " + err.Error())
	}
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"--no-cleanup", "-v"} // name of the testing db for this project
	RootCmd.SetOut(actual)
	RootCmd.SetArgs(append([]string{"setup", "db"}, args...))

	cleanupFolderrDbCmd(config, config.Database.DbName, ConfigDir)
	t.Log(actual.String())

	if !strings.HasPrefix(actual.String(), "Cleaned up ") {
		t.Error("Seems cleanup was useless. Why?")
	}
}
