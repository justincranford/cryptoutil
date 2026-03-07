// Copyright (c) 2025 Justin Cranford

// Package tls_minimum_version verifies that all TLS configurations use
// TLS 1.3 as the minimum version. Configurations that explicitly set
// MinVersion to a value below tls.VersionTLS13 are flagged as violations.
package tls_minimum_version

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// forbiddenMinVersions are TLS MinVersion values below TLS 1.3.
var forbiddenMinVersions = []string{
	"tls.VersionTLS10",
	"tls.VersionTLS11",
	"tls.VersionTLS12",
	"VersionTLS10",
	"VersionTLS11",
	"VersionTLS12",
}

// Check verifies TLS minimum version from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir verifies TLS minimum version under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking TLS minimum version requirements...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	var violations []string

	walkErr := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
				if name == cryptoutilSharedMagic.CICDExcludeDirGit || name == cryptoutilSharedMagic.CICDExcludeDirVendor || name == "test-output" || name == "archived" {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileViolations, scanErr := scanFileForTLSVersion(path, projectRoot)
		if scanErr != nil {
			return scanErr
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("filesystem walk failed: %w", walkErr)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d TLS minimum version violations", len(violations))
	}

	logger.Log("TLS minimum version check passed")

	return nil
}

// scanFileForTLSVersion scans a Go file for forbidden TLS MinVersion values.
func scanFileForTLSVersion(filePath, projectRoot string) ([]string, error) {
	f, err := os.Open(filePath) //nolint:gosec // filePath from filepath.Walk, controlled
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	rel, _ := filepath.Rel(projectRoot, filePath)

	var violations []string

	lineNum := 0

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue // Skip comment lines.
		}

		// Look for MinVersion assignments with a forbidden value.
		if !strings.Contains(line, "MinVersion") {
			continue
		}

		for _, forbidden := range forbiddenMinVersions {
			if strings.Contains(line, forbidden) {
				violations = append(violations, fmt.Sprintf(
					"%s:%d: TLS MinVersion below TLS 1.3: %s (use tls.VersionTLS13)",
					rel, lineNum, strings.TrimSpace(line)))

				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file %s: %w", filePath, err)
	}

	return violations, nil
}
