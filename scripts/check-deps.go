package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Run go list -u -m all to check for outdated dependencies
	cmd := exec.Command("go", "list", "-u", "-m", "all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	outdated := []string{}

	// Check for lines containing [v...] indicating available updates
	for _, line := range lines {
		if strings.Contains(line, "[v") && strings.Contains(line, "]") {
			outdated = append(outdated, line)
		}
	}

	if len(outdated) > 0 {
		fmt.Fprintln(os.Stderr, "Found outdated Go dependencies:")
		for _, dep := range outdated {
			fmt.Fprintln(os.Stderr, dep)
		}
		fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")
		os.Exit(1) // Fail to block push
	}

	fmt.Fprintln(os.Stderr, "All Go dependencies are up to date.")
}
