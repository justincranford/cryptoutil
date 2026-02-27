// Copyright (c) 2025 Justin Cranford

// Package magic_aliases detects const alias redeclarations of magic package constants.
// An alias is a const declaration like `localName = cryptoutilSharedMagic.MagicName`
// that adds indirection without value. Callers should use the magic constant directly.
package magic_aliases

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
)

// Injectable functions for testing defensive error paths.
var (
	magicAliasesAbsFn  = filepath.Abs
	magicAliasesWalkFn = filepath.Walk
)

// aliasViolation records one const alias redeclaration.
type aliasViolation struct {
	File      string
	Line      int
	LocalName string
	MagicName string
}

// Check is a LinterFunc that scans all Go files outside the magic package
// for const declarations that alias a cryptoutilSharedMagic constant.
// Exported aliases (uppercase first letter) are allowed as package API boundaries.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckMagicAliasesInDir(logger, lintGoCommon.MagicDefaultDir, ".")
}

// CheckMagicAliasesInDir is the testable implementation with explicit directory arguments.
func CheckMagicAliasesInDir(logger *cryptoutilCmdCicdCommon.Logger, magicDir, rootDir string) error {
	logger.Log("Checking for const alias redeclarations of magic package constants...")

	absMagicDir, err := magicAliasesAbsFn(magicDir)
	if err != nil {
		return fmt.Errorf("cannot resolve magic dir: %w", err)
	}

	absRootDir, err := magicAliasesAbsFn(rootDir)
	if err != nil {
		return fmt.Errorf("cannot resolve root dir: %w", err)
	}

	var violations []aliasViolation

	var walkErrors []string

	walkErr := magicAliasesWalkFn(absRootDir, func(path string, info os.FileInfo, walkFileErr error) error {
		if walkFileErr != nil {
			walkErrors = append(walkErrors, fmt.Sprintf("walk error at %s: %v", path, walkFileErr))

			return nil
		}

		if info.IsDir() {
			if path == absMagicDir {
				return filepath.SkipDir
			}

			relDir, _ := filepath.Rel(absRootDir, path)
			if lintGoCommon.MagicShouldSkipPath(relDir) {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || lintGoCommon.IsMagicGeneratedFile(filepath.Base(path)) {
			return nil
		}

		relPath, _ := filepath.Rel(absRootDir, path)

		fileViolations := findAliasesInFile(path, relPath)
		violations = append(violations, fileViolations...)

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("directory walk failed: %w", walkErr)
	}

	if len(walkErrors) > 0 {
		return fmt.Errorf("walk errors: %s", strings.Join(walkErrors, "; "))
	}

	if len(violations) == 0 {
		logger.Log("✅ magic-aliases: no const alias redeclarations found")

		return nil
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}

		return violations[i].Line < violations[j].Line
	})

	var sb strings.Builder

	fmt.Fprintf(&sb, "magic-aliases: %d unexported const alias redeclaration(s) found (use cryptoutilSharedMagic.X directly)\n\n", len(violations))

	for _, v := range violations {
		fmt.Fprintf(&sb, "  %s:%d  %s = cryptoutilSharedMagic.%s\n", v.File, v.Line, v.LocalName, v.MagicName)
	}

	// magic-aliases is informational: it logs violations but does not block CI.
	// Alias redeclarations are code smell, not correctness issues.
	logger.Log(sb.String())

	return nil
}

// findAliasesInFile parses a Go source file and returns violations for any
// unexported const declaration whose value is a selector expression referencing
// cryptoutilSharedMagic.
func findAliasesInFile(absPath, relPath string) []aliasViolation {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		return nil
	}

	// Find the import alias used for the magic package.
	magicAlias := findMagicImportAlias(file)
	if magicAlias == "" {
		return nil
	}

	var violations []aliasViolation

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.CONST {
			return true
		}

		for _, spec := range decl.Specs {
			vspec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, nameIdent := range vspec.Names {
				// Skip exported aliases - they serve as package API boundaries.
				if nameIdent.IsExported() {
					continue
				}

				if i >= len(vspec.Values) {
					continue
				}

				sel, ok := vspec.Values[i].(*ast.SelectorExpr)
				if !ok {
					continue
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok || ident.Name != magicAlias {
					continue
				}

				pos := fset.Position(nameIdent.Pos())
				violations = append(violations, aliasViolation{
					File:      relPath,
					Line:      pos.Line,
					LocalName: nameIdent.Name,
					MagicName: sel.Sel.Name,
				})
			}
		}

		return true
	})

	return violations
}

// findMagicImportAlias returns the alias used for the magic package import,
// or empty string if not imported.
func findMagicImportAlias(file *ast.File) string {
	for _, imp := range file.Imports {
		if imp.Path == nil {
			continue
		}

		path := strings.Trim(imp.Path.Value, `"`)
		if path == "cryptoutil/internal/shared/magic" {
			if imp.Name != nil {
				return imp.Name.Name
			}

			return "magic"
		}
	}

	return ""
}
