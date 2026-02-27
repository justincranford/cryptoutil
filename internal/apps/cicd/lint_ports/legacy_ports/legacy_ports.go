// Copyright (c) 2025 Justin Cranford

// Package legacy_ports provides linting for legacy port usage in cryptoutil services.
package legacy_ports

import (
	"bufio"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintPortsCommon "cryptoutil/internal/apps/cicd/lint_ports/common"
)

// legacyPortPattern matches port numbers in various contexts.
var legacyPortPattern = regexp.MustCompile(`\b(\d{4,5})\b`)

// Injectable functions for testing defensive error paths.
var (
	legacyPortsAllFn     = lintPortsCommon.AllLegacyPorts
	legacyPortsFindAllFn = legacyPortPattern.FindAllStringSubmatch
)

// Check checks for legacy port usage in all relevant files.
// Returns an error if any legacy ports are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	legacyPorts := legacyPortsAllFn()
	if len(legacyPorts) == 0 {
		logger.Log("No legacy ports defined, skipping legacy port check")

		return nil
	}

	logger.Log(fmt.Sprintf("Checking for legacy ports: %v", legacyPorts))

	var violations []lintPortsCommon.Violation

	// Check Go files.
	for _, file := range filesByExtension["go"] {
		fileViolations := CheckFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	// Check YAML files.
	for _, file := range filesByExtension["yml"] {
		fileViolations := CheckFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	for _, file := range filesByExtension["yaml"] {
		fileViolations := CheckFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	// Check Markdown files.
	for _, file := range filesByExtension["md"] {
		fileViolations := CheckFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		printViolations(violations)

		return fmt.Errorf("lint-ports failed: %d legacy port violations found", len(violations))
	}

	logger.Log("  ✅ No legacy port violations")

	return nil
}

// CheckFile checks a single file for legacy port usage.
func CheckFile(filePath string, legacyPorts []uint16) []lintPortsCommon.Violation {
	// Skip this package itself (port definitions are legitimate here).
	if strings.Contains(filePath, "lint_ports") {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var violations []lintPortsCommon.Violation

	scanner := bufio.NewScanner(file)
	lineNum := 0
	prevLine := "" // Track previous line for comment context.

	legacyPortSet := make(map[uint16]bool)
	for _, p := range legacyPorts {
		legacyPortSet[p] = true
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Find all potential port numbers in the line.
		matches := legacyPortsFindAllFn(line, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			portNum, err := strconv.ParseUint(match[1], cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.RealmMinTokenLengthBytes)
			if err != nil {
				continue
			}

			port := uint16(portNum)

			// Skip if it's an OTEL collector port in OTEL-related files, lines, or comments.
			// Check file path OR current line content OR previous line (comment) for OTEL-related terms.
			if lintPortsCommon.IsOtelCollectorPort(port) && (lintPortsCommon.IsOtelRelatedFile(filePath) || lintPortsCommon.IsOtelRelatedContent(line) || lintPortsCommon.IsOtelRelatedContent(prevLine)) {
				continue
			}

			// Check if this is a legacy port.
			if legacyPortSet[port] {
				// Get service name for the legacy port.
				serviceName := GetServiceForLegacyPort(port)
				violations = append(violations, lintPortsCommon.Violation{
					File:    filePath,
					Line:    lineNum,
					Content: strings.TrimSpace(line),
					Port:    port,
					Reason:  fmt.Sprintf("Legacy port %d found (service: %s). Use standardized port instead.", port, serviceName),
				})
			}
		}

		prevLine = line // Store current line for next iteration.
	}

	return violations
}

// GetServiceForLegacyPort returns the service name that used the given legacy port.
func GetServiceForLegacyPort(port uint16) string {
	for _, cfg := range lintPortsCommon.ServicePorts {
		for _, legacyPort := range cfg.LegacyPorts {
			if legacyPort == port {
				return cfg.Name
			}
		}
	}

	return "unknown"
}

func printViolations(violations []lintPortsCommon.Violation) {
	fmt.Println()
	fmt.Println("❌ PORT VIOLATIONS: Legacy ports detected")
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
	fmt.Println("💡 Fix: Replace legacy ports with standardized ports:")
	fmt.Println("   sm-kms: 8000-8002 (was 8080-8082)")
	fmt.Println("   pki-ca: 8100 (was 8050, 8443)")
	fmt.Println("   identity-authz: 8200 (was 8100, 18000)")
	fmt.Println("   identity-idp: 8300-8301 (was 8110-8112, 18100)")
	fmt.Println("   identity-rs: 8400 (was 8120-8122, 18200)")
	fmt.Println("   identity-rp: 8500 (was 8130-8132, 18300)")
	fmt.Println("   identity-spa: 8600 (was 8140-8142, 18400)")
	fmt.Println("   sm-im: 8700-8702 (was 8070-8072, 8888-8890)")
	fmt.Println("   jose-ja: 8800 (was 8060, 9443, 8092)")
	fmt.Println()
}
