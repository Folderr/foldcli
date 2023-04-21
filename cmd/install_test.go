package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Never run parallel. It fucks up Viper
func TestInstall(t *testing.T) {
	os.Setenv("test", "true")
	actual := &bytes.Buffer{}
	cmd, args, err := rootCmd.Find([]string{"install", "--dry"})
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	// init command usage: init [directory] [repository]
	// we'll use github.com/Folderr/Docs here as its a public repository
	rootCmd.SetArgs([]string{"install", "--dry"})
	rootCmd.SetOut(actual)
	_, err = cmd.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Errorf(`Command "folderr install %v" failed because of error, %v`, args[0], err)
	}
	suffix := []string{
		"Clone successful",
		"Checkout successful",
		"Install seems to have gone correctly.",
		`To build Folderr go to "` + config.directory + `" and type "npm run build:production"`,
	}
	for _, i := range suffix {
		if !strings.Contains(actual.String(), i) {
			t.Error(
				`Command "folderr install" did not produce expected out`,
				`\nExpected `+i+` and did not get that.`,
			)
		}
	}
}
