// Copyright (c) 2025 Justin Cranford

// Package no_local_closed_db_helper enforces that createClosedDatabase-style helpers
// are not defined outside the shared testdb package.
// Use testdb.NewClosedSQLiteDB() from internal/apps/framework/service/testing/testdb instead.
package no_local_closed_db_helper

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// allowedPathFragment is the testdb package path where closed-DB helpers are permitted.
const allowedPathFragment = "testing/testdb/"

// bannedFunctions contains lowercase prefix patterns for locally-defined closed-DB helpers.
var bannedFunctions = []string{
	"func createcloseddatabase(",
	"func createcloseddb(",
	"func createclosedservicedependencies(",
	"func createcloseddbhandler(",
}

// Check walks all test files in "." and reports closed-DB helper violations.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir walks test files from rootDir and reports violations.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for local createClosedDatabase-style helpers outside testdb package...")

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

		// Skip the testdb package - it is the canonical location for these helpers.
		if strings.Contains(filepath.ToSlash(path), allowedPathFragment) {
			return nil
		}

		testFiles = append(testFiles, path)

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking test files: %w", err)
	}

	return CheckFiles(logger, testFiles)
}

// CheckFiles checks the provided list of test files for violations.
func CheckFiles(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	if len(testFiles) == 0 {
		logger.Log("No test files to check for local closed-DB helpers")

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
		logger.Log(fmt.Sprintf("Found %d local closed-DB helper violation(s)", totalViolations))
		fmt.Fprintln(os.Stderr, "Use testdb.NewClosedSQLiteDB() from internal/apps/framework/service/testing/testdb instead.")

		return fmt.Errorf("found %d local closed-DB helper violation(s)", totalViolations)
	}

	logger.Log("\u2705 No local closed-DB helpers found")

	return nil
}

// CheckFile scans a single file for banned closed-DB helper function definitions.
func CheckFile(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("error reading file: %v", err)}
	}

	var violations []string

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		line := strings.ToLower(strings.TrimSpace(scanner.Text()))

		for _, banned := range bannedFunctions {
			if strings.HasPrefix(line, banned) {
				violations = append(violations, fmt.Sprintf("line %d: banned private function %q - use testdb.NewClosedSQLiteDB() instead", lineNum, scanner.Text()))
			}
		}
	}

	return violations
}
