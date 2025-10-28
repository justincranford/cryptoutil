// Package main provides the command-line interface for CI/CD quality control checks.
//
// This executable wraps the cicd package to provide a CLI for running automated checks
// to ensure code quality, dependency freshness, and workflow consistency.
//
// Usage:
//
//	cicd <command> [<command>...]
//	go run cmd/cicd/main.go <command> [<command>...]
//
// Examples:
//
//	cicd go-update-direct-dependencies
//	cicd github-workflow-lint
//	cicd go-update-direct-dependencies github-workflow-lint
package main

import (
	"cryptoutil/internal/cmd/cicd"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: cicd <command> [command...]\n\nCommands:\n  go-update-direct-dependencies    - Check direct Go dependencies only\n  go-update-all-dependencies       - Check all Go dependencies (direct + transitive)\n  go-check-circular-package-dependencies          - Check for circular dependencies in Go packages\n  github-workflow-lint             - Validate GitHub Actions workflow naming and structure\n  go-enforce-any                        - Custom Go source code fixes (any -> any, etc.)\n  go-enforce-test-patterns            - Enforce test patterns (UUIDv7 usage, testify assertions)\n  all-enforce-utf8            - Enforce UTF-8 encoding without BOM\n\nExamples:\n  cicd go-update-direct-dependencies\n  cicd go-update-all-dependencies\n  cicd go-check-circular-package-dependencies\n  cicd github-workflow-lint\n  cicd go-enforce-any\n  cicd go-enforce-test-patterns\n  cicd all-enforce-utf8\n  cicd go-update-direct-dependencies github-workflow-lint\n")
		os.Exit(1)
	}

	// Extract commands (skip program name)
	commands := os.Args[1:]

	// Run the CI/CD checks
	if err := cicd.Run(commands); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
