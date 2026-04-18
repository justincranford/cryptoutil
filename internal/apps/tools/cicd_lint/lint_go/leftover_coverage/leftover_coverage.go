// Copyright (c) 2025 Justin Cranford

// Package leftover_coverage detects banned "coverage-oriented" test file names
// such as *_coverage_test.go, *_comprehensive_test.go, *_gaps_test.go, etc.
// These names describe coverage intent rather than test content and are banned
// by the project test-file naming standards.
//
// Exception: package test files whose filename matches the package directory name
// (e.g. cicd_coverage/cicd_coverage_test.go) are allowed because the package name
// itself is the semantic identifier.
package leftover_coverage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// bannedSuffixes are the disallowed test file name patterns (without the _test.go suffix).
var bannedSuffixes = []string{
	"_coverage",
	"_coverage2",
	"_comprehensive",
	"_gaps",
	"_coverage_gaps",
	"_highcov",
	"_extra",
	"_additional",
	"_edge_cases",
}

// Check validates that no banned test file names exist in the repository.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", filepath.Walk)
}

// CheckInDir is the testable implementation that accepts explicit walk function.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, walkFn func(string, filepath.WalkFunc) error) error {
	logger.Log("Checking for banned coverage-oriented test file names...")

	var violations []string

	walkErr := walkFn(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		base := filepath.Base(path)

		// Only consider test files.
		if !strings.HasSuffix(base, "_test.go") {
			return nil
		}

		// Strip _test.go suffix to get the stem.
		stem := strings.TrimSuffix(base, "_test.go")

		// Check if stem ends with a banned suffix.
		banned := false

		for _, suffix := range bannedSuffixes {
			if strings.HasSuffix(stem, suffix) {
				banned = true

				break
			}
		}

		if !banned {
			return nil
		}

		// Exception: filename matches parent directory name.
		// e.g. cicd_coverage/cicd_coverage_test.go is allowed.
		// Use the absolute path to reliably determine the parent directory name.
		absPath, absErr := filepath.Abs(path)
		if absErr == nil {
			parentDir := filepath.Base(filepath.Dir(absPath))
			if stem == parentDir {
				return nil
			}
		}

		relPath := path
		if rel, relErr := filepath.Rel(rootDir, path); relErr == nil {
			relPath = rel
		}

		violations = append(violations, relPath)

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("directory walk failed: %w", walkErr)
	}

	if len(violations) == 0 {
		logger.Log("✅ No banned coverage-oriented test file names found")

		return nil
	}

	for _, v := range violations {
		logger.Log(fmt.Sprintf("VIOLATION: banned test file name: %s", v))
	}

	return fmt.Errorf("leftover-coverage: %d banned test file name(s) found:\n%s",
		len(violations), strings.Join(violations, "\n"))
}
