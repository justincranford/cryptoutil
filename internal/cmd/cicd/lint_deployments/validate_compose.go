package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComposeValidationResult holds the outcome of compose file validation.
type ComposeValidationResult struct {
	Path     string
	Valid    bool
	Errors   []string
	Warnings []string
}

// composeFile represents a Docker Compose file structure for validation.
type composeFile struct {
	Include  []composeInclude          `yaml:"include"`
	Services map[string]composeService `yaml:"services"`
	Secrets  map[string]composeSecret  `yaml:"secrets"`
	Networks map[string]any            `yaml:"networks"`
	Volumes  map[string]any            `yaml:"volumes"`
}

// composeInclude represents an include directive.
type composeInclude struct {
	Path string `yaml:"path"`
}

// composeService represents a service in a compose file.
type composeService struct {
	Image       string              `yaml:"image"`
	Build       any                 `yaml:"build"`
	Ports       []string            `yaml:"ports"`
	Volumes     []string            `yaml:"volumes"`
	Environment any                 `yaml:"environment"`
	Secrets     []any               `yaml:"secrets"`
	DependsOn   any                 `yaml:"depends_on"`
	Healthcheck *composeHealthcheck `yaml:"healthcheck"`
	Networks    any                 `yaml:"networks"`
	Command     any                 `yaml:"command"`
	Deploy      any                 `yaml:"deploy"`
	Profiles    []string            `yaml:"profiles"`
	WorkingDir  string              `yaml:"working_dir"`
	Entrypoint  any                 `yaml:"entrypoint"`
	ShmSize     string              `yaml:"shm_size"`
}

// composeHealthcheck represents a healthcheck configuration.
type composeHealthcheck struct {
	Test        any    `yaml:"test"`
	Interval    string `yaml:"interval"`
	Timeout     string `yaml:"timeout"`
	Retries     int    `yaml:"retries"`
	StartPeriod string `yaml:"start_period"`
}

// composeSecret represents a secret definition.
type composeSecret struct {
	File     string `yaml:"file"`
	External bool   `yaml:"external"`
}

// credentialKeyPatterns matches environment variable keys that indicate credentials.
var credentialKeyPatterns = regexp.MustCompile(`(?i)^(.*_)?(PASSWORD|PASSWD|SECRET|TOKEN|API_KEY|PRIVATE_KEY)$`)

// fileReferencePattern matches Docker secret file references (safe pattern).
var fileReferencePattern = regexp.MustCompile(`(?i)_FILE$`)

// ValidateComposeFile validates a single compose file against 7 validation types.
func ValidateComposeFile(composePath string) (*ComposeValidationResult, error) {
	result := &ComposeValidationResult{
		Path:  composePath,
		Valid: true,
	}

	// Validation 1: Schema validation (YAML parsing + include resolution).
	compose, parseErr := parseComposeWithIncludes(composePath)
	if parseErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("YAML parse error: %v", parseErr))
		result.Valid = false

		return result, nil
	}

	if len(compose.Services) == 0 {
		result.Errors = append(result.Errors, "no services defined in compose file")
		result.Valid = false

		return result, nil
	}

	// Validations 2-7.
	validatePortConflicts(compose, result)
	validateHealthChecks(compose, result)
	validateDependencyChains(compose, result)
	validateSecretReferences(compose, result)
	validateNoHardcodedCredentials(compose, result)
	validateBindMountSecurity(compose, result)

	return result, nil
}

// parseComposeWithIncludes parses a compose file and merges secrets/services from included files.
func parseComposeWithIncludes(composePath string) (*composeFile, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file %s: %w", composePath, err)
	}

	var compose composeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("YAML parse error in %s: %w", composePath, err)
	}

	// Resolve includes to merge secrets and services for reference validation.
	baseDir := filepath.Dir(composePath)
	for _, inc := range compose.Include {
		mergeIncludedFile(baseDir, inc.Path, &compose)
	}

	return &compose, nil
}

// mergeIncludedFile reads an included compose file and merges its secrets/services.
func mergeIncludedFile(baseDir string, includePath string, compose *composeFile) {
	if includePath == "" {
		return
	}

	fullPath := filepath.Join(baseDir, includePath)

	includeData, err := os.ReadFile(fullPath)
	if err != nil {
		return // Include not found is non-blocking.
	}

	var included composeFile
	if err := yaml.Unmarshal(includeData, &included); err != nil {
		return
	}

	if compose.Secrets == nil {
		compose.Secrets = make(map[string]composeSecret)
	}

	for name, secret := range included.Secrets {
		if _, exists := compose.Secrets[name]; !exists {
			compose.Secrets[name] = secret
		}
	}

	if compose.Services == nil {
		compose.Services = make(map[string]composeService)
	}

	for name, svc := range included.Services {
		if _, exists := compose.Services[name]; !exists {
			compose.Services[name] = svc
		}
	}
}

// validatePortConflicts checks for duplicate host port bindings across services.
func validatePortConflicts(compose *composeFile, result *ComposeValidationResult) {
	hostPorts := make(map[string]string)
	serviceNames := sortedServiceNames(compose)

	for _, name := range serviceNames {
		svc := compose.Services[name]

		for _, portMapping := range svc.Ports {
			hostPort := extractHostPort(portMapping)
			if hostPort == "" {
				continue
			}

			if existingService, exists := hostPorts[hostPort]; exists {
				result.Errors = append(result.Errors,
					fmt.Sprintf("port conflict: host port %s used by both '%s' and '%s'",
						hostPort, existingService, name))
				result.Valid = false
			} else {
				hostPorts[hostPort] = name
			}
		}
	}
}

// portMappingPartsIPHostContainer is the number of parts in an ip:host:container port mapping.
const portMappingPartsIPHostContainer = 3

// extractHostPort extracts the host port from a port mapping string.
func extractHostPort(portMapping string) string {
	portMapping = strings.Trim(portMapping, `"'`)
	parts := strings.Split(portMapping, ":")

	switch len(parts) {
	case 1:
		return "" // Just container port.
	case 2:
		return parts[0] // host:container.
	case portMappingPartsIPHostContainer:
		return parts[0] + ":" + parts[1] // ip:host:container.
	default:
		return ""
	}
}

// validateHealthChecks verifies all non-exempt services have health checks.
func validateHealthChecks(compose *composeFile, result *ComposeValidationResult) {
	for _, name := range sortedServiceNames(compose) {
		svc := compose.Services[name]
		if isExemptFromHealthcheck(name, &svc) {
			continue
		}

		if svc.Healthcheck == nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("service '%s' missing healthcheck", name))
			result.Valid = false
		}
	}
}

// isExemptFromHealthcheck returns true for ephemeral services.
func isExemptFromHealthcheck(name string, svc *composeService) bool {
	if strings.HasPrefix(name, "builder-") || strings.HasPrefix(name, "healthcheck-") ||
		strings.HasPrefix(name, "setup-") || strings.HasPrefix(name, "init-") {
		return true
	}

	// Exact name matches for known init containers.
	if strings.HasSuffix(name, "-setup") || strings.HasSuffix(name, "-init") {
		return true
	}

	if svc.Entrypoint != nil {
		if strings.Contains(fmt.Sprintf("%v", svc.Entrypoint), "echo") {
			return true
		}
	}

	return false
}

// validateDependencyChains checks that depends_on references exist.
func validateDependencyChains(compose *composeFile, result *ComposeValidationResult) {
	for _, name := range sortedServiceNames(compose) {
		svc := compose.Services[name]

		for _, dep := range extractDependencies(&svc) {
			if _, exists := compose.Services[dep]; !exists {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("service '%s' depends on '%s' which is not defined locally (may come from include)",
						name, dep))
			}
		}
	}
}

// extractDependencies extracts dependency service names from depends_on field.
func extractDependencies(svc *composeService) []string {
	if svc.DependsOn == nil {
		return nil
	}

	var deps []string

	switch v := svc.DependsOn.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				deps = append(deps, s)
			}
		}
	case map[string]any:
		for dep := range v {
			deps = append(deps, dep)
		}
	}

	sort.Strings(deps)

	return deps
}

// validateSecretReferences checks service secret refs are defined at top level.
func validateSecretReferences(compose *composeFile, result *ComposeValidationResult) {
	for _, name := range sortedServiceNames(compose) {
		svc := compose.Services[name]

		for _, secretRef := range svc.Secrets {
			secretName := extractSecretName(secretRef)
			if secretName == "" {
				continue
			}

			if compose.Secrets == nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("service '%s' references secret '%s' but no secrets section defined",
						name, secretName))
				result.Valid = false

				continue
			}

			if _, defined := compose.Secrets[secretName]; !defined {
				result.Errors = append(result.Errors,
					fmt.Sprintf("service '%s' references undefined secret '%s'",
						name, secretName))
				result.Valid = false
			}
		}
	}
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

// infrastructureServices are services excluded from credential checks per ARCHITECTURE.md Section 12.6.
// "Infrastructure deployments excluded" from Docker secrets requirement.
var infrastructureServices = map[string]bool{
	"grafana-otel-lgtm":                           true,
	"opentelemetry-collector-contrib":             true,
	"healthcheck-opentelemetry-collector-contrib": true,
}

// validateNoHardcodedCredentials checks for hardcoded credentials in env vars.
func validateNoHardcodedCredentials(compose *composeFile, result *ComposeValidationResult) {
	for _, name := range sortedServiceNames(compose) {
		if infrastructureServices[name] {
			continue
		}

		svc := compose.Services[name]

		for key, value := range extractEnvironmentVars(&svc) {
			if fileReferencePattern.MatchString(key) {
				continue
			}

			if credentialKeyPatterns.MatchString(key) && value != "" &&
				!strings.HasPrefix(value, "$") && !strings.HasPrefix(value, "/run/secrets/") {
				result.Errors = append(result.Errors,
					fmt.Sprintf("service '%s': environment variable '%s' appears to contain hardcoded credentials",
						name, key))
				result.Valid = false
			}
		}
	}
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
						fmt.Sprintf("service '%s': dangerous bind mount detected: %s",
							name, volume))
					result.Valid = false
				}
			}
		}
	}
}

// sortedServiceNames returns service names in sorted order for deterministic output.
func sortedServiceNames(compose *composeFile) []string {
	names := make([]string, 0, len(compose.Services))
	for name := range compose.Services {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// FormatComposeValidationResult formats a ComposeValidationResult for display.
func FormatComposeValidationResult(result *ComposeValidationResult) string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(&sb, "Compose Validation: %s\n", result.Path)

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
