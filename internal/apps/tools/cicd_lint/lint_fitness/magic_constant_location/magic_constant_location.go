// Copyright (c) 2025 Justin Cranford

// Package magic_constant_location enforces that magic constants are declared
// in internal/shared/magic/ and NOT as package-local const declarations
// (ENG-HANDBOOK.md Section 11.1.4 Magic Values Organization).
package magic_constant_location

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// portRangeMin is the minimum port-like integer value flagged as suspicious.
var portRangeMin = cryptoutilSharedMagic.PortRangeSuspiciousMin

// portRangeMax is the maximum port-like integer value flagged as suspicious.
var portRangeMax = cryptoutilSharedMagic.PortRangeSuspiciousMax

// constLineRegexp matches Go const declarations with integer values.
var constLineRegexp = regexp.MustCompile(`^\s*(?:const\s+)?(\w+)\s*(?:\w+)?\s*=\s*(\d+)\s*(?://.*)?$`)

// Check runs the magic-constant-location check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates that no non-magic packages declare const values in the
// suspicious port range (1000-65535). Reports violations as warnings (informational).
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking magic constant locations...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check magic-constant-location: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			logger.Log(fmt.Sprintf("  WARNING: %s", v))
		}

		logger.Log(fmt.Sprintf("magic-constant-location: %d suspicious const(s) found outside internal/shared/magic/ (informational — fix incrementally)", len(violations)))
	} else {
		logger.Log("magic-constant-location: all magic constants are in internal/shared/magic/")
	}

	return nil
}

// FindViolationsInDir scans Go source files outside internal/shared/magic/ for
// const declarations with integer values in the suspicious port range.
func FindViolationsInDir(rootDir string) ([]string, error) {
	var violations []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			return skipDir(rootDir, path)
		}

		if !isEligibleGoFile(path) {
			return nil
		}

		relPath, relErr := filepath.Rel(rootDir, path)
		if relErr != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, relErr)
		}

		fileViolations, scanErr := scanFileForSuspiciousConsts(path, filepath.ToSlash(relPath))
		if scanErr != nil {
			return scanErr
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	sort.Strings(violations)

	return violations, nil
}

// skipDir returns filepath.SkipDir for directories that should not be scanned.
func skipDir(rootDir, path string) error {
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	slashRel := filepath.ToSlash(rel)

	// Skip the magic package itself.
	if strings.HasPrefix(slashRel, "internal/shared/magic") {
		return filepath.SkipDir
	}

	// Skip vendor, .git, testdata, api (generated code).
	base := filepath.Base(path)
	switch base {
	case cryptoutilSharedMagic.CICDExcludeDirGit, cryptoutilSharedMagic.CICDExcludeDirVendor, "testdata":
		return filepath.SkipDir
	}

	// Skip api/ at project root (generated code).
	if slashRel == "api" {
		return filepath.SkipDir
	}

	return nil
}

// isEligibleGoFile returns true for non-test, non-generated Go source files.
func isEligibleGoFile(path string) bool {
	if !strings.HasSuffix(path, ".go") {
		return false
	}

	base := filepath.Base(path)

	// Skip test files.
	if strings.HasSuffix(base, "_test.go") {
		return false
	}

	// Skip generated files.
	if strings.HasSuffix(base, ".gen.go") || strings.HasSuffix(base, "_gen.go") {
		return false
	}

	return true
}

// scanFileForSuspiciousConsts scans a Go file for const declarations with
// integer values in the suspicious port range.
func scanFileForSuspiciousConsts(absPath, relPath string) ([]string, error) {
	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", relPath, err)
	}

	defer func() { _ = file.Close() }()

	var violations []string

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inConstBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track const blocks.
		if strings.HasPrefix(trimmed, "const (") || trimmed == "const(" {
			inConstBlock = true

			continue
		}

		if inConstBlock && trimmed == ")" {
			inConstBlock = false

			continue
		}

		// Check standalone const or const block entries.
		if !inConstBlock && !strings.HasPrefix(trimmed, "const ") {
			continue
		}

		matches := constLineRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		constName := matches[1]
		valueStr := matches[2]

		value, parseErr := strconv.Atoi(valueStr)
		if parseErr != nil {
			continue
		}

		if value >= portRangeMin && value <= portRangeMax {
			violations = append(violations, fmt.Sprintf(
				"%s:%d: const %s = %d is in suspicious port range (%d-%d); move to internal/shared/magic/",
				relPath, lineNum, constName, value, portRangeMin, portRangeMax,
			))
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", relPath, scanErr)
	}

	return violations, nil
}
