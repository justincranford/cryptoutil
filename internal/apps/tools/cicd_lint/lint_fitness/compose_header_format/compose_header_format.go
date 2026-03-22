// Copyright (c) 2025 Justin Cranford

// Package compose_header_format validates that every product-service compose.yml
// has the correct comment header:
//   - Line 3 (1-indexed): "# {PS-ID-UPPER} Docker Compose Configuration"
//   - Line 5 (1-indexed): "# SERVICE-level deployment for {Display Name}."
//
// This ensures the compose file header accurately describes the service and
// prevents drift when services are renamed or restructured.
package compose_header_format

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Check validates compose file header format from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates compose file header format under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking compose file header format...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkComposeHeader(rootDir, ps)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("compose header format violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("compose-header-format: all 10 product-services have correct compose headers")

	return nil
}

// checkComposeHeader reads the first CICDComposeHeaderLinesToCheck lines of the compose.yml and verifies
// line 3 and line 5 match expected format.
func checkComposeHeader(rootDir string, ps lintFitnessRegistry.ProductService) []string {
	var violations []string

	composePath := filepath.Join(rootDir, "deployments", ps.PSID, "compose.yml")

	lines, err := readFirstNLines(composePath, cryptoutilSharedMagic.CICDComposeHeaderLinesToCheck)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read deployments/%s/compose.yml: %v", ps.PSID, ps.PSID, err)}
	}

	if len(lines) < cryptoutilSharedMagic.CICDComposeHeaderLinesToCheck {
		return []string{fmt.Sprintf("%s: deployments/%s/compose.yml has fewer than 5 lines", ps.PSID, ps.PSID)}
	}

	// Line 3 (index 2): "# {PS-ID-UPPER} Docker Compose Configuration"
	expectedLine3 := "# " + strings.ToUpper(ps.PSID) + " Docker Compose Configuration"
	if lines[cryptoutilSharedMagic.CICDComposeLine3Index] != expectedLine3 {
		violations = append(violations, fmt.Sprintf(
			"%s: deployments/%s/compose.yml line 3: got %q, want %q",
			ps.PSID, ps.PSID, lines[cryptoutilSharedMagic.CICDComposeLine3Index], expectedLine3,
		))
	}

	// Line 5 (index 4): "# SERVICE-level deployment for {Display Name}."
	expectedLine5 := "# SERVICE-level deployment for " + ps.DisplayName + "."
	if lines[cryptoutilSharedMagic.CICDComposeLine5Index] != expectedLine5 {
		violations = append(violations, fmt.Sprintf(
			"%s: deployments/%s/compose.yml line 5: got %q, want %q",
			ps.PSID, ps.PSID, lines[cryptoutilSharedMagic.CICDComposeLine5Index], expectedLine5,
		))
	}

	return violations
}

// readFirstNLines reads up to n lines from the file at path.
func readFirstNLines(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	defer func() { _ = f.Close() }()

	var lines []string

	scanner := bufio.NewScanner(f)

	for scanner.Scan() && len(lines) < n {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	return lines, nil
}
