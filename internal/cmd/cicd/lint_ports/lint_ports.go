// Copyright (c) 2025 Justin Cranford

// Package lint_ports validates port assignments across cryptoutil codebase.
// Ensures legacy ports are not used and ports match the standardized scheme.
package lint_ports

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

// Violation represents a port configuration violation.
type Violation struct {
	File    string
	Line    int
	Content string
	Port    uint16
	Reason  string
}

// HealthViolation represents a health path configuration violation.
type HealthViolation struct {
	File    string
	Line    int
	Content string
	Reason  string
}

// legacyPortPattern matches port numbers in various contexts.
var legacyPortPattern = regexp.MustCompile(`\b(\d{4,5})\b`)

// healthPathPattern matches health check paths in Dockerfiles and compose files.
var healthPathPattern = regexp.MustCompile(`(/[a-zA-Z0-9/_-]+)`)

// Lint checks all relevant files for legacy port usage violations.
// Returns an error if any legacy ports are found.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running port validation lint...")

	var allErrors []error

	// Check for legacy ports.
	if err := lintLegacyPorts(logger, filesByExtension); err != nil {
		allErrors = append(allErrors, err)
	}

	// Check for host port range violations in compose files.
	if err := lintHostPortRanges(logger, filesByExtension); err != nil {
		allErrors = append(allErrors, err)
	}

	// Check for health path violations.
	if err := lintHealthPaths(logger, filesByExtension); err != nil {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		// Return the first error to preserve specific error messages for backwards compatibility.
		return allErrors[0]
	}

	logger.Log("✅ lint-ports passed: all validations successful")

	return nil
}

// checkFile checks a single file for legacy port usage.
func checkFile(filePath string, legacyPorts []uint16) []Violation {
	// Skip this package itself (port definitions are legitimate here).
	if strings.Contains(filePath, "lint_ports") {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var violations []Violation

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
		matches := legacyPortPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			portNum, err := strconv.ParseUint(match[1], 10, 16)
			if err != nil {
				continue
			}

			port := uint16(portNum)

			// Skip if it's an OTEL collector port in OTEL-related files, lines, or comments.
			// Check file path OR current line content OR previous line (comment) for OTEL-related terms.
			if IsOtelCollectorPort(port) && (isOtelRelatedFile(filePath) || isOtelRelatedContent(line) || isOtelRelatedContent(prevLine)) {
				continue
			}

			// Check if this is a legacy port.
			if legacyPortSet[port] {
				// Get service name for the legacy port.
				serviceName := getServiceForLegacyPort(port)
				violations = append(violations, Violation{
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

// isOtelRelatedFile checks if a file is related to OpenTelemetry configuration.
func isOtelRelatedFile(filePath string) bool {
	lowerPath := strings.ToLower(filePath)

	return strings.Contains(lowerPath, "otel") ||
		strings.Contains(lowerPath, "opentelemetry") ||
		strings.Contains(lowerPath, "telemetry")
}

// isOtelRelatedContent checks if a line of code contains OTEL-related terms.
// This catches cases like constant definitions with "Otel" in the name.
func isOtelRelatedContent(line string) bool {
	lowerLine := strings.ToLower(line)

	return strings.Contains(lowerLine, "otel") ||
		strings.Contains(lowerLine, "opentelemetry") ||
		strings.Contains(lowerLine, "telemetry")
}

// getServiceForLegacyPort returns the service name that used the given legacy port.
func getServiceForLegacyPort(port uint16) string {
	for _, cfg := range ServicePorts {
		for _, legacyPort := range cfg.LegacyPorts {
			if legacyPort == port {
				return cfg.Name
			}
		}
	}

	return "unknown"
}

func printViolations(violations []Violation) {
	fmt.Println()
	fmt.Println("❌ PORT VIOLATIONS: Legacy ports detected")
	fmt.Println(strings.Repeat("=", LineSeparatorLength))

	for _, v := range violations {
		fmt.Printf("\n📁 File: %s\n", v.File)
		fmt.Printf("📍 Line: %d\n", v.Line)
		fmt.Printf("🔢 Port: %d\n", v.Port)
		fmt.Printf("📝 Content: %s\n", v.Content)
		fmt.Printf("⚠️  Reason: %s\n", v.Reason)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", LineSeparatorLength))
	fmt.Println("💡 Fix: Replace legacy ports with standardized ports:")
	fmt.Println("   cipher-im: 8070-8072 (was 8888-8890)")
	fmt.Println("   jose-ja: 8060 (was 9443, 8092)")
	fmt.Println("   pki-ca: 8050 (was 8443)")
	fmt.Println()
}

// lintLegacyPorts checks for legacy port usage in all relevant files.
func lintLegacyPorts(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	legacyPorts := AllLegacyPorts()
	if len(legacyPorts) == 0 {
		logger.Log("No legacy ports defined, skipping legacy port check")

		return nil
	}

	logger.Log(fmt.Sprintf("Checking for legacy ports: %v", legacyPorts))

	var violations []Violation

	// Check Go files.
	for _, file := range filesByExtension["go"] {
		fileViolations := checkFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	// Check YAML files.
	for _, file := range filesByExtension["yml"] {
		fileViolations := checkFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	for _, file := range filesByExtension["yaml"] {
		fileViolations := checkFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	// Check Markdown files.
	for _, file := range filesByExtension["md"] {
		fileViolations := checkFile(file, legacyPorts)
		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		printViolations(violations)

		return fmt.Errorf("lint-ports failed: %d legacy port violations found", len(violations))
	}

	logger.Log("  ✅ No legacy port violations")

	return nil
}
