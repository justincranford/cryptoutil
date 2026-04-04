// Copyright (c) 2025 Justin Cranford

// Package lint_pythontest provides linting for Python test files in CI/CD pipelines.
// Sub-linters validate Python test files for cryptoutil project standards.
package lint_pythontest

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// LinterFunc is a function type for individual Python test file linters.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, pythonFiles []string) error

// registeredLinters holds all linters to run as part of lint-python-test.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"unittest-antipattern", checkUnittestAntipattern},
}

// Lint runs all registered Python test file linters.
// It filters the provided files to only include test_*.py and *_test.py files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running Python test linters...")

	pythonFiles := filterPythonTestFiles(filesByExtension["py"])

	if len(pythonFiles) == 0 {
		logger.Log("lint-python-test completed (no Python test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d Python test files to lint", len(pythonFiles)))

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, pythonFiles); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-python-test completed with %d errors", len(errors)))

		msgs := make([]string, len(errors))
		for i, e := range errors {
			msgs[i] = e.Error()
		}

		return fmt.Errorf("lint-python-test failed: %s", strings.Join(msgs, "; "))
	}

	logger.Log("lint-python-test completed successfully")

	return nil
}

// filterPythonTestFiles returns only test_*.py and *_test.py files.
func filterPythonTestFiles(pyFiles []string) []string {
	var testFiles []string

	for _, f := range pyFiles {
		base := f

		// Use last path component for matching.
		if idx := strings.LastIndexAny(f, "/\\"); idx >= 0 {
			base = f[idx+1:]
		}

		if strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") {
			testFiles = append(testFiles, f)
		}
	}

	return testFiles
}

// unittestPattern matches unittest.TestCase inheritance and self.assert* method calls.
var unittestPattern = regexp.MustCompile(`class\s+\w+\s*\(\s*unittest\.TestCase\s*\)|from\s+unittest\s+import\s+TestCase|self\.(assert\w+)\s*\(`)

// checkUnittestAntipattern checks Python test files for unittest-based testing patterns.
// cryptoutil uses pytest style: use @pytest.mark.parametrize, not unittest.TestCase.
func checkUnittestAntipattern(logger *cryptoutilCmdCicdCommon.Logger, pythonFiles []string) error {
	logger.Log("Checking for unittest antipatterns...")

	totalIssues := 0

	for _, filePath := range pythonFiles {
		issues := CheckUnittestAntipattern(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d unittest antipattern violations", totalIssues))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Fix: Use pytest style instead of unittest.TestCase:")
		fmt.Fprintln(os.Stderr, "  1. Replace 'class MyTest(unittest.TestCase)' with standalone test functions")
		fmt.Fprintln(os.Stderr, "  2. Replace 'self.assertEqual(a, b)' with 'assert a == b'")
		fmt.Fprintln(os.Stderr, "  3. Use '@pytest.mark.parametrize' for parameterized tests")

		return fmt.Errorf("found %d unittest antipattern violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ All Python test files use pytest style (no unittest.TestCase)")

	logger.Log("Unittest antipattern check completed")

	return nil
}

// CheckUnittestAntipattern checks a Python test file for unittest-based testing antipatterns.
func CheckUnittestAntipattern(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		matches := unittestPattern.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			if strings.Contains(match[0], "unittest.TestCase") {
				issues = append(issues, fmt.Sprintf("%s:%d: unittest.TestCase inheritance → use pytest standalone functions", filePath, lineNum+1))
			} else if strings.Contains(match[0], "from unittest import TestCase") {
				issues = append(issues, fmt.Sprintf("%s:%d: unittest.TestCase import → use pytest instead", filePath, lineNum+1))
			} else if len(match) > 1 && match[1] != "" {
				issues = append(issues, fmt.Sprintf("%s:%d: self.%s() method → use pytest assert patterns", filePath, lineNum+1, match[1]))
			}
		}
	}

	return issues
}
