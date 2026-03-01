package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TelemetryValidationResult holds results from ValidateTelemetry.
type TelemetryValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// otlpConfigEntry represents OTLP settings extracted from a single config file.
type otlpConfigEntry struct {
	FilePath    string
	Enabled     bool
	Service     string
	Endpoint    string
	Environment string
}

// ValidateTelemetry validates OTLP telemetry configuration across config files
// in a directory. It checks individual field formats and cross-file consistency.
func ValidateTelemetry(configDir string) (*TelemetryValidationResult, error) {
	result := &TelemetryValidationResult{Valid: true}

	info, err := os.Stat(configDir)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateTelemetry] Config directory not found: %s", configDir))
		result.Valid = false

		return result, nil //nolint:nilerr // Error aggregation pattern: validation errors collected in result.Errors, nil Go error allows validator pipeline to continue.
	}

	if !info.IsDir() {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateTelemetry] Path is not a directory: %s", configDir))
		result.Valid = false

		return result, nil
	}

	entries := collectOTLPEntries(configDir, result)
	validateOTLPEndpoints(entries, result)
	validateOTLPServiceNames(entries, result)
	validateOTLPConsistency(entries, result)

	return result, nil
}

// collectOTLPEntries reads all config files in configDir and extracts OTLP settings.
func collectOTLPEntries(configDir string, result *TelemetryValidationResult) []otlpConfigEntry {
	dirEntries, err := os.ReadDir(configDir)
	if err != nil {
		return nil
	}

	var entries []otlpConfigEntry

	for _, entry := range dirEntries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}

		configPath := filepath.Join(configDir, entry.Name())

		parsed := parseOTLPConfig(configPath)
		if parsed == nil {
			continue
		}

		entries = append(entries, *parsed)
	}

	return entries
}

// parseOTLPConfig reads a config file and extracts OTLP fields.
func parseOTLPConfig(configPath string) *otlpConfigEntry {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil
	}

	otlpVal, ok := config["otlp"]
	if !ok {
		return nil
	}

	enabled, ok := otlpVal.(bool)
	if !ok {
		return nil
	}

	entry := &otlpConfigEntry{
		FilePath: configPath,
		Enabled:  enabled,
	}

	if svc, ok := config["otlp-service"]; ok {
		if s, ok := svc.(string); ok {
			entry.Service = s
		}
	}

	if ep, ok := config["otlp-endpoint"]; ok {
		if s, ok := ep.(string); ok {
			entry.Endpoint = s
		}
	}

	if env, ok := config["otlp-environment"]; ok {
		if s, ok := env.(string); ok {
			entry.Environment = s
		}
	}

	return entry
}

// validateOTLPEndpoints checks that each OTLP endpoint has a valid URL format.
func validateOTLPEndpoints(entries []otlpConfigEntry, result *TelemetryValidationResult) {
	for _, entry := range entries {
		if !entry.Enabled {
			continue
		}

		if entry.Endpoint == "" {
			result.Errors = append(result.Errors,
				fmt.Sprintf("[ValidateTelemetry] '%s': otlp-endpoint is empty but otlp is enabled",
					filepath.Base(entry.FilePath)))
			result.Valid = false

			continue
		}

		validateEndpointFormat(entry, result)
	}
}

// validateEndpointFormat checks that an endpoint string is a valid URL with host and port.
func validateEndpointFormat(entry otlpConfigEntry, result *TelemetryValidationResult) {
	parsed, err := url.Parse(entry.Endpoint)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateTelemetry] '%s': invalid otlp-endpoint URL: %s",
				filepath.Base(entry.FilePath), entry.Endpoint))
		result.Valid = false

		return
	}

	host := parsed.Hostname()
	port := parsed.Port()

	if host == "" {
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidateTelemetry] '%s': otlp-endpoint missing host: %s",
				filepath.Base(entry.FilePath), entry.Endpoint))
		result.Valid = false
	}

	if port == "" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("[ValidateTelemetry] '%s': otlp-endpoint missing port (using default): %s",
				filepath.Base(entry.FilePath), entry.Endpoint))
	}

	if parsed.Scheme != "" && parsed.Scheme != cryptoutilSharedMagic.ProtocolHTTP && parsed.Scheme != cryptoutilSharedMagic.ProtocolHTTPS && parsed.Scheme != "grpc" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("[ValidateTelemetry] '%s': otlp-endpoint unusual scheme '%s': %s",
				filepath.Base(entry.FilePath), parsed.Scheme, entry.Endpoint))
	}
}

// validateOTLPServiceNames checks for duplicate service names across configs.
func validateOTLPServiceNames(entries []otlpConfigEntry, result *TelemetryValidationResult) {
	seen := make(map[string]string) // service -> first file

	for _, entry := range entries {
		if !entry.Enabled || entry.Service == "" {
			continue
		}

		if firstFile, exists := seen[entry.Service]; exists {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("[ValidateTelemetry] Duplicate otlp-service '%s' in '%s' and '%s'",
					entry.Service, filepath.Base(firstFile), filepath.Base(entry.FilePath)))
		} else {
			seen[entry.Service] = entry.FilePath
		}
	}
}

// validateOTLPConsistency checks that all enabled configs use the same collector endpoint.
func validateOTLPConsistency(entries []otlpConfigEntry, result *TelemetryValidationResult) {
	endpoints := make(map[string][]string) // endpoint -> list of files

	for _, entry := range entries {
		if !entry.Enabled || entry.Endpoint == "" {
			continue
		}

		normalized := normalizeEndpoint(entry.Endpoint)
		endpoints[normalized] = append(endpoints[normalized], filepath.Base(entry.FilePath))
	}

	if len(endpoints) > 1 {
		var details []string

		for ep, files := range endpoints {
			details = append(details, fmt.Sprintf("  %s: %s", ep, strings.Join(files, ", ")))
		}

		result.Warnings = append(result.Warnings,
			fmt.Sprintf("[ValidateTelemetry] Inconsistent otlp-endpoints across configs:\n%s",
				strings.Join(details, "\n")))
	}
}

// normalizeEndpoint strips trailing slashes for comparison.
func normalizeEndpoint(endpoint string) string {
	return strings.TrimRight(endpoint, "/")
}

// FormatTelemetryValidationResult formats the result for human-readable output.
func FormatTelemetryValidationResult(result *TelemetryValidationResult) string {
	if result == nil {
		return "No telemetry validation result"
	}

	var sb strings.Builder

	if result.Valid {
		sb.WriteString("Telemetry validation: PASSED\n")
	} else {
		sb.WriteString("Telemetry validation: FAILED\n")
	}

	for _, e := range result.Errors {
		sb.WriteString(fmt.Sprintf("  ERROR: %s\n", e))
	}

	for _, w := range result.Warnings {
		sb.WriteString(fmt.Sprintf("  WARNING: %s\n", w))
	}

	return sb.String()
}
