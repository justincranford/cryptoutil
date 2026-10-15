// Copyright (c) 2025-2026 Justin Cranford.
package lint_deployments

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// credentialKeyPatterns matches environment variable keys that indicate credentials.
var credentialKeyPatterns = regexp.MustCompile(`(?i)^(.*_)?(PASSWORD|PASS|PASSWD|SECRET|TOKEN|KEY)$`)

// infrastructureServices are services excluded from credential checks per ENG-HANDBOOK.md Section 13.3.
// "Infrastructure deployments excluded" from Docker secrets requirement.
var infrastructureServices = map[string]bool{
	cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm:             true,
	cryptoutilSharedMagic.IME2EOtelCollectorContainer:              true,
	cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib: true,
}

// extractSecretName extracts the secret name from a secret reference.
func extractSecretName(secretRef any) string {
	switch v := secretRef.(type) {
	case string:
		return v
	case map[string]any:
		if source, ok := v["source"]; ok {
			if s, ok := source.(string); ok {
				return s
			}
		}
	}

	return ""
}

// validateNoHardcodedCredentials ensures credential environment variables use mounted Docker secrets.
func validateNoHardcodedCredentials(compose *composeFile, result *ComposeValidationResult) {
	for _, name := range sortedServiceNames(compose) {
		if infrastructureServices[name] {
			continue
		}

		svc := compose.Services[name]

		for key, value := range extractEnvironmentVars(&svc) {
			validateCredentialEnvironmentVariable(name, key, value, &svc, compose, result)
		}

		envFiles, err := extractComposeEnvFiles(svc.EnvFile)
		if err != nil {
			appendCredentialValidationError(result,
				fmt.Sprintf("service '%s': invalid env_file declaration: %v", name, err))

			continue
		}

		for _, envFile := range envFiles {
			envFilePath, err := resolveComposeEnvFilePath(svc.composeDir, envFile.path)
			if err != nil {
				appendCredentialValidationError(result,
					fmt.Sprintf("service '%s': cannot validate env_file %q: %v", name, envFile.path, err))

				continue
			}

			envFileContents, err := os.ReadFile(envFilePath)
			if err != nil {
				if !envFile.required && errors.Is(err, os.ErrNotExist) {
					continue
				}

				appendCredentialValidationError(result,
					fmt.Sprintf("service '%s': failed to read env_file %q: %v", name, envFile.path, err))

				continue
			}

			for _, variable := range parseComposeEnvFile(string(envFileContents)) {
				validateCredentialEnvironmentVariable(name, variable.key, variable.value, &svc, compose, result)
			}
		}
	}
}

type composeEnvFile struct {
	path     string
	required bool
}

type composeEnvironmentVariable struct {
	key   string
	value string
}

func extractComposeEnvFiles(value any) ([]composeEnvFile, error) {
	var envFiles []composeEnvFile

	appendEnvFile := func(value any) error {
		switch envFile := value.(type) {
		case string:
			if envFile == "" {
				return fmt.Errorf("env_file path must not be empty")
			}

			envFiles = append(envFiles, composeEnvFile{path: envFile, required: true})
		case map[string]any:
			path, ok := envFile["path"].(string)
			if !ok || path == "" {
				return fmt.Errorf("env_file mapping must have a non-empty string path")
			}

			required := true

			if configuredRequired, exists := envFile["required"]; exists {
				var ok bool

				required, ok = configuredRequired.(bool)
				if !ok {
					return fmt.Errorf("env_file required value must be a boolean")
				}
			}

			envFiles = append(envFiles, composeEnvFile{path: path, required: required})
		default:
			return fmt.Errorf("env_file entry must be a string or mapping")
		}

		return nil
	}

	switch envFilesValue := value.(type) {
	case nil:
		return nil, nil
	case string:
		if err := appendEnvFile(envFilesValue); err != nil {
			return nil, err
		}
	case []any:
		for _, envFile := range envFilesValue {
			if err := appendEnvFile(envFile); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("env_file must be a string or sequence")
	}

	return envFiles, nil
}

func resolveComposeEnvFilePath(composeDir, envFilePath string) (string, error) {
	if strings.Contains(envFilePath, "$") {
		return "", fmt.Errorf("path contains unresolved variable interpolation")
	}

	if filepath.IsAbs(envFilePath) {
		return envFilePath, nil
	}

	return filepath.Join(composeDir, envFilePath), nil
}

func parseComposeEnvFile(contents string) []composeEnvironmentVariable {
	var variables []composeEnvironmentVariable

	for _, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")

		separator := strings.IndexAny(line, "=:")
		if separator < 0 {
			variables = append(variables, composeEnvironmentVariable{key: strings.Trim(line, `"'`)})

			continue
		}

		key := strings.Trim(strings.TrimSpace(line[:separator]), `"'`)

		value := strings.TrimSpace(line[separator+1:])
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}

		variables = append(variables, composeEnvironmentVariable{key: key, value: value})
	}

	return variables
}

func validateCredentialEnvironmentVariable(name, key, value string, svc *composeService, compose *composeFile, result *ComposeValidationResult) {
	upperKey := strings.ToUpper(key)
	fileReference := strings.HasSuffix(upperKey, "_FILE")

	credentialKey := key
	if fileReference {
		credentialKey = strings.TrimSuffix(upperKey, "_FILE")
	}

	if !credentialKeyPatterns.MatchString(credentialKey) {
		return
	}

	if fileReference && isDockerSecretFileReference(value, svc, compose) {
		return
	}

	appendCredentialValidationError(result,
		fmt.Sprintf("service '%s': environment variable '%s' appears to contain hardcoded credentials or is not backed by a declared, mounted Docker secret through its _FILE variant",
			name, key))
}

func appendCredentialValidationError(result *ComposeValidationResult, message string) {
	result.Errors = append(result.Errors, message+" | See: ENG-HANDBOOK.md Section 13.3")
	result.Valid = false
}

func isDockerSecretFileReference(value string, service *composeService, compose *composeFile) bool {
	prefix := cryptoutilSharedMagic.DockerSecretsDir + "/"
	if !strings.HasPrefix(value, prefix) {
		return false
	}

	secretName := strings.TrimPrefix(value, prefix)
	if !isDockerSecretName(secretName) {
		return false
	}

	for _, secretRef := range service.Secrets {
		sourceName := extractSecretName(secretRef)

		secret, defined := compose.Secrets[sourceName]
		if sourceName == "" || !defined || (!secret.External && strings.TrimSpace(secret.File) == "") {
			continue
		}

		if dockerSecretTargetName(secretRef) == secretName {
			return true
		}
	}

	return false
}

func dockerSecretTargetName(secretRef any) string {
	targetName := extractSecretName(secretRef)
	if secretMapping, ok := secretRef.(map[string]any); ok {
		if configuredTarget, ok := secretMapping["target"].(string); ok && configuredTarget != "" {
			targetName = configuredTarget
		}
	}

	prefix := cryptoutilSharedMagic.DockerSecretsDir + "/"
	if trimmedTarget, found := strings.CutPrefix(targetName, prefix); found {
		targetName = trimmedTarget
	}

	if !isDockerSecretName(targetName) {
		return ""
	}

	return targetName
}

func isDockerSecretName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\")
}

// extractEnvironmentVars extracts environment variables from a service.
func extractEnvironmentVars(svc *composeService) map[string]string {
	envVars := make(map[string]string)

	if svc.Environment == nil {
		return envVars
	}

	switch v := svc.Environment.(type) {
	case map[string]any:
		for key, val := range v {
			if val != nil {
				envVars[key] = fmt.Sprintf("%v", val)
			} else {
				envVars[key] = ""
			}
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				parts := strings.SplitN(s, "=", 2)
				if len(parts) == 2 {
					envVars[parts[0]] = parts[1]
				} else {
					envVars[parts[0]] = ""
				}
			}
		}
	}

	return envVars
}

// validateBindMountSecurity checks for dangerous bind mounts.
func validateBindMountSecurity(compose *composeFile, result *ComposeValidationResult) {
	dangerousMounts := []string{
		"/var/run/docker.sock",
		"/run/docker.sock",
		"/etc/shadow",
		"/etc/passwd",
	}

	for _, name := range sortedServiceNames(compose) {
		svc := compose.Services[name]

		for _, volume := range svc.Volumes {
			for _, dangerous := range dangerousMounts {
				if strings.Contains(volume, dangerous) {
					result.Errors = append(result.Errors,
						fmt.Sprintf("service '%s': dangerous bind mount detected: %s | See: ENG-HANDBOOK.md Section 12",
							name, volume))
					result.Valid = false
				}
			}
		}
	}
}
