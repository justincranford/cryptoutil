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

const defaultSequentialCommentWindow = 10

// sequentialCommentPattern matches explicit sequential documentation in test functions.
var sequentialCommentPattern = regexp.MustCompile(`//\s*Sequential:`)

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

// testFuncPattern matches test function declarations (captures the function name).
var testFuncPattern = regexp.MustCompile(`func\s+(Test\w+)\s*\(\s*t\s+\*testing\.T\s*\)`)

// parallelPattern matches t.Parallel() calls.
var parallelPattern = regexp.MustCompile(`t\.Parallel\(\)`)

// CheckParallelUsage checks a test file for missing t.Parallel() calls.
// Each Test function is checked individually: t.Parallel() must appear in that
// function's own body section (text between its declaration and the next top-level
// Test function declaration, or EOF).
func CheckParallelUsage(filePath string) []string {
content, err := os.ReadFile(filePath)
if err != nil {
return []string{fmt.Sprintf("Error reading file: %v", err)}
}

contentStr := string(content)

// Find all test function declarations with their positions.
funcMatches := testFuncPattern.FindAllStringSubmatchIndex(contentStr, -1)

if len(funcMatches) == 0 {
return nil
}

var issues []string

for i, match := range funcMatches {
funcName := contentStr[match[2]:match[3]]
funcBodyStart := match[1] // Position right after the function signature.

// Body ends just before the next top-level Test function or at EOF.
var funcBodyEnd int

if i+1 < len(funcMatches) {
funcBodyEnd = funcMatches[i+1][0]
} else {
funcBodyEnd = len(contentStr)
}

funcSection := contentStr[funcBodyStart:funcBodyEnd]

if !parallelPattern.MatchString(funcSection) {
// Skip if function is explicitly documented as sequential.
// Check 10 lines before the function declaration for a "// Sequential:" comment.
funcLineNum := strings.Count(contentStr[:match[0]], "\n")
commentStart := max(0, strings.LastIndex(contentStr[:match[0]], "func "))
// Find the 10-line window before the function
allLines := strings.Split(contentStr[:match[0]], "\n")
lineCount := len(allLines)
windowStart := max(0, lineCount-defaultSequentialCommentWindow)
commentWindow := strings.Join(allLines[windowStart:], "\n")

if sequentialCommentPattern.MatchString(commentWindow) {
continue
}

issues = append(issues, fmt.Sprintf("Test function %s (line %d) is missing t.Parallel()", funcName, funcLineNum+1))
_ = commentStart // suppress unused var warning
}
}

return issues
}
