// Copyright (c) 2025 Justin Cranford

// Package format_gotest provides Go test file formatting utilities for CICD workflows.
package format_gotest

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	formatGoTestTHelper "cryptoutil/internal/apps/cicd/format_gotest/thelper"
)

// FormatterFunc is a function type for individual Go test file formatters.
// Each formatter receives a logger and root directory, returning counts and error.
type FormatterFunc func(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) (processed, modified, fixed int, err error)

// registeredFormatters holds all formatters to run as part of format-go-test.
var registeredFormatters = []struct {
	name      string
	formatter FormatterFunc
}{
	{"thelper", formatGoTestTHelper.Fix},
}

// Format runs all registered Go test file formatters on the current directory.
// Returns an error if any formatter finds issues.
func Format(logger *cryptoutilCmdCicdCommon.Logger) error {
	return FormatDir(logger, ".")
}

// FormatDir runs all registered Go test file formatters on the specified directory.
// Returns an error if any formatter finds issues.
func FormatDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Running Go test formatters...")

	var errors []error

	var totalProcessed, totalModified, totalFixed int

	for _, f := range registeredFormatters {
		logger.Log(fmt.Sprintf("Running formatter: %s", f.name))

		processed, modified, fixed, err := f.formatter(logger, rootDir)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", f.name, err))
		} else {
			totalProcessed += processed
			totalModified += modified
			totalFixed += fixed
			logger.Log(fmt.Sprintf("%s: processed=%d, modified=%d, fixed=%d", f.name, processed, modified, fixed))
		}
	}

	logger.Log(fmt.Sprintf("format-go-test totals: processed=%d, modified=%d, fixed=%d", totalProcessed, totalModified, totalFixed))

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("format-go-test completed with %d errors", len(errors)))

		return fmt.Errorf("format-go-test failed with %d errors", len(errors))
	}

	logger.Log("format-go-test completed successfully")

	return nil
}
