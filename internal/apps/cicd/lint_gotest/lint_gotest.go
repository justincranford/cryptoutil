// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"fmt"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoTestBindAddressSafety "cryptoutil/internal/apps/cicd/lint_gotest/bind_address_safety"
	lintGoTestNoHardcodedPasswords "cryptoutil/internal/apps/cicd/lint_gotest/no_hardcoded_passwords"
	lintGoTestParallelTests "cryptoutil/internal/apps/cicd/lint_gotest/parallel_tests"
	lintGoTestRequireOverAssert "cryptoutil/internal/apps/cicd/lint_gotest/require_over_assert"
	lintGoTestTestPatterns "cryptoutil/internal/apps/cicd/lint_gotest/test_patterns"
)

// LinterFunc is a function type for individual Go test file linters.
// Each linter receives a logger and a list of test files, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error

// registeredLinters holds all linters to run as part of lint-go-test.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"test-patterns", lintGoTestTestPatterns.Check},
	{"bind-address-safety", lintGoTestBindAddressSafety.Check},
	{"require-over-assert", lintGoTestRequireOverAssert.Check},
	{"parallel-tests", lintGoTestParallelTests.Check},
	{"no-hardcoded-passwords", lintGoTestNoHardcodedPasswords.Check},
}

// Lint runs all registered Go test file linters.
// It filters the provided files to only include *_test.go files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running Go test linters...")

	// Filter to *_test.go files only from Go files.
	testFiles := filterTestFiles(filesByExtension)

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

// filterTestFiles returns only *_test.go files from the Go files in the map.
func filterTestFiles(filesByExtension map[string][]string) []string {
	var testFiles []string

	// Get Go files and filter for test files.
	for _, f := range filesByExtension["go"] {
		if strings.HasSuffix(f, "_test.go") {
			testFiles = append(testFiles, f)
		}
	}

	return testFiles
}
