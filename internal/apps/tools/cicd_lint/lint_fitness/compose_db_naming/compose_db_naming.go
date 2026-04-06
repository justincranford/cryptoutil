// Copyright (c) 2025 Justin Cranford

// Package compose_db_naming validates that no product-service compose.yml
// contains a per-PS-ID PostgreSQL service "{PS-ID}-db-postgres-1".
//
// After the shared-postgres migration (Framework v8), per-PS-ID postgres services
// were replaced by the shared postgres-leader/postgres-follower services defined
// in deployments/shared-postgres/compose.yml.  This linter acts as a regression
// guard to ensure the old pattern is never re-introduced.
package compose_db_naming

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

// Check validates compose db naming from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates that no PS-ID compose file contains a per-PS-ID DB service.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking compose db naming...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkNoPerPSIDDB(rootDir, ps.PSID)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose db naming violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-db-naming: all 10 product-services correctly omit per-PS-ID postgres (shared-postgres tier)")

	return nil
}

// checkNoPerPSIDDB verifies that the PS-ID compose.yml does NOT contain
// the legacy "{PS-ID}-db-postgres-1" service.  PostgreSQL is now provided by
// the shared-postgres tier and included via include: directives.
func checkNoPerPSIDDB(rootDir, psID string) []string {
	composePath := filepath.Join(rootDir, "deployments", psID, "compose.yml")

	data, err := os.ReadFile(composePath)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read deployments/%s/compose.yml: %v", psID, psID, err)}
	}

	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return []string{fmt.Sprintf("%s: cannot parse deployments/%s/compose.yml: %v", psID, psID, err)}
	}

	dbServiceName := lintFitnessRegistry.DBServiceName(psID)
	if _, ok := cf.Services[dbServiceName]; ok {
		return []string{fmt.Sprintf(
			"%s: deployments/%s/compose.yml must not contain per-PS-ID service %q (use shared-postgres tier instead)",
			psID, psID, dbServiceName,
		)}
	}

	return nil
}
