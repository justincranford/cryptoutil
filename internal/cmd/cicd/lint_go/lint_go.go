// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

// LinterFunc is a function type for individual Go linters.
// Each linter receives a logger, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger) error

// registeredLinters holds all linters to run as part of lint-go.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"circular-deps", checkCircularDeps},
}

// Lint runs all registered Go linters.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Running Go linters...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-go completed with %d errors", len(errors)))

		return fmt.Errorf("lint-go failed with %d errors", len(errors))
	}

	logger.Log("lint-go completed successfully")

	return nil
}
