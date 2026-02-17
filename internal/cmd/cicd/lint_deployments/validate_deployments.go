package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

		if foundProducts < productsWithOneServiceCount {
			result.Warnings = append(result.Warnings,
				"Suite should include all 4 products (sm, pki, cipher, jose) via PRODUCT-level compose")
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
	serviceNames := []string{
		"sm-kms", "pki-ca", "cipher-im", "jose-ja",
		"identity-authz", "identity-idp", "identity-rp", "identity-rs", "identity-spa",
	}

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

