package go_fix_thelper

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
)

// helperFunctionPatterns are naming patterns that indicate test helper functions.
var helperFunctionPatterns = []string{
	"setup",
	"check",
	"assert",
	"verify",
	"helper",
	"create",
	"build",
	"mock",
}

// Fix adds t.Helper() to test helper functions that are missing it.
// Returns the number of files processed, modified, and issues fixed.
func Fix(logger *cryptoutilCmd.Logger, rootDir string) (int, int, int, error) {
	var processed, modified, issuesFixed int

	if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process test files.
		if info.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		changed, fixes, err := fixTHelperInFile(logger, path)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", path, err)
		}

		processed++
		if changed {
			modified++
			issuesFixed += fixes
			logger.Log(fmt.Sprintf("Added t.Helper() to %d functions in: %s", fixes, path))
		}

		return nil
	}); err != nil {
		return processed, modified, issuesFixed, fmt.Errorf("failed to walk directory: %w", err)
	}

	return processed, modified, issuesFixed, nil
}

// fixTHelperInFile adds t.Helper() to test helper functions in a single file.
func fixTHelperInFile(logger *cryptoutilCmd.Logger, filePath string) (bool, int, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	fixCount := 0

	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Check if this is a test helper function.
		if !isTestHelperFunction(funcDecl) {
			return true
		}

		// Check if function already has t.Helper() call.
		if hasTHelperCall(funcDecl) {
			return true
		}

		// Find the testing.T parameter name.
		testingParam := getTestingParam(funcDecl)
		if testingParam == "" {
			return true
		}

		// Add t.Helper() as the first statement in the function body.
		if funcDecl.Body != nil {
			helperCall := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: testingParam},
						Sel: &ast.Ident{Name: "Helper"},
					},
				},
			}

			// Prepend t.Helper() to the function body.
			funcDecl.Body.List = append([]ast.Stmt{helperCall}, funcDecl.Body.List...)
			fixCount++
		}

		return true
	})

	if fixCount > 0 {
		// Write the modified AST back to the file.
		file, err := os.Create(filePath)
		if err != nil {
			return false, fixCount, fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close() //nolint:errcheck // Defer close is best-effort

		if err := printer.Fprint(file, fset, node); err != nil {
			return false, fixCount, fmt.Errorf("failed to write file: %w", err)
		}

		return true, fixCount, nil
	}

	return false, 0, nil
}

// isTestHelperFunction checks if a function is a test helper function.
func isTestHelperFunction(funcDecl *ast.FuncDecl) bool {
	funcName := strings.ToLower(funcDecl.Name.Name)

	// Check naming patterns.
	for _, pattern := range helperFunctionPatterns {
		if strings.HasPrefix(funcName, pattern) {
			return true
		}
	}

	return false
}

// hasTHelperCall checks if a function already has a t.Helper() call.
func hasTHelperCall(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil {
		return false
	}

	for _, stmt := range funcDecl.Body.List {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}

		callExpr, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			continue
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		if selExpr.Sel.Name == "Helper" {
			return true
		}
	}

	return false
}

// getTestingParam finds the testing.T parameter name in a function.
func getTestingParam(funcDecl *ast.FuncDecl) string {
	if funcDecl.Type == nil || funcDecl.Type.Params == nil {
		return ""
	}

	for _, field := range funcDecl.Type.Params.List {
		// Check if the parameter type is *testing.T.
		starExpr, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}

		selExpr, ok := starExpr.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		xIdent, ok := selExpr.X.(*ast.Ident)
		if !ok || xIdent.Name != "testing" {
			continue
		}

		if selExpr.Sel.Name == "T" {
			// Found testing.T parameter.
			if len(field.Names) > 0 {
				return field.Names[0].Name
			}
		}
	}

	return ""
}
