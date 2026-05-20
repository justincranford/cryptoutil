package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	postgresSharedDirName                  = "shared-postgres"
	postgresEnvFileName                    = ".env.postgres"
	postgresSecretsDirKey                  = "POSTGRES_SECRETS_DIR"
	postgresSecretsTemplateReferencePrefix = "${POSTGRES_SECRETS_DIR:-./secrets}/"
	postgresComposePathSection             = "ENG-HANDBOOK.md Section 12.6"
)

// PostgresSecretsDirSyncResult contains sync validation output.
type PostgresSecretsDirSyncResult struct {
	Path   string
	Valid  bool
	Errors []string
}

// ValidatePostgresSecretsDirSync checks shared-postgres reference usage and per-PS-ID env values.
func ValidatePostgresSecretsDirSync(deploymentsDir string) (*PostgresSecretsDirSyncResult, error) {
	result := &PostgresSecretsDirSyncResult{
		Path:  deploymentsDir,
		Valid: true,
	}

	validateSharedPostgresComposeReference(deploymentsDir, result)

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePostgresSecretsDirSync] Failed to read deployments directory '%s': %v", deploymentsDir, err))
		result.Valid = false

		return result, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		deploymentName := entry.Name()
		if classifyDeployment(deploymentName) != DeploymentTypeProductService {
			continue
		}

		validateServicePostgresSecretsDir(deploymentsDir, deploymentName, result)
	}

	return result, nil
}

func validateSharedPostgresComposeReference(deploymentsDir string, result *PostgresSecretsDirSyncResult) {
	sharedComposePath := filepath.Join(deploymentsDir, postgresSharedDirName, composeFileName)
	if _, err := os.Stat(sharedComposePath); os.IsNotExist(err) {
		return
	}

	sharedComposeData, err := os.ReadFile(sharedComposePath)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePostgresSecretsDirSync] Failed to read shared-postgres compose file '%s': %v", sharedComposePath, err))
		result.Valid = false

		return
	}

	sharedComposeText := string(sharedComposeData)
	if !strings.Contains(sharedComposeText, postgresSecretsTemplateReferencePrefix) {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePostgresSecretsDirSync] shared-postgres compose missing '%s' secret reference pattern | See: %s",
				postgresSecretsTemplateReferencePrefix, postgresComposePathSection))
		result.Valid = false
	}
}

func validateServicePostgresSecretsDir(deploymentsDir, deploymentName string, result *PostgresSecretsDirSyncResult) {
	envPath := filepath.Join(deploymentsDir, deploymentName, postgresEnvFileName)

	envData, err := os.ReadFile(envPath)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePostgresSecretsDirSync] Failed to read '%s': %v", envPath, err))
		result.Valid = false

		return
	}

	actualValue, found := parseEnvVarValue(string(envData), postgresSecretsDirKey)
	if !found {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePostgresSecretsDirSync] Missing %s in '%s'", postgresSecretsDirKey, envPath))
		result.Valid = false

		return
	}

	expectedValue := fmt.Sprintf("../%s/secrets", deploymentName)
	if actualValue != expectedValue {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePostgresSecretsDirSync] %s mismatch in '%s': got '%s', want '%s'",
				postgresSecretsDirKey, envPath, actualValue, expectedValue))
		result.Valid = false
	}
}

func parseEnvVarValue(content, key string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if !strings.HasPrefix(trimmed, key+"=") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			return "", false
		}

		return strings.TrimSpace(parts[1]), true
	}

	return "", false
}

// FormatPostgresSecretsDirSyncResult renders sync validation output.
func FormatPostgresSecretsDirSyncResult(result *PostgresSecretsDirSyncResult) string {
	if result == nil {
		return "[ValidatePostgresSecretsDirSync] nil result"
	}

	if result.Valid {
		return fmt.Sprintf("[ValidatePostgresSecretsDirSync] PASS: %s", result.Path)
	}

	if len(result.Errors) == 0 {
		return fmt.Sprintf("[ValidatePostgresSecretsDirSync] FAIL: %s", result.Path)
	}

	return strings.Join(result.Errors, "\n")
}
