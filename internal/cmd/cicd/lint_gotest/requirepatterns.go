// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

// enforceRequireOverAssert enforces that tests use testify's require package instead of assert.
// require.* fails immediately on assertion failure (fail-fast pattern).
// assert.* continues execution after failure (accumulates errors).
// For most unit tests, fail-fast with require is preferred.
func enforceRequireOverAssert(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing testify require over assert...")

	filteredTestFiles := filterExcludedTestFiles(testFiles)

	if len(filteredTestFiles) == 0 {
		logger.Log("Require over assert enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for assert usage", len(filteredTestFiles)))

	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := checkAssertUsage(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d assert usage violations", totalIssues))
		fmt.Fprintln(os.Stderr, "Please use require.* instead of assert.* for fail-fast testing pattern.")

		return fmt.Errorf("found %d assert usage violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ All test files use require over assert")

	logger.Log("Require over assert enforcement completed")

	return nil
}

// assertPattern matches assert.* calls that should be require.* calls.
var assertPattern = regexp.MustCompile(`\bassert\.(NoError|Error|Nil|NotNil|Equal|NotEqual|True|False|Contains|NotContains|Len|Empty|NotEmpty|Greater|Less|GreaterOrEqual|LessOrEqual|Eventually|Never|ErrorIs|ErrorAs|ErrorContains|Fail|FailNow|Implements|IsType|JSONEq|Panics|PanicsWithValue|PanicsWithError|NoPanic|Regexp|NotRegexp|Same|NotSame|Subset|NotSubset|WithinDuration|WithinRange|Zero|NotZero|FileExists|NoFileExists|DirExists|NoDirExists)\b`)

// checkAssertUsage checks a test file for assert.* usage that should be require.*.
func checkAssertUsage(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)

	// Check for assert.* usage.
	if assertPattern.MatchString(contentStr) {
		matches := assertPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of assert.* - should use require.* for fail-fast testing", len(matches)))
	}

	// Also check for assert import without require import.
	hasAssertImport := strings.Contains(contentStr, `"github.com/stretchr/testify/assert"`)
	hasRequireImport := strings.Contains(contentStr, `"github.com/stretchr/testify/require"`)

	if hasAssertImport && !hasRequireImport {
		issues = append(issues, "Test file imports testify/assert but not testify/require - prefer require for fail-fast testing")
	}

	return issues
}

// enforceParallelTests enforces that test functions call t.Parallel().
func enforceParallelTests(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing t.Parallel() in tests...")

	filteredTestFiles := filterExcludedTestFiles(testFiles)

	if len(filteredTestFiles) == 0 {
		logger.Log("t.Parallel() enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for t.Parallel()", len(filteredTestFiles)))

	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := checkParallelUsage(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d missing t.Parallel() violations", totalIssues))
		fmt.Fprintln(os.Stderr, "Please add t.Parallel() to test functions for concurrent execution.")

		return fmt.Errorf("found %d missing t.Parallel() violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ All test functions use t.Parallel()")

	logger.Log("t.Parallel() enforcement completed")

	return nil
}

// testFuncPattern matches test function declarations.
var testFuncPattern = regexp.MustCompile(`func\s+(Test\w+)\s*\(\s*t\s+\*testing\.T\s*\)`)

// parallelPattern matches t.Parallel() calls.
var parallelPattern = regexp.MustCompile(`t\.Parallel\(\)`)

// checkParallelUsage checks a test file for missing t.Parallel() calls.
func checkParallelUsage(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)

	// Find all test function declarations.
	testFuncs := testFuncPattern.FindAllStringSubmatch(contentStr, -1)

	// Check if t.Parallel() is called at least once.
	hasParallel := parallelPattern.MatchString(contentStr)

	// If there are test functions but no t.Parallel(), flag it.
	if len(testFuncs) > 0 && !hasParallel {
		funcNames := make([]string, 0, len(testFuncs))
		for _, match := range testFuncs {
			if len(match) > 1 {
				funcNames = append(funcNames, match[1])
			}
		}

		issues = append(issues, fmt.Sprintf("No t.Parallel() found in file with %d test functions: %s", len(testFuncs), strings.Join(funcNames, ", ")))
	}

	return issues
}

// enforceHardcodedPasswords enforces that tests don't contain hardcoded passwords.
func enforceHardcodedPasswords(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing no hardcoded passwords in tests...")

	filteredTestFiles := filterExcludedTestFiles(testFiles)

	if len(filteredTestFiles) == 0 {
		logger.Log("Hardcoded password enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for hardcoded passwords", len(filteredTestFiles)))

	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := checkHardcodedPasswords(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d hardcoded password violations", totalIssues))
		fmt.Fprintln(os.Stderr, "Please use dynamic passwords: password := googleUuid.NewV7().String()")

		return fmt.Errorf("found %d hardcoded password violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ No hardcoded passwords found in tests")

	logger.Log("Hardcoded password enforcement completed")

	return nil
}

// hardcodedPasswordPatterns matches common hardcoded password patterns.
var hardcodedPasswordPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)password\s*[:=]+\s*["'](?:test123|password|secret|admin|12345|qwerty|letmein|welcome|passw0rd)["']`),
	regexp.MustCompile(`(?i)secret\s*[:=]+\s*["'](?:test|secret|admin|12345|mysecret)["']`),
	regexp.MustCompile(`(?i)apiKey\s*[:=]+\s*["'](?:test|secret|12345|sk-test)["']`),
}

// checkHardcodedPasswords checks a test file for hardcoded passwords.
func checkHardcodedPasswords(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)

	for _, pattern := range hardcodedPasswordPatterns {
		if pattern.MatchString(contentStr) {
			matches := pattern.FindAllString(contentStr, -1)
			issues = append(issues, fmt.Sprintf("Found %d instances of hardcoded passwords - use uuid.NewV7().String() instead", len(matches)))
		}
	}

	return issues
}

// filterExcludedTestFiles filters out test files that should be excluded from linting.
func filterExcludedTestFiles(testFiles []string) []string {
	filteredTestFiles := make([]string, 0, len(testFiles))

	for _, path := range testFiles {
		// Exclude cicd test files as they contain deliberate patterns for testing cicd functionality.
		// Also exclude edge_cases_test.go, testmain_test.go, e2e_test.go, sessions_test.go, admin_test.go.
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

	return filteredTestFiles
}
