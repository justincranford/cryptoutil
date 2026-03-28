// Copyright (c) 2025 Justin Cranford

// Package dockerfile_labels validates that all Dockerfiles in deployments/
// contain correct OCI labels:
//
//   - org.opencontainers.image.title must contain the deployment name (PS-ID, product, or suite ID)
//   - org.opencontainers.image.description must be present (non-empty)
//
// The linter only checks Dockerfiles that have at least one OCI label line.
// Dockerfiles without any labels are reported as violations (labels are required).
package dockerfile_labels

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// requiredLabels lists the OCI label keys that must be present in every Dockerfile.
var requiredLabels = []string{
	"org.opencontainers.image.title",
	"org.opencontainers.image.description",
}

// Check validates Dockerfile labels from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates Dockerfile labels under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking Dockerfile OCI labels...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check Dockerfile labels: %w", err)
	}

	if len(violations) > 0 {
		return fmt.Errorf("dockerfile label violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("dockerfile-labels: all Dockerfiles have correct OCI labels")

	return nil
}

// FindViolationsInDir scans deployments/ for Dockerfile label violations.
func FindViolationsInDir(rootDir string) ([]string, error) {
	deploymentsDir := filepath.Join(rootDir, "deployments")

	entries, err := os.ReadDir(deploymentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployments/ directory: %w", err)
	}

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

		v := validateDockerfileLabels(dockerfilePath, deploymentName)
		violations = append(violations, v...)
	}

	return violations, nil
}

// validateDockerfileLabels checks a single Dockerfile for required OCI labels.
func validateDockerfileLabels(dockerfilePath, deploymentName string) []string {
	labels, err := extractLabels(dockerfilePath)
	if err != nil {
		return []string{fmt.Sprintf("%s: failed to read Dockerfile: %v", deploymentName, err)}
	}

	var violations []string

	// Check required labels are present.
	for _, required := range requiredLabels {
		if _, exists := labels[required]; !exists {
			violations = append(violations, fmt.Sprintf("%s: missing required label %q", deploymentName, required))
		}
	}

	// Validate title contains deployment name.
	if title, exists := labels["org.opencontainers.image.title"]; exists {
		if !titleContainsDeploymentName(title, deploymentName) {
			violations = append(violations, fmt.Sprintf("%s: image.title %q does not contain deployment name %q", deploymentName, title, deploymentName))
		}
	}

	return violations
}

// titleContainsDeploymentName checks if the title references the deployment name.
// The title should contain the deployment name (e.g., "cryptoutil-sm-kms" contains "sm-kms").
// Hyphens and spaces are treated as equivalent for comparison (e.g. "CryptoUtil Suite" matches "cryptoutil").
func titleContainsDeploymentName(title, deploymentName string) bool {
	normalizedTitle := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	normalizedName := strings.ToLower(strings.ReplaceAll(deploymentName, " ", "-"))

	return strings.Contains(normalizedTitle, normalizedName)
}

// extractLabels parses a Dockerfile and extracts OCI label key-value pairs.
// Handles both single-line and multi-line LABEL directives.
func extractLabels(dockerfilePath string) (map[string]string, error) {
	file, err := os.Open(dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Dockerfile: %w", err)
	}

	defer func() { _ = file.Close() }()

	labels := make(map[string]string)
	scanner := bufio.NewScanner(file)

	var continuation bool

	var currentLine string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if continuation {
			currentLine += " " + trimmed

			if strings.HasSuffix(trimmed, "\\") {
				currentLine = strings.TrimSuffix(currentLine, "\\")

				continue
			}

			continuation = false

			parseLabelsFromLine(currentLine, labels)
			currentLine = ""

			continue
		}

		if !strings.HasPrefix(trimmed, "LABEL ") {
			continue
		}

		labelContent := strings.TrimPrefix(trimmed, "LABEL ")

		if strings.HasSuffix(labelContent, "\\") {
			continuation = true
			currentLine = strings.TrimSuffix(labelContent, "\\")

			continue
		}

		parseLabelsFromLine(labelContent, labels)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan Dockerfile: %w", err)
	}

	return labels, nil
}

// parseLabelsFromLine extracts key=value pairs from a LABEL directive line.
// Handles quoted values with spaces (e.g., key="value with spaces").
func parseLabelsFromLine(line string, labels map[string]string) {
	remaining := strings.TrimSpace(line)

	for remaining != "" {
		// Find key=value start.
		eqIdx := strings.Index(remaining, "=")
		if eqIdx < 0 {
			break
		}

		key := strings.TrimSpace(remaining[:eqIdx])
		remaining = remaining[eqIdx+1:]

		var value string

		remaining = strings.TrimSpace(remaining)

		if strings.HasPrefix(remaining, "\"") {
			// Quoted value — find closing quote.
			remaining = remaining[1:]

			closeIdx := strings.Index(remaining, "\"")
			if closeIdx >= 0 {
				value = remaining[:closeIdx]
				remaining = strings.TrimSpace(remaining[closeIdx+1:])
			} else {
				value = remaining
				remaining = ""
			}
		} else {
			// Unquoted value — take until next space.
			spaceIdx := strings.Index(remaining, " ")
			if spaceIdx >= 0 {
				value = remaining[:spaceIdx]
				remaining = strings.TrimSpace(remaining[spaceIdx+1:])
			} else {
				value = remaining
				remaining = ""
			}
		}

		labels[key] = value
	}
}
