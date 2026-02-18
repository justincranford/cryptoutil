package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MirrorResult holds the result of structural mirror validation.
type MirrorResult struct {
	Valid          bool     `json:"valid"`
	MissingMirrors []string `json:"missing_mirrors,omitempty"`
	Orphans        []string `json:"orphans,omitempty"`
	Excluded       []string `json:"excluded,omitempty"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
}

// excludedDeployments lists infrastructure deployments that do not require a configs/ counterpart.
var excludedDeployments = map[string]bool{
	"shared-postgres":  true,
	"shared-citus":     true,
	"shared-telemetry": true,
	"archived":         true,
	"template":         true,
}

// ValidateStructuralMirror validates that every deployment directory has a corresponding configs directory.
// Direction: deployments → configs (one-way). Orphaned configs are warnings, not errors.
func ValidateStructuralMirror(deploymentsDir string, configsDir string) (*MirrorResult, error) {
	if _, err := os.Stat(deploymentsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("deployments directory does not exist: %s", deploymentsDir)
	}

	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("configs directory does not exist: %s", configsDir)
	}

	result := &MirrorResult{
		Valid: true,
	}

	// Get deployment directories.
	deployDirs, err := getSubdirectories(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployment directories: %w", err)
	}

	// Get config directories.
	configDirs, err := getSubdirectories(configsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list config directories: %w", err)
	}

	configDirSet := make(map[string]bool, len(configDirs))
	for _, d := range configDirs {
		configDirSet[d] = true
	}

	deployDirSet := make(map[string]bool, len(deployDirs))
	for _, d := range deployDirs {
		deployDirSet[d] = true
	}

	// Check each deployment directory has a configs counterpart.
	for _, deployDir := range deployDirs {
		if excludedDeployments[deployDir] {
			result.Excluded = append(result.Excluded, deployDir)

			continue
		}

		// Map deployment dir to expected config dir.
		configName := mapDeploymentToConfig(deployDir)
		if !configDirSet[configName] {
			result.MissingMirrors = append(result.MissingMirrors, deployDir)
			result.Errors = append(result.Errors,
				fmt.Sprintf("deployment '%s' has no configs counterpart (expected configs/%s/)", deployDir, configName))
			result.Valid = false
		}
	}

	// Check for orphaned config directories (warnings only).
	for _, configDir := range configDirs {
		// Check if any deployment maps to this config.
		found := false

		for _, deployDir := range deployDirs {
			if excludedDeployments[deployDir] {
				continue
			}

			if mapDeploymentToConfig(deployDir) == configDir {
				found = true

				break
			}
		}

		if !found {
			result.Orphans = append(result.Orphans, configDir)
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("config directory '%s' has no corresponding deployment (orphaned)", configDir))
		}
	}

	sort.Strings(result.MissingMirrors)
	sort.Strings(result.Orphans)
	sort.Strings(result.Excluded)
	sort.Strings(result.Errors)
	sort.Strings(result.Warnings)

	return result, nil
}

// deploymentToConfigMapping maps deployment directory names to their expected configs directory names.
// This handles naming differences between deployments/ and configs/ directories.
var deploymentToConfigMapping = map[string]string{
	"pki":    "ca",
	"pki-ca": "ca",
	"sm":     "sm",
	"sm-kms": "sm",
}

// mapDeploymentToConfig maps a deployment directory name to its expected configs directory name.
// Rules:
//   - Uses explicit mapping table for known naming differences.
//   - PRODUCT-SERVICE (e.g., "jose-ja") -> product name (e.g., "jose").
//   - PRODUCT (e.g., "cipher") -> same name (e.g., "cipher").
//   - SUITE (e.g., "cryptoutil-suite") -> "cryptoutil".
func mapDeploymentToConfig(deployDir string) string {
	// Check explicit mapping first.
	if mapped, ok := deploymentToConfigMapping[deployDir]; ok {
		return mapped
	}

	// Map PRODUCT-SERVICE to just the product name since configs use product-level dirs.
	parts := strings.SplitN(deployDir, "-", 2)
	if len(parts) == 2 {
		return parts[0]
	}

	return deployDir
}

// getSubdirectories returns the names of immediate subdirectories of the given path.
func getSubdirectories(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	sort.Strings(dirs)

	return dirs, nil
}

// FormatMirrorResult formats a MirrorResult for human-readable output.
func FormatMirrorResult(result *MirrorResult) string {
	var sb strings.Builder

	sb.WriteString("=== Structural Mirror Validation ===\n\n")

	if result.Valid {
		sb.WriteString("PASS: All deployment directories have configs counterparts\n\n")
	} else {
		sb.WriteString("FAIL: Missing config mirrors found\n\n")
	}

	if len(result.Excluded) > 0 {
		sb.WriteString(fmt.Sprintf("Excluded (%d):\n", len(result.Excluded)))

		for _, e := range result.Excluded {
			sb.WriteString(fmt.Sprintf("  - %s (infrastructure/template)\n", e))
		}

		sb.WriteString("\n")
	}

	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("Errors (%d):\n", len(result.Errors)))

		for _, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("  ✗ %s\n", e))
		}

		sb.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("Warnings (%d):\n", len(result.Warnings)))

		for _, w := range result.Warnings {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", w))
		}

		sb.WriteString("\n")
	}

	var (
		deploymentsDir = filepath.Clean("deployments")
		configsDir     = filepath.Clean("configs")
	)

	sb.WriteString(fmt.Sprintf("Summary: deployments=%s configs=%s valid=%t missing=%d orphans=%d excluded=%d\n",
		deploymentsDir, configsDir, result.Valid, len(result.MissingMirrors), len(result.Orphans), len(result.Excluded)))

	return sb.String()
}
