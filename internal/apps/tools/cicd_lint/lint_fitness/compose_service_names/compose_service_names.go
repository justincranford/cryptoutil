// Copyright (c) 2025 Justin Cranford

// Package compose_service_names validates that every product-service compose.yml
// contains the four required service definitions and no unrecognised service names:
//   - {PS-ID}-app-sqlite-1
//   - {PS-ID}-app-postgres-1
//   - {PS-ID}-app-postgres-2
//   - {PS-ID}-db-postgres-1
//
// Service names are validated via set-membership against the canonical list computed
// by ValidComposeServiceNames() and DBServiceName(), which are derived from the entity
// registry.  Any service name that is not in the computed valid set is rejected.
package compose_service_names

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// composeFile represents the top-level structure of a compose.yml file.
type composeFile struct {
	Services map[string]any `yaml:"services"`
}

// Check validates compose service names from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates compose service names under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking compose service names...")

	// Build a set of all valid service names from the registry once,
	// then reuse it for every PS-ID compose file check.
	validSet := buildValidServiceSet()

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkServiceNames(rootDir, ps.PSID, validSet)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose service name violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-service-names: all 10 product-services have required compose service names")

	return nil
}

// buildValidServiceSet computes the complete set of valid compose service names
// from the entity registry.  The set contains:
//   - All app variant names: {PS-ID}-app-sqlite-1, {PS-ID}-app-postgres-1, {PS-ID}-app-postgres-2
//   - All DB service names: {PS-ID}-db-postgres-1
func buildValidServiceSet() map[string]struct{} {
	validNames := lintFitnessRegistry.ValidComposeServiceNames()
	allPS := lintFitnessRegistry.AllProductServices()
	set := make(map[string]struct{})

	for _, name := range validNames {
		set[name] = struct{}{}
	}

	for _, ps := range allPS {
		set[lintFitnessRegistry.DBServiceName(ps.PSID)] = struct{}{}
	}

	return set
}

// checkServiceNames verifies the compose.yml for the given psID:
//  1. All 4 required services are present.
//  2. Every listed service name is a member of the valid set (set-membership check).
func checkServiceNames(rootDir, psID string, validSet map[string]struct{}) []string {
	var violations []string

	composePath := filepath.Join(rootDir, "deployments", psID, "compose.yml")

	data, err := os.ReadFile(composePath)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read deployments/%s/compose.yml: %v", psID, psID, err)}
	}

	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return []string{fmt.Sprintf("%s: cannot parse deployments/%s/compose.yml: %v", psID, psID, err)}
	}

	requiredServices := []string{
		lintFitnessRegistry.ComposeServiceName(psID, lintFitnessRegistry.ComposeVariantSQLite1),
		lintFitnessRegistry.ComposeServiceName(psID, lintFitnessRegistry.ComposeVariantPostgres1),
		lintFitnessRegistry.ComposeServiceName(psID, lintFitnessRegistry.ComposeVariantPostgres2),
		lintFitnessRegistry.DBServiceName(psID),
	}

	// Check all required services are present.
	for _, svc := range requiredServices {
		if _, ok := cf.Services[svc]; !ok {
			violations = append(violations, fmt.Sprintf(
				"%s: deployments/%s/compose.yml missing required service %q",
				psID, psID, svc,
			))
		}
	}

	// Set-membership check: for services that use the PS-ID prefix,
	// verify their full name is in the valid set.  Infrastructure services
	// like "healthcheck-secrets", "builder-{psID}", and "pki-init" that do
	// NOT carry the PS-ID prefix are excluded from this check.
	prefix := psID + "-"

	for svc := range cf.Services {
		if !strings.HasPrefix(svc, prefix) {
			continue // skip infrastructure/helper services not scoped to this PS-ID
		}

		if _, valid := validSet[svc]; !valid {
			violations = append(violations, fmt.Sprintf(
				"%s: deployments/%s/compose.yml contains unrecognised service %q (not in valid service name set)",
				psID, psID, svc,
			))
		}
	}

	return violations
}
