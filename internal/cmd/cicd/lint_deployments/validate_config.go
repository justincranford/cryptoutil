package lint_deployments

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigValidationResult holds validation results for a single config file.
type ConfigValidationResult struct {
	Path     string
	Valid    bool
	Errors   []string
	Warnings []string
}

// minPort is the minimum valid port number.
const minPort = 1

// maxPort is the maximum valid port number.
const maxPort = 65535

// mandatoryAdminBindAddress is the only acceptable admin bind address.
const mandatoryAdminBindAddress = "127.0.0.1"

// statusPass and statusFail are shared status labels for validation output formatting.
const (
	statusPass = "PASS"
	statusFail = "FAIL"
)

// mandatoryProtocol is the only acceptable protocol.
const mandatoryProtocol = "https"

// ValidateConfigFile validates the content of a single config YAML file.
// It checks YAML syntax, format validation, policy enforcement, and secret references.
func ValidateConfigFile(configPath string) (*ConfigValidationResult, error) {
	result := &ConfigValidationResult{
		Path:  configPath,
		Valid: true,
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("cannot read file: %s", err))

		return result, nil
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("YAML parse error: %s", err))

		return result, nil
	}

	if len(config) == 0 {
		result.Warnings = append(result.Warnings, "config file is empty")

		return result, nil
	}

	validateBindAddresses(config, result)
	validatePorts(config, result)
	validateProtocols(config, result)
	validateAdminBindPolicy(config, result)
	validateConfigSecretRefs(config, result)
	validateOTLPConfig(config, result)

	return result, nil
}

// validateBindAddresses checks that bind address values are valid IPv4 addresses.
func validateBindAddresses(config map[string]any, result *ConfigValidationResult) {
	addressKeys := []string{"bind-public-address", "bind-private-address"}

	for _, key := range addressKeys {
		val, ok := config[key]
		if !ok {
			continue
		}

		strVal, ok := val.(string)
		if !ok {
			result.Errors = append(result.Errors,
				fmt.Sprintf("'%s' must be a string, got %T", key, val))
			result.Valid = false

			continue
		}

		if net.ParseIP(strVal) == nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("'%s' is not a valid IP address: %q", key, strVal))
			result.Valid = false
		}
	}
}

// validatePorts checks that port values are within valid range 1-65535.
func validatePorts(config map[string]any, result *ConfigValidationResult) {
	portKeys := []string{"bind-public-port", "bind-private-port"}

	for _, key := range portKeys {
		val, ok := config[key]
		if !ok {
			continue
		}

		port, ok := toInt(val)
		if !ok {
			result.Errors = append(result.Errors,
				fmt.Sprintf("'%s' must be an integer, got %T", key, val))
			result.Valid = false

			continue
		}

		if port < minPort || port > maxPort {
			result.Errors = append(result.Errors,
				fmt.Sprintf("'%s' must be between %d and %d, got %d", key, minPort, maxPort, port))
			result.Valid = false
		}
	}
}

// validateProtocols checks that protocol values are "https".
func validateProtocols(config map[string]any, result *ConfigValidationResult) {
	protocolKeys := []string{"bind-public-protocol", "bind-private-protocol"}

	for _, key := range protocolKeys {
		val, ok := config[key]
		if !ok {
			continue
		}

		strVal, ok := val.(string)
		if !ok {
			result.Errors = append(result.Errors,
				fmt.Sprintf("'%s' must be a string, got %T", key, val))
			result.Valid = false

			continue
		}

		if strVal != mandatoryProtocol {
			result.Errors = append(result.Errors,
				fmt.Sprintf("'%s' must be %q, got %q", key, mandatoryProtocol, strVal))
			result.Valid = false
		}
	}
}

// validateAdminBindPolicy enforces that bind-private-address is always 127.0.0.1.
func validateAdminBindPolicy(config map[string]any, result *ConfigValidationResult) {
	val, ok := config["bind-private-address"]
	if !ok {
		return
	}

	strVal, ok := val.(string)
	if !ok {
		return // Type error already caught by validateBindAddresses.
	}

	if strVal != mandatoryAdminBindAddress {
		result.Errors = append(result.Errors,
			fmt.Sprintf("POLICY VIOLATION: 'bind-private-address' MUST be %q, got %q (admin must never be exposed)",
				mandatoryAdminBindAddress, strVal))
		result.Valid = false
	}
}

// validateConfigSecretRefs checks that database URLs use Docker secret file references.
func validateConfigSecretRefs(config map[string]any, result *ConfigValidationResult) {
	val, ok := config["database-url"]
	if !ok {
		return
	}

	strVal, ok := val.(string)
	if !ok {
		return
	}

	if strVal == "" {
		return
	}

	// Safe patterns: file:///run/secrets/ or sqlite://.
	if strings.HasPrefix(strVal, "file:///run/secrets/") ||
		strings.HasPrefix(strVal, "sqlite://") ||
		strings.HasPrefix(strVal, ":memory:") {
		return
	}

	// Inline credentials are not acceptable.
	if strings.Contains(strVal, "postgres://") || strings.Contains(strVal, "postgresql://") {
		result.Errors = append(result.Errors,
			"'database-url' contains inline database credentials; use 'file:///run/secrets/' instead")
		result.Valid = false

		return
	}

	result.Warnings = append(result.Warnings,
		fmt.Sprintf("'database-url' has unexpected format: %q", strVal))
}

// validateOTLPConfig checks OTLP telemetry settings for consistency.
func validateOTLPConfig(config map[string]any, result *ConfigValidationResult) {
	otlpEnabled, ok := config["otlp"]
	if !ok {
		return
	}

	enabled, ok := otlpEnabled.(bool)
	if !ok {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("'otlp' should be a boolean, got %T", otlpEnabled))

		return
	}

	if !enabled {
		return
	}

	requiredWhenEnabled := []string{"otlp-service", "otlp-endpoint"}

	for _, key := range requiredWhenEnabled {
		if _, exists := config[key]; !exists {
			result.Errors = append(result.Errors,
				fmt.Sprintf("'%s' is required when 'otlp' is true", key))
			result.Valid = false
		}
	}
}

// toInt converts a YAML value to an integer.
// YAML parsers may return int, int64, float64, or string values.
// String values (e.g., from extractHostPort) are converted via strconv.Atoi.
func toInt(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}

		return n, true
	default:
		return 0, false
	}
}

// FormatConfigValidationResult formats a ConfigValidationResult for display.
func FormatConfigValidationResult(result *ConfigValidationResult) string {
	var sb strings.Builder

	status := statusPass
	if !result.Valid {
		status = statusFail
	}

	sb.WriteString(fmt.Sprintf("[%s] %s\n", status, result.Path))

	for _, e := range result.Errors {
		sb.WriteString(fmt.Sprintf("  ERROR: %s\n", e))
	}

	for _, w := range result.Warnings {
		sb.WriteString(fmt.Sprintf("  WARNING: %s\n", w))
	}

	return sb.String()
}
