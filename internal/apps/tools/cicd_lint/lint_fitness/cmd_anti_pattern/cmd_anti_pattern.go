// Copyright (c) 2025 Justin Cranford

// Package cmd_anti_pattern validates that cmd/ directories follow the canonical
// naming convention. Every cmd/ directory must be one of:
//
//   - A PS-ID (e.g. "identity-authz", "sm-kms") from the entity registry
//   - A product name (e.g. "identity", "sm") from the entity registry
//   - The suite name (e.g. "cryptoutil")
//   - A documented infrastructure tool (cicd-lint, workflow)
//
// Directories like "identity-compose" or "sm-run" that embed a product prefix
// but are not registered PS-IDs are the anti-patterns this check prevents.
package cmd_anti_pattern

import (
"fmt"
"os"
"path/filepath"
"strings"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Check validates cmd/ directory naming from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir validates cmd/ directory naming under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking cmd/ directory naming...")

violations, err := FindViolationsInDir(rootDir)
if err != nil {
return fmt.Errorf("failed to check cmd anti-pattern: %w", err)
}

if len(violations) > 0 {
return fmt.Errorf("cmd/ anti-pattern violations:\n%s", strings.Join(violations, "\n"))
}

logger.Log("cmd-anti-pattern: all cmd/ directories follow canonical naming")

return nil
}

// FindViolationsInDir scans cmd/ under rootDir and returns naming violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
allowedNames := buildAllowedSet()

cmdDir := filepath.Join(rootDir, "cmd")

entries, err := os.ReadDir(cmdDir)
if err != nil {
return nil, fmt.Errorf("failed to read cmd/ directory: %w", err)
}

var violations []string

for _, entry := range entries {
if !entry.IsDir() {
continue
}

name := entry.Name()

if !allowedNames[name] {
violations = append(violations, fmt.Sprintf("cmd/%s: unknown cmd directory (not a PS-ID, product, suite, or documented infra tool)", name))
}
}

return violations, nil
}

// buildAllowedSet constructs the set of all permitted cmd/ directory names.
func buildAllowedSet() map[string]bool {
allowed := make(map[string]bool)

// All registered PS-IDs (e.g. "identity-authz", "sm-kms").
for _, ps := range lintFitnessRegistry.AllProductServices() {
allowed[ps.PSID] = true
}

// All registered product names (e.g. "identity", "sm").
for _, p := range lintFitnessRegistry.AllProducts() {
allowed[p.ID] = true
}

// Suite name (e.g. "cryptoutil").
for _, s := range lintFitnessRegistry.AllSuites() {
allowed[s.ID] = true
}

// Documented infrastructure tools.
allowed[cryptoutilSharedMagic.CICDCmdDirCicdLint] = true
allowed[cryptoutilSharedMagic.CICDCmdDirWorkflow] = true

return allowed
}
