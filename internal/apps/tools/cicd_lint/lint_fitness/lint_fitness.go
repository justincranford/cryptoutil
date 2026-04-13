// Copyright (c) 2025 Justin Cranford

// Package lint_fitness runs all registered architecture fitness functions.
// Fitness functions verify that the codebase conforms to architectural
// invariants defined in ENG-HANDBOOK.md.
package lint_fitness

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessAdminBindAddress "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/admin_bind_address"
	lintFitnessAPIPathRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/api_path_registry"
	lintFitnessArchiveDetector "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/archive_detector"
	lintFitnessBannedProductNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/banned_product_names"
	lintFitnessBindAddressSafety "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/bind_address_safety"
	lintFitnessCGOFreeSQLite "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cgo_free_sqlite"
	lintFitnessCheckSkeletonPlaceholders "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/check_skeleton_placeholders"
	lintFitnessCIDCoverage "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cicd_coverage"
	lintFitnessCircularDeps "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/circular_deps"
	lintFitnessCmdAntiPattern "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cmd_anti_pattern"
	lintFitnessCmdEntryWhitelist "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cmd_entry_whitelist"
	lintFitnessCmdMainPattern "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cmd_main_pattern"
	lintFitnessComposeDBNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_db_naming"
	lintFitnessComposeEntrypointUniformity "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_entrypoint_uniformity"
	lintFitnessComposeHeaderFormat "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_header_format"
	lintFitnessComposePortFormula "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_port_formula"
	lintFitnessComposeServiceNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_service_names"
	lintFitnessComposeTierOverrideIntegrity "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/compose_tier_override_integrity"
	lintFitnessConfigOverlayFreshness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/config_overlay_freshness"
	lintFitnessConfigRules "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/config_rules"
	lintFitnessConfigsDeploymentsConsistency "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/configs_deployments_consistency"
	lintFitnessConfigsEmptyDir "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/configs_empty_dir"
	lintFitnessConfigsNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/configs_naming"
	lintFitnessCrossServiceImportIsolation "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cross_service_import_isolation"
	lintFitnessDatabaseKeyUniformity "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/database_key_uniformity"
	lintFitnessDeploymentDirCompleteness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/deployment_dir_completeness"
	lintFitnessDockerfileHealthcheck "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/dockerfile_healthcheck"
	lintFitnessDockerfileLabels "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/dockerfile_labels"
	lintFitnessDomainLayerIsolation "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/domain_layer_isolation"
	lintFitnessEntityRegistryCompleteness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/entity_registry_completeness"
	lintFitnessEntityRegistrySchema "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/entity_registry_schema"
	lintFitnessFileSizeLimits "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/file_size_limits"
	lintFitnessFitnessRegistryCompleteness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/fitness_registry_completeness"
	lintFitnessGenConfigInitialisms "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/gen_config_initialisms"
	lintFitnessHealthEndpointPresence "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/health_endpoint_presence"
	lintFitnessHealthPathCompleteness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/health_path_completeness"
	lintFitnessImportAliasFormula "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/import_alias_formula"
	lintFitnessInfraToolNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/infra_tool_naming"
	lintFitnessInsecureSkipVerify "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/insecure_skip_verify"
	lintFitnessLegacyDirDetection "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/legacy_dir_detection"
	lintFitnessMagicConstantLocation "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/magic_constant_location"
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
	lintFitnessOTLPServiceNamePattern "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/otlp_service_name_pattern"
	lintFitnessParallelTests "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/parallel_tests"
	lintFitnessPKICAProfileSchema "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/pki_ca_profile_schema"
	lintFitnessProductStructure "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/product_structure"
	lintFitnessProductWiring "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/product_wiring"
	lintFitnessRequireAPIDir "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/require_api_dir"
	lintFitnessRequireFrameworkNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/require_framework_naming"
	lintFitnessRootJunkDetection "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/root_junk_detection"
	lintFitnessSecretContent "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/secret_content"
	lintFitnessSecretNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/secret_naming"
	lintFitnessSecretsCompliance "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/secrets_compliance"
	lintFitnessServiceContractCompliance "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/service_contract_compliance"
	lintFitnessServiceStructure "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/service_structure"
	lintFitnessStandaloneConfigOTLPNames "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/standalone_config_otlp_names"
	lintFitnessStandaloneConfigPresence "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/standalone_config_presence"
	lintFitnessSubcommandCompleteness "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/subcommand_completeness"
	lintFitnessTemplateConsistency "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/template_consistency"
	lintFitnessTemplateDrift "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/template_drift"
	lintFitnessTestFileSuffixStructure "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/test_file_suffix_structure"
	lintFitnessTestPatterns "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/test_patterns"
	lintFitnessTLSMinimumVersion "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/tls_minimum_version"
	lintFitnessUnsealSecretContent "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/unseal_secret_content"
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
	{"insecure-skip-verify", lintFitnessInsecureSkipVerify.Check},
	{"migration-numbering", lintFitnessMigrationNumbering.Check},
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
	// New fitness checks (added in Phase 1 of framework-v7).
	{"entity-registry-schema", lintFitnessEntityRegistrySchema.Check},
	// New fitness checks (added in Phase 2 of framework-v7).
	{"fitness-registry-completeness", lintFitnessFitnessRegistryCompleteness.Check},
	{"test-file-suffix-structure", lintFitnessTestFileSuffixStructure.Check},
	{"import-alias-formula", lintFitnessImportAliasFormula.Check},
	{"pki-ca-profile-schema", lintFitnessPKICAProfileSchema.Check},
	// New fitness checks (added in Phase 3 of framework-v4).
	{"banned-product-names", lintFitnessBannedProductNames.Check},
	{"legacy-dir-detection", lintFitnessLegacyDirDetection.Check},
	// New fitness checks (added in Phase 4 of framework-v4).
	{"deployment-dir-completeness", lintFitnessDeploymentDirCompleteness.Check},
	// New fitness checks (added in Phase 5 of framework-v4).
	{"compose-header-format", lintFitnessComposeHeaderFormat.Check},
	{"compose-service-names", lintFitnessComposeServiceNames.Check},
	{"compose-tier-override-integrity", lintFitnessComposeTierOverrideIntegrity.Check},
	{"compose-port-formula", lintFitnessComposePortFormula.Check},
	{"compose-db-naming", lintFitnessComposeDBNaming.Check},
	{"compose-entrypoint-uniformity", lintFitnessComposeEntrypointUniformity.Check},
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
	{"cmd-anti-pattern", lintFitnessCmdAntiPattern.Check},
	{"configs-deployments-consistency", lintFitnessConfigsDeploymentsConsistency.Check},
	{"configs-empty-dir", lintFitnessConfigsEmptyDir.Check},
	{"configs-naming", lintFitnessConfigsNaming.Check},                    // New fitness checks (added in Phase 4 of framework-v7).
	{"config-overlay-freshness", lintFitnessConfigOverlayFreshness.Check}, // New fitness checks (added in Phase 8 of framework-v6).
	{"database-key-uniformity", lintFitnessDatabaseKeyUniformity.Check},
	{"dockerfile-healthcheck", lintFitnessDockerfileHealthcheck.Check},
	{"dockerfile-labels", lintFitnessDockerfileLabels.Check},
	{"secret-content", lintFitnessSecretContent.Check},
	{"secret-naming", lintFitnessSecretNaming.Check},
	{"secrets-compliance", lintFitnessSecretsCompliance.Check},
	{"unseal-secret-content", lintFitnessUnsealSecretContent.Check},
	// New fitness checks (added in Phase 6 of framework-v7).
	{"api-path-registry", lintFitnessAPIPathRegistry.Check},
	{"health-path-completeness", lintFitnessHealthPathCompleteness.Check},
	{"subcommand-completeness", lintFitnessSubcommandCompleteness.Check},
	// New fitness checks (added in documentation-audit pass).
	{"cmd-entry-whitelist", lintFitnessCmdEntryWhitelist.Check},
	{"infra-tool-naming", lintFitnessInfraToolNaming.Check},
	{"magic-constant-location", lintFitnessMagicConstantLocation.Check},
	{"root-junk-detection", lintFitnessRootJunkDetection.Check},
	{"template-consistency", lintFitnessTemplateConsistency.Check},
	// New fitness checks (added in Phase 8 of framework-v9; rewritten in framework-v10).
	{"template-compliance", lintFitnessTemplateDrift.CheckTemplateCompliance},
	// New fitness checks (added in Phase 8 of framework-v9): supplementary config rules.
	{"config-key-naming", lintFitnessConfigRules.CheckKeyNaming},
	{"config-header-identity", lintFitnessConfigRules.CheckHeaderIdentity},
	{"config-instance-minimal", lintFitnessConfigRules.CheckInstanceMinimal},
	{"config-common-complete", lintFitnessConfigRules.CheckCommonComplete},
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
		for _, e := range errors {
			logger.Log(fmt.Sprintf("FAILED: %s", e))
		}

		logger.Log(fmt.Sprintf("lint-fitness completed with %d errors", len(errors)))

		return fmt.Errorf("lint-fitness failed with %d errors", len(errors))
	}

	logger.Log("lint-fitness completed successfully")

	return nil
}
