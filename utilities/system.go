package utilities

import (
	"fmt"
	"os/exec"
	"strings"
)

func FindSystemCommandVersion(command string, hasPrefix bool, prefix string) (string, error) {
	execCmd, err := FindSystemCommand(command, []string{"-v"})
	if err != nil {
		return "", err
	}
	output, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("It would seem I encountered an error. Is %v installed?\n", command)
		println("Here's the error:", err)
		return "", err
	}
	fmt.Printf("Printing output from executing \"%v -v\"\n", command)
	// transform output into integers.
	out := string(output)
	out = strings.TrimSpace(out)
	println(out)
	if !hasPrefix {
		return out, nil
	}
	if strings.HasPrefix(out, prefix) {
		out, _ = strings.CutPrefix(out, prefix)
	} else {
		fmt.Printf("Got unexpected output from running \"%v -v\". Contact developers.\n", command)
		println("Contact developers at https://github.com/Folderr/Folderr-CLI/issues")
		println("Output:", out)
		return "", nil
	}
	return out, nil
}

func FindSystemCommand(command string, args []string) (*exec.Cmd, error) {
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		fmt.Printf("I can't find %v. Is %v installed?\n", command, command)
		println("Error for debug purposes:", err)
		return nil, err
	}
	return exec.Command(cmdPath, args...), nil
}
