// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// enforceTestPatterns enforces test patterns including UUIDv7 usage and testify assertions.
// It checks all test files for proper patterns and returns an error if violations are found.
func enforceTestPatterns(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing Go test patterns...")

	// Filter out cicd test files to prevent self-modification.
	filteredTestFiles := make([]string, 0, len(testFiles))

	for _, path := range testFiles {
		// Exclude cicd test files as they contain deliberate patterns for testing cicd functionality.
		// Also exclude edge_cases_test.go (may need hardcoded UUIDs), testmain_test.go (legitimate t.Fatalf()),
		// e2e_test.go (may have placeholders), sessions_test.go (may have test data UUIDs),
		// and admin_test.go (may have timeout t.Fatalf() in setup).
		if strings.HasSuffix(path, "cicd_test.go") ||
			strings.HasSuffix(path, "cicd.go") ||
			strings.HasSuffix(path, "cicd_enforce_test_patterns_test.go") ||
			strings.HasSuffix(path, "cicd_enforce_test_patterns_integration_test.go") ||
			strings.HasSuffix(path, "cicd_run_integration_test.go") ||
			strings.Contains(path, "lint_gotest") ||
			strings.HasSuffix(path, "_edge_cases_test.go") ||
			strings.HasSuffix(path, "testmain_test.go") ||
			strings.HasSuffix(path, "e2e_test.go") ||
			strings.HasSuffix(path, "sessions_test.go") ||
			strings.HasSuffix(path, "admin_test.go") {
			continue
		}

		filteredTestFiles = append(filteredTestFiles, path)
	}

	if len(filteredTestFiles) == 0 {
		logger.Log("Test pattern enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check", len(filteredTestFiles)))

	// Check each test file.
	totalIssues := 0

	for _, filePath := range filteredTestFiles {
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

		return fmt.Errorf("found %d test pattern violations across %d files", totalIssues, len(filteredTestFiles))
	}

	fmt.Fprintln(os.Stderr, "\n✅ All test files follow established patterns")

	logger.Log("Test pattern enforcement completed")

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

	// Pattern 1: Check for UUIDv7 usage.
	// Look for uuid.New() instead of uuid.NewV7().
	if strings.Contains(contentStr, "uuid.New()") {
		issues = append(issues, "Found uuid.New() - should use uuid.NewV7() for concurrency safety")
	}

	// Pattern 2: Check for hardcoded UUIDs (basic pattern).
	uuidPattern := regexp.MustCompile(cryptoutilSharedMagic.StringUUIDRegexPattern)
	if uuidPattern.MatchString(contentStr) {
		issues = append(issues, "Found hardcoded UUID - consider using uuid.NewV7() for test data")
	}

	// Pattern 3: Check for testify usage patterns.
	// Look for t.Errorf/t.Fatalf that should use require/assert.
	// Use a more sophisticated pattern to avoid matching string literals.
	if cryptoutilSharedMagic.TestErrorfPattern.MatchString(contentStr) {
		matches := cryptoutilSharedMagic.TestErrorfPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Errorf() - should use require.Errorf() or assert.Errorf()", len(matches)))
	}

	if cryptoutilSharedMagic.TestFatalfPattern.MatchString(contentStr) {
		matches := cryptoutilSharedMagic.TestFatalfPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Fatalf() - should use require.Fatalf() or assert.Fatalf()", len(matches)))
	}

	// Pattern 4: Check for testify imports if testify assertions are used.
	hasTestifyUsage := strings.Contains(contentStr, "require.") || strings.Contains(contentStr, "assert.")
	hasTestifyImport := strings.Contains(contentStr, "github.com/stretchr/testify")

	if hasTestifyUsage && !hasTestifyImport {
		issues = append(issues, "Test file uses testify assertions but doesn't import testify")
	}

	return issues
}
