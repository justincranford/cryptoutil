// Copyright (c) 2025-2026 Justin Cranford.
package test_patterns

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// enforceTestPatterns enforces test patterns including UUIDv7 usage and testify assertions.
// It checks all test files for proper patterns and returns an error if violations are found.
func CheckFiles(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Enforcing Go test patterns...")

	// Check all test files for banned filename patterns first.
	bannedNameIssues := checkBannedFilenames(testFiles)

	// Filter out cicd test files to prevent self-modification.
	filteredTestFiles := make([]string, 0, len(testFiles))

	for _, path := range testFiles {
		// Exclude cicd test files as they contain deliberate patterns for testing cicd functionality.
		// Also exclude edge_cases_test.go (may need hardcoded UUIDs), testmain_test.go (legitimate t.Fatalf()),
		// e2e_test.go (may have placeholders), sessions_test.go (may have test data UUIDs),
		// and admin_test.go (may have timeout t.Fatalf() in setup).
		if strings.HasSuffix(path, "cicd_test.go") ||
			strings.HasSuffix(path, "cicd.go") ||
			strings.HasSuffix(path, "cicd_enforce_test_patterns_test.go") ||
			strings.HasSuffix(path, "cicd_enforce_test_patterns_integration_test.go") ||
			strings.HasSuffix(path, "cicd_run_integration_test.go") ||
			strings.Contains(path, "lint_gotest") ||
			strings.Contains(path, "lint_fitness") ||
			strings.HasSuffix(path, "testmain_test.go") ||
			strings.HasSuffix(path, "e2e_test.go") ||
			strings.HasSuffix(path, "sessions_test.go") ||
			strings.HasSuffix(path, "admin_test.go") {
			continue
		}

		filteredTestFiles = append(filteredTestFiles, path)
	}

	if len(filteredTestFiles) == 0 {
		logger.Log("Test pattern enforcement completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check", len(filteredTestFiles)))

	// Check each test file.
	totalIssues := 0

	for _, filePath := range filteredTestFiles {
		issues := CheckTestFile(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d test pattern violations", totalIssues))
		fmt.Fprintln(os.Stderr, "Please fix the issues above to follow established test patterns.")

		return fmt.Errorf("found %d test pattern violations across %d files", totalIssues, len(filteredTestFiles))
	}

	if len(bannedNameIssues) > 0 {
		for _, issue := range bannedNameIssues {
			fmt.Fprintf(os.Stderr, "  - %s\n", issue)
		}

		logger.Log(fmt.Sprintf("Found %d banned test filename violations", len(bannedNameIssues)))

		return fmt.Errorf("found %d banned test filename violations", len(bannedNameIssues))
	}

	logger.LogWithPrefix("test-patterns", "✅ All test files follow established patterns")

	logger.Log("Test pattern enforcement completed")

	return nil
}

// checkTestFile checks a single test file for proper test patterns.
// It returns a slice of issues found, empty if the file follows all patterns.
func CheckTestFile(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)

	// Pattern 1: Check for UUIDv7 usage.
	// Look for uuid.New() instead of uuid.NewV7().
	if strings.Contains(contentStr, "uuid.New()") {
		issues = append(issues, "Found uuid.New() - should use uuid.NewV7() for concurrency safety")
	}

	// Pattern 2: Check for hardcoded UUIDs (basic pattern).
	uuidPattern := regexp.MustCompile(cryptoutilSharedMagic.StringUUIDRegexPattern)
	if uuidPattern.MatchString(contentStr) {
		issues = append(issues, "Found hardcoded UUID - consider using uuid.NewV7() for test data")
	}

	// Pattern 3: Check for testing.T failure calls in Go syntax, not fixture strings.
	errorfCount, fatalfCount := countTestingFailureCalls(content, filePath)
	if errorfCount > 0 {
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Errorf() - use assert.Failf() to preserve non-fatal behavior", errorfCount))
	}

	if fatalfCount > 0 {
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Fatalf() - use require.FailNowf() to preserve fatal behavior", fatalfCount))
	}

	// Pattern 4: Check for testify imports if testify assertions are used.
	hasTestifyUsage := strings.Contains(contentStr, "require.") || strings.Contains(contentStr, "assert.")
	hasTestifyImport := strings.Contains(contentStr, "github.com/stretchr/testify")

	if hasTestifyUsage && !hasTestifyImport {
		issues = append(issues, "Test file uses testify assertions but doesn't import testify")
	}

	return issues
}

func countTestingFailureCalls(content []byte, filePath string) (int, int) {
	file, err := parser.ParseFile(token.NewFileSet(), filePath, content, 0)
	if err != nil {
		return 0, 0
	}

	var errorfCount, fatalfCount int

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}

		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Name != "t" {
			return true
		}

		switch selector.Sel.Name {
		case cryptoutilSharedMagic.CICDTestMethodErrorf:
			errorfCount++
		case cryptoutilSharedMagic.CICDTestMethodFatalf:
			fatalfCount++
		}

		return true
	})

	return errorfCount, fatalfCount
}

// bannedTestFilenameSuffixes lists nonsense test filename suffixes that describe
// coverage intent rather than test content.
var bannedTestFilenameSuffixes = []string{
	"_coverage_test.go",
	"_coverage2_test.go",
	"_comprehensive_test.go",
	"_gaps_test.go",
	"_coverage_gaps_test.go",
	"_highcov_test.go",
	"_extra_test.go",
	"_additional_test.go",
	"_edge_cases_test.go",
}

// checkBannedFilenames checks for test files with nonsense filenames.
// Files where the filename matches the package directory name are exempt.
func checkBannedFilenames(testFiles []string) []string {
	var issues []string

	for _, filePath := range testFiles {
		fileName := filepath.Base(filePath)
		dirName := filepath.Base(filepath.Dir(filePath))

		// Exempt: filename matches package directory name (e.g., propagation_coverage/propagation_coverage_test.go).
		fileBase := strings.TrimSuffix(fileName, "_test.go")
		if fileBase == dirName {
			continue
		}

		for _, suffix := range bannedTestFilenameSuffixes {
			if strings.HasSuffix(fileName, suffix) {
				issues = append(issues, fmt.Sprintf("Banned test filename pattern %q: %s (rename to describe test content, not coverage intent)", suffix, filePath))

				break
			}
		}
	}

	return issues
}

// Check runs the linter by discovering all _test.go files in the repository.
// Returns an error if any violations are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	var testFiles []string

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if d.Name() == cryptoutilSharedMagic.CICDExcludeDirVendor || d.Name() == cryptoutilSharedMagic.CICDExcludeDirGit {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking test files: %w", err)
	}

	return CheckFiles(logger, testFiles)
}
