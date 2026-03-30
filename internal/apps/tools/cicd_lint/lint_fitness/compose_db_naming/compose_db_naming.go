// Copyright (c) 2025 Justin Cranford

// Package compose_db_naming validates that every product-service compose.yml
// has the PostgreSQL database service named "{PS-ID}-db-postgres-1" with:
//   - container_name: {PS-ID}-postgres
//   - hostname: {PS-ID}-postgres
//
// This ensures the PostgreSQL container naming matches the canonical PS-ID and
// prevents drift when services are renamed.
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

// serviceConfig represents the relevant fields of a compose service definition.
type serviceConfig struct {
ContainerName string `yaml:"container_name"`
Hostname      string `yaml:"hostname"`
}

// composeFile represents the top-level structure of a compose.yml file.
type composeFile struct {
Services map[string]serviceConfig `yaml:"services"`
}

// Check validates compose db naming from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir validates compose db naming under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking compose db naming...")

var violations []string

for _, ps := range lintFitnessRegistry.AllProductServices() {
v := checkDBNaming(rootDir, ps.PSID)
violations = append(violations, v...)
}

if len(violations) > 0 {
return fmt.Errorf("compose db naming violations:\n%s", strings.Join(violations, "\n"))
}

logger.Log("compose-db-naming: all 10 product-services have correct PostgreSQL container names")

return nil
}

// checkDBNaming verifies the PostgreSQL service in the compose.yml has the
// correct container_name and hostname.
func checkDBNaming(rootDir, psID string) []string {
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

	dbServiceName := lintFitnessRegistry.DBServiceName(psID)
svc, ok := cf.Services[dbServiceName]

if !ok {
return []string{fmt.Sprintf("%s: deployments/%s/compose.yml missing service %q", psID, psID, dbServiceName)}
}

expectedName := lintFitnessRegistry.PostgresServiceName(psID)

if svc.ContainerName != expectedName {
violations = append(violations, fmt.Sprintf(
"%s: deployments/%s/compose.yml service %s: container_name: got %q, want %q",
psID, psID, dbServiceName, svc.ContainerName, expectedName,
))
}

if svc.Hostname != expectedName {
violations = append(violations, fmt.Sprintf(
"%s: deployments/%s/compose.yml service %s: hostname: got %q, want %q",
psID, psID, dbServiceName, svc.Hostname, expectedName,
))
}

return violations
}
