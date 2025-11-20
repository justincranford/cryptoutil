package cicd

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

// goFixTHelper adds t.Helper() to test helper functions that are missing it.
// Test helper functions are identified by naming patterns (setup*, check*, assert*, verify*, helper*).
func goFixTHelper(logger *LogUtil, files []string) error {
	logger.Log("Starting thelper auto-fix")

	goFiles := filterGoFiles(files)
	// Only process test files
	testFiles := filterTestFiles(goFiles)
	logger.Log(fmt.Sprintf("Processing %d test files", len(testFiles)))

	totalFixCount := 0
	for _, file := range testFiles {
		fixCount, err := fixTHelperInFile(file)
		if err != nil {
			return fmt.Errorf("failed to fix thelper in %s: %w", file, err)
		}
		totalFixCount += fixCount
	}

	if totalFixCount > 0 {
		logger.Log(fmt.Sprintf("Added t.Helper() to %d test helper functions", totalFixCount))
		return fmt.Errorf("added t.Helper() to %d test helper functions - please review changes", totalFixCount)
	}

	logger.Log("No test helper functions needed t.Helper()")
	return nil
}

// filterTestFiles returns only _test.go files from the given list.
func filterTestFiles(files []string) []string {
	testFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			testFiles = append(testFiles, file)
		}
	}
	return testFiles
}

// fixTHelperInFile adds t.Helper() to test helper functions missing it.
func fixTHelperInFile(filePath string) (int, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file: %w", err)
	}

	fixCount := 0
	modified := false

	// Find all function declarations
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if this is a test helper function
		if !isTestHelperFunction(funcDecl) {
			continue
		}

		// Check if function already has t.Helper() call
		if hasTHelperCall(funcDecl) {
			continue
		}

		// Add t.Helper() as first statement
		if addTHelperCall(funcDecl) {
			fixCount++
			modified = true
		}
	}

	// Write back to file if modified
	if modified {
		// Verify AST has the expected t.Helper() calls before formatting
		actualHelperCount := countTHelperCalls(node)
		if actualHelperCount != fixCount {
			return 0, fmt.Errorf("AST modification verification failed: expected %d t.Helper() calls, found %d", fixCount, actualHelperCount)
		}

		var buf bytes.Buffer
		cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
		if err := cfg.Fprint(&buf, fset, node); err != nil {
			return 0, fmt.Errorf("failed to format code: %w", err)
		}

		// Write formatted output
		if err := os.WriteFile(filePath, buf.Bytes(), 0o600); err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}

		// Verify output file has expected t.Helper() calls
		outputContent := buf.String()
		outputHelperCount := strings.Count(outputContent, "t.Helper()")
		if outputHelperCount != fixCount {
			return 0, fmt.Errorf("output verification failed: expected %d t.Helper() calls in output, found %d", fixCount, outputHelperCount)
		}
	}

	return fixCount, nil
}

// countTHelperCalls counts t.Helper() calls in the AST.
func countTHelperCalls(node *ast.File) int {
	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		exprStmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}

		callExpr, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if selectorExpr.Sel.Name == "Helper" {
			if ident, ok := selectorExpr.X.(*ast.Ident); ok {
				if ident.Name == "t" || ident.Name == "tb" {
					count++
				}
			}
		}

		return true
	})
	return count
}

// isTestHelperFunction checks if a function is a test helper based on naming patterns.
func isTestHelperFunction(funcDecl *ast.FuncDecl) bool {
	funcName := funcDecl.Name.Name

	// Exclude actual test functions (Test*, Benchmark*, Fuzz*, Example*)
	if strings.HasPrefix(funcName, "Test") || strings.HasPrefix(funcName, "Benchmark") ||
		strings.HasPrefix(funcName, "Fuzz") || strings.HasPrefix(funcName, "Example") {
		return false
	}

	// Check if function name matches helper patterns (case-insensitive)
	funcNameLower := strings.ToLower(funcName)
	helperPrefixes := []string{"setup", "check", "assert", "verify", "helper", "create", "build", "make"}
	matchesPattern := false
	for _, prefix := range helperPrefixes {
		if strings.HasPrefix(funcNameLower, prefix) {
			matchesPattern = true
			break
		}
	}

	if !matchesPattern {
		return false
	}

	// Must have *testing.T parameter
	return hasTestingTParam(funcDecl)
}

// hasTestingTParam checks if function has *testing.T parameter.
func hasTestingTParam(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil {
		return false
	}

	for _, param := range funcDecl.Type.Params.List {
		// Check if parameter type is *testing.T
		if starExpr, ok := param.Type.(*ast.StarExpr); ok {
			if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
				if ident, ok := selectorExpr.X.(*ast.Ident); ok {
					if ident.Name == "testing" && selectorExpr.Sel.Name == "T" {
						return true
					}
				}
			}
		}
	}

	return false
}

// hasTHelperCall checks if function body contains t.Helper() call.
func hasTHelperCall(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil {
		return false
	}

	// Look for t.Helper() in function body
	found := false
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		// Look for expression statements
		exprStmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}

		// Look for call expressions
		callExpr, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's a selector expression (t.Helper)
		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check if the selector is "Helper"
		if selectorExpr.Sel.Name != "Helper" {
			return true
		}

		// Check if the receiver is "t" (or any *testing.T variable)
		if ident, ok := selectorExpr.X.(*ast.Ident); ok {
			if ident.Name == "t" || ident.Name == "tb" {
				found = true
				return false
			}
		}

		return true
	})

	return found
}

// addTHelperCall adds t.Helper() as the first statement in the function body.
func addTHelperCall(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil {
		return false
	}

	// Find the testing.T parameter name
	tParamName := findTestingTParamName(funcDecl)
	if tParamName == "" {
		return false
	}

	// Create t.Helper() call expression
	helperCall := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: tParamName},
				Sel: &ast.Ident{Name: "Helper"},
			},
		},
	}

	// Insert t.Helper() as first statement
	funcDecl.Body.List = append([]ast.Stmt{helperCall}, funcDecl.Body.List...)

	return true
}

// findTestingTParamName finds the name of the *testing.T parameter.
func findTestingTParamName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Type.Params == nil {
		return ""
	}

	for _, param := range funcDecl.Type.Params.List {
		// Check if parameter type is *testing.T
		if starExpr, ok := param.Type.(*ast.StarExpr); ok {
			if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
				if ident, ok := selectorExpr.X.(*ast.Ident); ok {
					if ident.Name == "testing" && selectorExpr.Sel.Name == "T" {
						// Return the parameter name
						if len(param.Names) > 0 {
							return param.Names[0].Name
						}
					}
				}
			}
		}
	}

	return ""
}
