// Copyright (c) 2025 Justin Cranford

// Package compose_service_names validates that every product-service compose.yml
// contains the four required service definitions:
//   - {PS-ID}-app-sqlite-1
//   - {PS-ID}-app-postgres-1
//   - {PS-ID}-app-postgres-2
//   - {PS-ID}-db-postgres-1
//
// This ensures compose service naming matches the canonical PS-ID and prevents
// drift when services are renamed.
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

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkServiceNames(rootDir, ps.PSID)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose service name violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-service-names: all 10 product-services have required compose service names")

	return nil
}

// checkServiceNames verifies the 4 required service names are present in the compose.yml.
func checkServiceNames(rootDir, psID string) []string {
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

	for _, svc := range requiredServices {
		if _, ok := cf.Services[svc]; !ok {
			violations = append(violations, fmt.Sprintf(
				"%s: deployments/%s/compose.yml missing required service %q",
				psID, psID, svc,
			))
		}
	}

	return violations
}
