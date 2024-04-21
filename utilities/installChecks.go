package utilities

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// Checks if Folderr is installed
//
// May not be compliant with future versions of Folderr. Oh well.
func IsFolderrInstalled(directory string) (bool, error) {
	fsExists := CheckIfDirExists(directory)
	if !fsExists {
		return false, fmt.Errorf("directory %q not found", directory)
	}

	_, err := os.Stat(filepath.Join(directory, "package.json"))

	// "hey does the one thing we need to install all of the right dependencies exist?"
	if err != nil {
		// "no? the fuck? this isn't Folderr"
		return false, err
	}

	// we just want to see if there's a git repository there
	_, err = git.PlainOpen(directory)
	if err != nil {
		fmt.Printf("%v\n", err)
		fmt.Println(err.Error())
		return false, err
	}

	// we will consider Folderr installed.
	return true, nil
}
