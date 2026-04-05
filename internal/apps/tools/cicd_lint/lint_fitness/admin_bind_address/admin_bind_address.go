// Copyright (c) 2025 Justin Cranford

// Package admin_bind_address verifies that admin/private server bindings use
// 127.0.0.1 (loopback), not 0.0.0.0. The admin port must never be exposed
// publicly. See ENG-HANDBOOK.md Section 5.3: admin binds to 127.0.0.1.
package admin_bind_address

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Check verifies admin bind address from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", filepath.Walk, os.Open)
}

// CheckInDir verifies admin bind address under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, walkFn func(string, filepath.WalkFunc) error, openFn func(string) (*os.File, error)) error {
	logger.Log("Checking admin bind address (must be 127.0.0.1, not 0.0.0.0)...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	var violations []string

	walkErr := walkFn(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == cryptoutilSharedMagic.CICDExcludeDirGit || name == cryptoutilSharedMagic.CICDExcludeDirVendor || name == cryptoutilSharedMagic.CICDExcludeDirTestOutput || name == "deployments" {
				return filepath.SkipDir
			}

			return nil
		}

		// Only scan Go source files; exclude test files which may contain "0.0.0.0"
		// as test fixture strings when testing the linter itself.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileViolations, scanErr := scanForAdminBindViolations(path, projectRoot, openFn)
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

		return fmt.Errorf("found %d admin bind address violations", len(violations))
	}

	logger.Log("Admin bind address check passed")

	return nil
}

// adminBindPatterns are source patterns that assign 0.0.0.0 to admin/private bind addresses.
var adminBindPatterns = []string{
	"BindPrivateAddress",
}

// scanForAdminBindViolations checks a Go file for admin bind address set to 0.0.0.0.
func scanForAdminBindViolations(filePath, projectRoot string, openFn func(string) (*os.File, error)) ([]string, error) {
	f, err := openFn(filePath) //nolint:gosec // filePath from filepath.Walk, controlled
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

		// Check for patterns like BindPrivateAddress set to 0.0.0.0.
		for _, pattern := range adminBindPatterns {
			if strings.Contains(line, pattern) && strings.Contains(line, cryptoutilSharedMagic.IPv4AnyAddress) {
				violations = append(violations, fmt.Sprintf(
					"%s:%d: admin bind address is 0.0.0.0 (use 127.0.0.1): %s",
					rel, lineNum, strings.TrimSpace(line)))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file %s: %w", filePath, err)
	}

	return violations, nil
}
