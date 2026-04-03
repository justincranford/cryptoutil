// Copyright (c) 2025 Justin Cranford

// Package fitness_registry_completeness validates that the lint-fitness-registry.yaml
// manifest is consistent with the fitness sub-linter directories on disk.
//
// It detects two categories of drift:
//   - Orphaned directories: exist under lint_fitness/ but not listed in the registry YAML.
//   - Missing directories: listed in the registry YAML but no directory exists on disk.
//
// The "registry" directory is excluded from the check (it is infrastructure, not a sub-linter).
package fitness_registry_completeness

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

// FitnessSubLinter represents one sub-linter entry from lint-fitness-registry.yaml.
type FitnessSubLinter struct {
	Name        string `yaml:"name"`
	Directory   string `yaml:"directory"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
}

// FitnessRegistry is the top-level structure of lint-fitness-registry.yaml.
type FitnessRegistry struct {
	SubLinters []FitnessSubLinter `yaml:"sub_linters"`
}

// registryDirName is excluded from the filesystem scan because it is infrastructure, not a sub-linter.
const registryDirName = "registry"

// findFitnessProjectRoot walks up from cwd to find the directory containing go.mod.
func findFitnessProjectRoot(getwdFn func() (string, error)) (string, error) {
	dir, err := getwdFn()
	if err != nil {
		return "", fmt.Errorf("getwd failed: %w", err)
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}

		dir = parent
	}
}

// LoadFitnessRegistry loads and parses the lint-fitness-registry.yaml manifest.
func LoadFitnessRegistry(rootDir string, readFileFn func(string) ([]byte, error)) (*FitnessRegistry, error) {
	path := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDFitnessRegistryFile))

	data, err := readFileFn(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", cryptoutilSharedMagic.CICDFitnessRegistryFile, err)
	}

	var reg FitnessRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", cryptoutilSharedMagic.CICDFitnessRegistryFile, err)
	}

	return &reg, nil
}

// CheckRegistryCompleteness validates consistency between the YAML registry and the filesystem.
// Returns orphaned (filesystem-only) and missing (YAML-only) directory names.
func CheckRegistryCompleteness(rootDir string, readFileFn func(string) ([]byte, error), readDirFn func(string) ([]os.DirEntry, error)) (orphaned, missing []string, err error) {
	reg, err := LoadFitnessRegistry(rootDir, readFileFn)
	if err != nil {
		return nil, nil, err
	}

	// Build set of directories declared in YAML.
	yamlDirs := make(map[string]bool, len(reg.SubLinters))
	for _, sl := range reg.SubLinters {
		yamlDirs[sl.Directory] = true
	}

	// Scan actual directories on disk.
	fitnessDir := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDLintFitnessDir))

	entries, readErr := readDirFn(fitnessDir)
	if readErr != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %w", cryptoutilSharedMagic.CICDLintFitnessDir, readErr)
	}

	fsDirs := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		if dirName == registryDirName {
			continue // exclude registry infrastructure directory
		}

		fsDirs[dirName] = true
	}

	// Find orphaned: in filesystem but not in YAML.
	for dir := range fsDirs {
		if !yamlDirs[dir] {
			orphaned = append(orphaned, dir)
		}
	}

	// Find missing: in YAML but not in filesystem.
	for dir := range yamlDirs {
		if !fsDirs[dir] {
			missing = append(missing, dir)
		}
	}

	sort.Strings(orphaned)
	sort.Strings(missing)

	return orphaned, missing, nil
}

// Check validates fitness sub-linter registry completeness from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, os.Getwd, os.ReadFile, os.ReadDir)
}

func check(logger *cryptoutilCmdCicdCommon.Logger, getwdFn func() (string, error), readFileFn func(string) ([]byte, error), readDirFn func(string) ([]os.DirEntry, error)) error {
	rootDir, err := findFitnessProjectRoot(getwdFn)
	if err != nil {
		return fmt.Errorf("fitness-registry-completeness: %w", err)
	}

	return checkInDir(logger, rootDir, readFileFn, readDirFn)
}

// CheckInDir validates fitness sub-linter registry completeness under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDir(logger, rootDir, os.ReadFile, os.ReadDir)
}

func checkInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readFileFn func(string) ([]byte, error), readDirFn func(string) ([]os.DirEntry, error)) error {
	logger.Log("Checking fitness sub-linter registry completeness...")

	reg, loadErr := LoadFitnessRegistry(rootDir, readFileFn)
	if loadErr != nil {
		return fmt.Errorf("fitness-registry-completeness: %w", loadErr)
	}

	orphaned, missing, err := CheckRegistryCompleteness(rootDir, readFileFn, readDirFn)
	if err != nil {
		return fmt.Errorf("fitness-registry-completeness: %w", err)
	}

	var violations []string

	for _, dir := range orphaned {
		violations = append(violations, fmt.Sprintf("  ORPHANED (in filesystem but not in registry): %s", dir))
	}

	for _, dir := range missing {
		violations = append(violations, fmt.Sprintf("  MISSING (in registry but not in filesystem): %s", dir))
	}

	sort.Strings(violations)

	if len(violations) > 0 {
		return fmt.Errorf("fitness-registry-completeness violations:\n%s",
			strings.Join(violations, "\n"))
	}

	logger.Log(fmt.Sprintf("fitness-registry-completeness: all %d sub-linters are registered", len(reg.SubLinters)))

	return nil
}
