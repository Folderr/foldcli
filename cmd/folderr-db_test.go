package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSetupFolderrDb(t *testing.T) {
	// set MONGO_URI before running this test
	if os.Getenv("MONGO_URI") == "" {
		t.Skip("No MONGO_URI environment variable provided. No DB operations available")
	}
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"folderr-cli-testing", "--no-cleanup", "-v"} // name of the testing db for this project
	rootCmd.SetOut(actual)
	rootCmd.SetArgs(append([]string{"setup", "folderr-db"}, args...))
	_, err := rootCmd.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Fatalf("Command folderr-cli setup folderr-db %v failed because of error, %v", args[0], err)
	}
	if strings.Contains(actual.String(), "Folderr appears to be setup") {
		t.Fatal("Using pre-setup version of Folderr. Test invalid.")
	}
	if !strings.Contains(actual.String(), "END RSA PRIVATE KEY") {
		t.Logf("Command folderr-cli setup folderr-db %v failed because unexpected output", args[0])
		t.FailNow()
	}
}

func TestSetupFolderrDbPresetup(t *testing.T) {
	// set MONGO_URI before running this test
	if os.Getenv("MONGO_URI") == "" {
		t.Skip("No MONGO_URI environment variable provided. No DB operations available")
	}
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"folderr-cli-testing", "--no-cleanup", "-v"} // name of the testing db for this project
	rootCmd.SetOut(actual)
	rootCmd.SetArgs(append([]string{"setup", "folderr-db"}, args...))
	_, err := rootCmd.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Fatalf("Command folderr-cli setup folderr-db %v failed because of error, %v", args[0], err)
	}
	t.Log(strings.Contains(actual.String(), "Folderr appears to be setup"))
	if !strings.Contains(actual.String(), "Folderr appears to be setup") {
		t.Fatal("Could not find pre-setup version.")
	}
}

func TestSetupFolderrDbCleanup(t *testing.T) {
	// set MONGO_URI before running this test
	if os.Getenv("MONGO_URI") == "" {
		t.Skip("No MONGO_URI environment variable provided. No DB operations available")
	}
	actual := &bytes.Buffer{}
	os.Setenv("test", "true")

	args := []string{"folderr-cli-testing", "--no-cleanup", "-v"} // name of the testing db for this project
	rootCmd.SetOut(actual)
	rootCmd.SetArgs(append([]string{"setup", "folderr-db"}, args...))

	cleanupFolderrDbCmd(args[0], ConfigDir)
	t.Log(actual.String())

	if !strings.HasPrefix(actual.String(), "Cleaned up ") {
		t.Error("Seems cleanup was useless. Why?")
	}
}
