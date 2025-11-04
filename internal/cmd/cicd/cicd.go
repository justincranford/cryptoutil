// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This package performs various automated checks to ensure code quality, dependency freshness,
// and workflow consistency. It is designed to run both locally (during development) and
// in CI/CD pipelines (via pre-push hooks and GitHub Actions).
//
// IMPORTANT: This file contains deliberate linter error patterns for testing cicd functionality.
// It MUST be excluded from all linting operations to prevent self-referencing errors.
// See .golangci.yml exclude-rules and cicd.go exclusion patterns for details.
//
// Exit Codes:
//
//	0: All checks passed
//	1: One or more checks failed (details printed to stderr)
package cicd

import (
	"fmt"
	"os"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// Run executes the specified CI/CD check commands.
// It takes a slice of command names and executes them sequentially.
// Returns an error if any command is unknown or if execution fails.
func Run(commands []string) error {
	// Create LogUtil for structured performance logging
	logger := NewLogUtil("Run")

	// Validate commands and determine if file walk is needed
	doFindAllFiles, err := validateCommands(commands)
	if err != nil {
		return err
	}

	var allFiles []string

	if doFindAllFiles {
		// Collect all files once for efficiency
		fileWalkStart := time.Now()

		allFiles, err = collectAllFiles()
		if err != nil {
			return fmt.Errorf("failed to collect files: %w", err)
		}

		logger.LogWithDetails(fileWalkStart, "filepath.Walk")
	}

	// Process all commands provided as arguments
	for i := range commands {
		command := commands[i]
		logger.Log(fmt.Sprintf("Executing command: %s", command))

		// Execute command with performance logging
		switch command {
		case "all-enforce-utf8":
			allEnforceUtf8(logger, allFiles)
		case "go-enforce-test-patterns":
			goEnforceTestPatterns(logger, allFiles)
		case "go-enforce-any":
			goEnforceAny(logger, allFiles)
		case "go-check-circular-package-dependencies":
			goCheckCircularPackageDeps(logger)
		case "go-update-direct-dependencies": // Best practice, only direct dependencies
			goUpdateDeps(logger, DepCheckDirect)
		case "go-update-all-dependencies": // Less practiced, direct & transient dependencies
			goUpdateDeps(logger, DepCheckAll)
		case "github-workflow-lint":
			checkWorkflowLint(logger, allFiles)
		}

		// Add a separator between multiple commands
		if i < len(commands)-1 {
			fmt.Fprintln(os.Stderr, "\n"+strings.Repeat("=", cryptoutilMagic.SeparatorLength)+"\n")
		}
	}

	logger.Log("Run completed")

	return nil
}
