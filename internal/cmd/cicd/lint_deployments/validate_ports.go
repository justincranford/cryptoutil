package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// PortValidationResult holds port validation outcomes.
type PortValidationResult struct {
	// Path is the deployment directory root.
	Path string
	// Valid indicates whether all validations passed (no errors).
	Valid bool
	// Errors contains critical validation failures.
	Errors []string
	// Warnings contains non-critical issues.
	Warnings []string
}

// Port range constants per deployment level.
//
// SERVICE level:  8000-8999  (isolated service deployment).
// PRODUCT level: 18000-18999 (isolated product deployment, SERVICE + 10000).
// SUITE level:   28000-28999 (full suite deployment, SERVICE + 20000).
//
// See ARCHITECTURE.md Section 3.4 Port Assignments & Networking.
const (
	servicePortMin = 8000
	servicePortMax = 8999
	productPortMin = 18000
	productPortMax = 18999
	suitePortMin   = 28000
	suitePortMax   = 28999
)

// ValidatePorts validates that ports in compose and config files follow the
// SERVICE/PRODUCT/SUITE deployment level pattern.
//
// deploymentLevel MUST be one of: "PRODUCT-SERVICE", "PRODUCT", "SUITE".
// deploymentName is the directory name (e.g., "sm-im", "sm-im", "cryptoutil-suite").
func ValidatePorts(deploymentPath, deploymentName, deploymentLevel string) (*PortValidationResult, error) {
	result := &PortValidationResult{
		Path:  deploymentPath,
		Valid: true,
	}

	info, statErr := os.Stat(deploymentPath)
	if statErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePorts] Path does not exist: %s", deploymentPath))

		return result, nil
	}

	if !info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePorts] Path is not a directory: %s", deploymentPath))

		return result, nil
	}

	// Validate compose file ports.
	composePath := filepath.Join(deploymentPath, "compose.yml")
	validateComposePortRanges(composePath, deploymentLevel, result)

	// Validate config file ports.
	configDir := filepath.Join(deploymentPath, "config")

	if info, err := os.Stat(configDir); err == nil && info.IsDir() {
		validateConfigPortRanges(configDir, deploymentName, deploymentLevel, result)
	}

	return result, nil
}

// validateComposePortRanges checks that host ports in compose files are within
// the correct range for the deployment level.
func validateComposePortRanges(composePath, deploymentLevel string, result *PortValidationResult) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		// Compose file absence handled by other validators.
		return
	}

	var compose composeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("[ValidatePorts] Cannot parse compose file: %s", composePath))

		return
	}

	minPort, maxPort := getPortRange(deploymentLevel)

	for serviceName, svc := range compose.Services {
		for _, portMapping := range svc.Ports {
			hostPort := extractHostPort(portMapping)
			if hostPort == "" {
				continue
			}

			port, err := strconv.Atoi(hostPort)
			if err != nil {
				continue
			}

			// Skip infrastructure ports (postgres 5432, grafana 3000, otel 4317/4318/14317/14318).
			if isInfrastructurePort(port) {
				continue
			}

			if port < minPort || port > maxPort {
				result.Valid = false
				result.Errors = append(result.Errors,
					fmt.Sprintf("[ValidatePorts] Service '%s' host port %d outside %s range [%d-%d]",
						serviceName, port, deploymentLevel, minPort, maxPort))
			}
		}
	}
}

// validateConfigPortRanges checks bind-public-port in config files against
// expected deployment level ranges.
func validateConfigPortRanges(configDir, deploymentName, deploymentLevel string, result *PortValidationResult) {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}

		configPath := filepath.Join(configDir, entry.Name())

		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		var config map[string]any
		if err := yaml.Unmarshal(data, &config); err != nil {
			continue
		}

		validateConfigPortValue(config, configPath, deploymentName, deploymentLevel, result)
	}
}

// validateConfigPortValue checks a single config file's bind-public-port value.
func validateConfigPortValue(config map[string]any, configPath, deploymentName, deploymentLevel string, result *PortValidationResult) {
	val, ok := config["bind-public-port"]
	if !ok {
		return
	}

	port, ok := toInt(val)
	if !ok {
		return
	}

	minPort, maxPort := getPortRange(deploymentLevel)

	if port < minPort || port > maxPort {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("[ValidatePorts] Config '%s' bind-public-port %d outside %s range [%d-%d]",
				filepath.Base(configPath), port, deploymentLevel, minPort, maxPort))
	}
}

// getPortRange returns the min and max port for a deployment level.
func getPortRange(deploymentLevel string) (int, int) {
	switch deploymentLevel {
	case DeploymentTypeProduct:
		return productPortMin, productPortMax
	case DeploymentTypeSuite:
		return suitePortMin, suitePortMax
	default:
		// PRODUCT-SERVICE and any unknown level default to SERVICE range.
		return servicePortMin, servicePortMax
	}
}

// isInfrastructurePort returns true for well-known infrastructure ports that
// should be excluded from deployment level range checks.
func isInfrastructurePort(port int) bool {
	infraPorts := map[int]bool{
		cryptoutilSharedMagic.JoseJAE2EGrafanaPort:  true, // Grafana UI.
		cryptoutilSharedMagic.JoseJAE2EOtelCollectorGRPCPort:  true, // OTLP gRPC (collector).
		cryptoutilSharedMagic.JoseJAE2EOtelCollectorHTTPPort:  true, // OTLP HTTP (collector).
		int(cryptoutilSharedMagic.DefaultPublicPortPostgres):  true, // PostgreSQL.
		int(cryptoutilSharedMagic.DefaultPublicPortOtelCollectorHealth): true, // OTel collector health.
		int(cryptoutilSharedMagic.PortGrafanaOTLPGRPC): true, // OTLP gRPC (forwarded).
		int(cryptoutilSharedMagic.PortGrafanaOTLPHTTP): true, // OTLP HTTP (forwarded).
	}

	return infraPorts[port]
}

// FormatPortValidationResult formats a PortValidationResult for display.
func FormatPortValidationResult(result *PortValidationResult) string {
	var sb strings.Builder

	if result.Valid {
		sb.WriteString(fmt.Sprintf("  PASS: Port validation: %s\n", result.Path))
	} else {
		sb.WriteString(fmt.Sprintf("  FAIL: Port validation: %s\n", result.Path))
	}

	for _, e := range result.Errors {
		sb.WriteString(fmt.Sprintf("    ERROR: %s\n", e))
	}

	for _, w := range result.Warnings {
		sb.WriteString(fmt.Sprintf("    WARN: %s\n", w))
	}

	return sb.String()
}
