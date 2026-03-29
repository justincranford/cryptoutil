// Copyright (c) 2025 Justin Cranford

// Package require_over_assert enforces use of testify require over assert.
package require_over_assert

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoTestCommon "cryptoutil/internal/apps/tools/cicd_lint/lint_gotest/common"
)

// enforceRequireOverAssert enforces that tests use testify's require package instead of assert.
// require.* fails immediately on assertion failure (fail-fast pattern).
// assert.* continues execution after failure (accumulates errors).
// For most unit tests, fail-fast with require is preferred.
func Check(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing testify require over assert...")

	filteredTestFiles := lintGoTestCommon.FilterExcludedTestFiles(testFiles)

	if len(filteredTestFiles) == 0 {
		logger.Log("Require over assert enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for assert usage", len(filteredTestFiles)))

	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := CheckAssertUsage(filePath)

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
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Fix: Replace assert.* calls with require.* equivalents for fail-fast testing.")
		fmt.Fprintln(os.Stderr, "  1. Change import: testify/assert → testify/require")
		fmt.Fprintln(os.Stderr, "  2. Change calls: assert.NoError → require.NoError, assert.Equal → require.Equal, etc.")

		return fmt.Errorf("found %d assert usage violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ All test files use require over assert")

	logger.Log("Require over assert enforcement completed")

	return nil
}

// assertPattern matches assert.* calls that should be require.* calls.
var assertPattern = regexp.MustCompile(`\bassert\.(NoError|Error|Nil|NotNil|Equal|NotEqual|True|False|Contains|NotContains|Len|Empty|NotEmpty|Greater|Less|GreaterOrEqual|LessOrEqual|Eventually|Never|ErrorIs|ErrorAs|ErrorContains|Fail|FailNow|Implements|IsType|JSONEq|Panics|PanicsWithValue|PanicsWithError|NoPanic|Regexp|NotRegexp|Same|NotSame|Subset|NotSubset|WithinDuration|WithinRange|Zero|NotZero|FileExists|NoFileExists|DirExists|NoDirExists)\b`)

// CheckAssertUsage checks a test file for assert.* usage that should be require.*.
func CheckAssertUsage(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)

	// Check for assert.* usage with line numbers.
	lines := strings.Split(contentStr, "\n")

	for lineNum, line := range lines {
		if matches := assertPattern.FindAllString(line, -1); len(matches) > 0 {
			for _, match := range matches {
				issues = append(issues, fmt.Sprintf("%s:%d: %s → use require.%s instead", filePath, lineNum+1, match, strings.TrimPrefix(match, "assert.")))
			}
		}
	}

	// Also check for assert import without require import.
	hasAssertImport := strings.Contains(contentStr, `"github.com/stretchr/testify/assert"`)
	hasRequireImport := strings.Contains(contentStr, `"github.com/stretchr/testify/require"`)

	if hasAssertImport && !hasRequireImport {
		issues = append(issues, "imports testify/assert but not testify/require | Fix: replace assert import with require import")
	}

	return issues
}
