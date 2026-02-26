// Copyright (c) 2025 Justin Cranford

// Package health_paths provides linting for health path configuration in Dockerfiles and compose files.
package health_paths

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

// healthPathPattern matches health check paths in Dockerfiles and compose files.
var healthPathPattern = regexp.MustCompile(`(/[a-zA-Z0-9/_-]+)`)

// legacyPortPattern matches port numbers in various contexts.
var legacyPortPattern = regexp.MustCompile(`\b(\d{4,5})\b`)

// lintHealthPaths checks for health path configuration violations.
// Validates that health checks use the standard path /admin/api/v1/livez on port 9090.
func Check(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Checking health path configurations...")

	var violations []lintPortsCommon.HealthViolation

	// Check Dockerfiles.
	for _, filePath := range filesByExtension["dockerfile"] {
		// Skip third-party services (Grafana, OTEL collector).
		if lintPortsCommon.IsOtelRelatedFile(filePath) {
			continue
		}

		fileViolations := CheckHealthPathsInDockerfile(filePath)
		violations = append(violations, fileViolations...)
	}

	// Check compose files.
	allYamlFiles := append(filesByExtension["yml"], filesByExtension["yaml"]...)

	for _, filePath := range allYamlFiles {
		if !lintPortsCommon.IsComposeFile(filePath) {
			continue
		}

		// Skip third-party services (Grafana, OTEL collector).
		if lintPortsCommon.IsOtelRelatedFile(filePath) {
			continue
		}

		fileViolations := CheckHealthPathsInCompose(filePath)
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
func CheckHealthPathsInDockerfile(filePath string) []lintPortsCommon.HealthViolation {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var violations []lintPortsCommon.HealthViolation

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
					if path == lintPortsCommon.StandardHealthPath {
						continue
					}

					// Check if it looks like a health endpoint path.
					if IsLikelyHealthPath(path) {
						violations = append(violations, lintPortsCommon.HealthViolation{
							File:    filePath,
							Line:    lineNum,
							Content: strings.TrimSpace(line),
							Reason:  fmt.Sprintf("Non-standard health path '%s'. Use '%s' instead", path, lintPortsCommon.StandardHealthPath),
						})
					}
				}
			}

			// Check for incorrect port.
			if match := legacyPortPattern.FindStringSubmatch(line); match != nil {
				port, err := strconv.ParseUint(match[1], cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.RealmMinTokenLengthBytes)
				if err == nil && uint16(port) != lintPortsCommon.StandardAdminPort {
					// Only flag if it looks like it's being used for a health check.
					if strings.Contains(strings.ToLower(line), "health") || strings.Contains(line, "livez") || strings.Contains(line, "readyz") {
						violations = append(violations, lintPortsCommon.HealthViolation{
							File:    filePath,
							Line:    lineNum,
							Content: strings.TrimSpace(line),
							Reason:  fmt.Sprintf("Health check should use admin port %d, found %d", lintPortsCommon.StandardAdminPort, port),
						})
					}
				}
			}
		}
	}

	return violations
}

// isLikelyHealthPath checks if a path looks like a health endpoint.
func IsLikelyHealthPath(path string) bool {
	lowerPath := strings.ToLower(path)

	return strings.Contains(lowerPath, "health") ||
		strings.Contains(lowerPath, "livez") ||
		strings.Contains(lowerPath, "readyz") ||
		strings.Contains(lowerPath, "alive") ||
		strings.Contains(lowerPath, "ready")
}

// checkHealthPathsInCompose checks a compose file for health path violations.
func CheckHealthPathsInCompose(filePath string) []lintPortsCommon.HealthViolation {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() { _ = file.Close() }()

	var violations []lintPortsCommon.HealthViolation

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
					if path == lintPortsCommon.StandardHealthPath {
						continue
					}

					// Check if it looks like a health path.
					if IsLikelyHealthPath(path) {
						violations = append(violations, lintPortsCommon.HealthViolation{
							File:    filePath,
							Line:    lineNum,
							Content: strings.TrimSpace(line),
							Reason:  fmt.Sprintf("Non-standard health path '%s'. Use '%s' instead", path, lintPortsCommon.StandardHealthPath),
						})
					}
				}
			}
		}
	}

	return violations
}

// printHealthViolations outputs health path violations.
func printHealthViolations(violations []lintPortsCommon.HealthViolation) {
	fmt.Println()
	fmt.Println("❌ HEALTH PATH VIOLATIONS: Non-standard health configurations")
	fmt.Println(strings.Repeat("=", lintPortsCommon.LineSeparatorLength))

	for _, v := range violations {
		fmt.Printf("\n📁 File: %s\n", v.File)
		fmt.Printf("📍 Line: %d\n", v.Line)
		fmt.Printf("📝 Content: %s\n", v.Content)
		fmt.Printf("⚠️  Reason: %s\n", v.Reason)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", lintPortsCommon.LineSeparatorLength))
	fmt.Printf("💡 Fix: Use standard health path '%s' on port %d\n", lintPortsCommon.StandardHealthPath, lintPortsCommon.StandardAdminPort)
	fmt.Println()
}
