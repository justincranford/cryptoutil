// Copyright (c) 2025 Justin Cranford

// Package lint_fitness runs all registered architecture fitness functions.
// Fitness functions verify that the codebase conforms to architectural
// invariants defined in ARCHITECTURE.md.
package lint_fitness

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessAdminBindAddress "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/admin_bind_address"
	lintFitnessArchiveDetector "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/archive_detector"
	lintFitnessBannedProductNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/banned_product_names"
	lintFitnessBindAddressSafety "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/bind_address_safety"
	lintFitnessCGOFreeSQLite "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cgo_free_sqlite"
	lintFitnessCheckSkeletonPlaceholders "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/check_skeleton_placeholders"
	lintFitnessCIDCoverage "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cicd_coverage"
	lintFitnessCircularDeps "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/circular_deps"
	lintFitnessCmdMainPattern "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cmd_main_pattern"
	lintFitnessComposeDBNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_db_naming"
	lintFitnessComposeHeaderFormat "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_header_format"
	lintFitnessComposeServiceNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_service_names"
	lintFitnessCrossServiceImportIsolation "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cross_service_import_isolation"
	lintFitnessCryptoRand "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/crypto_rand"
	lintFitnessDeploymentDirCompleteness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/deployment_dir_completeness"
	lintFitnessDomainLayerIsolation "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/domain_layer_isolation"
	lintFitnessEntityRegistryCompleteness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/entity_registry_completeness"
	lintFitnessFileSizeLimits "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/file_size_limits"
	lintFitnessGenConfigInitialisms "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/gen_config_initialisms"
	lintFitnessHealthEndpointPresence "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/health_endpoint_presence"
	lintFitnessInsecureSkipVerify "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/insecure_skip_verify"
	lintFitnessLegacyDirDetection "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/legacy_dir_detection"
	lintFitnessMagicE2EComposePath "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/magic_e2e_compose_path"
	lintFitnessMagicE2EContainerNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/magic_e2e_container_names"
	lintFitnessMigrationCommentHeaders "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/migration_comment_headers"
	lintFitnessMigrationNumbering "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/migration_numbering"
	lintFitnessMigrationRangeCompliance "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/migration_range_compliance"
	lintFitnessNoHardcodedPasswords "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/no_hardcoded_passwords"
	lintFitnessNoLocalClosedDBHelper "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/no_local_closed_db_helper"
	lintFitnessNoPostgresInNonE2E "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/no_postgres_in_non_e2e"
	lintFitnessNoUnitTestRealDB "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/no_unit_test_real_db"
	lintFitnessNoUnitTestRealServer "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/no_unit_test_real_server"
	lintFitnessNonFIPSAlgorithms "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/non_fips_algorithms"
	lintFitnessOTLPServiceNamePattern "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/otlp_service_name_pattern"
	lintFitnessParallelTests "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/parallel_tests"
	lintFitnessProductStructure "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/product_structure"
	lintFitnessProductWiring "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/product_wiring"
	lintFitnessRequireAPIDir "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/require_api_dir"
	lintFitnessRequireFrameworkNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/require_framework_naming"
	lintFitnessServiceContractCompliance "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/service_contract_compliance"
	lintFitnessServiceStructure "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/service_structure"
	lintFitnessStandaloneConfigOTLPNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/standalone_config_otlp_names"
	lintFitnessStandaloneConfigPresence "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/standalone_config_presence"
	lintFitnessTestPatterns "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/test_patterns"
	lintFitnessTLSMinimumVersion "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/tls_minimum_version"
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
	{"gen-config-initialisms", lintFitnessGenConfigInitialisms.Check},
	{"require-api-dir", lintFitnessRequireAPIDir.Check},
	{"require-framework-naming", lintFitnessRequireFrameworkNaming.Check},
	// New fitness checks (added in Phase 1 of framework-v4).
	{"otlp-service-name-pattern", lintFitnessOTLPServiceNamePattern.Check},
	// New fitness checks (added in Phase 2 of framework-v4).
	{"entity-registry-completeness", lintFitnessEntityRegistryCompleteness.Check},
	// New fitness checks (added in Phase 3 of framework-v4).
	{"banned-product-names", lintFitnessBannedProductNames.Check},
	{"legacy-dir-detection", lintFitnessLegacyDirDetection.Check},
	// New fitness checks (added in Phase 4 of framework-v4).
	{"deployment-dir-completeness", lintFitnessDeploymentDirCompleteness.Check},
	// New fitness checks (added in Phase 5 of framework-v4).
	{"compose-header-format", lintFitnessComposeHeaderFormat.Check},
	{"compose-service-names", lintFitnessComposeServiceNames.Check},
	{"compose-db-naming", lintFitnessComposeDBNaming.Check},
	// New fitness checks (added in Phase 6 of framework-v4).
	{"magic-e2e-container-names", lintFitnessMagicE2EContainerNames.Check},
	{"magic-e2e-compose-path", lintFitnessMagicE2EComposePath.Check},
	// New fitness checks (added in Phase 7 of framework-v4).
	{"standalone-config-presence", lintFitnessStandaloneConfigPresence.Check},
	{"standalone-config-otlp-names", lintFitnessStandaloneConfigOTLPNames.Check},
	// New fitness checks (added in Phase 8 of framework-v4).
	{"migration-comment-headers", lintFitnessMigrationCommentHeaders.Check},
	// New fitness check: validates cicd commands are covered in action, pre-commit, and CI workflow.
	{"cicd-coverage", lintFitnessCIDCoverage.Check},
	// New fitness checks (added in Phase 6 of framework-v5).
	{"archive-detector", lintFitnessArchiveDetector.Check},
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
