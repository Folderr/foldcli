package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Never run parallel. It fucks up viper.
func TestRoot(t *testing.T) {
	os.Setenv("test", "true")
	actual := &bytes.Buffer{}
	rootCmd.SetOut(actual)
	rootCmd.SetArgs([]string{"--dry"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatal("Root Command failed with error", err)
	}
	prefix := strings.HasPrefix(actual.String(), rootCmd.Long)
	suffix := strings.Contains(actual.String(), `Use "folderr [command] --help" for more information about a command.`)

	if !prefix || !suffix {
		t.Log(actual.String())
		t.Fatal("Root Command does not output expected help command")
	}
}
