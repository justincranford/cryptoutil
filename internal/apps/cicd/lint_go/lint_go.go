// Copyright (c) 2025 Justin Cranford

// Package lint_go runs all registered Go linters.
package lint_go

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCircularDeps "cryptoutil/internal/apps/cicd/lint_go/circular_deps"
	lintGoCGOFreeSQLite "cryptoutil/internal/apps/cicd/lint_go/cgo_free_sqlite"
	lintGoCmdMainPattern "cryptoutil/internal/apps/cicd/lint_go/cmd_main_pattern"
	lintGoCryptoRand "cryptoutil/internal/apps/cicd/lint_go/crypto_rand"
	lintGoInsecureSkipVerify "cryptoutil/internal/apps/cicd/lint_go/insecure_skip_verify"
	lintGoLeftoverCoverage "cryptoutil/internal/apps/cicd/lint_go/leftover_coverage"
	lintGoMagicAliases "cryptoutil/internal/apps/cicd/lint_go/magic_aliases"
	lintGoMagicDuplicates "cryptoutil/internal/apps/cicd/lint_go/magic_duplicates"
	lintGoMagicUsage "cryptoutil/internal/apps/cicd/lint_go/magic_usage"
	lintGoMigrationNumbering "cryptoutil/internal/apps/cicd/lint_go/migration_numbering"
	lintGoNonFIPSAlgorithms "cryptoutil/internal/apps/cicd/lint_go/non_fips_algorithms"
	lintGoNoUnaliasedCryptoutilImports "cryptoutil/internal/apps/cicd/lint_go/no_unaliased_cryptoutil_imports"
	lintGoProductStructure "cryptoutil/internal/apps/cicd/lint_go/product_structure"
	lintGoProductWiring "cryptoutil/internal/apps/cicd/lint_go/product_wiring"
	lintGoServiceStructure "cryptoutil/internal/apps/cicd/lint_go/service_structure"
	lintGoTestPresence "cryptoutil/internal/apps/cicd/lint_go/test_presence"
)

// LinterFunc is a function type for individual Go linters.
// Each linter receives a logger, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger) error

// registeredLinters holds all linters to run as part of lint-go.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"circular-deps", lintGoCircularDeps.Check},
	{"cgo-free-sqlite", lintGoCGOFreeSQLite.Check},
	{"cmd-main-pattern", lintGoCmdMainPattern.Check},
	{"non-fips-algorithms", lintGoNonFIPSAlgorithms.Check},
	{"no-unaliased-cryptoutil-imports", lintGoNoUnaliasedCryptoutilImports.Check},
	{"crypto-rand", lintGoCryptoRand.Check},
	{"insecure-skip-verify", lintGoInsecureSkipVerify.Check},
	{"leftover-coverage", lintGoLeftoverCoverage.Check},
	{"magic-aliases", lintGoMagicAliases.Check},
	{"magic-duplicates", lintGoMagicDuplicates.Check},
	{"magic-usage", lintGoMagicUsage.Check},
	{"migration-numbering", lintGoMigrationNumbering.Check},
	{"product-structure", lintGoProductStructure.Check},
	{"product-wiring", lintGoProductWiring.Check},
	{"service-structure", lintGoServiceStructure.Check},
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
