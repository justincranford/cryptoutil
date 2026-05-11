// Copyright (c) 2025-2026 Justin Cranford.
// Package testmain_integration_tag_policy enforces that testmain_test.go files
// do not carry build constraint directives.
//
// Per the testing standards, testmain_test.go files must not use build constraint
// directives of any form. This rule prevents package test entry-points from being
// accidentally excluded from a build variant, which causes the test runner to
// silently skip all package tests when a constraint is absent.
//
// Files scanned: every testmain_test.go found anywhere under internal/.
// Note: The apps_ps_id_template linter already enforces this rule for
// internal/apps/{PSID}/server/ and internal/apps/{PSID}/client/ packages.
// This linter extends coverage to internal/apps-framework/ and other subdirectories.
package testmain_integration_tag_policy

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
	// testmainFileName is the canonical filename for package test setup.
	testmainFileName = "testmain_test.go"

	// buildTagPrefix marks a modern (go 1.17+) build constraint line.
	buildTagPrefix = "//go:build"

	// legacyBuildTagPrefix marks a legacy (go 1.16 and earlier) build constraint line.
	legacyBuildTagPrefix = "// +build"
)

// Violation records one file that contains a forbidden build tag.
type Violation struct {
	File string
	Line int
	Tag  string
}

type readFileFunc func(string) ([]byte, error)

// Lint runs the linter from the current working directory.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
	return lintWithReader(logger, os.ReadFile)
}

// Check runs the linter from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return Lint(logger)
}

func lintWithReader(logger *cryptoutilCmdCicdCommon.Logger, readFileFn readFileFunc) error {
	return checkInDirWithReader(logger, ".", readFileFn)
}

// CheckInDir scans rootDir for testmain_test.go files that contain build tags.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithReader(logger, rootDir, os.ReadFile)
}

func checkInDirWithReader(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readFileFn readFileFunc) error {
	logger.Log("Checking that testmain_test.go files do not carry //go:build directives...")

	violations, err := findViolationsWithReader(rootDir, readFileFn)
	if err != nil {
		return fmt.Errorf("testmain-integration-tag-policy: directory walk failed: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "%s:%d: testmain_test.go must not have build tag %q\n",
				v.File, v.Line, v.Tag)
		}

		return fmt.Errorf("found %d testmain build tag policy violation(s)", len(violations))
	}

	logger.LogWithPrefix("testmain-integration-tag-policy", "✅ No testmain_test.go files carry build tags")

	return nil
}

// FindViolations returns all policy violations found under rootDir.
func FindViolations(rootDir string) ([]Violation, error) {
	return findViolationsWithReader(rootDir, os.ReadFile)
}

func findViolationsWithReader(rootDir string, readFileFn readFileFunc) ([]Violation, error) {
	var violations []Violation

	internalDir := filepath.Join(rootDir, "internal")

	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("internal/ directory not found at %s", internalDir)
	}

	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip hidden and vendor directories.
			if d.Name() == cryptoutilSharedMagic.CICDExcludeDirGit || d.Name() == cryptoutilSharedMagic.CICDExcludeDirVendor {
				return filepath.SkipDir
			}

			return nil
		}

		if d.Name() != testmainFileName {
			return nil
		}

		fileViolations, err := checkFile(path, readFileFn)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", internalDir, err)
	}

	return violations, nil
}

// checkFile returns violations for a single testmain_test.go file.
func checkFile(filePath string, readFileFn readFileFunc) ([]Violation, error) {
	content, err := readFileFn(filePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filePath, err)
	}

	var violations []Violation

	lines := strings.Split(string(content), "\n")
	for lineIndex, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if strings.HasPrefix(line, buildTagPrefix) || strings.HasPrefix(line, legacyBuildTagPrefix) {
			violations = append(violations, Violation{
				File: filePath,
				Line: lineIndex + 1,
				Tag:  line,
			})
		}
	}

	return violations, nil
}
