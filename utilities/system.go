package utilities

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func FindSystemCommandVersion(w io.Writer, command string, hasPrefix bool, prefix string) (string, error) {
	execCmd, err := FindSystemCommand(w, command, []string{"--version"})
	if err != nil {
		return "", err
	}
	output, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "It would seem I encountered an error. Is %v installed?\n", command)
		fmt.Fprintln(w, "Here's the error:", err)
		return "", err
	}
	fmt.Fprintf(w, "Printing output from executing \"%v -v\"\n", command)
	// transform output into integers.
	out := string(output)
	out = strings.TrimSpace(out)
	fmt.Fprintln(w, out)
	if !hasPrefix {
		return out, nil
	}
	if strings.HasPrefix(out, prefix) {
		out, _ = strings.CutPrefix(out, prefix)
	} else {
		fmt.Fprintf(w, "Got unexpected output from running \"%v -v\". Contact developers.\n", command)
		fmt.Fprintln(w, "Contact developers at https://github.com/Folderr/Folderr-CLI/issues")
		fmt.Fprintln(w, "Output:", out)
		return "", nil
	}
	return out, nil
}

func FindSystemCommand(w io.Writer, command string, args []string) (*exec.Cmd, error) {
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		fmt.Fprintf(w, "I can't find %v. Is %v installed?\n", command, command)
		println("Error for debug purposes:", err)
		return nil, err
	}
	return exec.Command(cmdPath, args...), nil
}
