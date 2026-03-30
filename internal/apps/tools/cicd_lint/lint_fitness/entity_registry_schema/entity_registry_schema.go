// Copyright (c) 2025 Justin Cranford

// Package entity_registry_schema validates the structural correctness of the
// canonical entity registry YAML file (api/cryptosuite-registry/registry.yaml).
//
// Validation runs through the registry loader, which enforces:
//   - Required fields present and non-empty
//   - PS-ID equals product + "-" + service
//   - internal_apps_dir matches PS-ID with trailing slash
//   - No duplicate suites, products, or PS-IDs
//   - No overlapping base_ports or pg_host_ports
//   - No overlapping or invalid migration_range_start/end values
package entity_registry_schema

import (
	"fmt"
	"path/filepath"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// Check validates the entity registry schema from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates the entity registry schema under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking entity registry schema...")

	path := filepath.Join(rootDir, "api", "cryptosuite-registry", "registry.yaml")

	if _, err := lintFitnessRegistry.LoadRegistry(path); err != nil {
		return fmt.Errorf("entity registry schema violations: %w", err)
	}

	logger.Log("entity-registry-schema: registry.yaml is structurally valid")

	return nil
}
