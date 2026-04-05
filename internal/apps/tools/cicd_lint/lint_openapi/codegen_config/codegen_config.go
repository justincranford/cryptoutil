// Copyright (c) 2025 Justin Cranford

// Package codegen_config validates oapi-codegen config files have the required base initialisms.
package codegen_config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// baseInitialisms lists the mandatory base initialisms per ENG-HANDBOOK.md Section 8.1.
var baseInitialisms = []string{
	"IDS",
	"JWT",
	"JWK",
	string(cryptoutilSharedMagic.SessionAlgorithmJWE),
	string(cryptoutilSharedMagic.SessionAlgorithmJWS),
	"OIDC",
	"SAML",
	"AES",
	"GCM",
	"CBC",
	cryptoutilSharedMagic.KeyTypeRSA,
	"EC",
	"HMAC",
	"SHA",
	"TLS",
	"IP",
	"AI",
	"ML",
	"KEM",
	"PEM",
	"DER",
	"DSA",
	"IKM",
}

// Check validates that all oapi-codegen config files contain the required base initialisms.
func Check(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Checking oapi-codegen config base initialisms...")

	yamlFiles := append(filesByExtension["yaml"], filesByExtension["yml"]...)
	if len(yamlFiles) == 0 {
		logger.Log("No YAML files to check")

		return nil
	}

	// Filter to only codegen config files in api/ directories.
	var configFiles []string

	for _, f := range yamlFiles {
		normalized := filepath.ToSlash(f)
		base := filepath.Base(f)

		if strings.Contains(normalized, "api/") && strings.HasPrefix(base, "openapi-gen_config") {
			configFiles = append(configFiles, f)
		}
	}

	if len(configFiles) == 0 {
		logger.Log("No oapi-codegen config files found in api/ directory")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d codegen config files to check", len(configFiles)))

	var violations []string

	for _, configFile := range configFiles {
		missing, err := findMissingInitialisms(configFile)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: %s", configFile, err.Error()))

			continue
		}

		if len(missing) > 0 {
			violations = append(violations, fmt.Sprintf("%s: missing base initialisms: %s", configFile, strings.Join(missing, ", ")))
		}
	}

	if len(violations) > 0 {
		fmt.Fprintln(os.Stderr, "\n❌ Found codegen config violations:")

		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}

		return fmt.Errorf("codegen-config: found %d violations", len(violations))
	}

	logger.Log("All codegen configs contain required base initialisms")

	return nil
}

// findMissingInitialisms reads a codegen config file and returns any missing base initialisms.
func findMissingInitialisms(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	foundInitialisms := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Match lines like "- IDS" or "- IDS    # comment".
		if !strings.HasPrefix(line, "- ") {
			continue
		}

		entry := strings.TrimPrefix(line, "- ")

		// Split on whitespace or comment to get the initialism.
		parts := strings.Fields(entry)
		if len(parts) > 0 {
			foundInitialisms[parts[0]] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	var missing []string

	for _, initialism := range baseInitialisms {
		if !foundInitialisms[initialism] {
			missing = append(missing, initialism)
		}
	}

	return missing, nil
}
