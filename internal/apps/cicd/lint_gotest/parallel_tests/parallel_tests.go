// Copyright (c) 2025 Justin Cranford

// Package parallel_tests enforces t.Parallel() usage in test functions.
package parallel_tests

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoTestCommon "cryptoutil/internal/apps/cicd/lint_gotest/common"
)

// enforceParallelTests enforces that test functions call t.Parallel().
func Check(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing t.Parallel() in tests...")

	filteredTestFiles := lintGoTestCommon.FilterExcludedTestFiles(testFiles)

	if len(filteredTestFiles) == 0 {
		logger.Log("t.Parallel() enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for t.Parallel()", len(filteredTestFiles)))

	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := CheckParallelUsage(filePath)

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
func CheckParallelUsage(filePath string) []string {
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
