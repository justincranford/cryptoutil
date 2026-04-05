// Copyright (c) 2025 Justin Cranford

// Package database_key_uniformity validates that no deployment config overlay file
// uses the deprecated nested "database:" key structure (with sub-keys like "type:"
// and "dsn:"). The framework standard is the flat "database-url:" key only.
//
// Scans all *.yml files in deployments/{ps-id}/config/ for all known PS-IDs.
// Any file whose top-level YAML contains a "database" key whose value is a mapping
// (rather than a scalar or absent) is a violation.
//
// See ENG-HANDBOOK.md Section 5.2 Service Builder Pattern and Section 7. Data Architecture
// for the database-url canonical key documentation.
package database_key_uniformity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Check validates database key uniformity from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.ReadDir, os.ReadFile)
}

// CheckInDir validates database key uniformity under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) error {
	logger.Log("Checking database key uniformity (no nested database: mapping allowed)...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkConfigDir(rootDir, ps.PSID, readDirFn, readFileFn)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("database key uniformity violations:\n%s\n\nRemediation: replace nested 'database: {type:, dsn:}' with the flat 'database-url: \"<url>\"' key", strings.Join(violations, "\n"))
	}

	logger.Log("database-key-uniformity: all deployment config files use the framework-standard database-url: key")

	return nil
}

// checkConfigDir scans all *.yml files in deployments/{psID}/config/ for violations.
func checkConfigDir(rootDir, psID string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) []string {
	configDir := filepath.Join(rootDir, "deployments", psID, "config")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return []string{fmt.Sprintf("%s: deployments/%s/config/ directory does not exist", psID, psID)}
	}

	entries, err := readDirFn(configDir)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read config dir: %s", psID, err)}
	}

	var violations []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		configPath := filepath.Join(configDir, entry.Name())

		v := checkFile(configPath, psID, entry.Name(), readFileFn)
		violations = append(violations, v...)
	}

	return violations
}

// checkFile parses one YAML file and returns a violation string if it contains
// a nested "database:" mapping.
func checkFile(configPath, psID, filename string, readFileFn func(string) ([]byte, error)) []string {
	data, err := readFileFn(configPath) //nolint:gosec // configPath constructed from controlled registry + dir walk
	if err != nil {
		return []string{fmt.Sprintf("%s: %s: cannot read file: %s", psID, filename, err)}
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return []string{fmt.Sprintf("%s: %s: YAML parse error: %s", psID, filename, err)}
	}

	if _, isMapping := config[cryptoutilSharedMagic.RealmStorageTypeDatabase].(map[string]any); isMapping {
		return []string{fmt.Sprintf("%s: %s: found deprecated nested 'database:' mapping; use 'database-url: \"<url>\"' instead", psID, filename)}
	}

	return nil
}
