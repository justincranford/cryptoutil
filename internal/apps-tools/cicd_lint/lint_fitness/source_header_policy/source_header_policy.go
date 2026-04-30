// Copyright (c) 2025-2026 Justin Cranford.

// Package source_header_policy validates repository source header conventions.
package source_header_policy

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	requiredSPDXLicense     = "AGPL-3.0-only"
	copyrightRangeMatchSize = 3
)

var (
	spdxHeaderPattern      = regexp.MustCompile(`(?m)^//\s*SPDX-License-Identifier:\s*(.+)\s*$`)
	copyrightHeaderPattern = regexp.MustCompile(`(?m)^//\s*Copyright \(c\)\s*(\d{4})(?:-(\d{4}))?\s+Justin Cranford\.?\s*$`)
)

// Check validates source header policy from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates source header policy under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	currentYear := time.Now().UTC().Year()

	return checkInDirWithYear(logger, rootDir, currentYear)
}

func checkInDirWithYear(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, currentYear int) error {
	logger.Log("Checking source header policy for SPDX and copyright year drift...")

	violations, err := findViolations(rootDir, currentYear)
	if err != nil {
		return fmt.Errorf("source-header-policy: %w", err)
	}

	if len(violations) > 0 {
		sort.Strings(violations)

		return fmt.Errorf("source-header-policy violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("source-header-policy: all Go source headers comply with repository policy")

	return nil
}

func findViolations(rootDir string, currentYear int) ([]string, error) {
	var violations []string

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch d.Name() {
			case cryptoutilSharedMagic.CICDExcludeDirVendor, cryptoutilSharedMagic.CICDExcludeDirGit:
				return filepath.SkipDir
			default:
				return nil
			}
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		fileViolations, err := checkFile(path, currentYear)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk source files: %w", err)
	}

	return violations, nil
}

func checkFile(path string, currentYear int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	header, err := readHeaderSection(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read header section %s: %w", path, err)
	}

	var violations []string

	if match := spdxHeaderPattern.FindStringSubmatch(header); len(match) == 2 {
		licenseID := strings.TrimSpace(match[1])
		if licenseID != requiredSPDXLicense {
			violations = append(violations, fmt.Sprintf("%s: SPDX-License-Identifier is %q; expected %q", path, licenseID, requiredSPDXLicense))
		}
	}

	if match := copyrightHeaderPattern.FindStringSubmatch(header); len(match) == copyrightRangeMatchSize {
		startYear, endYear, parseErr := parseCopyrightYears(match[1], match[2])
		if parseErr != nil {
			violations = append(violations, fmt.Sprintf("%s: invalid copyright header years: %v", path, parseErr))
		} else {
			if endYear < currentYear {
				violations = append(violations, fmt.Sprintf("%s: copyright year range ends at %d; current year is %d", path, endYear, currentYear))
			}

			if startYear > endYear {
				violations = append(violations, fmt.Sprintf("%s: copyright year range start %d is greater than end %d", path, startYear, endYear))
			}
		}
	}

	return violations, nil
}

func readHeaderSection(file *os.File) (string, error) {
	scanner := bufio.NewScanner(file)

	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "//"):
			lines = append(lines, line)
		case trimmed == "":
			lines = append(lines, line)
		case strings.HasPrefix(trimmed, "package "):
			return strings.Join(lines, "\n"), nil
		default:
			return strings.Join(lines, "\n"), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan header section: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}

func parseCopyrightYears(startText, endText string) (int, int, error) {
	var startYear int
	if _, err := fmt.Sscanf(startText, "%d", &startYear); err != nil {
		return 0, 0, fmt.Errorf("invalid start year %q", startText)
	}

	if strings.TrimSpace(endText) == "" {
		return startYear, startYear, nil
	}

	var endYear int
	if _, err := fmt.Sscanf(endText, "%d", &endYear); err != nil {
		return 0, 0, fmt.Errorf("invalid end year %q", endText)
	}

	return startYear, endYear, nil
}
