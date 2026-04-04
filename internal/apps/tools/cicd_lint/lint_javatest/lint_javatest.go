// Copyright (c) 2025 Justin Cranford

// Package lint_javatest provides linting for Java test files in CI/CD pipelines.
// Sub-linters validate Gatling simulation files for cryptoutil project standards.
package lint_javatest

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// LinterFunc is a function type for individual Java test file linters.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, javaFiles []string) error

// registeredLinters holds all linters to run as part of lint-java-test.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"insecure-random", checkInsecureRandom},
}

// Lint runs all registered Java test file linters.
// It uses all *.java files from the provided file map.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running Java test linters...")

	javaFiles := filesByExtension["java"]

	if len(javaFiles) == 0 {
		logger.Log("lint-java-test completed (no Java files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d Java files to lint", len(javaFiles)))

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, javaFiles); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-java-test completed with %d errors", len(errors)))

		msgs := make([]string, len(errors))
		for i, e := range errors {
			msgs[i] = e.Error()
		}

		return fmt.Errorf("lint-java-test failed: %s", strings.Join(msgs, "; "))
	}

	logger.Log("lint-java-test completed successfully")

	return nil
}

// insecureRandomPattern matches insecure RNG usage in Java files.
var insecureRandomPattern = regexp.MustCompile(`\bnew\s+Random\s*\(|Math\.random\s*\(`)

// checkInsecureRandom checks Java files for insecure random number generation.
// FIPS 140-3 compliance requires SecureRandom, not java.util.Random or Math.random().
func checkInsecureRandom(logger *cryptoutilCmdCicdCommon.Logger, javaFiles []string) error {
	logger.Log("Checking for insecure random number generation...")

	totalIssues := 0

	for _, filePath := range javaFiles {
		issues := CheckInsecureRandom(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d insecure RNG violations", totalIssues))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Fix: Replace insecure RNG with SecureRandom for FIPS 140-3 compliance:")
		fmt.Fprintln(os.Stderr, "  1. Replace 'new Random()' with 'new SecureRandom()'")
		fmt.Fprintln(os.Stderr, "  2. Replace 'Math.random()' with 'secureRandom.nextDouble()'")
		fmt.Fprintln(os.Stderr, "  3. Add import: 'import java.security.SecureRandom;'")

		return fmt.Errorf("found %d insecure RNG violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ All Java files use SecureRandom for random number generation")

	logger.Log("Insecure random check completed")

	return nil
}

// CheckInsecureRandom checks a Java file for insecure random number generation.
func CheckInsecureRandom(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		if insecureRandomPattern.MatchString(line) {
			issues = append(issues, fmt.Sprintf("%s:%d: insecure RNG → use new SecureRandom() instead", filePath, lineNum+1))
		}
	}

	return issues
}
