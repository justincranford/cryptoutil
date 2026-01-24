// Copyright (c) 2025 Justin Cranford
//
//

// Package main provides the CICD utilities entry point.
package main

import (
	"fmt"
	"os"

	cryptoutilCmdCicd "cryptoutil/internal/cmd/cicd"
)

// getUsage returns the usage information for the cicd command.
func getUsage() string {
	return "Usage: cicd <command> [command...]\n\nCommands:\n  go-update-direct-dependencies    - Check direct Go dependencies only\n  go-update-all-dependencies       - Check all Go dependencies (direct + transitive)\n  go-check-circular-package-dependencies          - Check for circular dependencies in Go packages\n  github-workflow-lint             - Validate GitHub Actions workflow naming and structure\n  go-enforce-any                        - Custom Go source code fixes (any -> any, etc.)\n  go-enforce-test-patterns            - Enforce test patterns (UUIDv7 usage, testify assertions)\n  all-enforce-utf8            - Enforce UTF-8 encoding without BOM\n\nExamples:\n  cicd go-update-direct-dependencies\n  cicd go-update-all-dependencies\n  cicd go-check-circular-package-dependencies\n  cicd github-workflow-lint\n  cicd go-enforce-any\n  cicd go-enforce-test-patterns\n  cicd all-enforce-utf8\n  cicd go-update-direct-dependencies github-workflow-lint\n"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, getUsage())
		os.Exit(1)
	}

	// Extract commands (skip program name)
	commands := os.Args[1:]

	// Run the CI/CD checks
	if err := cryptoutilCmdCicd.Run(commands); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
