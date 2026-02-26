// Copyright (c) 2025 Justin Cranford

// Package leftover_coverage detects and removes leftover coverage files.
package leftover_coverage

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// CoveragePatterns defines file patterns considered leftover coverage artifacts.
// These files should be placed in test-output/ or deleted after test runs.
var CoveragePatterns = []string{
	"*.out",
	"*.cov",
	"*.prof",
	"*coverage*.html",
	"*coverage*.txt",
}

// Check detects and auto-deletes leftover coverage files.
// Scans ALL directories including test-output/ per user decision.
// Returns error if files were found and deleted (to trigger CI awareness).
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir detects and auto-deletes leftover coverage files in rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for leftover coverage files...")

	var deletedFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Skip vendor and .git directories.
		if info.IsDir() && (info.Name() == cryptoutilSharedMagic.CICDExcludeDirVendor || info.Name() == cryptoutilSharedMagic.CICDExcludeDirGit) {
			return filepath.SkipDir
		}

		// Skip known output directories where coverage files are expected.
		if info.IsDir() && (info.Name() == "workflow-reports" || info.Name() == "test-output") {
			return filepath.SkipDir
		}

		// Skip directories.
		if info.IsDir() {
			return nil
		}

		// Check if file matches any coverage pattern.
		if MatchesCoveragePattern(info.Name()) {
			// Auto-delete the file per user decision.
			if err := os.Remove(path); err != nil {
				logger.Log(fmt.Sprintf("⚠️  WARNING: Failed to delete leftover coverage file: %s (error: %v)", path, err))
			} else {
				logger.Log(fmt.Sprintf("⚠️  WARNING: Deleted leftover coverage file: %s", path))
				deletedFiles = append(deletedFiles, path)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory tree: %w", err)
	}

	if len(deletedFiles) > 0 {
		printLeftoverCoverageViolations(deletedFiles)

		return fmt.Errorf("found and deleted %d leftover coverage files", len(deletedFiles))
	}

	logger.Log("✅ No leftover coverage files found")

	return nil
}

// MatchesCoveragePattern checks if a filename matches any coverage pattern.
func MatchesCoveragePattern(filename string) bool {
	for _, pattern := range CoveragePatterns {
		matched, err := filepath.Match(pattern, filename)
		if err == nil && matched {
			return true
		}

		// Handle patterns with wildcards in the middle (e.g., *coverage*.html).
		if strings.Contains(pattern, "*") {
			// Split pattern into parts around wildcards.
			parts := strings.Split(pattern, "*")
			allPartsMatch := true

			for _, part := range parts {
				if part == "" {
					continue
				}

				if !strings.Contains(strings.ToLower(filename), strings.ToLower(part)) {
					allPartsMatch = false

					break
				}
			}

			if allPartsMatch && len(parts) > 1 {
				return true
			}
		}
	}

	return false
}

// printLeftoverCoverageViolations prints formatted list of deleted files.
func printLeftoverCoverageViolations(deletedFiles []string) {
	fmt.Fprintf(os.Stderr, "\n⚠️  Deleted %d leftover coverage files:\n\n", len(deletedFiles))

	for _, f := range deletedFiles {
		fmt.Fprintf(os.Stderr, "  - %s\n", f)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Why this matters:")
	fmt.Fprintln(os.Stderr, "  - Coverage files should NOT be committed to git")
	fmt.Fprintln(os.Stderr, "  - Coverage files should be placed in test-output/ directory")
	fmt.Fprintln(os.Stderr, "  - LLM agents may create coverage files in unexpected locations")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Patterns detected:")

	for _, p := range CoveragePatterns {
		fmt.Fprintf(os.Stderr, "  - %s\n", p)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Fix:")
	fmt.Fprintln(os.Stderr, "  1. Use test-output/ directory for coverage files")
	fmt.Fprintln(os.Stderr, "  2. Add coverage files to .gitignore")
	fmt.Fprintln(os.Stderr, "  3. Run: go test -coverprofile=test-output/coverage.out ./...")
}
