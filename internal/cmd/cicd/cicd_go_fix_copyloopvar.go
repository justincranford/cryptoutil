package cicd

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

	"golang.org/x/mod/modfile"

	"cryptoutil/internal/cmd/cicd/common"
)

// goFixCopyLoopVar removes unnecessary loop variable copies (x := x) in Go 1.25+.
// In Go 1.25+, loop variables are automatically per-iteration, making explicit copies redundant.
func goFixCopyLoopVar(logger *common.Logger, files []string) error {
	logger.Log("Starting copyloopvar auto-fix")

	// Check Go version from go.mod
	goVersion, err := getGoVersion()
	if err != nil {
		return fmt.Errorf("failed to get Go version: %w", err)
	}
	logger.Log(fmt.Sprintf("Go version: %s", goVersion))

	// Only run for Go 1.25+
	if !isGo125OrHigher(goVersion) {
		logger.Log(fmt.Sprintf("Skipping: Go version %s < 1.25 (loop variables not per-iteration)", goVersion))
		return nil
	}

	goFiles := filterGoFiles(files)
	logger.Log(fmt.Sprintf("Processing %d Go files", len(goFiles)))

	totalFixCount := 0
	for _, file := range goFiles {
		fixCount, err := fixCopyLoopVarInFile(file)
		if err != nil {
			return fmt.Errorf("failed to fix copyloopvar in %s: %w", file, err)
		}
		totalFixCount += fixCount
	}

	if totalFixCount > 0 {
		logger.Log(fmt.Sprintf("Fixed %d loop variable copies", totalFixCount))
		return fmt.Errorf("fixed %d loop variable copies - please review changes", totalFixCount)
	}

	logger.Log("No loop variable copies needed fixing")
	return nil
}

// fixCopyLoopVarInFile removes unnecessary loop variable copies from a file.
func fixCopyLoopVarInFile(filePath string) (int, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file: %w", err)
	}

	fixCount := 0
	modified := false

	// Find all loop variable copy statements
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for range loops
		rangeStmt, ok := n.(*ast.RangeStmt)
		if !ok {
			return true
		}

		// Body is already *ast.BlockStmt, check if it has statements
		blockStmt := rangeStmt.Body
		if blockStmt == nil || len(blockStmt.List) == 0 {
			return true
		}

		// Collect loop variables (key, value)
		loopVars := make(map[string]bool)
		if rangeStmt.Key != nil {
			if ident, ok := rangeStmt.Key.(*ast.Ident); ok {
				loopVars[ident.Name] = true
			}
		}
		if rangeStmt.Value != nil {
			if ident, ok := rangeStmt.Value.(*ast.Ident); ok {
				loopVars[ident.Name] = true
			}
		}

		if len(loopVars) == 0 {
			return true
		}

		// Check first statements in loop body for x := x pattern
		stmtsToRemove := make(map[int]bool)
		for i, stmt := range blockStmt.List {
			assignStmt, ok := stmt.(*ast.AssignStmt)
			if !ok || assignStmt.Tok != token.DEFINE {
				continue
			}

			// Check if it's a simple assignment: x := x
			if len(assignStmt.Lhs) != 1 || len(assignStmt.Rhs) != 1 {
				continue
			}

			lhs, lhsOk := assignStmt.Lhs[0].(*ast.Ident)
			rhs, rhsOk := assignStmt.Rhs[0].(*ast.Ident)
			if !lhsOk || !rhsOk {
				continue
			}

			// Check if both sides are the same AND it's a loop variable
			if lhs.Name == rhs.Name && loopVars[lhs.Name] {
				stmtsToRemove[i] = true
				fixCount++
			}
		}

		// Remove the unnecessary statements
		if len(stmtsToRemove) > 0 {
			newStmts := make([]ast.Stmt, 0, len(blockStmt.List)-len(stmtsToRemove))
			for i, stmt := range blockStmt.List {
				if !stmtsToRemove[i] {
					newStmts = append(newStmts, stmt)
				}
			}
			blockStmt.List = newStmts
			modified = true
		}

		return true
	})

	// Write back to file if modified
	if modified {
		f, err := os.Create(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to create file: %w", err)
		}
		defer f.Close()

		if err := printer.Fprint(f, fset, node); err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return fixCount, nil
}

// getGoVersion reads the Go version from go.mod.
func getGoVersion() (string, error) {
	// Try current directory first, then walk up to find go.mod
	modPath := "go.mod"
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		// Not in current directory, try parent directories
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}

		// Walk up directory tree to find go.mod
		dir := cwd
		for {
			testPath := filepath.Join(dir, "go.mod")
			if _, err := os.Stat(testPath); err == nil {
				modPath = testPath
				break
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				// Reached root without finding go.mod
				return "", fmt.Errorf("go.mod not found in current directory or parents")
			}
			dir = parent
		}
	}

	data, err := os.ReadFile(modPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse go.mod: %w", err)
	}

	if modFile.Go == nil {
		return "", fmt.Errorf("no go version specified in go.mod")
	}

	return modFile.Go.Version, nil
}

// isGo125OrHigher checks if the Go version is 1.25 or higher.
func isGo125OrHigher(version string) bool {
	// Parse version (e.g., "1.25.4" or "1.25")
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

	// Check if >= 1.25
	return major > 1 || (major == 1 && minor >= 25)
}
