package install

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Folderr/foldcli/cmd"
	"github.com/Folderr/foldcli/utilities"
)

// Never run parallel. It fucks up Viper
func TestInstall(t *testing.T) {
	os.Setenv("test", "true")
	dir, err := utilities.GetConfigDir(true)
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	_, config, _, err := utilities.ReadConfig(dir, true)
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	actual := &bytes.Buffer{}
	command, args, err := cmd.RootCmd.Find([]string{"install", "--dry"})
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	// we'll use github.com/Folderr/Docs here as its a public repository
	cmd.RootCmd.SetArgs([]string{"install", "--dry"})
	cmd.RootCmd.SetOut(actual)
	_, err = command.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Errorf(`Command "`+utilities.Constants.RootCmdName+`install %v" failed because of error, %v`, args[0], err)
	}
	suffix := []string{
		"Clone successful",
		"Checkout successful",
		"Install seems to have gone correctly.",
		`To build Folderr go to "` + config.Directory + `" and type "npm run build:production"`,
	}
	for _, i := range suffix {
		if !strings.Contains(actual.String(), i) {
			t.Error(
				`Command "`+utilities.Constants.EnvPrefix+` install" did not produce expected out`,
				`\nExpected `+i+` and did not get that.`,
			)
		}
	}
}
