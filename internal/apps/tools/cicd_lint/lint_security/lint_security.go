// Copyright (c) 2025 Justin Cranford

// Package lint_security provides security linting for CI/CD pipelines.
// Sub-linters check for FIPS 140-3 compliance violations beyond what gosec covers.
package lint_security

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintSecurityBannedImports "cryptoutil/internal/apps/tools/cicd_lint/lint_security/banned_imports"
)

// LinterFunc is a function type for individual security linters.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error

// registeredLinters holds all linters to run as part of lint-security.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"banned-imports", lintSecurityBannedImports.Check},
}

// Lint runs all registered security linters on the provided files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running security linters...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, filesByExtension); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-security completed with %d errors", len(errors)))

		return fmt.Errorf("lint-security failed with %d errors", len(errors))
	}

	logger.Log("lint-security completed successfully")

	return nil
}
