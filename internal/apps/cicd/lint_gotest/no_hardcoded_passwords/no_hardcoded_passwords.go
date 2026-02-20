// Copyright (c) 2025 Justin Cranford

// Package no_hardcoded_passwords enforces that tests do not contain hardcoded passwords.
package no_hardcoded_passwords

import (
	"fmt"
	"os"
	"regexp"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoTestCommon "cryptoutil/internal/apps/cicd/lint_gotest/common"
)

// enforceHardcodedPasswords enforces that tests don't contain hardcoded passwords.
func Check(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing no hardcoded passwords in tests...")

	filteredTestFiles := lintGoTestCommon.FilterExcludedTestFiles(testFiles)

	if len(filteredTestFiles) == 0 {
		logger.Log("Hardcoded password enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for hardcoded passwords", len(filteredTestFiles)))

	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := CheckHardcodedPasswords(filePath)

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
func CheckHardcodedPasswords(filePath string) []string {
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
