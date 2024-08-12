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
	// this is broken. will fix later
	// os.Setenv(utilities.Constants.EnvPrefix+"CFG_TEMPDIR", dir)
	if err != nil {
		t.Fatal("Failed due to error", err)
	}
	_, sharedConfig, _, err = utilities.ReadConfig(dir, true)
	os.Setenv(utilities.Constants.EnvPrefix+"CFG_TEMPDIR", dir)
	os.Setenv(utilities.Constants.EnvPrefix+"FLDRR_TEMPDIR", sharedConfig.Directory)
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
	buildCmd := "npm run build:tsc"
	if strings.Contains(actual.String(), "SWC appears to be installed") || strings.Contains(actual.String(), "Both SWC and TypeScript are installed") {
		buildCmd = "npm run build"
	}
	suffix := []string{
		"Clone successful",
		"Checkout successful",
		"Install seems to have gone correctly.",
		`To build Folderr go to "` + sharedConfig.Directory + `" and type "` + buildCmd + `"`,
	}
	for _, i := range suffix {
		if !strings.Contains(actual.String(), i) {
			t.Error(
				`Command "`+utilities.Constants.RootCmdName+` install folderr" did not produce expected out`,
				`\nExpected `+i+` and did not get that.`,
			)
		}
	}

	t.Cleanup(func() {
		cfgTemp := os.Getenv(utilities.Constants.EnvPrefix + "CFG_TEMPDIR")
		fldrrTemp := os.Getenv(utilities.Constants.EnvPrefix + "FLDRR_TEMPDIR")

		if cfgTemp != "" {
			err := os.RemoveAll(cfgTemp)
			if err != nil {
				t.Logf("Ran into error when removing config directories: %v", err.Error())
			}
		}
		err = os.RemoveAll(fldrrTemp)
		if err != nil {
			t.Logf("Ran into error when removing folderr directories: %v", err.Error())
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Logf("Ran into error when removing folderr directories: %v", err.Error())
		}
		os.Unsetenv(utilities.Constants.EnvPrefix + "FLDRR_TEMPDIR")
		os.Unsetenv(utilities.Constants.EnvPrefix + "CFG_TEMPDIR")
	})
}
