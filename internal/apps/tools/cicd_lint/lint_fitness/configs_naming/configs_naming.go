// Copyright (c) 2025 Justin Cranford

// Package configs_naming validates that the configs/ directory structure follows
// the canonical hierarchy:
//
//   - configs/{suite}/                  - suite-level configs (e.g. cryptoutil/)
//   - configs/{product}/                - product-level configs
//   - configs/{product}/{service}/      - service-level configs
//
// Top-level directories must be a known suite or product ID.
// Second-level directories that contain files prefixed with "{product}-" must
// be a known service for that product. Directories without such files (e.g.
// configs/identity/policies/) are treated as product-level special directories
// and are allowed.
package configs_naming

import (
"fmt"
"io/fs"
"os"
"path/filepath"
"strings"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// Check validates configs/ directory structure from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir validates configs/ directory structure under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking configs/ naming and structure...")

violations, err := FindViolationsInDir(rootDir)
if err != nil {
return fmt.Errorf("failed to check configs naming: %w", err)
}

if len(violations) > 0 {
return fmt.Errorf("configs/ naming violations:\n%s", strings.Join(violations, "\n"))
}

logger.Log("configs-naming: configs/ structure is valid")

return nil
}

// FindViolationsInDir scans configs/ under rootDir and returns all naming violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
suiteIDs, productIDs, servicesByProduct := buildRegistrySets()

configsDir := filepath.Join(rootDir, "configs")

topEntries, err := os.ReadDir(configsDir)
if err != nil {
return nil, fmt.Errorf("failed to read configs/ directory: %w", err)
}

var violations []string

for _, entry := range topEntries {
if !entry.IsDir() {
continue
}

name := entry.Name()

if suiteIDs[name] {
continue
}

if !productIDs[name] {
violations = append(violations, fmt.Sprintf("configs/%s: unknown product or suite directory (not in entity registry)", name))

continue
}

v := checkProductServiceDirs(configsDir, name, servicesByProduct[name])
violations = append(violations, v...)
}

return violations, nil
}

// checkProductServiceDirs validates second-level directories under configs/{product}/.
// Returns violations for unknown service directories that contain PS-ID-prefixed files.
func checkProductServiceDirs(configsDir, product string, validServices map[string]bool) []string {
productDir := filepath.Join(configsDir, product)

entries, _ := os.ReadDir(productDir)

var violations []string

for _, entry := range entries {
if !entry.IsDir() {
continue
}

service := entry.Name()

if validServices[service] {
continue
}

if hasProductPrefixFiles(filepath.Join(productDir, service), product+"-") {
violations = append(violations, fmt.Sprintf("configs/%s/%s: unknown service directory contains %s-* files (not in entity registry)", product, service, product))
}
}

return violations
}

// hasProductPrefixFiles returns true if dir contains any .yml file starting with prefix.
func hasProductPrefixFiles(dir, prefix string) bool {
found := false

_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
if err != nil {
return err
}

if d.IsDir() {
return nil
}

if strings.HasPrefix(d.Name(), prefix) && strings.HasSuffix(d.Name(), ".yml") {
found = true

return filepath.SkipAll
}

return nil
})

return found
}

// buildRegistrySets constructs lookup maps from the entity registry.
func buildRegistrySets() (suiteIDs map[string]bool, productIDs map[string]bool, servicesByProduct map[string]map[string]bool) {
suiteIDs = make(map[string]bool)

for _, s := range lintFitnessRegistry.AllSuites() {
suiteIDs[s.ID] = true
}

productIDs = make(map[string]bool)

for _, p := range lintFitnessRegistry.AllProducts() {
productIDs[p.ID] = true
}

servicesByProduct = make(map[string]map[string]bool)

for _, ps := range lintFitnessRegistry.AllProductServices() {
if servicesByProduct[ps.Product] == nil {
servicesByProduct[ps.Product] = make(map[string]bool)
}

servicesByProduct[ps.Product][ps.Service] = true
}

return suiteIDs, productIDs, servicesByProduct
}
