// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

// cryptoViolation represents a security violation in crypto usage.
type cryptoViolation struct {
	File    string
	Line    int
	Issue   string
	Content string
}

// checkCryptoRand verifies that code uses crypto/rand, not math/rand for security operations.
func checkCryptoRand(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking crypto/rand vs math/rand usage...")

	violations, err := findMathRandViolations()
	if err != nil {
		return fmt.Errorf("failed to check math/rand usage: %w", err)
	}

	if len(violations) > 0 {
		printCryptoViolations("math/rand", violations)

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

// findMathRandViolations scans Go files for math/rand usage.
func findMathRandViolations() ([]cryptoViolation, error) {
	var violations []cryptoViolation

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories.
		if d.IsDir() {
			if d.Name() == excludeDirVendor || d.Name() == excludeDirGit {
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

		fileViolations, err := checkFileForMathRand(path)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})

	return violations, err
}

// checkFileForMathRand checks a single file for math/rand usage.
func checkFileForMathRand(filePath string) ([]cryptoViolation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var violations []cryptoViolation

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
		return nil, err
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

		if inImportBlock && trimmedLine == ")" {
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

				violations = append(violations, cryptoViolation{
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

					violations = append(violations, cryptoViolation{
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

// checkInsecureSkipVerify verifies that code doesn't disable TLS certificate verification.
func checkInsecureSkipVerify(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking for InsecureSkipVerify usage...")

	violations, err := findInsecureSkipVerifyViolations()
	if err != nil {
		return fmt.Errorf("failed to check InsecureSkipVerify: %w", err)
	}

	if len(violations) > 0 {
		printCryptoViolations("InsecureSkipVerify", violations)

		return fmt.Errorf("found %d InsecureSkipVerify violations - never disable TLS certificate verification", len(violations))
	}

	logger.Log("✅ InsecureSkipVerify validation passed")

	return nil
}

// insecureSkipVerifyPattern matches InsecureSkipVerify set to true in Go code.
var insecureSkipVerifyPattern = regexp.MustCompile(`InsecureSkipVerify\s*:\s*true`)

// findInsecureSkipVerifyViolations scans Go files for disabling TLS verification.
func findInsecureSkipVerifyViolations() ([]cryptoViolation, error) {
	var violations []cryptoViolation

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories.
		if d.IsDir() {
			if d.Name() == excludeDirVendor || d.Name() == excludeDirGit {
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

		fileViolations, err := checkFileForInsecureSkipVerify(path)
		if err != nil {
			return err
		}

		violations = append(violations, fileViolations...)

		return nil
	})

	return violations, err
}

// checkFileForInsecureSkipVerify checks a single file for TLS verification disabled.
func checkFileForInsecureSkipVerify(filePath string) ([]cryptoViolation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var violations []cryptoViolation

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

			violations = append(violations, cryptoViolation{
				File:    filePath,
				Line:    lineNum,
				Issue:   "disables TLS certificate verification",
				Content: strings.TrimSpace(line),
			})
		}
	}

	return violations, scanner.Err()
}

// printCryptoViolations prints crypto-related violations to stderr.
func printCryptoViolations(category string, violations []cryptoViolation) {
	fmt.Fprintf(os.Stderr, "\n❌ Found %d %s violations:\n", len(violations), category)

	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "  %s:%d: %s\n", v.File, v.Line, v.Issue)
		fmt.Fprintf(os.Stderr, "    > %s\n", v.Content)
	}

	fmt.Fprintln(os.Stderr)
}
