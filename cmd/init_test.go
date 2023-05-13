package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Never run parallel. It fucks up Viper
func TestInit(t *testing.T) {
	os.Setenv("test", "true")
	repo := "https://github.com/Folderr/Docs"
	if getToken() != "" {
		repo = "https://github.com/Folderr/Folderr"
	}
	actual := &bytes.Buffer{}
	lDir := ConfigDir
	if ConfigDir == "" {
		llDir, err := os.MkdirTemp(os.TempDir(), "Folderr-")
		if err != nil {
			t.Fatal("Failed because couldn't make temp directory")
		}
		lDir = llDir
	}
	args := []string{lDir, repo, "--dry", "-o"}
	cmd, args, err := rootCmd.Find(append([]string{"init"}, args...))
	rootCmd.SetOut(actual)
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	// init command usage: init [directory] [repository]
	// we'll use github.com/Folderr/Docs here as its a public repository

	rootCmd.SetArgs(append([]string{"init"}, args...))
	_, err = cmd.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Errorf(`Command "folderr init %v %v %v %v" failed because of error, %v`, args[0], args[1], args[2], args[3], err)
	}
	suffix := []string{"It looks like your Folderr CLI is initialized!", "No changes were made."}
	if !strings.Contains(actual.String(), suffix[0]) || !strings.Contains(actual.String(), suffix[1]) {
		t.Errorf(`Unexpected output from "folderr init %v %v %v %v"`, args[0], args[1], args[2], args[3])
	}
}
