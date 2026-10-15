// Copyright (c) 2025-2026 Justin Cranford.
package lint_deployments

import (
	"fmt"
	"regexp"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// credentialKeyPatterns matches environment variable keys that indicate credentials.
var credentialKeyPatterns = regexp.MustCompile(`(?i)^(.*_)?(PASSWORD|PASS|PASSWD|SECRET|TOKEN|API_KEY|PRIVATE_KEY)$`)

var envVariableNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// infrastructureServices are services excluded from credential checks per ENG-HANDBOOK.md Section 12.6.
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

// validateNoHardcodedCredentials checks for hardcoded credentials in env vars.
func validateNoHardcodedCredentials(compose *composeFile, result *ComposeValidationResult) {
	for _, name := range sortedServiceNames(compose) {
		if infrastructureServices[name] {
			continue
		}

		svc := compose.Services[name]

		for key, value := range extractEnvironmentVars(&svc) {
			upperKey := strings.ToUpper(key)
			fileReference := strings.HasSuffix(upperKey, "_FILE")

			credentialKey := key
			if fileReference {
				credentialKey = strings.TrimSuffix(upperKey, "_FILE")
			}

			if !credentialKeyPatterns.MatchString(credentialKey) {
				continue
			}

			if fileReference && isDockerSecretFileReference(value) {
				continue
			}

			if !fileReference && (value == "" || isSafeEnvironmentVariableReference(value) || isDockerSecretFileReference(value)) {
				continue
			}

			result.Errors = append(result.Errors,
				fmt.Sprintf("service '%s': environment variable '%s' appears to contain hardcoded credentials | See: ENG-HANDBOOK.md Section 12.6",
					name, key))
			result.Valid = false
		}
	}
}

func isDockerSecretFileReference(value string) bool {
	prefix := cryptoutilSharedMagic.DockerSecretsDir + "/"
	if !strings.HasPrefix(value, prefix) {
		return false
	}

	secretName := strings.TrimPrefix(value, prefix)

	return secretName != "" && secretName != "." && secretName != ".." && !strings.ContainsAny(secretName, "/\\")
}

func isSafeEnvironmentVariableReference(value string) bool {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		expression := value[2 : len(value)-1]

		nameEnd := strings.IndexAny(expression, ":-+?")
		if nameEnd < 0 {
			nameEnd = len(expression)
		}

		if !envVariableNamePattern.MatchString(expression[:nameEnd]) {
			return false
		}

		operator := expression[nameEnd:]
		switch {
		case operator == "":
			return true
		case strings.HasPrefix(operator, "?") || strings.HasPrefix(operator, ":?"):
			return true
		case operator == "-", operator == ":-", operator == "+", operator == ":+":
			return true
		default:
			return false
		}
	}

	return strings.HasPrefix(value, "$") && envVariableNamePattern.MatchString(strings.TrimPrefix(value, "$"))
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
