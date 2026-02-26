// Copyright (c) 2025 Justin Cranford

// Package insecure_skip_verify verifies that code does not disable TLS certificate verification.
package insecure_skip_verify

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)


// Check verifies that code doesn't disable TLS certificate verification.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks for InsecureSkipVerify usage in rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for InsecureSkipVerify usage...")

	violations, err := FindInsecureSkipVerifyViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check InsecureSkipVerify: %w", err)
	}

	if len(violations) > 0 {
		lintGoCommon.PrintCryptoViolations("InsecureSkipVerify", violations)

		return fmt.Errorf("found %d InsecureSkipVerify violations - never disable TLS certificate verification", len(violations))
	}

	logger.Log("✅ InsecureSkipVerify validation passed")

	return nil
}

// insecureSkipVerifyPattern matches InsecureSkipVerify set to true in Go code.
var insecureSkipVerifyPattern = regexp.MustCompile(`InsecureSkipVerify\s*:\s*true`)

// FindInsecureSkipVerifyViolationsInDir scans Go files in rootDir for disabling TLS verification.
func FindInsecureSkipVerifyViolationsInDir(rootDir string) ([]lintGoCommon.CryptoViolation, error) {
	var violations []lintGoCommon.CryptoViolation

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories.
		if d.IsDir() {
			if d.Name() == cryptoutilSharedMagic.CICDExcludeDirVendor || d.Name() == cryptoutilSharedMagic.CICDExcludeDirGit {
				return filepath.SkipDir
			}

			return nil
		}

		// Only check .go files.
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files - tests may use InsecureSkipVerify for local testing.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip test helper directories - they legitimately need InsecureSkipVerify.
		if strings.Contains(path, "/testing/") ||
			strings.Contains(path, "/testutil/") ||
			strings.Contains(path, "/test/") ||
			strings.Contains(path, "/demo/") ||
			strings.Contains(path, "cmd/demo/") ||
			strings.Contains(path, "cmd/identity-demo/") ||
			strings.Contains(path, "lint_go/") {
			return nil
		}

		fileViolations, err := CheckFileForInsecureSkipVerify(path)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return violations, fmt.Errorf("failed to walk directory %s for InsecureSkipVerify check: %w", rootDir, err)
	}

	return violations, nil
}

// CheckFileForInsecureSkipVerify checks a single file for TLS verification disabled.
func CheckFileForInsecureSkipVerify(filePath string) ([]lintGoCommon.CryptoViolation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	defer func() { _ = file.Close() }()

	var violations []lintGoCommon.CryptoViolation

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if insecureSkipVerifyPattern.MatchString(line) {
			// Skip if line contains nolint comment (intentional suppression).
			if strings.Contains(line, "nolint") {
				continue
			}

			violations = append(violations, lintGoCommon.CryptoViolation{
				File:    filePath,
				Line:    lineNum,
				Issue:   "disables TLS certificate verification",
				Content: strings.TrimSpace(line),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return violations, fmt.Errorf("error reading file: %w", err)
	}

	return violations, nil
}
