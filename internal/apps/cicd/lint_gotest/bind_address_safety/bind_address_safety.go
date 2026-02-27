// Copyright (c) 2025 Justin Cranford

// Package lint_gotest provides linting utilities for Go test files.
package bind_address_safety

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// enforceBindAddressSafety enforces proper bind address usage in test files.
// Detects 0.0.0.0 usage which triggers Windows Firewall prompts.
// Returns an error if violations are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing bind address safety in test files...")

	// Filter out cicd test files and url_test.go (legitimate use cases).
	filteredTestFiles := make([]string, 0, len(testFiles))

	for _, path := range testFiles {
		// Exclude cicd test files, url_test.go (legitimate URL parsing tests),
		// validation_test.go files (test validation of 0.0.0.0 rejection),
		// and config_test.go files (test configuration validation).
		if strings.HasSuffix(path, "cicd_test.go") ||
			strings.HasSuffix(path, "cicd.go") ||
			strings.Contains(path, "lint_gotest") ||
			strings.HasSuffix(path, "url_test.go") ||
			strings.HasSuffix(path, "_validation_test.go") ||
				strings.HasSuffix(path, "config_test.go") ||
				strings.HasSuffix(path, "config_validate_test.go") {
			continue
		}

		filteredTestFiles = append(filteredTestFiles, path)
	}

	if len(filteredTestFiles) == 0 {
		logger.Log("Bind address safety check completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check", len(filteredTestFiles)))

	// Check each test file.
	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := CheckBindAddressSafety(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d bind address violations", totalIssues))
		fmt.Fprintln(os.Stderr, "\nWhy this matters:")
		fmt.Fprintln(os.Stderr, "  - Binding to 0.0.0.0 in tests triggers Windows Firewall exception prompts")
		fmt.Fprintln(os.Stderr, "  - This blocks CI/CD automation and disrupts developer workflow")
		fmt.Fprintln(os.Stderr, "  - Tests should bind to 127.0.0.1 (loopback only)")
		fmt.Fprintln(os.Stderr, "  - Use cryptoutilConfig.NewTestConfig(\"127.0.0.1\", 0, true) for safe test configs")
		fmt.Fprintln(os.Stderr, "\nPlease fix the issues above to prevent Windows Firewall prompts.")

		return fmt.Errorf("found %d bind address violations across %d files", totalIssues, len(filteredTestFiles))
	}

	fmt.Fprintln(os.Stderr, "\n✅ All test files use safe bind addresses (127.0.0.1)")

	logger.Log("Bind address safety check completed")

	return nil
}

// checkBindAddressSafety checks a single test file for bind address violations.
// Returns a slice of issues found, empty if the file is safe.
func CheckBindAddressSafety(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Pattern 1: Direct "0.0.0.0" string usage.
	if strings.Contains(contentStr, "\"0.0.0.0\"") {
		// Find line numbers for better diagnostics.
		for i, line := range lines {
			if strings.Contains(line, "\"0.0.0.0\"") {
				issues = append(issues, fmt.Sprintf("Line %d: Found \"0.0.0.0\" - should use \"127.0.0.1\" in tests", i+1))
			}
		}
	}

	// Pattern 2: Blank BindPublicAddress or BindPrivateAddress (defaults to 0.0.0.0).
	blankBindPattern := regexp.MustCompile(`Bind(Public|Private)Address:\s*""`)
	if blankBindPattern.MatchString(contentStr) {
		for i, line := range lines {
			if blankBindPattern.MatchString(line) {
				issues = append(issues, fmt.Sprintf("Line %d: Blank bind address defaults to 0.0.0.0 - explicitly set to \"127.0.0.1\"", i+1))
			}
		}
	}

	// Pattern 3: Direct struct instantiation without NewTestConfig.
	structPattern := regexp.MustCompile(`&cryptoutilConfig\.ServiceTemplateServerSettings\{`)
	if structPattern.MatchString(contentStr) {
		// Check if NewTestConfig is also present (safe pattern).
		hasNewTestConfig := strings.Contains(contentStr, "NewTestConfig")

		if !hasNewTestConfig {
			for i, line := range lines {
				if structPattern.MatchString(line) {
					issues = append(issues, fmt.Sprintf("Line %d: Direct ServiceTemplateServerSettings{} creation - use NewTestConfig() for safe defaults", i+1))
				}
			}
		}
	}

	// Pattern 4: net.Listen with empty or 0.0.0.0 address.
	// Match: net.Listen("tcp", ":0") or net.Listen("tcp4", ":0").
	netListenPattern := regexp.MustCompile(`net\.Listen\s*\(\s*"[^"]*",\s*":`)
	if netListenPattern.MatchString(contentStr) {
		for i, line := range lines {
			if netListenPattern.MatchString(line) {
				issues = append(issues, fmt.Sprintf("Line %d: net.Listen with \":0\" or \":port\" binds to 0.0.0.0 - use \"127.0.0.1:0\"", i+1))
			}
		}
	}

	return issues
}
