// Copyright (c) 2025 Justin Cranford

// Package lint_ports validates port assignments across cryptoutil codebase.
// This file contains health path validation functions.
package lint_ports

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// lintHealthPaths checks for health path configuration violations.
// Validates that health checks use the standard path /admin/api/v1/livez on port 9090.
func lintHealthPaths(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Checking health path configurations...")

	var violations []HealthViolation

	// Check Dockerfiles.
	for _, filePath := range filesByExtension["dockerfile"] {
		// Skip third-party services (Grafana, OTEL collector).
		if isOtelRelatedFile(filePath) {
			continue
		}

		fileViolations := checkHealthPathsInDockerfile(filePath)
		violations = append(violations, fileViolations...)
	}

	// Check compose files.
	allYamlFiles := append(filesByExtension["yml"], filesByExtension["yaml"]...)

	for _, filePath := range allYamlFiles {
		if !isComposeFile(filePath) {
			continue
		}

		// Skip third-party services (Grafana, OTEL collector).
		if isOtelRelatedFile(filePath) {
			continue
		}

		fileViolations := checkHealthPathsInCompose(filePath)
		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		printHealthViolations(violations)

		return fmt.Errorf("lint-ports: %d health path violations found", len(violations))
	}

	logger.Log("  ✅ No health path violations")

	return nil
}

// checkHealthPathsInDockerfile checks a Dockerfile for health path violations.
func checkHealthPathsInDockerfile(filePath string) []HealthViolation {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var violations []HealthViolation

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Pattern to match HEALTHCHECK commands.
	healthcheckPattern := regexp.MustCompile(`(?i)HEALTHCHECK`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for HEALTHCHECK directive.
		if healthcheckPattern.MatchString(line) {
			// Check if it contains a health path.
			if paths := healthPathPattern.FindAllString(line, -1); len(paths) > 0 {
				for _, path := range paths {
					// Skip if it's the correct path.
					if path == StandardHealthPath {
						continue
					}

					// Check if it looks like a health endpoint path.
					if isLikelyHealthPath(path) {
						violations = append(violations, HealthViolation{
							File:    filePath,
							Line:    lineNum,
							Content: strings.TrimSpace(line),
							Reason:  fmt.Sprintf("Non-standard health path '%s'. Use '%s' instead", path, StandardHealthPath),
						})
					}
				}
			}

			// Check for incorrect port.
			if match := legacyPortPattern.FindStringSubmatch(line); match != nil {
				port, err := strconv.ParseUint(match[1], 10, 16)
				if err == nil && uint16(port) != StandardAdminPort {
					// Only flag if it looks like it's being used for a health check.
					if strings.Contains(strings.ToLower(line), "health") || strings.Contains(line, "livez") || strings.Contains(line, "readyz") {
						violations = append(violations, HealthViolation{
							File:    filePath,
							Line:    lineNum,
							Content: strings.TrimSpace(line),
							Reason:  fmt.Sprintf("Health check should use admin port %d, found %d", StandardAdminPort, port),
						})
					}
				}
			}
		}
	}

	return violations
}

// isLikelyHealthPath checks if a path looks like a health endpoint.
func isLikelyHealthPath(path string) bool {
	lowerPath := strings.ToLower(path)

	return strings.Contains(lowerPath, "health") ||
		strings.Contains(lowerPath, "livez") ||
		strings.Contains(lowerPath, "readyz") ||
		strings.Contains(lowerPath, "alive") ||
		strings.Contains(lowerPath, "ready")
}

// checkHealthPathsInCompose checks a compose file for health path violations.
func checkHealthPathsInCompose(filePath string) []HealthViolation {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var violations []HealthViolation

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inHealthcheck := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Detect healthcheck section.
		if strings.Contains(line, "healthcheck:") {
			inHealthcheck = true

			continue
		}

		// Exit healthcheck section on non-indented line (not starting with whitespace).
		if inHealthcheck && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			inHealthcheck = false
		}

		// Check health paths in healthcheck section or lines with curl/wget commands.
		if inHealthcheck || strings.Contains(line, "curl") || strings.Contains(line, "wget") {
			// Look for health-related paths.
			if paths := healthPathPattern.FindAllString(line, -1); len(paths) > 0 {
				for _, path := range paths {
					// Skip correct path.
					if path == StandardHealthPath {
						continue
					}

					// Check if it looks like a health path.
					if isLikelyHealthPath(path) {
						violations = append(violations, HealthViolation{
							File:    filePath,
							Line:    lineNum,
							Content: strings.TrimSpace(line),
							Reason:  fmt.Sprintf("Non-standard health path '%s'. Use '%s' instead", path, StandardHealthPath),
						})
					}
				}
			}
		}
	}

	return violations
}

// printHealthViolations outputs health path violations.
func printHealthViolations(violations []HealthViolation) {
	fmt.Println()
	fmt.Println("❌ HEALTH PATH VIOLATIONS: Non-standard health configurations")
	fmt.Println(strings.Repeat("=", LineSeparatorLength))

	for _, v := range violations {
		fmt.Printf("\n📁 File: %s\n", v.File)
		fmt.Printf("📍 Line: %d\n", v.Line)
		fmt.Printf("📝 Content: %s\n", v.Content)
		fmt.Printf("⚠️  Reason: %s\n", v.Reason)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", LineSeparatorLength))
	fmt.Printf("💡 Fix: Use standard health path '%s' on port %d\n", StandardHealthPath, StandardAdminPort)
	fmt.Println()
}
