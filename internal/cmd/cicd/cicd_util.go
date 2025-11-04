// Package cicd provides common utilities for CI/CD quality control checks.
package cicd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

const timeFormat = time.RFC3339Nano

type LogUtil struct {
	startTime time.Time
}

func NewLogUtil(operation string) *LogUtil {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[CICD] start=%s\n", start.Format(timeFormat))

	return &LogUtil{startTime: start}
}

func (l *LogUtil) Log(message string) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s: %s\n", now.Sub(l.startTime), now.Format(timeFormat), message)
}

func (l *LogUtil) LogWithDetails(operationStart time.Time, message string) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v start=%s end=%s: %s\n", now.Sub(operationStart), operationStart.Format(timeFormat), now.Format(timeFormat), message)
}

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

func validateCommands(commands []string) (bool, error) {
	logger := NewLogUtil("validateCommands")

	if len(commands) == 0 {
		logger.LogWithDetails(logger.startTime, "validateCommands: empty commands")

		return false, fmt.Errorf("%s", getUsageMessage())
	}

	var errs []error

	commandCounts := make(map[string]int)

	for _, command := range commands {
		if cryptoutilMagic.ValidCommands[command] {
			commandCounts[command]++
		} else {
			errs = append(errs, fmt.Errorf("unknown command: %s\n\n%s", command, getUsageMessage()))
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
		logger.LogWithDetails(logger.startTime, "validateCommands: validation errors")

		return false, fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	logger.LogWithDetails(logger.startTime, "validateCommands: success")

	doFindAllFiles := commandCounts["all-enforce-utf8"] > 0 ||
		commandCounts["go-enforce-test-patterns"] > 0 ||
		commandCounts["go-enforce-any"] > 0 ||
		commandCounts["github-workflow-lint"] > 0

	return doFindAllFiles, nil
}

// collectAllFiles walks the now directory and collects all file paths.
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
