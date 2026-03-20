// Copyright (c) 2025 Justin Cranford

// Package no_postgres_in_non_e2e enforces that PostgreSQL test containers are only used in E2E tests.
// Unit tests and integration tests MUST use testdb.NewInMemorySQLiteDB() instead.
// PostgreSQL containers are allowed only in files ending with _e2e_test.go or containing //go:build e2e.
// See ARCHITECTURE.md Section 10.3 Integration Testing Strategy (D19).
package no_postgres_in_non_e2e

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// bannedPatterns contains lowercase patterns indicating a PostgreSQL container is being started.
// PostgreSQL containers are forbidden in non-E2E tests. Use testdb.NewInMemorySQLiteDB() instead.
var bannedPatterns = []string{
	"postgres.runcontainer(",
	"postgres.run(",
	"postgresmodule.run(",
	".newpostgrestestcontainer(",
	".requirenewpostgrestestcontainer(",
}

// allowedSuffixes lists test file suffixes permitted to use PostgreSQL containers.
var allowedSuffixes = []string{
	"_e2e_test.go",
}

// allowedPathFragments lists path segments permitted to use PostgreSQL containers.
var allowedPathFragments = []string{
	"testing/testdb/",
	"shared/container/",
	"shared/database/",
	"lint_fitness/",
	"lint_gotest/",
}

// maxBuildTagLines is the number of header lines scanned for a //go:build e2e directive.
const maxBuildTagLines = 10

// Check walks all test files in "." and reports PostgreSQL container violations.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir walks test files from rootDir and reports PostgreSQL container violations.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for PostgreSQL containers in non-E2E tests...")

	var testFiles []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch d.Name() {
			case cryptoutilSharedMagic.CICDExcludeDirGit, cryptoutilSharedMagic.CICDExcludeDirVendor:
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		normalized := filepath.ToSlash(path)

		for _, suffix := range allowedSuffixes {
			if strings.HasSuffix(normalized, suffix) {
				return nil
			}
		}

		for _, fragment := range allowedPathFragments {
			if strings.Contains(normalized, fragment) {
				return nil
			}
		}

		testFiles = append(testFiles, path)

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking test files: %w", err)
	}

	return CheckFiles(logger, testFiles)
}

// CheckFiles checks the provided list of test files for PostgreSQL container violations.
func CheckFiles(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	if len(testFiles) == 0 {
		logger.Log("No non-E2E test files to check for PostgreSQL containers")

		return nil
	}

	totalViolations := 0

	for _, filePath := range testFiles {
		issues := CheckFile(filePath)

		if len(issues) > 0 {
			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "%s: %s\n", filePath, issue)
			}

			totalViolations += len(issues)
		}
	}

	if totalViolations > 0 {
		logger.Log(fmt.Sprintf("Found %d PostgreSQL container violation(s) in non-E2E tests", totalViolations))
		fmt.Fprintln(os.Stderr, "Use testdb.NewInMemorySQLiteDB() for unit/integration tests or place in _e2e_test.go files.")
		fmt.Fprintln(os.Stderr, "See ARCHITECTURE.md Section 10.3 Integration Testing Strategy (D19).")

		return fmt.Errorf("found %d PostgreSQL container violation(s) in non-E2E tests", totalViolations)
	}

	logger.Log("\u2705 No PostgreSQL containers found in non-E2E tests")

	return nil
}

// CheckFile scans a single file for banned PostgreSQL container patterns.
// Files with //go:build e2e tag are allowed to use PostgreSQL containers.
func CheckFile(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("error reading file: %v", err)}
	}

	if hasE2EBuildTag(string(content)) {
		return nil
	}

	var violations []string

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		rawLine := scanner.Text()
		line := strings.ToLower(strings.TrimSpace(rawLine))

		// Skip comment lines.
		if strings.HasPrefix(line, "//") {
			continue
		}

		for _, pattern := range bannedPatterns {
			if strings.Contains(line, pattern) {
				violations = append(violations, fmt.Sprintf(
					"line %d: PostgreSQL container %q - use testdb.NewInMemorySQLiteDB() or place in _e2e_test.go",
					lineNum, rawLine))

				break
			}
		}
	}

	return violations
}

// hasE2EBuildTag checks the first maxBuildTagLines header lines for a //go:build e2e directive.
func hasE2EBuildTag(content string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	linesRead := 0

	for scanner.Scan() && linesRead < maxBuildTagLines {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "//go:build") && strings.Contains(line, "e2e") {
			return true
		}

		linesRead++
	}

	return false
}
