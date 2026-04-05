// Copyright (c) 2025 Justin Cranford

// Package function_var_redeclaration detects package-level var declarations
// that redeclare functions from other packages using the `Fn` naming suffix
// (e.g. var walkFn = filepath.Walk). This seam-injection anti-pattern is banned
// in production code; use function-parameter injection instead.
package function_var_redeclaration

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoCommon "cryptoutil/internal/apps/tools/cicd_lint/lint_go/common"
)

// violation records one package-level function-var redeclaration.
type violation struct {
	File    string
	Line    int
	VarName string
	Pkg     string
	Ident   string
}

// Check is a LinterFunc that scans all Go source files for package-level
// var declarations whose name ends in "Fn" and whose initialiser is a bare
// SelectorExpr (pkg.Func). This is the project convention for seam vars.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", filepath.Walk)
}

// CheckInDir is the testable implementation that accepts explicit fn dependencies.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, walkFn func(string, filepath.WalkFunc) error) error {
	logger.Log("Checking for package-level function-var redeclarations...")

	var violations []violation

	walkErr := walkFn(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			relDir, _ := filepath.Rel(rootDir, path)
			if lintGoCommon.MagicShouldSkipPath(relDir) {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		base := filepath.Base(path)

		// Skip test files and export_test.go — seam exposure is legitimate there.
		if strings.HasSuffix(base, "_test.go") {
			return nil
		}

		fileViolations := checkFile(path)
		violations = append(violations, fileViolations...)

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("directory walk failed: %w", walkErr)
	}

	if len(violations) == 0 {
		logger.Log("✅ No package-level function-var redeclarations found")

		return nil
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "function-var-redeclaration: %d violation(s) found\n\n", len(violations))

	for _, v := range violations {
		fmt.Fprintf(&sb, "  %s:%d  var %s = %s.%s\n", v.File, v.Line, v.VarName, v.Pkg, v.Ident)
	}

	fmt.Fprint(&sb, "\nReplace with fn-parameter injection: pass the function as a parameter instead.\n")

	logger.Log(sb.String())

	return fmt.Errorf("function-var-redeclaration: %d violation(s) found", len(violations))
}

// checkFile parses one Go source file and returns violations.
func checkFile(filePath string) []violation {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil // skip unparseable files silently.
	}

	var violations []violation

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			// For token.VAR GenDecls the parser always produces *ast.ValueSpec.
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, val := range valSpec.Values {
				sel, ok := val.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					// X is another SelectorExpr (e.g., outer.inner.Method) — not a simple pkg.Func.
					continue
				}

				varName := valSpec.Names[i].Name

				// Only flag vars whose name ends in "Fn" — the project-wide convention
				// for function seam vars (e.g. walkFn, absFn, marshalFn).
				// Non-Fn vars like `var defaultCertPEM = magic.SomeConst` are legitimate.
				if !strings.HasSuffix(varName, "Fn") {
					continue
				}

				violations = append(violations, violation{
					File:    filePath,
					Line:    fset.Position(genDecl.Pos()).Line,
					VarName: varName,
					Pkg:     pkgIdent.Name,
					Ident:   sel.Sel.Name,
				})
			}
		}
	}

	return violations
}
