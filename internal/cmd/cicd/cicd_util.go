// Package cicd provides common utilities for CI/CD quality control checks.
//
// This file contains shared types, constants, and utility functions used across
// different CI/CD commands. It provides common functionality for performance timing,
// file operations, command validation, and caching.
package cicd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// getUsageMessage returns the usage message for the cicd command.
func getUsageMessage() string {
	return `Usage: cicd <command> [command...]

Commands:
  all-enforce-utf8                       - Enforce UTF-8 encoding without BOM
  go-enforce-test-patterns               - Enforce test patterns (UUIDv7 usage, testify assertions)
  go-enforce-any                         - Custom Go source code fixes (any -> any, etc.)
  go-check-circular-package-dependencies - Check for circular dependencies in Go packages
  go-update-direct-dependencies          - Check direct Go dependencies only
  go-update-all-dependencies             - Check all Go dependencies (direct + transitive)
  github-workflow-lint                   - Validate GitHub Actions workflow naming and structure, and check for outdated actions`
}

// validateCommands validates the provided commands for duplicates, mutually exclusive combinations,
// and empty command lists. Returns doFindAllFiles flag and any validation error.
func validateCommands(commands []string) (bool, error) {
	// Start performance timing
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] validateCommands started at %s\n", start.Format(time.RFC3339Nano))

	// Check for empty commands first (also handles nil slices since len(nil) == 0)
	if len(commands) == 0 {
		end := time.Now()
		fmt.Fprintf(os.Stderr, "[PERF] validateCommands: start=%s end=%s duration=%v (empty commands)\n",
			start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), end.Sub(start))

		return false, fmt.Errorf("%s", getUsageMessage())
	}

	doFindAllFiles := false

	var errs []error

	commandCounts := make(map[string]int)

	// Count command occurrences and determine if file walk is needed
	for _, command := range commands {
		if cryptoutilMagic.ValidCommands[command] {
			commandCounts[command]++
		} else {
			errs = append(errs, fmt.Errorf("unknown command: %s\n\n%s", command, getUsageMessage()))
		}
	}

	// Compute doFindAllFiles after counting all commands
	if commandCounts["all-enforce-utf8"] > 0 ||
		commandCounts["go-enforce-test-patterns"] > 0 ||
		commandCounts["go-enforce-any"] > 0 ||
		commandCounts["github-workflow-lint"] > 0 {
		doFindAllFiles = true
	}

	// Check for duplicate commands
	for command, count := range commandCounts {
		if count > 1 {
			errs = append(errs, fmt.Errorf("command '%s' specified %d times - each command can only be used once", command, count))
		}
	}

	// Check for mutually exclusive commands
	if commandCounts["go-update-direct-dependencies"] > 0 && commandCounts["go-update-all-dependencies"] > 0 {
		errs = append(errs, fmt.Errorf("commands 'go-update-direct-dependencies' and 'go-update-all-dependencies' cannot be used together - choose one dependency update mode"))
	}

	if len(errs) > 0 {
		end := time.Now()
		fmt.Fprintf(os.Stderr, "[PERF] validateCommands: duration=%v start=%s end=%s (validation errors)\n",
			end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))

		return false, fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	end := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] validateCommands: duration=%v start=%s end=%s (success)\n",
		end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))

	return doFindAllFiles, nil
}

// collectAllFiles walks the current directory and collects all file paths.
// Returns a slice of all file paths found.
func collectAllFiles() ([]string, error) {
	var allFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return allFiles, nil
}
