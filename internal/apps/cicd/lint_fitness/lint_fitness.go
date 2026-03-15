// Copyright (c) 2025 Justin Cranford

// Package lint_fitness runs all registered architecture fitness functions.
// Fitness functions verify that the codebase conforms to architectural
// invariants defined in ARCHITECTURE.md.
package lint_fitness

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintFitnessAdminBindAddress "cryptoutil/internal/apps/cicd/lint_fitness/admin_bind_address"
	lintFitnessBindAddressSafety "cryptoutil/internal/apps/cicd/lint_fitness/bind_address_safety"
	lintFitnessCGOFreeSQLite "cryptoutil/internal/apps/cicd/lint_fitness/cgo_free_sqlite"
	lintFitnessCheckSkeletonPlaceholders "cryptoutil/internal/apps/cicd/lint_fitness/check_skeleton_placeholders"
	lintFitnessCircularDeps "cryptoutil/internal/apps/cicd/lint_fitness/circular_deps"
	lintFitnessCmdMainPattern "cryptoutil/internal/apps/cicd/lint_fitness/cmd_main_pattern"
	lintFitnessCrossServiceImportIsolation "cryptoutil/internal/apps/cicd/lint_fitness/cross_service_import_isolation"
	lintFitnessCryptoRand "cryptoutil/internal/apps/cicd/lint_fitness/crypto_rand"
	lintFitnessDomainLayerIsolation "cryptoutil/internal/apps/cicd/lint_fitness/domain_layer_isolation"
	lintFitnessFileSizeLimits "cryptoutil/internal/apps/cicd/lint_fitness/file_size_limits"
	lintFitnessHealthEndpointPresence "cryptoutil/internal/apps/cicd/lint_fitness/health_endpoint_presence"
	lintFitnessInsecureSkipVerify "cryptoutil/internal/apps/cicd/lint_fitness/insecure_skip_verify"
	lintFitnessMigrationNumbering "cryptoutil/internal/apps/cicd/lint_fitness/migration_numbering"
	lintFitnessMigrationRangeCompliance "cryptoutil/internal/apps/cicd/lint_fitness/migration_range_compliance"
	lintFitnessNoHardcodedPasswords "cryptoutil/internal/apps/cicd/lint_fitness/no_hardcoded_passwords"
	lintFitnessNoLocalClosedDBHelper "cryptoutil/internal/apps/cicd/lint_fitness/no_local_closed_db_helper"
	lintFitnessNoPostgresInNonE2E "cryptoutil/internal/apps/cicd/lint_fitness/no_postgres_in_non_e2e"
	lintFitnessNoUnitTestRealDB "cryptoutil/internal/apps/cicd/lint_fitness/no_unit_test_real_db"
	lintFitnessNoUnitTestRealServer "cryptoutil/internal/apps/cicd/lint_fitness/no_unit_test_real_server"
	lintFitnessNonFIPSAlgorithms "cryptoutil/internal/apps/cicd/lint_fitness/non_fips_algorithms"
	lintFitnessParallelTests "cryptoutil/internal/apps/cicd/lint_fitness/parallel_tests"
	lintFitnessProductStructure "cryptoutil/internal/apps/cicd/lint_fitness/product_structure"
	lintFitnessProductWiring "cryptoutil/internal/apps/cicd/lint_fitness/product_wiring"
	lintFitnessServiceContractCompliance "cryptoutil/internal/apps/cicd/lint_fitness/service_contract_compliance"
	lintFitnessServiceStructure "cryptoutil/internal/apps/cicd/lint_fitness/service_structure"
	lintFitnessTestPatterns "cryptoutil/internal/apps/cicd/lint_fitness/test_patterns"
	lintFitnessTLSMinimumVersion "cryptoutil/internal/apps/cicd/lint_fitness/tls_minimum_version"
)

// LinterFunc is a function type for individual architecture fitness linters.
// Each linter receives a logger, returning an error if fitness violations found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger) error

// registeredLinters holds all architecture fitness linters to run as part of lint-fitness.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	// Architecture checks (migrated from lint_go).
	{"cgo-free-sqlite", lintFitnessCGOFreeSQLite.Check},
	{"circular-deps", lintFitnessCircularDeps.Check},
	{"cmd-main-pattern", lintFitnessCmdMainPattern.Check},
	{"crypto-rand", lintFitnessCryptoRand.Check},
	{"insecure-skip-verify", lintFitnessInsecureSkipVerify.Check},
	{"migration-numbering", lintFitnessMigrationNumbering.Check},
	{"non-fips-algorithms", lintFitnessNonFIPSAlgorithms.Check},
	{"product-structure", lintFitnessProductStructure.Check},
	{"product-wiring", lintFitnessProductWiring.Check},
	{"service-structure", lintFitnessServiceStructure.Check},
	// Architecture checks (migrated from lint_gotest).
	{"bind-address-safety", lintFitnessBindAddressSafety.Check},
	{"no-hardcoded-passwords", lintFitnessNoHardcodedPasswords.Check},
	{"parallel-tests", lintFitnessParallelTests.Check},
	{"test-patterns", lintFitnessTestPatterns.Check},
	// Architecture checks (migrated from lint_skeleton).
	{"check-skeleton-placeholders", lintFitnessCheckSkeletonPlaceholders.Check},
	// New fitness checks (added in Phase 4).
	{"cross-service-import-isolation", lintFitnessCrossServiceImportIsolation.Check},
	{"domain-layer-isolation", lintFitnessDomainLayerIsolation.Check},
	{"file-size-limits", lintFitnessFileSizeLimits.Check},
	{"health-endpoint-presence", lintFitnessHealthEndpointPresence.Check},
	{"tls-minimum-version", lintFitnessTLSMinimumVersion.Check},
	{"admin-bind-address", lintFitnessAdminBindAddress.Check},
	{"service-contract-compliance", lintFitnessServiceContractCompliance.Check},
	{"migration-range-compliance", lintFitnessMigrationRangeCompliance.Check},
	{"no-local-closed-db-helper", lintFitnessNoLocalClosedDBHelper.Check},
	{"no-postgres-in-non-e2e", lintFitnessNoPostgresInNonE2E.Check},
	{"no-unit-test-real-db", lintFitnessNoUnitTestRealDB.Check},
	{"no-unit-test-real-server", lintFitnessNoUnitTestRealServer.Check},
}

// Lint runs all registered architecture fitness linters.
// Returns an error if any linter finds violations.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Running architecture fitness functions...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running fitness linter: %s", l.name))

		if err := l.linter(logger); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-fitness completed with %d errors", len(errors)))

		return fmt.Errorf("lint-fitness failed with %d errors", len(errors))
	}

	logger.Log("lint-fitness completed successfully")

	return nil
}
