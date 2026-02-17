package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	if structType == DeploymentTypeProductService {
		validateConfigFiles(basePath, deploymentName, result)
	}

	// Check PRODUCT-level specific requirements
	if structType == DeploymentTypeProduct {
		validateProductSecrets(basePath, deploymentName, result)
	}

	// Check SUITE-level specific requirements
	if structType == DeploymentTypeSuite {
		validateSuiteSecrets(basePath, result)
	}

	// Check delegation pattern (SUITE → PRODUCT → SERVICE)
	checkDelegationPattern(basePath, deploymentName, structType, result)

	// Check browser/service credentials for all services
	checkBrowserServiceCredentials(basePath, deploymentName, structType, result)

	// Check OTLP protocol overrides
	checkOTLPProtocolOverride(basePath, deploymentName, structType, result)

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

// validateProductSecrets validates PRODUCT-level hash_pepper.secret and .never files.
// See: docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md Section 5.2 "Layered Pepper Strategy".
func validateProductSecrets(basePath string, productName string, result *ValidationResult) {
	expectedSecret := fmt.Sprintf("%s-hash_pepper.secret", productName)
	secretPath := filepath.Join(basePath, "secrets", expectedSecret)

	if _, err := os.Stat(secretPath); os.IsNotExist(err) {
		result.MissingSecrets = append(result.MissingSecrets, expectedSecret)
		result.Errors = append(result.Errors,
			fmt.Sprintf("PRODUCT-level deployment MUST have secrets/%s for shared SSO/federation within product", expectedSecret))
		result.Valid = false
	}

	// Check for .never files (documents prohibition)
	neverFiles := []string{
		fmt.Sprintf("%s-unseal_1of5.secret.never", productName),
		fmt.Sprintf("%s-unseal_2of5.secret.never", productName),
		fmt.Sprintf("%s-unseal_3of5.secret.never", productName),
		fmt.Sprintf("%s-unseal_4of5.secret.never", productName),
		fmt.Sprintf("%s-unseal_5of5.secret.never", productName),
		fmt.Sprintf("%s-postgres_username.secret.never", productName),
		fmt.Sprintf("%s-postgres_password.secret.never", productName),
		fmt.Sprintf("%s-postgres_database.secret.never", productName),
		fmt.Sprintf("%s-postgres_url.secret.never", productName),
	}

	for _, neverFile := range neverFiles {
		neverPath := filepath.Join(basePath, "secrets", neverFile)
		if _, err := os.Stat(neverPath); os.IsNotExist(err) {
			result.MissingSecrets = append(result.MissingSecrets, neverFile)
			result.Errors = append(result.Errors,
				fmt.Sprintf("Missing secrets/%s - documents that these secrets MUST NOT be shared at PRODUCT level", neverFile))
			result.Valid = false
		}
	}
}

// validateSuiteSecrets validates SUITE-level hash_pepper.secret and .never files.
// See: docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md Section 5.2 "Layered Pepper Strategy".
func validateSuiteSecrets(basePath string, result *ValidationResult) {
	expectedSecret := "cryptoutil-hash_pepper.secret"
	secretPath := filepath.Join(basePath, "secrets", expectedSecret)

	if _, err := os.Stat(secretPath); os.IsNotExist(err) {
		result.MissingSecrets = append(result.MissingSecrets, expectedSecret)
		result.Errors = append(result.Errors,
			fmt.Sprintf("SUITE-level deployment MUST have secrets/%s for cross-product SSO/federation", expectedSecret))
		result.Valid = false
	}

	// Check for .never files (documents prohibition)
	neverFiles := []string{
		"cryptoutil-unseal_1of5.secret.never",
		"cryptoutil-unseal_2of5.secret.never",
		"cryptoutil-unseal_3of5.secret.never",
		"cryptoutil-unseal_4of5.secret.never",
		"cryptoutil-unseal_5of5.secret.never",
		"cryptoutil-postgres_username.secret.never",
		"cryptoutil-postgres_password.secret.never",
		"cryptoutil-postgres_database.secret.never",
		"cryptoutil-postgres_url.secret.never",
	}

	for _, neverFile := range neverFiles {
		neverPath := filepath.Join(basePath, "secrets", neverFile)
		if _, err := os.Stat(neverPath); os.IsNotExist(err) {
			result.MissingSecrets = append(result.MissingSecrets, neverFile)
			result.Errors = append(result.Errors,
				fmt.Sprintf("Missing secrets/%s - documents that these secrets MUST NOT be shared at SUITE level", neverFile))
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

