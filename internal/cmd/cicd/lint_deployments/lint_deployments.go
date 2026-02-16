package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DeploymentStructure defines expected directory structure for each deployment type.
type DeploymentStructure struct {
	Name              string
	RequiredDirs      []string
	RequiredFiles     []string
	OptionalFiles     []string
	RequiredSecrets   []string
	AllowedExtensions []string
}

// GetExpectedStructures returns validation rules for different deployment types.
// See: docs/ARCHITECTURE.md Section 12.4 "Deployment Structure Validation".
func GetExpectedStructures() map[string]DeploymentStructure {
	return map[string]DeploymentStructure{
		"PRODUCT-SERVICE": {
			Name:          "PRODUCT-SERVICE deployment (e.g., jose-ja, cipher-im)",
			RequiredDirs:  []string{"secrets", "config"},
			RequiredFiles: []string{"compose.yml", "Dockerfile"},
			OptionalFiles: []string{"compose.demo.yml", "otel-collector-config.yaml", "README.md"},
			RequiredSecrets: []string{
				"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
				"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
				"postgres_url.secret", "postgres_username.secret",
				"postgres_password.secret", "postgres_database.secret",
			},
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
		},
		"template": {
			Name:          "Template deployment (deployments/template/)",
			RequiredDirs:  []string{"secrets"},
			RequiredFiles: []string{"compose.yml"},
			RequiredSecrets: []string{
				"unseal_1of5.secret", "unseal_2of5.secret", "unseal_3of5.secret",
				"unseal_4of5.secret", "unseal_5of5.secret", "hash_pepper_v3.secret",
				"postgres_url.secret", "postgres_username.secret",
				"postgres_password.secret", "postgres_database.secret",
			},
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
		},
		"infrastructure": {
			Name:              "Infrastructure deployment (postgres, citus, telemetry)",
			RequiredDirs:      []string{},
			RequiredFiles:     []string{"compose.yml"},
			OptionalFiles:     []string{"init-db.sql", "init-citus.sql", "README.md"},
			RequiredSecrets:   []string{}, // Infrastructure secrets are optional
			AllowedExtensions: []string{".yml", ".yaml", ".sql", ".md"},
		},
	}
}

// ValidationResult holds validation outcome for a directory.
type ValidationResult struct {
	Path           string
	Type           string
	Valid          bool
	MissingDirs    []string
	MissingFiles   []string
	MissingSecrets []string
	Errors         []string
	Warnings       []string
}

// ValidateDeploymentStructure validates a deployment directory against expected structure.
func ValidateDeploymentStructure(basePath string, deploymentName string, structType string) (*ValidationResult, error) {
	structures := GetExpectedStructures()

	expected, ok := structures[structType]
	if !ok {
		return nil, fmt.Errorf("unknown structure type: %s", structType)
	}

	result := &ValidationResult{
		Path:  basePath,
		Type:  structType,
		Valid: true,
	}

	// Check required directories
	for _, dir := range expected.RequiredDirs {
		dirPath := filepath.Join(basePath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			result.MissingDirs = append(result.MissingDirs, dir)
			result.Valid = false
		}
	}

	// Check required files
	for _, file := range expected.RequiredFiles {
		filePath := filepath.Join(basePath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			result.MissingFiles = append(result.MissingFiles, file)
			result.Valid = false
		}
	}

	// Check required secrets
	if len(expected.RequiredSecrets) > 0 {
		secretsPath := filepath.Join(basePath, "secrets")
		if _, err := os.Stat(secretsPath); err == nil {
			for _, secret := range expected.RequiredSecrets {
				secretPath := filepath.Join(secretsPath, secret)
				if _, err := os.Stat(secretPath); os.IsNotExist(err) {
					result.MissingSecrets = append(result.MissingSecrets, secret)
					result.Valid = false
				}
			}
		}
	}

	// Check config files for PRODUCT-SERVICE deployments
	if structType == "PRODUCT-SERVICE" {
		validateConfigFiles(basePath, deploymentName, result)
	}

	// Check for hardcoded credentials in ALL deployment types
	checkHardcodedCredentials(basePath, result)

	return result, nil
}

// validateConfigFiles checks config directory for required files and deprecated patterns.
// Strict enforcement mode: all violations are errors that block CI/CD.
// See: docs/ARCHITECTURE.md Section 12.4.5 "Config File Naming Strategy".
// See: docs/ARCHITECTURE.md Section 12.4.7 "Linter Validation Modes".
func validateConfigFiles(basePath string, deploymentName string, result *ValidationResult) {
	configPath := filepath.Join(basePath, "config")

	// Extract PRODUCT-SERVICE parts (e.g., "sm-kms" -> product="sm", service="kms").
	parts := strings.Split(deploymentName, "-")
	if len(parts) < 2 {
		result.Errors = append(result.Errors,
			fmt.Sprintf("Cannot validate config files: deployment name '%s' does not match PRODUCT-SERVICE pattern", deploymentName))
		result.Valid = false

		return
	}

	productService := deploymentName // Full name like "sm-kms".

	// Define required standard config files per Section 12.4.5.
	requiredConfigs := []string{
		fmt.Sprintf("%s-app-common.yml", productService),       // sm-kms-app-common.yml.
		fmt.Sprintf("%s-app-sqlite-1.yml", productService),     // sm-kms-app-sqlite-1.yml.
		fmt.Sprintf("%s-app-postgresql-1.yml", productService), // sm-kms-app-postgresql-1.yml.
		fmt.Sprintf("%s-app-postgresql-2.yml", productService), // sm-kms-app-postgresql-2.yml.
	}

	// Check for required config files (strict enforcement: missing = error).
	for _, configFile := range requiredConfigs {
		configFilePath := filepath.Join(configPath, configFile)
		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Missing required config file: %s (see Section 12.4.5)", configFile))
			result.Valid = false
		}
	}

	// Check for deprecated files (strict enforcement: presence = error, Section 12.4.6).
	deprecatedFiles := []struct {
		deprecated  string
		replacement string
	}{
		{"demo-seed.yml", fmt.Sprintf("%s-demo.yml", productService)},
		{"integration.yml", fmt.Sprintf("%s-e2e.yml", productService)},
	}

	for _, df := range deprecatedFiles {
		deprecatedPath := filepath.Join(configPath, df.deprecated)
		if _, err := os.Stat(deprecatedPath); err == nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("DEPRECATED file must be removed: %s (rename to %s, see Section 12.4.6)", df.deprecated, df.replacement))
			result.Valid = false
		}
	}

	// Check for non-conformant config filenames (strict enforcement: mismatch = error).
	entries, err := os.ReadDir(configPath)
	if err != nil {
		return // Config directory doesn't exist or not readable.
	}

	expectedPrefix := productService + "-"

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Skip known non-YAML files.
		if filename == "README.md" || filename == ".gitkeep" {
			continue
		}

		// Check if YAML file.
		if !strings.HasSuffix(filename, ".yml") && !strings.HasSuffix(filename, ".yaml") {
			continue
		}

		// Check if matches expected pattern.
		if !strings.HasPrefix(filename, expectedPrefix) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Config file '%s' does not match required pattern '%s*.yml' (see Section 12.4.5)", filename, expectedPrefix))
			result.Valid = false
		}
	}
}

// checkHardcodedCredentials scans compose.yml files for hardcoded database credentials.
// CRITICAL: ALL database credentials MUST use Docker secrets, NEVER hardcoded values.
func checkHardcodedCredentials(basePath string, result *ValidationResult) {
	composeFiles := []string{"compose.yml", "compose.yaml"}

	for _, filename := range composeFiles {
		composePath := filepath.Join(basePath, filename)

		content, err := os.ReadFile(composePath)
		if err != nil {
			continue // File doesn't exist, skip
		}

		text := string(content)

		lineNumber := 0
		for _, line := range strings.Split(text, "\n") {
			lineNumber++
			lineTrimmed := strings.TrimSpace(line)

			// Check for hardcoded POSTGRES_USER (not POSTGRES_USER_FILE)
			if strings.Contains(lineTrimmed, "POSTGRES_USER:") && !strings.Contains(lineTrimmed, "POSTGRES_USER_FILE:") {
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s:%d: Hardcoded POSTGRES_USER detected. Use POSTGRES_USER_FILE with Docker secrets instead (see deployments/sm-kms/compose.yml for pattern)", filename, lineNumber))
				result.Valid = false
			}

			// Check for hardcoded POSTGRES_PASSWORD (not POSTGRES_PASSWORD_FILE)
			if strings.Contains(lineTrimmed, "POSTGRES_PASSWORD:") && !strings.Contains(lineTrimmed, "POSTGRES_PASSWORD_FILE:") && !strings.Contains(lineTrimmed, "#") {
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s:%d: Hardcoded POSTGRES_PASSWORD detected. Use POSTGRES_PASSWORD_FILE with Docker secrets instead", filename, lineNumber))
				result.Valid = false
			}

			// Check for hardcoded POSTGRES_DB (not POSTGRES_DB_FILE)
			if strings.Contains(lineTrimmed, "POSTGRES_DB:") && !strings.Contains(lineTrimmed, "POSTGRES_DB_FILE:") && !strings.Contains(lineTrimmed, "#") {
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s:%d: Hardcoded POSTGRES_DB detected. Use POSTGRES_DB_FILE with Docker secrets instead", filename, lineNumber))
				result.Valid = false
			}

			// Check for hardcoded database URLs in connection strings
			if strings.Contains(lineTrimmed, "postgresql://") || strings.Contains(lineTrimmed, "postgres://") {
				if !strings.Contains(line, "file:///run/secrets/") && !strings.Contains(line, "#") {
					result.Errors = append(result.Errors,
						fmt.Sprintf("%s:%d: Hardcoded database URL detected. Use 'file:///run/secrets/postgres_url.secret' instead", filename, lineNumber))
					result.Valid = false
				}
			}
		}
	}
}

// ValidateAllDeployments validates all deployments in the given root directory.
func ValidateAllDeployments(deploymentsRoot string) ([]ValidationResult, error) {
	var results []ValidationResult
	// Service deployments (PRODUCT-SERVICE pattern)
	serviceNames := []string{
		"jose-ja", "cipher-im", "pki-ca", "sm-kms",
		"identity-authz", "identity-idp", "identity-rp", "identity-rs", "identity-spa",
	}

	for _, svc := range serviceNames {
		svcPath := filepath.Join(deploymentsRoot, svc)
		if _, err := os.Stat(svcPath); err == nil {
			result, err := ValidateDeploymentStructure(svcPath, svc, "PRODUCT-SERVICE")
			if err != nil {
				return nil, fmt.Errorf("failed to validate %s: %w", svc, err)
			}

			results = append(results, *result)
		}
	}

	// Template deployment
	templatePath := filepath.Join(deploymentsRoot, "template")
	if _, err := os.Stat(templatePath); err == nil {
		result, err := ValidateDeploymentStructure(templatePath, "template", "template")
		if err != nil {
			return nil, fmt.Errorf("failed to validate template: %w", err)
		}

		results = append(results, *result)
	}

	// Infrastructure deployments
	infraNames := []string{"postgres", "citus", "telemetry", "compose"}
	for _, infra := range infraNames {
		infraPath := filepath.Join(deploymentsRoot, infra)
		if _, err := os.Stat(infraPath); err == nil {
			result, err := ValidateDeploymentStructure(infraPath, infra, "infrastructure")
			if err != nil {
				return nil, fmt.Errorf("failed to validate %s: %w", infra, err)
			}

			results = append(results, *result)
		}
	}

	return results, nil
}

// FormatResults formats validation results for human-readable output.
func FormatResults(results []ValidationResult) string {
	var sb strings.Builder

	validCount := 0

	for _, r := range results {
		if r.Valid {
			validCount++
		}
	}

	sb.WriteString(fmt.Sprintf("Validated %d deployments: %d valid, %d with issues\n\n", len(results), validCount, len(results)-validCount))

	// Sort results: invalid first, then by path
	sorted := make([]ValidationResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Valid != sorted[j].Valid {
			return !sorted[i].Valid // Invalid first
		}

		return sorted[i].Path < sorted[j].Path
	})

	for _, r := range sorted {
		status := "✅ VALID"
		if !r.Valid {
			status = "❌ INVALID"
		}

		sb.WriteString(fmt.Sprintf("%s %s (%s)\n", status, filepath.Base(r.Path), r.Type))

		if len(r.MissingDirs) > 0 {
			sb.WriteString(fmt.Sprintf("  Missing directories: %s\n", strings.Join(r.MissingDirs, ", ")))
		}

		if len(r.MissingFiles) > 0 {
			sb.WriteString(fmt.Sprintf("  Missing files: %s\n", strings.Join(r.MissingFiles, ", ")))
		}

		if len(r.MissingSecrets) > 0 {
			sb.WriteString(fmt.Sprintf("  Missing secrets: %s\n", strings.Join(r.MissingSecrets, ", ")))
		}

		if len(r.Errors) > 0 {
			for _, err := range r.Errors {
				sb.WriteString(fmt.Sprintf("  ERROR: %s\n", err))
			}
		}

		if len(r.Warnings) > 0 {
			for _, warn := range r.Warnings {
				sb.WriteString(fmt.Sprintf("  WARN: %s\n", warn))
			}
		}
	}

	return sb.String()
}
