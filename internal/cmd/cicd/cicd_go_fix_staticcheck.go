// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"
	"runtime"
	"strings"

	"cryptoutil/internal/cmd/cicd/common"
	fixAll "cryptoutil/internal/cmd/cicd/fix/all"
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

	// Get Go version from runtime.
	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	// Call fix/all package.
	processed, modified, issuesFixed, err := fixAll.Fix(logger, rootDir, goVersion)
	if err != nil {
		return fmt.Errorf("go-fix-all failed: %w", err)
	}

	logger.Log(fmt.Sprintf("go-fix-all completed: processed=%d, modified=%d, fixed=%d", processed, modified, issuesFixed))

	return nil
}
