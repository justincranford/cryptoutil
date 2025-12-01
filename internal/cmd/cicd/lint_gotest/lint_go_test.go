// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

// LinterFunc is a function type for individual Go test file linters.
// Each linter receives a logger and a list of files, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error

// registeredLinters holds all linters to run as part of lint-go-test.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"test-patterns", enforceTestPatterns},
}

// Lint runs all registered Go test file linters.
// It filters the provided files to only include *_test.go files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, allFiles []string) error {
	logger.Log("Running Go test linters...")

	// Filter to *_test.go files only.
	testFiles := filterTestFiles(allFiles)

	if len(testFiles) == 0 {
		logger.Log("lint-go-test completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to lint", len(testFiles)))

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, testFiles); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-go-test completed with %d errors", len(errors)))

		return fmt.Errorf("lint-go-test failed with %d errors", len(errors))
	}

	logger.Log("lint-go-test completed successfully")

	return nil
}

// filterTestFiles returns only *_test.go files from the input.
func filterTestFiles(allFiles []string) []string {
	var testFiles []string

	for _, f := range allFiles {
		if len(f) > 8 && f[len(f)-8:] == "_test.go" {
			testFiles = append(testFiles, f)
		}
	}

	return testFiles
}
