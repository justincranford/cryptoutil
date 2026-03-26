// Copyright (c) 2025 Justin Cranford

// Package root_junk_detection enforces that no junk files exist at the project root.
// Junk files include compiled binaries (*.exe, *.test.exe), Python scripts (*.py),
// and test-output coverage files (coverage*). These must never be committed to
// the repository (ARCHITECTURE.md Section A.2 and Section 13.9).
package root_junk_detection

import (
	"fmt"
	"os"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// bannedSuffixes is the list of file extensions that must not exist at project root.
var bannedSuffixes = []string{
	".exe",
	".py",
	".test.exe",
}

// bannedPrefixes is the list of file name prefixes that must not exist at project root.
var bannedPrefixes = []string{
	"coverage",
}

// Check runs the root-junk-detection check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks rootDir for junk files at the project root level.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking root directory for junk files...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check root directory for junk files: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  junk file at root: %s\n", v)
		}

		return fmt.Errorf("root-junk-detection: found %d junk file(s) at project root that must be removed", len(violations))
	}

	logger.Log("root-junk-detection: no junk files found at project root")

	return nil
}

// FindViolationsInDir lists rootDir entries and returns the names of any banned files.
func FindViolationsInDir(rootDir string) ([]string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read root directory %s: %w", rootDir, err)
	}

	var violations []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if isBannedRootFile(name) {
			violations = append(violations, name)
		}
	}

	return violations, nil
}

// isBannedRootFile returns true if the file name matches any banned pattern.
func isBannedRootFile(name string) bool {
	nameLower := strings.ToLower(name)

	for _, suffix := range bannedSuffixes {
		if strings.HasSuffix(nameLower, suffix) {
			return true
		}
	}

	for _, prefix := range bannedPrefixes {
		if strings.HasPrefix(nameLower, prefix) {
			return true
		}
	}

	return false
}
