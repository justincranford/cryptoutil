// Copyright (c) 2025-2026 Justin Cranford.
// Package parallel_tests enforces t.Parallel() usage in test functions.
package parallel_tests

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
	lintGoTestCommon "cryptoutil/internal/apps-tools/cicd_lint/lint_gotest/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// sequentialCommentPattern matches explicit sequential documentation in test functions.
var sequentialCommentPattern = regexp.MustCompile(`//\s*Sequential:`)

// enforceParallelTests enforces that test functions call t.Parallel().
func CheckFiles(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
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

	logger.LogWithPrefix("parallel-tests", "✅ All test functions use t.Parallel()")

	logger.Log("t.Parallel() enforcement completed")

	return nil
}

// CheckParallelUsage checks a test file for missing t.Parallel() calls.
// Only declarations and calls in Go syntax count; source text embedded in strings does not.
func CheckParallelUsage(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	contentStr := string(content)
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, filePath, content, parser.ParseComments)
	if err != nil {
		return []string{fmt.Sprintf("Error parsing file: %v", err)}
	}

	var issues []string

	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || !strings.HasPrefix(function.Name.Name, "Test") || function.Body == nil {
			continue
		}

		parameterName := testingTParameter(function)
		if parameterName == "" || hasParallelCall(function.Body, parameterName) || hasSequentialComment(contentStr, fileSet, function) {
			continue
		}

		issues = append(issues, fmt.Sprintf("Test function %s (line %d) is missing t.Parallel()", function.Name.Name, fileSet.Position(function.Pos()).Line))
	}

	return issues
}

func testingTParameter(function *ast.FuncDecl) string {
	if function.Type.Params == nil {
		return ""
	}

	for _, field := range function.Type.Params.List {
		pointer, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}

		selector, ok := pointer.X.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "T" {
			continue
		}

		packageName, ok := selector.X.(*ast.Ident)
		if ok && packageName.Name == "testing" && len(field.Names) > 0 {
			return field.Names[0].Name
		}
	}

	return ""
}

func hasParallelCall(body *ast.BlockStmt, parameterName string) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		if _, isNestedFunction := node.(*ast.FuncLit); isNestedFunction {
			return false
		}

		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Parallel" {
			return true
		}

		parameter, ok := selector.X.(*ast.Ident)
		if ok && parameter.Name == parameterName {
			found = true
		}

		return true
	})

	return found
}

func hasSequentialComment(content string, fileSet *token.FileSet, function *ast.FuncDecl) bool {
	functionOffset := fileSet.Position(function.Pos()).Offset
	priorLines := strings.Split(content[:functionOffset], "\n")
	windowStart := max(0, len(priorLines)-cryptoutilSharedMagic.SequentialCommentWindow)

	return sequentialCommentPattern.MatchString(strings.Join(priorLines[windowStart:], "\n"))
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
