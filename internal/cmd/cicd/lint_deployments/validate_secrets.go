package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SecretValidationResult holds results of secret validation.
type SecretValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// secretKeyPatterns are field name substrings that indicate credential fields.
var secretKeyPatterns = []string{
	"password", "passwd", "secret", cryptoutilSharedMagic.ParamToken, "api-key",
	"api_key", "private-key", "private_key", "pepper",
}

// minSecretLengthRaw is the minimum acceptable length in bytes for raw secret files.
// Decision 15:E per quizme-v3 Q3.
const minSecretLengthRaw = 32

// secretFilePatterns are substrings in secret file names that indicate high-entropy credential content.
var secretFilePatterns = []string{
	"password", "passwd", "pepper", "private_key", "private-key",
	"api_key", "api-key", "secret_key", "secret-key",
	"unseal", "hash_pepper", "hash-pepper",
}

// safeConfigPrefixes are prefixes for config values that indicate safe external references.
var safeConfigPrefixes = []string{
	"file:///run/secrets/",
	cryptoutilSharedMagic.FileURIScheme,
	"sqlite://",
	cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
}

// ValidateSecrets validates secret files and checks for inline secrets in configs.
func ValidateSecrets(deploymentPath string) (*SecretValidationResult, error) {
	result := &SecretValidationResult{Valid: true}

	info, statErr := os.Stat(deploymentPath)
	if statErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("path not found: %s", deploymentPath))

		return result, nil
	}

	if !info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("path is not a directory: %s", deploymentPath))

		return result, nil
	}

	secretsDir := filepath.Join(deploymentPath, "secrets")

	if dirInfo, err := os.Stat(secretsDir); err == nil && dirInfo.IsDir() {
		validateSecretFileLengths(secretsDir, result)
	}

	configsDir := filepath.Join(deploymentPath, "configs")

	if dirInfo, err := os.Stat(configsDir); err == nil && dirInfo.IsDir() {
		validateConfigInlineSecrets(configsDir, result)
	}

	composePath := findComposeFile(deploymentPath)
	if composePath != "" {
		validateComposeInlineSecrets(composePath, result)
	}

	return result, nil
}

// validateSecretFileLengths checks that secret files meet minimum length requirements.
func validateSecretFileLengths(secretsDir string, result *SecretValidationResult) {
	entries, err := os.ReadDir(secretsDir)
	if err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("cannot read secrets directory: %s", secretsDir))

		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isSecretFile(name) {
			continue
		}

		filePath := filepath.Join(secretsDir, name)

		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("cannot read secret file: %s", name))

			continue
		}

		trimmed := strings.TrimSpace(string(content))

		if !isHighEntropySecretFile(name) {
			continue
		}

		checkSecretLength(name, trimmed, result)
	}
}

// isSecretFile returns true if the filename indicates a secret file.
func isSecretFile(name string) bool {
	return strings.HasSuffix(name, ".secret") || strings.HasSuffix(name, ".secret.never")
}

// isHighEntropySecretFile returns true if the secret file name indicates it should contain
// high-entropy content (passwords, peppers, keys) rather than identifiers (usernames, database names).
func isHighEntropySecretFile(name string) bool {
	lower := strings.ToLower(name)

	for _, pattern := range secretFilePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// checkSecretLength validates that a secret value meets minimum length requirements.
func checkSecretLength(name string, value string, result *SecretValidationResult) {
	rawLen := len(value)
	if rawLen == 0 {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("secret file '%s' is empty", name))

		return
	}

	if rawLen >= minSecretLengthRaw {
		return
	}

	result.Warnings = append(result.Warnings,
		fmt.Sprintf("secret file '%s' has %d bytes (minimum recommended: %d)",
			name, rawLen, minSecretLengthRaw))
}

// validateConfigInlineSecrets checks config files for inline credential values.
func validateConfigInlineSecrets(configsDir string, result *SecretValidationResult) {
	entries, err := os.ReadDir(configsDir)
	if err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("cannot read configs directory: %s", configsDir))

		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !isYAMLFile(entry.Name()) {
			continue
		}

		filePath := filepath.Join(configsDir, entry.Name())
		checkConfigFileForInlineSecrets(filePath, entry.Name(), result)
	}
}

// checkConfigFileForInlineSecrets reads a config file and checks for inline credential values.
func checkConfigFileForInlineSecrets(filePath string, fileName string, result *SecretValidationResult) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("cannot read config file: %s", fileName))

		return
	}

	var config map[string]any
	if yamlErr := yaml.Unmarshal(data, &config); yamlErr != nil {
		return
	}

	scanConfigMapForSecrets(config, fileName, "", result)
}

// scanConfigMapForSecrets recursively scans a config map for inline credential values.
func scanConfigMapForSecrets(config map[string]any, fileName string, prefix string, result *SecretValidationResult) {
	for key, val := range config {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if nested, ok := val.(map[string]any); ok {
			scanConfigMapForSecrets(nested, fileName, fullKey, result)

			continue
		}

		strVal, ok := val.(string)
		if !ok {
			continue
		}

		if !isSecretFieldName(key) {
			continue
		}

		if isSafeReference(strVal) {
			continue
		}

		if strVal == "" {
			continue
		}

		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("config '%s': field '%s' appears to contain an inline secret; "+
				"use 'file:///run/secrets/' reference or move to external vault", fileName, fullKey))
	}
}

// isSecretFieldName returns true if the field name matches known credential patterns.
func isSecretFieldName(name string) bool {
	lower := strings.ToLower(name)

	for _, pattern := range secretKeyPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// isSafeReference returns true if the value uses a safe external reference pattern.
func isSafeReference(value string) bool {
	for _, prefix := range safeConfigPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}

// findComposeFile locates the compose file in a deployment directory.
func findComposeFile(deploymentPath string) string {
	candidates := []string{"compose.yml", "compose.yaml", "docker-compose.yml", "docker-compose.yaml"}

	for _, name := range candidates {
		path := filepath.Join(deploymentPath, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// validateComposeInlineSecrets checks compose environment variables for inline credential values.
func validateComposeInlineSecrets(composePath string, result *SecretValidationResult) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("cannot read compose file: %s", filepath.Base(composePath)))

		return
	}

	var compose composeFile
	if yamlErr := yaml.Unmarshal(data, &compose); yamlErr != nil {
		return
	}

	fileName := filepath.Base(composePath)

	for svcName, svc := range compose.Services {
		envMap, ok := svc.Environment.(map[string]any)
		if !ok {
			continue
		}

		checkServiceEnvForInlineSecrets(svcName, envMap, fileName, result)
	}
}

// checkServiceEnvForInlineSecrets checks a service's environment variables for inline secrets.
func checkServiceEnvForInlineSecrets(svcName string, env map[string]any, fileName string, result *SecretValidationResult) {
	for key, val := range env {
		strVal, ok := val.(string)
		if !ok {
			continue
		}

		if !isSecretFieldName(key) {
			continue
		}

		if isSafeReference(strVal) {
			continue
		}

		if strings.HasPrefix(strVal, "/run/secrets/") {
			continue
		}

		if strVal == "" {
			continue
		}

		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("compose '%s': service '%s' env '%s' contains inline secret; "+
				"use Docker secrets with _FILE suffix", fileName, svcName, key))
	}
}

// FormatSecretValidationResult formats a SecretValidationResult for display.
func FormatSecretValidationResult(result *SecretValidationResult) string {
	if result == nil {
		return "Secret Validation: SKIP (no result)"
	}

	var sb strings.Builder

	status := statusPass
	if !result.Valid {
		status = statusFail
	}

	sb.WriteString(fmt.Sprintf("Secret Validation: %s\n", status))

	for _, e := range result.Errors {
		sb.WriteString(fmt.Sprintf("  ERROR: %s\n", e))
	}

	for _, w := range result.Warnings {
		sb.WriteString(fmt.Sprintf("  WARN: %s\n", w))
	}

	return sb.String()
}
