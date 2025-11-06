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
	"errors"
	"fmt"
	"os"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

// Run executes the specified CI/CD check commands.
func Run(commands []string) error {
	logger := NewLogUtil("Run")

	doListAllFiles, err := validateCommands(commands)
	if err != nil {
		return err
	}

	logger.Log("validateCommands completed")

	var allFiles []string

	if doListAllFiles {
		allFiles, err = cryptoutilFiles.ListAllFiles(".")
		if err != nil {
			return fmt.Errorf("failed to collect files: %w", err)
		}

		logger.Log("collectAllFiles completed")
	}

	logger.Log(fmt.Sprintf("Executing %d commands", len(commands)))

	for i, command := range commands {
		logger.Log(fmt.Sprintf("Executing command: %s", command))

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
			goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
		case "go-update-all-dependencies": // Less practiced, direct & transient dependencies
			goUpdateDeps(logger, cryptoutilMagic.DepCheckAll)
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

func validateCommands(commands []string) (bool, error) {
	logger := NewLogUtil("validateCommands")

	if len(commands) == 0 {
		logger.Log("validateCommands: empty commands")

		return false, fmt.Errorf("%s", cryptoutilMagic.UsageCICD)
	}

	var errs []error

	commandCounts := make(map[string]int)

	for _, command := range commands {
		if cryptoutilMagic.ValidCommands[command] {
			commandCounts[command]++
		} else {
			errs = append(errs, fmt.Errorf("unknown command: %s\n\n%s", command, cryptoutilMagic.UsageCICD))
		}
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
		logger.Log("validateCommands: validation errors")

		return false, fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	logger.Log("validateCommands: success")

	doListAllFiles := commandCounts["all-enforce-utf8"] > 0 ||
		commandCounts["go-enforce-test-patterns"] > 0 ||
		commandCounts["go-enforce-any"] > 0 ||
		commandCounts["github-workflow-lint"] > 0

	return doListAllFiles, nil
}
