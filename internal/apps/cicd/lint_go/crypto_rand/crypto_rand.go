// Copyright (c) 2025 Justin Cranford

// Package crypto_rand verifies that code uses crypto/rand, not math/rand.
package crypto_rand

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

const (
	importBlockEndMarker = ")"
)

// Check verifies that code uses crypto/rand, not math/rand for security operations.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks for math/rand usage in rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking crypto/rand vs math/rand usage...")

	violations, err := FindMathRandViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check math/rand usage: %w", err)
	}

	if len(violations) > 0 {
		lintGoCommon.PrintCryptoViolations("math/rand", violations)

		return fmt.Errorf("found %d math/rand violations - use crypto/rand for cryptographic operations", len(violations))
	}

	logger.Log("✅ crypto/rand validation passed")

	return nil
}

// mathRandImportPattern matches math/rand import.
var mathRandImportPattern = regexp.MustCompile(`["']math/rand["']`)

// mathRandUsagePatterns matches direct math/rand usage.
var mathRandUsagePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brand\.(Seed|Int|Intn|Int31|Int31n|Int63|Int63n|Uint32|Uint64|Float32|Float64|ExpFloat64|NormFloat64|Perm|Shuffle|Read)\b`),
	regexp.MustCompile(`\brand\.New\b`),
	regexp.MustCompile(`\brand\.NewSource\b`),
}

// FindMathRandViolationsInDir scans Go files in rootDir for math/rand usage.
func FindMathRandViolationsInDir(rootDir string) ([]lintGoCommon.CryptoViolation, error) {
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

		// Skip test files - they may legitimately use math/rand for non-crypto purposes.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip test helper directories - they legitimately may use math/rand for non-crypto purposes.
		if strings.Contains(path, "/testing/") ||
			strings.Contains(path, "/testutil/") ||
			strings.Contains(path, "/test/") ||
			strings.Contains(path, "/demo/") ||
			strings.Contains(path, "cmd/demo/") ||
			strings.Contains(path, "cmd/identity-demo/") ||
			strings.Contains(path, "lint_go/") {
			return nil
		}

		fileViolations, err := CheckFileForMathRand(path)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if err != nil {
		return violations, fmt.Errorf("failed to walk directory %s for math/rand check: %w", rootDir, err)
	}

	return violations, nil
}

// CheckFileForMathRand checks a single file for math/rand usage.
func CheckFileForMathRand(filePath string) ([]lintGoCommon.CryptoViolation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	defer func() { _ = file.Close() }()

	var violations []lintGoCommon.CryptoViolation

	// First pass: collect all lines and check for nolint comments.
	scanner := bufio.NewScanner(file)

	var lines []string

	hasNolintForMathRand := false

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		// Check if file has nolint comment that covers math/rand usage.
		// gosec is the standard linter that catches math/rand issues.
		if strings.Contains(line, "nolint") && (strings.Contains(line, "math/rand") || strings.Contains(line, "gosec")) {
			hasNolintForMathRand = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// If file has explicit nolint for math/rand or gosec, skip entirely.
	if hasNolintForMathRand {
		return nil, nil
	}

	// Second pass: check for violations.
	hasMathRandImport := false
	inImportBlock := false

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Track import blocks.
		if strings.HasPrefix(trimmedLine, "import (") {
			inImportBlock = true

			continue
		}

		if inImportBlock && trimmedLine == importBlockEndMarker {
			inImportBlock = false

			continue
		}

		// Check for math/rand import.
		if inImportBlock || strings.HasPrefix(trimmedLine, "import ") {
			if mathRandImportPattern.MatchString(line) {
				// Check if it's aliased to crand (acceptable for crypto workaround).
				if strings.Contains(line, "crand") {
					continue
				}

				// Skip if line contains nolint comment.
				if strings.Contains(line, "nolint") {
					continue
				}

				hasMathRandImport = true

				violations = append(violations, lintGoCommon.CryptoViolation{
					File:    filePath,
					Line:    lineNum + 1,
					Issue:   "imports math/rand instead of crypto/rand",
					Content: trimmedLine,
				})
			}
		}

		// Only check for usage if we know there's a math/rand import.
		if hasMathRandImport {
			for _, pattern := range mathRandUsagePatterns {
				if pattern.MatchString(line) {
					// Skip if line contains nolint comment.
					if strings.Contains(line, "nolint") {
						continue
					}

					violations = append(violations, lintGoCommon.CryptoViolation{
						File:    filePath,
						Line:    lineNum + 1,
						Issue:   "uses math/rand function",
						Content: trimmedLine,
					})
				}
			}
		}
	}

	return violations, nil
}
