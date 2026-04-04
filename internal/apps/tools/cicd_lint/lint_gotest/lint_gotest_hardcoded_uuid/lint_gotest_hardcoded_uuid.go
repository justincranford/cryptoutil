// Copyright (c) 2025 Justin Cranford

// Package lint_gotest_hardcoded_uuid enforces that test files do not use
// uuid.MustParse with hardcoded literal strings. Per ARCHITECTURE.md §10.2,
// tests must use googleUuid.NewV7() for dynamic test data, or googleUuid.UUID{}
// (nil UUID) and googleUuid.UUID{0xff, 0xff, ...} (max UUID) for edge cases.
// Hardcoded literal UUIDs cause test coupling and fragile fixtures.
package lint_gotest_hardcoded_uuid

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoTestCommon "cryptoutil/internal/apps/tools/cicd_lint/lint_gotest/common"
)

// hardcodedUUIDPattern matches uuid.MustParse("...") with any import alias prefix
// (e.g., googleUuid.MustParse("..."), uuid.MustParse("...")).
var hardcodedUUIDPattern = regexp.MustCompile(`\w+\.MustParse\("([0-9a-fA-F-]+)"\)`)

// Check scans test files for hardcoded uuid.MustParse literal string calls.
// Returns an error if any violations are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Checking for hardcoded uuid.MustParse literals in test files...")

	filteredFiles := lintGoTestCommon.FilterExcludedTestFiles(testFiles)

	if len(filteredFiles) == 0 {
		logger.Log("Hardcoded UUID check completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for hardcoded UUID literals", len(filteredFiles)))

	totalIssues := 0

	for _, filePath := range filteredFiles {
		issues := checkHardcodedUUIDs(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d hardcoded UUID literal violations", totalIssues))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Fix: Replace hardcoded uuid.MustParse(\"...\") with dynamic UUID generation.")
		fmt.Fprintln(os.Stderr, "  - Dynamic test data:  googleUuid.NewV7()")
		fmt.Fprintln(os.Stderr, "  - Nil UUID edge case: googleUuid.UUID{}")
		fmt.Fprintln(os.Stderr, "  - Max UUID edge case: googleUuid.UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}")

		return fmt.Errorf("found %d hardcoded UUID literal violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ No hardcoded uuid.MustParse literals found")

	logger.Log("Hardcoded UUID check completed")

	return nil
}

// checkHardcodedUUIDs returns violation messages for each hardcoded MustParse call found.
func checkHardcodedUUIDs(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if hardcodedUUIDPattern.MatchString(line) {
			issues = append(issues, fmt.Sprintf("line %d: hardcoded uuid.MustParse literal: %s", i+1, strings.TrimSpace(line)))
		}
	}

	return issues
}
