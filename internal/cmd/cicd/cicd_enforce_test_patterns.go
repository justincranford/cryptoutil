// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	"cryptoutil/internal/cmd/cicd/common"
)

// goEnforceTestPatterns enforces test patterns including UUIDv7 usage and testify assertions.
// It checks all test files for proper patterns and returns an error if violations are found.
func goEnforceTestPatterns(logger *common.Logger, allFiles []string) error {
	logger.Log("Enforcing Go test patterns...")

	// Find all test files
	var testFiles []string

	for _, path := range allFiles {
		if strings.HasSuffix(path, "_test.go") {
			// Exclude cicd_test.go and cicd.go as they contain deliberate patterns for testing cicd functionality
			if strings.HasSuffix(path, "cicd_test.go") ||
				strings.HasSuffix(path, "cicd.go") ||
				strings.HasSuffix(path, "cicd_enforce_test_patterns_test.go") ||
				strings.HasSuffix(path, "cicd_enforce_test_patterns_integration_test.go") ||
				strings.HasSuffix(path, "cicd_run_integration_test.go") {
				continue
			}

			testFiles = append(testFiles, path)
		}
	}

	if len(testFiles) == 0 {
		logger.Log("goEnforceTestPatterns completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check", len(testFiles)))

	// Check each test file
	totalIssues := 0

	for _, filePath := range testFiles {
		issues := checkTestFile(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d test pattern violations", totalIssues))
		fmt.Fprintln(os.Stderr, "Please fix the issues above to follow established test patterns.")

		return fmt.Errorf("found %d test pattern violations across %d files", totalIssues, len(testFiles))
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All test files follow established patterns")
	}

	logger.Log("goEnforceTestPatterns completed")

	return nil
}

// checkTestFile checks a single test file for proper test patterns.
// It returns a slice of issues found, empty if the file follows all patterns.
func checkTestFile(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)

	// Pattern 1: Check for UUIDv7 usage
	// Look for uuid.New() instead of uuid.NewV7()
	if strings.Contains(contentStr, "uuid.New()") {
		issues = append(issues, "Found uuid.New() - should use uuid.NewV7() for concurrency safety")
	}

	// Pattern 2: Check for hardcoded UUIDs (basic pattern)
	uuidPattern := regexp.MustCompile(cryptoutilMagic.StringUUIDRegexPattern)
	if uuidPattern.MatchString(contentStr) {
		issues = append(issues, "Found hardcoded UUID - consider using uuid.NewV7() for test data")
	}

	// Pattern 3: Check for testify usage patterns
	// Look for t.Errorf/t.Fatalf that should use require/assert
	// Use a more sophisticated pattern to avoid matching string literals
	if cryptoutilMagic.TestErrorfPattern.MatchString(contentStr) {
		matches := cryptoutilMagic.TestErrorfPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Errorf() - should use require.Errorf() or assert.Errorf()", len(matches)))
	}

	if cryptoutilMagic.TestFatalfPattern.MatchString(contentStr) {
		matches := cryptoutilMagic.TestFatalfPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Fatalf() - should use require.Fatalf() or assert.Fatalf()", len(matches)))
	}

	// Pattern 4: Check for testify imports if testify assertions are used
	hasTestifyUsage := strings.Contains(contentStr, "require.") || strings.Contains(contentStr, "assert.")
	hasTestifyImport := strings.Contains(contentStr, "github.com/stretchr/testify")

	if hasTestifyUsage && !hasTestifyImport {
		issues = append(issues, "Test file uses testify assertions but doesn't import testify")
	}

	return issues
}
