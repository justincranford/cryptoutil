package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplatePatternResult holds template pattern validation results.
type TemplatePatternResult struct {
	// Path is the template directory root.
	Path string
	// Valid indicates whether all validations passed (no errors).
	Valid bool
	// Errors contains critical validation failures.
	Errors []string
	// Warnings contains non-critical issues.
	Warnings []string
}

// Required compose template files that MUST exist in deployments/template/.
var requiredTemplateComposeFiles = []string{
	"compose.yml",
	"compose-cryptoutil-PRODUCT-SERVICE.yml",
	"compose-cryptoutil-PRODUCT.yml",
	"compose-cryptoutil.yml",
}

// Required config template files in deployments/template/config/.
var requiredTemplateConfigFiles = []string{
	"template-app-common.yml",
	"template-app-sqlite-1.yml",
	"template-app-postgresql-1.yml",
	"template-app-postgresql-2.yml",
}

// Required secret files in deployments/template/secrets/.
var requiredTemplateSecretFiles = []string{
	"hash_pepper_v3.secret",
	"unseal_1of5.secret",
	"unseal_2of5.secret",
	"unseal_3of5.secret",
	"unseal_4of5.secret",
	"unseal_5of5.secret",
	"postgres_database.secret",
	"postgres_password.secret",
	"postgres_username.secret",
	"postgres_url.secret",
}

// Placeholder patterns that MUST appear in SERVICE-level compose template.
var requiredServicePlaceholders = []string{
	"PRODUCT-SERVICE",
	"XXXX",
}

// Placeholder patterns that MUST appear in PRODUCT-level compose template.
var requiredProductPlaceholders = []string{
	"PRODUCT",
}

// ValidateTemplatePattern validates the template directory for naming, structure,
// and value correctness.
//
// Checks:
//  1. Required files and directories exist (compose, config, secrets).
//  2. Compose template files contain required placeholder patterns.
//  3. Config template files use consistent naming conventions.
func ValidateTemplatePattern(templatePath string) (*TemplatePatternResult, error) {
	result := &TemplatePatternResult{
		Path:  templatePath,
		Valid: true,
	}

	info, statErr := os.Stat(templatePath)
	if statErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateTemplatePattern] Template path does not exist: %s", templatePath))

		return result, nil //nolint:nilerr // Error aggregation pattern: validation errors collected in result.Errors, nil Go error allows validator pipeline to continue.
	}

	if !info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateTemplatePattern] Template path is not a directory: %s", templatePath))

		return result, nil
	}

	validateRequiredTemplateFiles(templatePath, result)
	validateTemplatePlaceholders(templatePath, result)
	validateTemplateConfigNaming(templatePath, result)

	return result, nil
}

// validateRequiredTemplateFiles checks that all required files exist.
func validateRequiredTemplateFiles(templatePath string, result *TemplatePatternResult) {
	// Check compose files.
	for _, f := range requiredTemplateComposeFiles {
		path := filepath.Join(templatePath, f)
		if _, err := os.Stat(path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("[ValidateTemplatePattern] Missing required compose file: %s", f))
		}
	}

	// Check config directory and files.
	configDir := filepath.Join(templatePath, "config")

	if info, err := os.Stat(configDir); err != nil || !info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors,
			"[ValidateTemplatePattern] Missing required config/ directory")
	} else {
		for _, f := range requiredTemplateConfigFiles {
			path := filepath.Join(configDir, f)
			if _, err := os.Stat(path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors,
					fmt.Sprintf("[ValidateTemplatePattern] Missing required config file: config/%s", f))
			}
		}
	}

	// Check secrets directory and files.
	secretsDir := filepath.Join(templatePath, "secrets")

	if info, err := os.Stat(secretsDir); err != nil || !info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors,
			"[ValidateTemplatePattern] Missing required secrets/ directory")
	} else {
		for _, f := range requiredTemplateSecretFiles {
			path := filepath.Join(secretsDir, f)
			if _, err := os.Stat(path); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors,
					fmt.Sprintf("[ValidateTemplatePattern] Missing required secret file: secrets/%s", f))
			}
		}
	}
}

// validateTemplatePlaceholders verifies compose template files contain expected
// placeholder patterns (PRODUCT-SERVICE, XXXX, etc.).
func validateTemplatePlaceholders(templatePath string, result *TemplatePatternResult) {
	// Check SERVICE-level compose template.
	serviceCompose := filepath.Join(templatePath, "compose-cryptoutil-PRODUCT-SERVICE.yml")

	data, err := os.ReadFile(serviceCompose)
	if err != nil {
		// Missing file already reported by validateRequiredTemplateFiles.
		return
	}

	content := string(data)
	for _, placeholder := range requiredServicePlaceholders {
		if !strings.Contains(content, placeholder) {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("[ValidateTemplatePattern] Service compose template missing placeholder '%s': %s",
					placeholder, "compose-cryptoutil-PRODUCT-SERVICE.yml"))
		}
	}

	// Check PRODUCT-level compose template.
	productCompose := filepath.Join(templatePath, "compose-cryptoutil-PRODUCT.yml")

	data, err = os.ReadFile(productCompose)
	if err != nil {
		return
	}

	content = string(data)
	for _, placeholder := range requiredProductPlaceholders {
		if !strings.Contains(content, placeholder) {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("[ValidateTemplatePattern] Product compose template missing placeholder '%s': %s",
					placeholder, "compose-cryptoutil-PRODUCT.yml"))
		}
	}
}

// validateTemplateConfigNaming ensures config files follow the naming convention:
// template-app-<variant>.yml (e.g., template-app-common.yml, template-app-sqlite-1.yml).
func validateTemplateConfigNaming(templatePath string, result *TemplatePatternResult) {
	configDir := filepath.Join(templatePath, "config")

	entries, err := os.ReadDir(configDir)
	if err != nil {
		// Missing config dir already reported by validateRequiredTemplateFiles.
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isYAMLFile(name) {
			continue
		}

		if !strings.HasPrefix(name, "template-app-") {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("[ValidateTemplatePattern] Config file does not follow 'template-app-*.yml' naming: config/%s", name))
		}
	}
}

// FormatTemplatePatternResult formats a TemplatePatternResult for display.
func FormatTemplatePatternResult(result *TemplatePatternResult) string {
	var sb strings.Builder

	if result.Valid {
		sb.WriteString(fmt.Sprintf("  PASS: Template pattern validation: %s\n", result.Path))
	} else {
		sb.WriteString(fmt.Sprintf("  FAIL: Template pattern validation: %s\n", result.Path))
	}

	for _, e := range result.Errors {
		sb.WriteString(fmt.Sprintf("    ERROR: %s\n", e))
	}

	for _, w := range result.Warnings {
		sb.WriteString(fmt.Sprintf("    WARN: %s\n", w))
	}

	return sb.String()
}
