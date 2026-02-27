// Copyright (c) 2025 Justin Cranford

// Package lint_text provides text linting utilities for CI/CD pipelines.
package lint_text

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintTextUTF8 "cryptoutil/internal/apps/cicd/lint_text/utf8"
)

// LinterFunc is a function type for individual text linters.
// Each linter receives a logger and a map of files by extension, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error

// registeredLinters holds all linters to run as part of lint-text.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"utf8", lintTextUTF8.Check},
}

// Lint runs all registered text linters on the provided files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running text linters...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, filesByExtension); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-text completed with %d errors", len(errors)))

		return fmt.Errorf("lint-text failed with %d errors", len(errors))
	}

	logger.Log("lint-text completed successfully")

	return nil
}
