// Copyright (c) 2025 Justin Cranford

// Package host_port_ranges provides linting for host port range assignments in Docker Compose files.
package host_port_ranges

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintPortsCommon "cryptoutil/internal/apps/cicd/lint_ports/common"
)

// lintHostPortRanges checks compose files for host port range violations.
// Validates that host ports in port mappings are within the allocated ranges for each service.
func Check(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Checking host port ranges in compose files...")

	var violations []lintPortsCommon.Violation

	// Only check compose files (YAML).
	allYamlFiles := append(filesByExtension["yml"], filesByExtension["yaml"]...)

	for _, filePath := range allYamlFiles {
		// Only check compose files.
		if !lintPortsCommon.IsComposeFile(filePath) {
			continue
		}

		fileViolations := CheckHostPortRangesInFile(filePath)
		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		printHostPortViolations(violations)

		return fmt.Errorf("lint-ports: %d host port range violations found", len(violations))
	}

	logger.Log("  ✅ No host port range violations")

	return nil
}


// checkHostPortRangesInFile checks a single compose file for host port range violations.
func CheckHostPortRangesInFile(filePath string) []lintPortsCommon.Violation {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var violations []lintPortsCommon.Violation

	scanner := bufio.NewScanner(file)
	lineNum := 0
	currentService := ""
	inServicesBlock := false

	// Pattern to match "services:" at root level.
	servicesPattern := regexp.MustCompile(`^services:\s*$`)

	// Pattern to match service definitions like "  service-name:" (2 spaces indentation under services).
	servicePattern := regexp.MustCompile(`^\s{2}([a-zA-Z][a-zA-Z0-9_-]*):\s*$`)

	// Pattern to match port mappings like "- 8080:8080" or "- "8080:8080"".
	portMappingPattern := regexp.MustCompile(`-\s*"?(\d+):(\d+)"?`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check if we're entering the services block.
		if servicesPattern.MatchString(line) {
			inServicesBlock = true

			continue
		}

		// Reset if we hit a new top-level key (no indentation).
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			inServicesBlock = false
			currentService = ""

			continue
		}

		// Check if this is a service definition (only when in services block).
		if inServicesBlock {
			if match := servicePattern.FindStringSubmatch(line); match != nil {
				currentService = match[1]

				continue
			}
		}

		// Check for port mappings.
		if match := portMappingPattern.FindStringSubmatch(line); match != nil {
			hostPort, err := strconv.ParseUint(match[1], cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.RealmMinTokenLengthBytes)
			if err != nil {
				continue
			}

			port := uint16(hostPort)

			// Validate host port is in valid range for the service.
			if currentService != "" {
				serviceCfg := GetServiceConfig(currentService)
				if serviceCfg != nil && !IsPortInValidRange(port, serviceCfg) {
					violations = append(violations, lintPortsCommon.Violation{
						File:    filePath,
						Line:    lineNum,
						Content: strings.TrimSpace(line),
						Port:    port,
						Reason:  fmt.Sprintf("Host port %d is outside valid range for service '%s'. Expected: %v or admin %d", port, currentService, serviceCfg.PublicPorts, serviceCfg.AdminPort),
					})
				}
			}
		}
	}

	return violations
}

// getServiceConfig returns the port configuration for a service based on its name.
func GetServiceConfig(serviceName string) *lintPortsCommon.ServicePortConfig {
	// Try exact match first.
	if cfg, ok := lintPortsCommon.ServicePorts[serviceName]; ok {
		return &cfg
	}

	// Try partial match (e.g., "sm-im-postgres" should match "sm-im").
	for key, cfg := range lintPortsCommon.ServicePorts {
		if strings.HasPrefix(serviceName, key) {
			return &cfg
		}
	}

	return nil
}

// isPortInValidRange checks if a port is within the valid range for a service.
func IsPortInValidRange(port uint16, cfg *lintPortsCommon.ServicePortConfig) bool {
	// Check if it's the admin port.
	if port == cfg.AdminPort {
		return true
	}

	// Check if it's one of the public ports.
	for _, p := range cfg.PublicPorts {
		if port == p {
			return true
		}
	}

	// Check if it's in the extended range (100-port blocks per service).
	// Public ports define the base, allow up to +99 for the full range.
	if len(cfg.PublicPorts) > 0 {
		basePort := cfg.PublicPorts[0]
		if port >= basePort && port < basePort+cryptoutilSharedMagic.JoseJAMaxMaterials {
			return true
		}
	}

	return false
}

// printHostPortViolations outputs host port range violations.
func printHostPortViolations(violations []lintPortsCommon.Violation) {
	fmt.Println()
	fmt.Println("❌ HOST PORT RANGE VIOLATIONS: Ports outside allocated ranges")
	fmt.Println(strings.Repeat("=", lintPortsCommon.LineSeparatorLength))

	for _, v := range violations {
		fmt.Printf("\n📁 File: %s\n", v.File)
		fmt.Printf("📍 Line: %d\n", v.Line)
		fmt.Printf("🔢 Port: %d\n", v.Port)
		fmt.Printf("📝 Content: %s\n", v.Content)
		fmt.Printf("⚠️  Reason: %s\n", v.Reason)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", lintPortsCommon.LineSeparatorLength))
	fmt.Println("💡 Fix: Use ports within the allocated range for each service")
	fmt.Println()
}
