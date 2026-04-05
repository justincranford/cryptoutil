// Copyright (c) 2025 Justin Cranford

// Package root_junk_detection enforces that no junk files or directories exist at the project root.
// Junk files include compiled binaries (*.exe, *.test.exe), Python scripts (*.py),
// and test-output coverage files (coverage*). Junk directories include coverage
// output directories (*_coverage, cover). These must never be committed to
// the repository (ENG-HANDBOOK.md Section A.2 and Section 13.9).
package root_junk_detection

import (
	"fmt"
	"os"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// bannedFileSuffixes is the list of file extensions that must not exist at project root.
var bannedFileSuffixes = []string{
	".exe",
	".py",
	".test.exe",
}

// bannedFilePrefixes is the list of file name prefixes that must not exist at project root.
var bannedFilePrefixes = []string{
	"coverage",
}

// bannedDirSuffixes is the list of directory name suffixes that must not exist at project root.
var bannedDirSuffixes = []string{
	"_coverage",
}

// bannedDirExactNames is the list of exact directory names that must not exist at project root.
var bannedDirExactNames = []string{
	"cover",
}

// Check runs the root-junk-detection check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks rootDir for junk files at the project root level.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking root directory for junk files and directories...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check root directory for junk entries: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  junk entry at root: %s (remove this to fix)\n", v)
		}

		return fmt.Errorf("root-junk-detection: found %d junk entry/entries at project root that must be removed", len(violations))
	}

	logger.Log("root-junk-detection: no junk entries found at project root")

	return nil
}

// FindViolationsInDir lists rootDir entries and returns the names of any banned files or directories.
func FindViolationsInDir(rootDir string) ([]string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read root directory %s: %w", rootDir, err)
	}

	var violations []string

	for _, entry := range entries {
		name := entry.Name()

		switch {
		case entry.IsDir() && isBannedRootDir(name):
			violations = append(violations, name+"/")
		case !entry.IsDir() && isBannedRootFile(name):
			violations = append(violations, name)
		}
	}

	return violations, nil
}

// isBannedRootFile returns true if the file name matches any banned file pattern.
func isBannedRootFile(name string) bool {
	nameLower := strings.ToLower(name)

	for _, suffix := range bannedFileSuffixes {
		if strings.HasSuffix(nameLower, suffix) {
			return true
		}
	}

	for _, prefix := range bannedFilePrefixes {
		if strings.HasPrefix(nameLower, prefix) {
			return true
		}
	}

	return false
}

// isBannedRootDir returns true if the directory name matches any banned directory pattern.
func isBannedRootDir(name string) bool {
	nameLower := strings.ToLower(name)

	for _, suffix := range bannedDirSuffixes {
		if strings.HasSuffix(nameLower, suffix) {
			return true
		}
	}

	for _, exact := range bannedDirExactNames {
		if nameLower == exact {
			return true
		}
	}

	return false
}
