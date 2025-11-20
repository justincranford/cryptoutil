package go_fix_copyloopvar

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
)

const (
	minGoVersion = "1.22" // Go 1.22+ has automatic loop variable scoping.
)

// Fix removes unnecessary loop variable copies in Go 1.25+ code.
// Returns the number of files processed, modified, and issues fixed.
func Fix(logger *cryptoutilCmd.Logger, rootDir string, goVersion string) (int, int, int, error) {
	// Check Go version.
	if !isGoVersionSupported(goVersion) {
		logger.Log(fmt.Sprintf("Skipping: Go version %s < %s (automatic loop variable scoping)", goVersion, minGoVersion))

		return 0, 0, 0, nil
	}

	logger.Log(fmt.Sprintf("Analyzing Go %s code for unnecessary loop variable copies", goVersion))

	var processed, modified, issuesFixed int

	if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-Go files, test files, and generated files.
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			strings.HasSuffix(path, "_gen.go") || strings.Contains(path, "openapi_gen_") {
			return nil
		}

		changed, fixes, err := fixCopyLoopVarInFile(logger, path)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", path, err)
		}

		processed++
		if changed {
			modified++
			issuesFixed += fixes
			logger.Log(fmt.Sprintf("Removed %d unnecessary loop variable copies in: %s", fixes, path))
		}

		return nil
	}); err != nil {
		return processed, modified, issuesFixed, fmt.Errorf("failed to walk directory: %w", err)
	}

	return processed, modified, issuesFixed, nil
}

// fixCopyLoopVarInFile removes unnecessary loop variable copies from a single file.
func fixCopyLoopVarInFile(logger *cryptoutilCmd.Logger, filePath string) (bool, int, error) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	fixCount := 0

	ast.Inspect(node, func(n ast.Node) bool {
		rangeStmt, ok := n.(*ast.RangeStmt)
		if !ok {
			return true
		}

		// Check if the range statement body contains the loop variable copy pattern.
		body := rangeStmt.Body
		if body == nil || len(body.List) == 0 {
			return true
		}

		// Check if the first statement is a loop variable copy.
		assign, ok := body.List[0].(*ast.AssignStmt)
		if !ok {
			return true
		}

		if isLoopVarCopy(rangeStmt, assign) {
			// Remove the assignment statement.
			body.List = body.List[1:]
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
		defer file.Close()

		if err := printer.Fprint(file, fset, node); err != nil {
			return false, fixCount, fmt.Errorf("failed to write file: %w", err)
		}

		return true, fixCount, nil
	}

	return false, 0, nil
}

// isLoopVarCopy checks if an assignment is a loop variable copy (e.g., `x := x`).
func isLoopVarCopy(rangeStmt *ast.RangeStmt, assign *ast.AssignStmt) bool {
	// Check for short variable declaration with :=
	if assign.Tok != token.DEFINE {
		return false
	}

	// Must have exactly one LHS and one RHS.
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return false
	}

	lhs, ok1 := assign.Lhs[0].(*ast.Ident)

	rhs, ok2 := assign.Rhs[0].(*ast.Ident)
	if !ok1 || !ok2 {
		return false
	}

	// Check if LHS and RHS are the same identifier.
	if lhs.Name != rhs.Name {
		return false
	}

	// Verify that the RHS identifier is the range loop variable.
	if rangeStmt.Value != nil {
		if valueIdent, ok := rangeStmt.Value.(*ast.Ident); ok {
			if valueIdent.Name == rhs.Name {
				return true
			}
		}
	}

	if rangeStmt.Key != nil {
		if keyIdent, ok := rangeStmt.Key.(*ast.Ident); ok {
			if keyIdent.Name == rhs.Name {
				return true
			}
		}
	}

	return false
}

// isGoVersionSupported checks if the Go version supports automatic loop variable scoping.
func isGoVersionSupported(version string) bool {
	// Parse version string (e.g., "1.25.4" -> [1, 25, 4])
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	// Compare with minimum version (1.22)
	minParts := strings.Split(minGoVersion, ".")
	minMajor, _ := strconv.Atoi(minParts[0])
	minMinor, _ := strconv.Atoi(minParts[1])

	if major > minMajor {
		return true
	}

	if major == minMajor && minor >= minMinor {
		return true
	}

	return false
}
