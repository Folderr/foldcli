package install

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Folderr/foldcli/utilities"
)

// Never run parallel. It fucks up Viper
func TestInstall(t *testing.T) {
	os.Setenv("test", "true")
	dir, err := utilities.GetConfigDir(true)
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	_, sharedConfig, _, err = utilities.ReadConfig(dir, true)
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	actual := &bytes.Buffer{}
	command, args, err := installCmd.Find([]string{"folderr", "--dry"})
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	// we'll use github.com/Folderr/Docs here as its a public repository
	command.Root().SetOut(actual)
	command.Root().SetArgs([]string{"install", "folderr", "--dry"})
	_, err = command.ExecuteC()
	t.Log(actual.String())
	if err != nil {
		t.Errorf(`Command "`+utilities.Constants.RootCmdName+` install folderr %v" failed because of error, %v`, args[0], err)
	}
	suffix := []string{
		"Clone successful",
		"Checkout successful",
		"Install seems to have gone correctly.",
		`To build Folderr go to "` + sharedConfig.Directory + `" and type "npm run build:production"`,
	}
	for _, i := range suffix {
		if !strings.Contains(actual.String(), i) {
			t.Error(
				`Command "`+utilities.Constants.RootCmdName+` install folderr" did not produce expected out`,
				`\nExpected `+i+` and did not get that.`,
			)
		}
	}
}
