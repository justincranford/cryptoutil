// Copyright (c) 2025 Justin Cranford

// Package dockerfile_healthcheck validates that all Dockerfiles in deployments/
// use the built-in PS-ID CLI livez subcommand for HEALTHCHECK instead of
// third-party tools like wget or curl.
//
// The canonical HEALTHCHECK pattern is:
//
//	HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
//	    CMD /app/<ps-id> livez || exit 1
//
// This eliminates external tool dependencies each container image and leverages
// the framework's built-in health check client which handles self-signed TLS
// certificates via InsecureSkipVerify when no CA cert is provided.
package dockerfile_healthcheck

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilCmdCicdRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// bannedHealthcheckTools lists tools that MUST NOT appear in HEALTHCHECK commands.
var bannedHealthcheckTools = []string{"wget", "curl"}

// Check validates Dockerfile healthchecks from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates Dockerfile healthchecks under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking Dockerfile HEALTHCHECK instructions...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check Dockerfile healthchecks: %w", err)
	}

	if len(violations) > 0 {
		return fmt.Errorf("dockerfile healthcheck violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("dockerfile-healthcheck: all Dockerfiles use built-in PS-ID livez subcommand")

	return nil
}

// FindViolationsInDir scans deployments/ for Dockerfile HEALTHCHECK violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
	deploymentsDir := filepath.Join(rootDir, "deployments")

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployments/ directory: %w", err)
	}

	psidBinaryMap := buildPSIDBinaryMap()

	var violations []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		deploymentName := entry.Name()
		dockerfilePath := filepath.Join(deploymentsDir, deploymentName, "Dockerfile")

		if _, statErr := os.Stat(dockerfilePath); os.IsNotExist(statErr) {
			continue
		}

		v := validateDockerfileHealthcheck(dockerfilePath, deploymentName, psidBinaryMap)
		violations = append(violations, v...)
	}

	return violations, nil
}

// buildPSIDBinaryMap returns a map of deployment-name → expected binary name.
// For PS-ID services, the binary is /app/<ps-id>.
// Suite deployments map to /app/<suite-id>.
func buildPSIDBinaryMap() map[string]string {
	pss := cryptoutilCmdCicdRegistry.AllProductServices()
	suites := cryptoutilCmdCicdRegistry.AllSuites()
	binaryMap := make(map[string]string, len(pss)+len(suites))

	for _, ps := range pss {
		binaryMap[ps.PSID] = fmt.Sprintf("/app/%s", ps.PSID)
	}

	// Suite-level deployments use the suite binary.
	for _, s := range suites {
		binaryMap[s.ID] = fmt.Sprintf("/app/%s", s.ID)
	}

	return binaryMap
}

// validateDockerfileHealthcheck checks a single Dockerfile for HEALTHCHECK violations.
func validateDockerfileHealthcheck(dockerfilePath, deploymentName string, psidBinaryMap map[string]string) []string {
	healthcheckCMD, err := extractHealthcheckCMD(dockerfilePath)
	if err != nil {
		return []string{fmt.Sprintf("%s: failed to read Dockerfile: %v", deploymentName, err)}
	}

	// Dockerfiles without HEALTHCHECK are not PS-ID services (e.g. shared-postgres).
	if healthcheckCMD == "" {
		return nil
	}

	var violations []string

	// Check for banned tools.
	for _, tool := range bannedHealthcheckTools {
		if containsTool(healthcheckCMD, tool) {
			violations = append(violations, fmt.Sprintf(
				"%s: HEALTHCHECK uses banned tool %q; use built-in PS-ID livez subcommand instead",
				deploymentName, tool,
			))
		}
	}

	// Check for expected PS-ID livez pattern.
	expectedBinary, isPSIDOrSuite := psidBinaryMap[deploymentName]
	if isPSIDOrSuite {
		expectedCMD := fmt.Sprintf("%s livez || exit 1", expectedBinary)
		if !strings.Contains(healthcheckCMD, expectedCMD) {
			violations = append(violations, fmt.Sprintf(
				"%s: HEALTHCHECK CMD should be %q, got %q",
				deploymentName, expectedCMD, healthcheckCMD,
			))
		}
	}

	return violations
}

// containsTool checks if cmd contains the given tool name as a standalone command
// (not as a substring of another word).
func containsTool(cmd, tool string) bool {
	// Check for tool appearing as a complete word.
	lower := strings.ToLower(cmd)
	toolLower := strings.ToLower(tool)

	idx := strings.Index(lower, toolLower)
	if idx < 0 {
		return false
	}

	// Check left boundary (start of string or non-alphanumeric).
	if idx > 0 {
		prev := lower[idx-1]
		if isAlphaNumeric(prev) {
			return false
		}
	}

	// Check right boundary (end of string or non-alphanumeric).
	end := idx + len(toolLower)
	if end < len(lower) {
		next := lower[end]
		if isAlphaNumeric(next) {
			return false
		}
	}

	return true
}

// isAlphaNumeric returns true if the byte is a letter or digit.
func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// extractHealthcheckCMD parses a Dockerfile and returns the HEALTHCHECK CMD content.
// Returns empty string if no HEALTHCHECK instruction is found.
// Handles multi-line HEALTHCHECK instructions with line continuations (\).
func extractHealthcheckCMD(dockerfilePath string) (string, error) {
	file, err := os.Open(dockerfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open Dockerfile: %w", err)
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	var inHealthcheck bool

	var healthcheckLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if inHealthcheck {
			healthcheckLines = append(healthcheckLines, trimmed)

			if !strings.HasSuffix(trimmed, "\\") {
				inHealthcheck = false
			}

			continue
		}

		if strings.HasPrefix(trimmed, "HEALTHCHECK ") {
			inHealthcheck = strings.HasSuffix(trimmed, "\\")
			healthcheckLines = append(healthcheckLines, trimmed)

			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to scan Dockerfile: %w", err)
	}

	if len(healthcheckLines) == 0 {
		return "", nil
	}

	// Join multi-line HEALTHCHECK into single string, removing continuations.
	return joinHealthcheckLines(healthcheckLines), nil
}

// joinHealthcheckLines joins HEALTHCHECK continuation lines into a single string.
func joinHealthcheckLines(lines []string) string {
	var parts []string

	for _, line := range lines {
		cleaned := strings.TrimSuffix(strings.TrimSpace(line), "\\")
		cleaned = strings.TrimSpace(cleaned)

		if cleaned != "" {
			parts = append(parts, cleaned)
		}
	}

	return strings.Join(parts, " ")
}
