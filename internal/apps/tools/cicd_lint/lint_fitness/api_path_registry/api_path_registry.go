// Copyright (c) 2025 Justin Cranford

// Package api_path_registry validates that the OpenAPI spec paths declared in each
// service's api/{ps-id}/openapi_spec*.yaml file(s) match the api_resources declared
// in the entity registry (api/cryptosuite-registry/registry.yaml).
//
// The linter reports:
//   - Paths declared in the registry but not found in any spec file (missing from spec).
//   - Paths found in the spec but not declared in the registry (undeclared in registry).
//
// Services with no api_resources in the registry are skipped because they have
// no OpenAPI spec files.
package api_path_registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilLintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// Check validates API path registry from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.ReadDir, os.ReadFile)
}

// CheckInDir validates API path registry under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) error {
	logger.Log("Checking API path registry consistency...")

	// Build psID → resources map from the registry.
	apiResourcesByPSID := make(map[string][]string)

	for _, info := range cryptoutilLintFitnessRegistry.AllAPIResources() {
		if len(info.Resources) > 0 {
			apiResourcesByPSID[info.PSID] = info.Resources
		}
	}

	var violations []string

	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		resources := apiResourcesByPSID[ps.PSID]
		if len(resources) == 0 {
			// No api_resources declared — service intentionally has no OpenAPI spec.
			continue
		}

		apiDir := filepath.Join(rootDir, "api", ps.PSID)

			specPaths, err := collectSpecPaths(apiDir, readDirFn, readFileFn)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: %v", ps.PSID, err))

			continue
		}

		registryPaths := toSet(resources)
		psViolations := comparePaths(ps.PSID, registryPaths, specPaths)
		violations = append(violations, psViolations...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("api-path-registry violations:\n%s", strings.Join(violations, "\n"))
	}

	return nil
}

// collectSpecPaths reads all openapi_spec*.yaml files in apiDir (excluding *_components.yaml)
// and returns the union of all path keys found across the spec files.
func collectSpecPaths(apiDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) (map[string]struct{}, error) {
	entries, err := readDirFn(apiDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read api directory %s: %w", apiDir, err)
	}

	result := make(map[string]struct{})
	foundSpec := false

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !isSpecFile(name) {
			continue
		}

		foundSpec = true

		filePath := filepath.Join(apiDir, name)

			paths, readErr := parseSpecPaths(filePath, readFileFn)
		if readErr != nil {
			return nil, fmt.Errorf("cannot parse %s: %w", name, readErr)
		}

		for p := range paths {
			result[p] = struct{}{}
		}
	}

	if !foundSpec {
		return nil, fmt.Errorf("no openapi_spec*.yaml files found (excluding *_components.yaml)")
	}

	return result, nil
}

// isSpecFile returns true if the filename is an OpenAPI spec file to include in path extraction.
// Excludes *_components.yaml (schema definitions, no paths) and *gen_config*.yaml (tool config).
func isSpecFile(name string) bool {
	if !strings.HasPrefix(name, "openapi_spec") {
		return false
	}

	if !strings.HasSuffix(name, ".yaml") {
		return false
	}

	if strings.HasSuffix(name, "_components.yaml") {
		return false
	}

	if strings.Contains(name, "gen_config") {
		return false
	}

	return true
}

// parseSpecPaths reads a YAML file and returns the set of path keys under the `paths:` top-level key.
func parseSpecPaths(filePath string, readFileFn func(string) ([]byte, error)) (map[string]struct{}, error) {
	data, err := readFileFn(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var doc map[string]any

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("cannot parse YAML: %w", err)
	}

	paths, ok := doc["paths"]
	if !ok {
		// This spec file has no paths block (e.g., only components).
		return make(map[string]struct{}), nil
	}

	pathsMap, ok := paths.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("paths block is not a YAML mapping")
	}

	result := make(map[string]struct{}, len(pathsMap))

	for k := range pathsMap {
		result[k] = struct{}{}
	}

	return result, nil
}

// comparePaths compares registry-declared paths with spec paths and returns any violations.
func comparePaths(psID string, registryPaths, specPaths map[string]struct{}) []string {
	var violations []string

	// Missing from spec: declared in registry but absent from spec.
	var missingFromSpec []string

	for p := range registryPaths {
		if _, ok := specPaths[p]; !ok {
			missingFromSpec = append(missingFromSpec, p)
		}
	}

	// Undeclared in registry: present in spec but absent from registry.
	var undeclaredInRegistry []string

	for p := range specPaths {
		if _, ok := registryPaths[p]; !ok {
			undeclaredInRegistry = append(undeclaredInRegistry, p)
		}
	}

	sort.Strings(missingFromSpec)
	sort.Strings(undeclaredInRegistry)

	for _, p := range missingFromSpec {
		violations = append(violations, fmt.Sprintf("%s: path %q declared in registry but missing from OpenAPI spec", psID, p))
	}

	for _, p := range undeclaredInRegistry {
		violations = append(violations, fmt.Sprintf("%s: path %q in OpenAPI spec but not declared in registry api_resources", psID, p))
	}

	return violations
}

// toSet converts a string slice to a set (map[string]struct{}).
func toSet(items []string) map[string]struct{} {
	result := make(map[string]struct{}, len(items))

	for _, item := range items {
		result[item] = struct{}{}
	}

	return result
}
