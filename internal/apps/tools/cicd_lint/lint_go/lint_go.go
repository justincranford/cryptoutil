// Copyright (c) 2025 Justin Cranford

// Package lint_go runs all registered Go linters.
package lint_go

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoFunctionVarRedeclaration "cryptoutil/internal/apps/tools/cicd_lint/lint_go/function_var_redeclaration"
	lintGoLeftoverCoverage "cryptoutil/internal/apps/tools/cicd_lint/lint_go/leftover_coverage"
	lintGoMagicAliases "cryptoutil/internal/apps/tools/cicd_lint/lint_go/magic_aliases"
	lintGoMagicDuplicates "cryptoutil/internal/apps/tools/cicd_lint/lint_go/magic_duplicates"
	lintGoMagicUsage "cryptoutil/internal/apps/tools/cicd_lint/lint_go/magic_usage"
	lintGoNoUnaliasedCryptoutilImports "cryptoutil/internal/apps/tools/cicd_lint/lint_go/no_unaliased_cryptoutil_imports"
	lintGoTestPresence "cryptoutil/internal/apps/tools/cicd_lint/lint_go/test_presence"
)

// LinterFunc is a function type for individual Go linters.
// Each linter receives a logger, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger) error

// registeredLinters holds all linters to run as part of lint-go.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"function-var-redeclaration", lintGoFunctionVarRedeclaration.Check},
	{"leftover-coverage", lintGoLeftoverCoverage.Check},
	{"magic-aliases", lintGoMagicAliases.Check},
	{"magic-duplicates", lintGoMagicDuplicates.Check},
	{"magic-usage", lintGoMagicUsage.Check},
	{"no-unaliased-cryptoutil-imports", lintGoNoUnaliasedCryptoutilImports.Check},
	{"test-presence", lintGoTestPresence.Check},
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
