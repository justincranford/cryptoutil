package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// KebabCaseValidationResult holds the outcome of kebab-case YAML field validation.
type KebabCaseValidationResult struct {
	Path     string
	Valid    bool
	Errors   []string
	Warnings []string
}

// kebabCaseConfigFields lists YAML paths that must follow kebab-case convention.
// These are field values inside config files (not file/dir names - see ValidateNaming).
var kebabCaseConfigFields = []string{
	"service.name",
}

// ValidateKebabCase validates YAML config field values follow kebab-case convention.
// Checks service.name and similar identity fields inside *.yml/*.yaml config files.
// For file/directory/compose-service name validation, see ValidateNaming.
func ValidateKebabCase(rootPath string) (*KebabCaseValidationResult, error) {
	result := &KebabCaseValidationResult{
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

	// Walk directory tree and validate YAML config field values.
	_ = filepath.Walk(rootPath, func(path string, fInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("error accessing path %s: %v", path, walkErr))

			return nil
		}

		// Only check YAML files that are not directories.
		if fInfo.IsDir() || !isYAMLFile(fInfo.Name()) {
			return nil
		}

		// Skip compose files (handled by ValidateNaming).
		name := fInfo.Name()
		if strings.HasPrefix(name, "compose") || strings.HasPrefix(name, "docker-compose") {
			return nil
		}

		relPath, _ := filepath.Rel(rootPath, path)
		validateConfigFieldKebabCase(path, relPath, result)

		return nil
	})

	return result, nil
}

// validateConfigFieldKebabCase reads a YAML config file and checks that specified fields follow kebab-case.
func validateConfigFieldKebabCase(configPath, relPath string, result *KebabCaseValidationResult) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"[ValidateKebabCase] Failed to read config file: %s", relPath))

		return
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"[ValidateKebabCase] Failed to parse YAML: %s", relPath))

		return
	}

	for _, fieldPath := range kebabCaseConfigFields {
		value := getNestedField(config, fieldPath)
		if value == "" {
			continue // Field not present in this config.
		}

		if !isKebabCase(value) {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"[ValidateKebabCase] Field '%s' value '%s' violates kebab-case - use '%s' (file: %s)",
				fieldPath, value, toKebabCase(value), relPath))
			result.Valid = false
		}
	}
}

// getNestedField retrieves a dot-separated field value from a nested map.
// Returns empty string if field is not found or is not a string.
func getNestedField(config map[string]any, fieldPath string) string {
	parts := strings.Split(fieldPath, ".")
	current := any(config)

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}

		current, ok = m[part]
		if !ok {
			return ""
		}
	}

	if s, ok := current.(string); ok {
		return s
	}

	return ""
}

// FormatKebabCaseValidationResult formats a KebabCaseValidationResult for display.
func FormatKebabCaseValidationResult(result *KebabCaseValidationResult) string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(&sb, "Kebab-Case Field Validation: %s\n", result.Path)

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
