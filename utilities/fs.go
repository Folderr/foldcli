package utilities

import (
	"errors"
	"os"
	"runtime"
	"strings"
)

func IsValidPath(input string) bool {
	if runtime.GOOS == "windows" { // Having to handle this because the NT kernel.
		if len(strings.SplitN(input, ":", 2)) < 1 {
			return false
		}
		if !strings.Contains(input, "/") && !strings.Contains(input, "\\\\") && !strings.Contains(input, "\\") {
			return false
		}
		return true
	}
	if strings.Contains(input, "/") {
		return true
	}
	return false
}

func ManipulateDir(input string) string {
	result := input
	if runtime.GOOS == "windows" && !strings.Contains(input, "\\\\") { // Because windows. Fuck you Windows.
		result = strings.ReplaceAll(result, "/", "\\")
	}
	return result
}

func CheckIfDirExists(input string) bool {
	if _, err := os.Stat(input); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return false
	}
}
