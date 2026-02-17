package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Deployment type constants.
const (
	DeploymentTypeSuite          = "SUITE"
	DeploymentTypeProduct        = "PRODUCT"
	DeploymentTypeProductService = "PRODUCT-SERVICE"
	DeploymentTypeInfrastructure = "infrastructure"
	DeploymentTypeTemplate       = "template"
)

// Single-service product count (sm, pki, cipher, jose).
const singleServiceProductCount = 4

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
		"template": {
			Name:          "Template deployment (deployments/template/)",
			RequiredDirs:  []string{"secrets"},
			RequiredFiles: []string{"compose.yml"},
			RequiredSecrets: []string{
				"hash_pepper_v3.secret", "unseal_1of5.secret", "unseal_2of5.secret",
				"unseal_3of5.secret", "unseal_4of5.secret", "unseal_5of5.secret",
				"postgres_username.secret", "postgres_password.secret",
				"postgres_database.secret", "postgres_url.secret",
			},
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
		},
		DeploymentTypeProductService: {
			Name:          "PRODUCT-SERVICE deployment (e.g., jose-ja, cipher-im)",
			RequiredDirs:  []string{"secrets", "config"},
			RequiredFiles: []string{"compose.yml", "Dockerfile"},
			OptionalFiles: []string{}, // no optional files
			RequiredSecrets: []string{
				"hash_pepper_v3.secret", "unseal_1of5.secret", "unseal_2of5.secret",
				"unseal_3of5.secret", "unseal_4of5.secret", "unseal_5of5.secret",
				"postgres_username.secret", "postgres_password.secret",
				"postgres_database.secret", "postgres_url.secret",
			},
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".md"},
		},
		DeploymentTypeProduct: {
			Name:              "PRODUCT-level deployment (e.g., identity, sm, pki, cipher, jose)",
			RequiredDirs:      []string{"secrets"},
			RequiredFiles:     []string{"compose.yml"},
			OptionalFiles:     []string{}, // no optional files
			RequiredSecrets:   []string{}, // Validated by validateProductSecrets() with product-specific prefixes
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".never", ".md"},
		},
		DeploymentTypeSuite: {
			Name:              "SUITE-level deployment (cryptoutil - all 9 services)",
			RequiredDirs:      []string{"secrets"},
			RequiredFiles:     []string{"compose.yml"},
			OptionalFiles:     []string{}, // no optional files
			RequiredSecrets:   []string{}, // Validated by validateSuiteSecrets() with suite-specific prefixes
			AllowedExtensions: []string{".yml", ".yaml", ".secret", ".never", ".md"},
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
			result, err := ValidateDeploymentStructure(svcPath, svc, DeploymentTypeProductService)
			if err != nil {
				return nil, fmt.Errorf("failed to validate %s: %w", svc, err)
			}

			results = append(results, *result)
		}
	}

	// PRODUCT-level deployments
	productNames := []string{"identity", "sm", "pki", "cipher", "jose"}
	for _, product := range productNames {
		productPath := filepath.Join(deploymentsRoot, product)
		if _, err := os.Stat(productPath); err == nil {
			result, err := ValidateDeploymentStructure(productPath, product, DeploymentTypeProduct)
			if err != nil {
				return nil, fmt.Errorf("failed to validate %s: %w", product, err)
			}

			results = append(results, *result)
		}
	}

	// SUITE-level deployment
	suitePath := filepath.Join(deploymentsRoot, "cryptoutil")
	if _, err := os.Stat(suitePath); err == nil {
		result, err := ValidateDeploymentStructure(suitePath, "cryptoutil", DeploymentTypeSuite)
		if err != nil {
			return nil, fmt.Errorf("failed to validate cryptoutil: %w", err)
		}

		results = append(results, *result)
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
	infraNames := []string{"shared-postgres", "shared-citus", "shared-telemetry", "compose"}
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

	// Cross-service validation: Database isolation check
	dbErrors := checkDatabaseIsolation(serviceNames, deploymentsRoot)
	if len(dbErrors) > 0 {
		dbResult := ValidationResult{
			Path:   deploymentsRoot,
			Type:   "DATABASE-ISOLATION",
			Valid:  false,
			Errors: dbErrors,
		}
		results = append(results, dbResult)
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

// checkDelegationPattern validates SUITE → PRODUCT → SERVICE delegation in compose includes.
// CRITICAL: Suite MUST delegate to products, products MUST delegate to services.
func checkDelegationPattern(basePath string, deploymentName string, structType string, result *ValidationResult) {
	if structType != DeploymentTypeSuite && structType != DeploymentTypeProduct {
		return
	}

	composePath := filepath.Join(basePath, "compose.yml")

	content, err := os.ReadFile(composePath)
	if err != nil {
		return
	}

	text := string(content)

	if structType == DeploymentTypeSuite {
		// Suite MUST include product-level compose files, NOT service-level
		invalidPatterns := []string{
			"../sm-kms/compose.yml",
			"../pki-ca/compose.yml",
			"../cipher-im/compose.yml",
			"../jose-ja/compose.yml",
		}
		validPatterns := []string{
			"../sm/compose.yml",
			"../pki/compose.yml",
			"../cipher/compose.yml",
			"../jose/compose.yml",
		}

		for _, invalid := range invalidPatterns {
			if strings.Contains(text, invalid) {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Suite compose.yml MUST delegate to PRODUCT-level (use %s, not %s)",
						strings.Replace(invalid, "-kms", "", 1),
						invalid))
				result.Valid = false
			}
		}

		// Check that it includes product-level
		foundProducts := 0

		for _, valid := range validPatterns {
			if strings.Contains(text, valid) {
				foundProducts++
			}
		}

		if foundProducts < singleServiceProductCount {
			result.Warnings = append(result.Warnings,
				"Suite should include all 4 single-service products via PRODUCT-level compose")
		}
	}

	if structType == DeploymentTypeProduct {
		// Product MUST include service-level compose files
		if deploymentName == "sm" && !strings.Contains(text, "../sm-kms/compose.yml") {
			result.Errors = append(result.Errors, "Product sm/compose.yml MUST include ../sm-kms/compose.yml")
			result.Valid = false
		}

		if deploymentName == "pki" && !strings.Contains(text, "../pki-ca/compose.yml") {
			result.Errors = append(result.Errors, "Product pki/compose.yml MUST include ../pki-ca/compose.yml")
			result.Valid = false
		}

		if deploymentName == "cipher" && !strings.Contains(text, "../cipher-im/compose.yml") {
			result.Errors = append(result.Errors, "Product cipher/compose.yml MUST include ../cipher-im/compose.yml")
			result.Valid = false
		}

		if deploymentName == "jose" && !strings.Contains(text, "../jose-ja/compose.yml") {
			result.Errors = append(result.Errors, "Product jose/compose.yml MUST include ../jose-ja/compose.yml")
			result.Valid = false
		}
	}
}

// checkDatabaseIsolation validates that each service has unique database credentials.
// CRITICAL: ALL services MUST have isolated database storage (unique db/username/password).
func checkDatabaseIsolation(deploymentsList []string, deploymentsRoot string) []string {
	serviceNames := []string{"sm-kms", "pki-ca", "cipher-im", "jose-ja",
		"identity-authz", "identity-idp", "identity-rp", "identity-rs", "identity-spa"}

	databaseNames := make(map[string][]string)
	usernames := make(map[string][]string)

	for _, svc := range serviceNames {
		dbPath := filepath.Join(deploymentsRoot, svc, "secrets", "postgres_database.secret")
		userPath := filepath.Join(deploymentsRoot, svc, "secrets", "postgres_username.secret")

		if dbContent, err := os.ReadFile(dbPath); err == nil {
			dbName := strings.TrimSpace(string(dbContent))
			databaseNames[dbName] = append(databaseNames[dbName], svc)
		}

		if userContent, err := os.ReadFile(userPath); err == nil {
			username := strings.TrimSpace(string(userContent))
			usernames[username] = append(usernames[username], svc)
		}
	}

	var errors []string

	for dbName, services := range databaseNames {
		if len(services) > 1 {
			errors = append(errors,
				fmt.Sprintf("Database '%s' shared by services: %v (MUST be unique per service)", dbName, services))
		}
	}

	for username, services := range usernames {
		if len(services) > 1 {
			errors = append(errors,
				fmt.Sprintf("Username '%s' shared by services: %v (MUST be unique per service)", username, services))
		}
	}

	return errors
}

// checkBrowserServiceCredentials validates that all services have browser/service credential files.
// CRITICAL: ALL services MUST have unique browser_username/password and service_username/password.
func checkBrowserServiceCredentials(basePath string, deploymentName string, structType string, result *ValidationResult) {
	if structType != DeploymentTypeProductService {
		return
	}

	requiredCredFiles := []string{
		"browser_username.secret",
		"browser_password.secret",
		"service_username.secret",
		"service_password.secret",
	}

	secretsPath := filepath.Join(basePath, "secrets")
	for _, credFile := range requiredCredFiles {
		credPath := filepath.Join(secretsPath, credFile)
		if _, err := os.Stat(credPath); os.IsNotExist(err) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Missing required credential file: secrets/%s", credFile))
			result.Valid = false
		}
	}
}

// checkOTLPProtocolOverride validates that config files do not override OTLP protocol.
// Services should use default protocol, not explicitly set grpc:// or http://.
func checkOTLPProtocolOverride(basePath string, deploymentName string, structType string, result *ValidationResult) {
	if structType != DeploymentTypeProductService {
		return
	}

	configPath := filepath.Join(basePath, "config")

	configFiles, err := filepath.Glob(filepath.Join(configPath, "*.yml"))
	if err != nil {
		return
	}

	for _, configFile := range configFiles {
		content, err := os.ReadFile(configFile)
		if err != nil {
			continue
		}

		text := string(content)

		lineNumber := 0
		for _, line := range strings.Split(text, "\n") {
			lineNumber++

			if strings.Contains(line, "otlp-endpoint:") {
				// Check if line contains protocol prefix
				if strings.Contains(line, "grpc://") || strings.Contains(line, "http://") {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("%s:%d: OTLP endpoint should not specify protocol (remove grpc:// or http://, use hostname:port only)",
							filepath.Base(configFile), lineNumber))
				}
			}
		}
	}
}
