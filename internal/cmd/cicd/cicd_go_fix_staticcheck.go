// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"
	"strings"

	"cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/fix/staticcheck"
)

// goFixStaticcheckErrorStrings fixes Go error strings that violate staticcheck ST1005 rules.
// It lowercases error strings that start with uppercase letters, except for known acronyms.
func goFixStaticcheckErrorStrings(logger *common.Logger, rootDir string) error {
	logger.Log("Starting staticcheck error string fixes")

	processed, modified, issuesFixed, err := staticcheck.Fix(logger, rootDir)
	if err != nil {
		return fmt.Errorf("failed to fix staticcheck errors: %w", err)
	}

	logger.Log(fmt.Sprintf("Processed: %d files, Modified: %d files, Fixed: %d issues", processed, modified, issuesFixed))

	return nil
}

// goFixAll runs all auto-fix commands in sequence.
// This is a convenience command that orchestrates all go-fix-* commands.
func goFixAll(logger *common.Logger, rootDir string) error {
	logger.Log("Starting go-fix-all: running all auto-fix commands")

	commands := []struct {
		name string
		fn   func(*common.Logger, string) error
	}{
		{"go-fix-staticcheck-error-strings", goFixStaticcheckErrorStrings},
	}

	var errors []error
	successCount := 0

	for _, cmd := range commands {
		logger.Log(fmt.Sprintf("Running: %s", cmd.name))

		err := cmd.fn(logger, rootDir)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s failed: %w", cmd.name, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("go-fix-all completed: %d succeeded, %d failed", successCount, len(errors)))

		// Join all errors.
		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}

		return fmt.Errorf("go-fix-all failures:\n%s", strings.Join(errMsgs, "\n"))
	}

	logger.Log(fmt.Sprintf("go-fix-all completed successfully: %d commands executed", successCount))

	return nil
}
