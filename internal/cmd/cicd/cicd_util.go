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

type LogUtil struct {
	startTime time.Time
}

func NewLogUtil(operation string) *LogUtil {
	start := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] start=%s\n", start.Format(cryptoutilMagic.TimeFormat))

	return &LogUtil{startTime: start}
}

func (l *LogUtil) Log(message string) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s: %s\n", now.Sub(l.startTime), now.Format(cryptoutilMagic.TimeFormat), message)
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

func listAllFiles() ([]string, error) {
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

// writeFile writes content to a file with the specified permissions.
// If permissions is 0, it defaults to CacheFilePermissions.
// The content can be a string or []byte.
func writeFile(filePath string, bytesOrString any, permissions os.FileMode) error {
	if permissions == 0 {
		return fmt.Errorf("missing file permissions")
	}

	var data []byte
	switch v := bytesOrString.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("content must be string or []byte, got %T", bytesOrString)
	}

	err := os.WriteFile(filePath, data, permissions)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
