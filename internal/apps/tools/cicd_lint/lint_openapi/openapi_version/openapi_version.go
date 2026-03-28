// Copyright (c) 2025 Justin Cranford

// Package openapi_version validates that OpenAPI spec files use version 3.0.3.
package openapi_version

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

const requiredVersion = "3.0.3"

// Check validates that all OpenAPI spec files in api/ use version 3.0.3.
func Check(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Checking OpenAPI spec versions...")

	yamlFiles := append(filesByExtension["yaml"], filesByExtension["yml"]...)
	if len(yamlFiles) == 0 {
		logger.Log("No YAML files to check")

		return nil
	}

	// Filter to only api/ directory OpenAPI spec files.
	var specFiles []string

	for _, f := range yamlFiles {
		normalized := filepath.ToSlash(f)
		base := filepath.Base(f)

		if strings.Contains(normalized, "api/") && strings.HasPrefix(base, "openapi_spec") {
			specFiles = append(specFiles, f)
		}
	}

	if len(specFiles) == 0 {
		logger.Log("No OpenAPI spec files found in api/ directory")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d OpenAPI spec files to check", len(specFiles)))

	var violations []string

	for _, specFile := range specFiles {
		version, err := extractOpenAPIVersion(specFile)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: %s", specFile, err.Error()))

			continue
		}

		if version == "" {
			// Component files (openapi_spec_components.yaml, openapi_spec_paths.yaml) may not have the openapi field.
			continue
		}

		if version != requiredVersion {
			violations = append(violations, fmt.Sprintf("%s: openapi version is %q, expected %q", specFile, version, requiredVersion))
		}
	}

	if len(violations) > 0 {
		fmt.Fprintln(os.Stderr, "\n❌ Found OpenAPI version violations:")

		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}

		return fmt.Errorf("openapi-version: found %d violations", len(violations))
	}

	logger.Log("All OpenAPI specs use version 3.0.3")

	return nil
}

// extractOpenAPIVersion reads the first few lines of a YAML file to find the openapi version field.
func extractOpenAPIVersion(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "openapi:") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "openapi:"))
			// Remove surrounding quotes if present.
			version = strings.Trim(version, "'\"")

			return version, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %w", err)
	}

	// No openapi field found — component/paths files may omit it.
	return "", nil
}
