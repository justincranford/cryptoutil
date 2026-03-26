// Copyright (c) 2025 Justin Cranford

// Package require_framework_naming prevents regression to the old
// internal/apps/template/ import path. After the framework rename
// (internal/apps/template/ -> internal/apps/framework/), any Go file
// importing internal/apps/template/ is a violation -- UNLESS the import
// is for the skeleton-template service (internal/apps/skeleton-template/).
package require_framework_naming

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// bannedImportFragment is the old framework import path fragment that must
	// not appear in any Go import (except skeleton-template).
	bannedImportFragment = "cryptoutil/internal/apps/template/"

	// allowedSkeletonPrefix is the skeleton-template import path prefix which
	// is exempt from this rule.
	allowedSkeletonPrefix = "cryptoutil/internal/apps/skeleton-template/"
)

var importLinePattern = regexp.MustCompile(`^\s+(?:\w+ )?"([^"]+)"`)

var singleImportPattern = regexp.MustCompile(`^import\s+"([^"]+)"`)

// Test seam: replaceable in tests to exercise unreachable OS-level error paths.
// See ARCHITECTURE.md Section 10.2.4 (Test Seam Injection Pattern).
var walkFn = filepath.Walk

// Check validates that no Go file imports the banned old framework path.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates that no Go file under rootDir imports the banned path.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for banned internal/apps/template/ imports...")

	var violations []string

	err := walkFn(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == cryptoutilSharedMagic.CICDExcludeDirGit || name == cryptoutilSharedMagic.CICDExcludeDirVendor || strings.HasPrefix(name, "_") {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileViolations, parseErr := checkFile(path)
		if parseErr != nil {
			return parseErr
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking directory tree: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d banned internal/apps/template/ imports (use internal/apps/framework/ instead)", len(violations))
	}

	logger.Log("require-framework-naming: no banned internal/apps/template/ imports found")

	return nil
}

// checkFile scans a single Go file for banned imports.
func checkFile(filePath string) ([]string, error) {
	f, err := os.Open(filePath) //nolint:gosec // filePath from filepath.Walk, controlled
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	var violations []string

	inImport := false
	inRawString := false

	scanner := bufio.NewScanner(f)

	lineNum := 0

	for scanner.Scan() {
		lineNum++

		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track raw string literals (backtick strings) to avoid false positives
		// from import-like patterns inside test fixture strings.
		backtickCount := strings.Count(line, "`")
		if backtickCount%2 != 0 {
			inRawString = !inRawString
		}

		if inRawString {
			continue
		}

		if trimmed == "import (" {
			inImport = true

			continue
		}

		if inImport && trimmed == ")" {
			inImport = false

			continue
		}

		if inImport {
			if m := importLinePattern.FindStringSubmatch(line); len(m) > 1 {
				imp := m[1]
				if isBannedImport(imp) {
					violations = append(violations, fmt.Sprintf(
						"%s:%d: imports banned path %q (use internal/apps/framework/ instead of internal/apps/template/)",
						filePath, lineNum, imp))
				}
			}
		} else if strings.HasPrefix(trimmed, `import "`) {
			if m := singleImportPattern.FindStringSubmatch(trimmed); len(m) > 1 {
				imp := m[1]
				if isBannedImport(imp) {
					violations = append(violations, fmt.Sprintf(
						"%s:%d: imports banned path %q (use internal/apps/framework/ instead of internal/apps/template/)",
						filePath, lineNum, imp))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file %s: %w", filePath, err)
	}

	return violations, nil
}

// isBannedImport returns true if the import path uses the old template path
// and is NOT the whitelisted skeleton-template path.
func isBannedImport(importPath string) bool {
	if !strings.Contains(importPath, bannedImportFragment) {
		return false
	}

	// Allow skeleton-template imports (internal/apps/skeleton-template/).
	if strings.Contains(importPath, allowedSkeletonPrefix) {
		return false
	}

	return true
}
