package cicd

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
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

	// Read original file content
	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	modifiedContent := string(originalContent)
	fixCount := 0

	// Find all function declarations that need t.Helper()
	var functionsToFix []struct {
		name      string
		paramName string
		bodyStart token.Pos
	}

	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		funcName := funcDecl.Name.Name

		// Check if this is a test helper function
		isHelper := isTestHelperFunction(funcDecl)

		if !isHelper {
			continue
		}

		// Check if function already has t.Helper() call
		hasHelper := hasTHelperCall(funcDecl)

		if hasHelper {
			continue
		}

		// Find the testing.T parameter name
		tParamName := findTestingTParamName(funcDecl)
		if tParamName == "" || funcDecl.Body == nil {
			continue
		}

		functionsToFix = append(functionsToFix, struct {
			name      string
			paramName string
			bodyStart token.Pos
		}{
			name:      funcName,
			paramName: tParamName,
			bodyStart: funcDecl.Body.Lbrace,
		})
	}

	// Apply fixes in reverse order to preserve positions
	for i := len(functionsToFix) - 1; i >= 0; i-- {
		fix := functionsToFix[i]
		bodyStartOffset := fset.Position(fix.bodyStart).Offset

		// Find the position of '{' in the source
		insertPos := bodyStartOffset + 1 // Right after '{'

		// Extract the content after '{' to determine proper indentation
		afterBrace := modifiedContent[insertPos:]

		// Find the first non-whitespace character to determine indentation
		indentation := "\t" // Default
		firstLineEnd := strings.IndexAny(afterBrace, "\n\r")
		if firstLineEnd != -1 && firstLineEnd+1 < len(afterBrace) {
			// Look at the next line to get indentation
			nextLineStart := firstLineEnd + 1
			if afterBrace[firstLineEnd] == '\r' && nextLineStart < len(afterBrace) && afterBrace[nextLineStart] == '\n' {
				nextLineStart++
			}
			indent := ""
			for i := nextLineStart; i < len(afterBrace) && (afterBrace[i] == ' ' || afterBrace[i] == '\t'); i++ {
				indent += string(afterBrace[i])
			}
			if indent != "" {
				indentation = indent
			}
		}

		// Insert t.Helper() call right after the opening brace
		helperCall := fmt.Sprintf("\n%s%s.Helper()", indentation, fix.paramName)
		modifiedContent = modifiedContent[:insertPos] + helperCall + modifiedContent[insertPos:]

		fixCount++
	}

	// Write back to file if modified
	if fixCount > 0 {
		// Format the modified content
		formatted, err := format.Source([]byte(modifiedContent))
		if err != nil {
			return 0, fmt.Errorf("failed to format modified code: %w", err)
		}

		if err := os.WriteFile(filePath, formatted, 0o600); err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return fixCount, nil
}// countTHelperCalls counts t.Helper() calls in the AST.
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
