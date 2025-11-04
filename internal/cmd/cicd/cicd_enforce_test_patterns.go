package cicd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// goEnforceTestPatterns enforces test patterns including UUIDv7 usage and testify assertions.
// It checks all test files for proper patterns and reports violations.
func goEnforceTestPatterns(allFiles []string) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] goEnforceTestPatterns started at %s\n", start.Format(time.RFC3339Nano))

	fmt.Fprintln(os.Stderr, "Enforcing test patterns (UUIDv7 usage, testify assertions)...")

	// Find all test files
	var testFiles []string

	for _, path := range allFiles {
		if strings.HasSuffix(path, "_test.go") {
			// Exclude cicd_test.go and cicd.go as they contain deliberate patterns for testing cicd functionality
			if strings.HasSuffix(path, "cicd_test.go") || strings.HasSuffix(path, "cicd.go") {
				continue
			}

			testFiles = append(testFiles, path)
		}
	}

	if len(testFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No test files found")

		end := time.Now()
		fmt.Fprintf(os.Stderr, "[PERF] goEnforceTestPatterns: duration=%v start=%s end=%s (no test files)\n",
			end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))

		return
	}

	fmt.Fprintf(os.Stderr, "Found %d test files to check\n", len(testFiles))

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
		fmt.Fprintf(os.Stderr, "\n❌ Found %d test pattern violations\n", totalIssues)
		fmt.Fprintln(os.Stderr, "Please fix the issues above to follow established test patterns.")
		os.Exit(1) // Fail the build
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All test files follow established patterns")
	}

	end := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] goEnforceTestPatterns: duration=%v start=%s end=%s files=%d issues=%d\n",
		end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), len(testFiles), totalIssues)
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
	errorfPattern := regexp.MustCompile(`(?m)^[\t ]*t\.Errorf\(`)
	if errorfPattern.MatchString(contentStr) {
		matches := errorfPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Errorf() - should use require.Errorf() or assert.Errorf()", len(matches)))
	}

	fatalfPattern := regexp.MustCompile(`(?m)^[\t ]*t\.Fatalf\(`)
	if fatalfPattern.MatchString(contentStr) {
		matches := fatalfPattern.FindAllString(contentStr, -1)
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
