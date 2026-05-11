// Copyright (c) 2025-2026 Justin Cranford.
// Package testmain_e2e_policy enforces canonical E2E TestMain imports.
package testmain_e2e_policy

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	e2eTestMainFileName = "testmain_e2e_test.go"
	requiredImportPath  = "service/test_orch_e2e"
	forbiddenImportPath = "service/testing/e2e_infra"
)

// Violation describes one policy violation in an E2E TestMain file.
type Violation struct {
	File   string
	Reason string
}

type (
	readFileFunc func(string) ([]byte, error)
	walkDirFunc  func(string, fs.WalkDirFunc) error
)

// Lint runs the policy check from the current working directory.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkInDirWithReader(logger, ".", os.ReadFile)
}

// Check runs the policy check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return Lint(logger)
}

// CheckInDir runs the policy check under a specific root directory.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithReader(logger, rootDir, os.ReadFile)
}

func checkInDirWithReader(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readFileFn readFileFunc) error {
	logger.Log("Checking E2E TestMain import policy...")

	violations, err := findViolationsWithReader(rootDir, readFileFn)
	if err != nil {
		return fmt.Errorf("testmain-e2e-policy: directory walk failed: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "%s: %s\n", v.File, v.Reason)
		}

		return fmt.Errorf("found %d testmain e2e policy violation(s)", len(violations))
	}

	logger.LogWithPrefix("testmain-e2e-policy", "✅ All E2E TestMain files import test_orch_e2e and avoid e2e_infra")

	return nil
}

// FindViolations returns policy violations under rootDir.
func FindViolations(rootDir string) ([]Violation, error) {
	return findViolationsWithReader(rootDir, os.ReadFile)
}

func findViolationsWithReader(rootDir string, readFileFn readFileFunc) ([]Violation, error) {
	return findViolationsWithDeps(rootDir, readFileFn, filepath.WalkDir)
}

func findViolationsWithDeps(rootDir string, readFileFn readFileFunc, walkDirFn walkDirFunc) ([]Violation, error) {
	var violations []Violation

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("internal/apps directory not found at %s", appsDir)
	}

	err := walkDirFn(appsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == cryptoutilSharedMagic.CICDExcludeDirGit || d.Name() == cryptoutilSharedMagic.CICDExcludeDirVendor {
				return filepath.SkipDir
			}

			return nil
		}

		if d.Name() != e2eTestMainFileName || filepath.Base(filepath.Dir(path)) != "e2e" {
			return nil
		}

		content, readErr := readFileFn(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		text := string(content)
		hasRequiredImport := strings.Contains(text, requiredImportPath)
		hasForbiddenImport := strings.Contains(text, forbiddenImportPath)

		switch {
		case !hasRequiredImport && hasForbiddenImport:
			violations = append(violations, Violation{
				File:   path,
				Reason: fmt.Sprintf("must import %q", requiredImportPath),
			})
			violations = append(violations, Violation{
				File:   path,
				Reason: fmt.Sprintf("must not import %q", forbiddenImportPath),
			})
		case !hasRequiredImport:
			violations = append(violations, Violation{
				File:   path,
				Reason: fmt.Sprintf("must import %q", requiredImportPath),
			})
		case hasForbiddenImport:
			violations = append(violations, Violation{
				File:   path,
				Reason: fmt.Sprintf("must not import %q", forbiddenImportPath),
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", appsDir, err)
	}

	return violations, nil
}
