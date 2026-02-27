// Copyright (c) 2025 Justin Cranford

// Package lint_golangci provides linting for golangci-lint configuration files.
package lint_golangci

import (
	"fmt"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGolangciConfig "cryptoutil/internal/apps/cicd/lint_golangci/golangci_config"
)

// LinterFunc is a function type for individual golangci-lint configuration linters.
// Each linter receives a logger and the files map, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error

// registeredLinters holds all linters to run as part of lint-golangci.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"golangci-config", lintGolangciConfig.Check},
}

// Lint runs all registered golangci-lint configuration linters.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running golangci-lint configuration linters...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, filesByExtension); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-golangci completed with %d errors", len(errors)))

		msgs := make([]string, len(errors))

		for i, e := range errors {
			msgs[i] = e.Error()
		}

		return fmt.Errorf("lint-golangci failed with %d errors: %s", len(errors), strings.Join(msgs, "; "))
	}

	logger.Log("lint-golangci completed successfully")

	return nil
}
