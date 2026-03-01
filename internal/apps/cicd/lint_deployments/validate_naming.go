package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// NamingValidationResult holds the outcome of naming validation.
type NamingValidationResult struct {
	Path     string
	Valid    bool
	Errors   []string
	Warnings []string
}

// kebabCasePattern matches valid kebab-case names.
// Allows: lowercase letters, digits, hyphens (NOT at start/end).
var kebabCasePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidateNaming validates all deployment/config names follow kebab-case convention.
// Validates:
// - Directory names (SERVICE, PRODUCT, SUITE levels)
// - File names (*.yml, *.yaml, docker-compose.yml)
// - Compose service names (parsed from *.yml files).
func ValidateNaming(rootPath string) (*NamingValidationResult, error) {
	result := &NamingValidationResult{
		Path:  rootPath,
		Valid: true,
	}

	// Check if root path exists.
	info, statErr := os.Stat(rootPath)
	if statErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("path does not exist: %s", rootPath))
		result.Valid = false

		return result, nil //nolint:nilerr // Error aggregation pattern: validation errors collected in result.Errors, nil Go error allows validator pipeline to continue.
	}

	// Root must be a directory.
	if !info.IsDir() {
		result.Errors = append(result.Errors, fmt.Sprintf("path is not a directory: %s", rootPath))
		result.Valid = false

		return result, nil
	}

	// Walk directory tree and validate names. Callback always returns nil to continue
	// walking, so Walk itself will not return an error (root was verified above).
	_ = filepath.Walk(rootPath, func(path string, fInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("error accessing path %s: %v", path, walkErr))
			result.Valid = false

			return nil // Continue walking.
		}

		// Skip root directory.
		if path == rootPath {
			return nil
		}

		// Get relative path from root (safe: Walk guarantees path is under rootPath).
		relPath, _ := filepath.Rel(rootPath, path)

		name := fInfo.Name()
		// Skip template directory entirely - it uses intentional uppercase
		// placeholders (PRODUCT-SERVICE, PRODUCT) for template substitution.
		if fInfo.IsDir() && name == DeploymentTypeTemplate {
			return filepath.SkipDir
		}
		// Validate directory names.
		if fInfo.IsDir() {
			if !isKebabCase(name) {
				result.Errors = append(result.Errors, fmt.Sprintf(
					"[ValidateNaming] Directory '%s' violates kebab-case - rename to '%s' (path: %s)",
					name, toKebabCase(name), relPath))
				result.Valid = false
			}

			return nil
		}

		// Validate file names (*.yml, *.yaml).
		if isYAMLFile(name) {
			// Allow docker-compose.yml as-is (valid kebab-case).
			if !isKebabCase(strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml")) {
				result.Errors = append(result.Errors, fmt.Sprintf(
					"[ValidateNaming] File '%s' violates kebab-case - rename to '%s' (path: %s)",
					name, toKebabCase(name), relPath))
				result.Valid = false
			}

			// Validate compose service names if this is a compose file.
			if strings.HasPrefix(name, "compose") || strings.HasPrefix(name, "docker-compose") {
				validateComposeServiceNames(path, relPath, result)
			}
		}

		return nil
	})

	return result, nil
}

// isKebabCase checks if a string follows kebab-case convention.
func isKebabCase(s string) bool {
	return kebabCasePattern.MatchString(s)
}

// toKebabCase converts a string to kebab-case suggestion.
// Simple heuristic: lowercase and replace non-alphanumeric with hyphens.
func toKebabCase(s string) string {
	// Remove file extensions for suggestion.
	base := strings.TrimSuffix(strings.TrimSuffix(s, ".yml"), ".yaml")

	// Convert to lowercase.
	base = strings.ToLower(base)

	// Replace underscores and spaces with hyphens.
	base = strings.ReplaceAll(base, "_", "-")
	base = strings.ReplaceAll(base, " ", "-")

	// Remove consecutive hyphens.
	for strings.Contains(base, "--") {
		base = strings.ReplaceAll(base, "--", "-")
	}

	// Trim leading/trailing hyphens.
	base = strings.Trim(base, "-")

	// Re-add extension if original had one.
	if strings.HasSuffix(s, ".yml") {
		return base + ".yml"
	}

	if strings.HasSuffix(s, ".yaml") {
		return base + ".yaml"
	}

	return base
}

// isYAMLFile checks if a filename has .yml or .yaml extension.
func isYAMLFile(name string) bool {
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

// validateComposeServiceNames parses a compose file and validates service names.
func validateComposeServiceNames(composePath, relPath string, result *NamingValidationResult) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"[ValidateNaming] Failed to read compose file for service name validation: %s", relPath))

		return
	}

	var compose struct {
		Services map[string]any `yaml:"services"`
	}

	if err := yaml.Unmarshal(data, &compose); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"[ValidateNaming] Failed to parse compose file for service name validation: %s", relPath))

		return
	}

	for serviceName := range compose.Services {
		if !isKebabCase(serviceName) {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"[ValidateNaming] Compose service '%s' violates kebab-case - rename to '%s' (file: %s)",
				serviceName, toKebabCase(serviceName), relPath))
			result.Valid = false
		}
	}
}

// FormatNamingValidationResult formats a NamingValidationResult for display.
func FormatNamingValidationResult(result *NamingValidationResult) string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(&sb, "Naming Validation: %s\n", result.Path)

	if result.Valid {
		sb.WriteString("  Status: PASS\n")
	} else {
		sb.WriteString("  Status: FAIL\n")
	}

	for _, err := range result.Errors {
		_, _ = fmt.Fprintf(&sb, "  ERROR: %s\n", err)
	}

	for _, warn := range result.Warnings {
		_, _ = fmt.Fprintf(&sb, "  WARNING: %s\n", warn)
	}

	return sb.String()
}
