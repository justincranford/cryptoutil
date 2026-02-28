// Copyright (c) 2025 Justin Cranford

// Package magic_usage verifies that magic constants are properly defined and used in magic files.
package magic_usage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintGoCommon "cryptoutil/internal/apps/cicd/lint_go/common"
)

// Injectable functions for testing defensive error paths.
var (
	magicUsageAbsFn  = filepath.Abs
	magicUsageWalkFn = filepath.Walk
)

// magicUsageKind classifies how a magic value appears outside the magic package.
type magicUsageKind string

const (
	// magicUsageKindRedefine means the value is the right-hand side of a const
	// declaration outside the magic package (should reference magic.XXX instead).
	magicUsageKindRedefine magicUsageKind = "const-redefine"

	// magicUsageKindLiteral means the value appears as a bare literal in non-const
	// code (should reference magic.XXX instead of repeating the literal).
	magicUsageKindLiteral magicUsageKind = "literal-use"
)

// magicUsageViolation records one occurrence of a magic value used outside the magic package.
type magicUsageViolation struct {
	File         string
	Line         int
	Kind         magicUsageKind
	LiteralValue string
	MagicName    string
}

// Check is a LinterFunc that builds an inventory of the magic package
// and then walks the project tree, flagging any Go source file that:
//   - uses a magic constant's literal value as a bare expression literal, or
//   - redeclares that value as a local const outside the magic package.
//
// This catches violations that fall through goconst (requires >=2 occurrences
// per file) and mnd (numbers only; strings ignored).
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckMagicUsageInDir(logger, lintGoCommon.MagicDefaultDir, ".")
}

// CheckMagicUsageInDir is the testable implementation with explicit directory arguments.
func CheckMagicUsageInDir(logger *cryptoutilCmdCicdCommon.Logger, magicDir, rootDir string) error {
	logger.Log("Checking for magic values used as literals outside the magic package...")

	inv, err := lintGoCommon.ParseMagicDir(magicDir)
	if err != nil {
		return fmt.Errorf("failed to parse magic package: %w", err)
	}

	if len(inv.Constants) == 0 {
		logger.Log("✅ magic-usage: magic package empty, nothing to check")

		return nil
	}

	absMagicDir, err := magicUsageAbsFn(magicDir)
	if err != nil {
		return fmt.Errorf("cannot resolve magic dir: %w", err)
	}

	absRootDir, err := magicUsageAbsFn(rootDir)
	if err != nil {
		return fmt.Errorf("cannot resolve root dir: %w", err)
	}

	var (
		violations []magicUsageViolation
		walkErrors []string
	)

	walkErr := magicUsageWalkFn(absRootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			walkErrors = append(walkErrors, fmt.Sprintf("walk error at %s: %v", path, err))

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
		if lintGoCommon.MagicShouldSkipPath(relPath) {
			return nil
		}

		isTestFile := strings.HasSuffix(path, "_test.go")
		fileViolations := scanMagicFile(path, relPath, inv, isTestFile)
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
		logger.Log("✅ magic-usage: no magic values used as literals outside the magic package")

		return nil
	}

	var (
		literalUseViolations    []magicUsageViolation
		constRedefineViolations []magicUsageViolation
	)

	for _, v := range violations {
		if v.Kind == magicUsageKindLiteral {
			literalUseViolations = append(literalUseViolations, v)
		} else {
			constRedefineViolations = append(constRedefineViolations, v)
		}
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "magic-usage: %d violation(s) found (%d literal-use [blocking], %d const-redefine [informational])\n\n",
		len(violations), len(literalUseViolations), len(constRedefineViolations))

	for _, v := range violations {
		fmt.Fprintf(&sb, "  %s:%d  [%s]  literal %s  →  use magic.%s\n",
			v.File, v.Line, v.Kind, v.LiteralValue, v.MagicName)
	}

	logger.Log(sb.String())

	if len(literalUseViolations) > 0 {
		return fmt.Errorf("found %d literal-use violations: inline magic constant values must reference magic.XXX instead of repeating the literal", len(literalUseViolations))
	}

	// const-redefine violations are informational: numeric literals may coincidentally
	// share a value with a magic constant but have different semantics (e.g., step counts,
	// timeouts, and buffer sizes that happen to equal port numbers or pool sizes).
	// Run 'cicd lint-go' to track progress as true const-redefine violations are addressed.
	return nil
}

// scanMagicFile parses one Go source file and returns all magic-usage violations.
// isTestFile controls whether test-only magic constants are checked.
func scanMagicFile(absPath, relPath string, inv *lintGoCommon.MagicInventory, isTestFile bool) []magicUsageViolation {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		return nil // skip unparseable files silently.
	}

	v := &magicUsageVisitor{
		fset:       fset,
		inv:        inv,
		relFile:    relPath,
		isTestFile: isTestFile,
	}

	ast.Walk(v, file)

	return v.violations
}

// magicUsageVisitor walks an AST recording BasicLit nodes whose value matches a magic constant.
type magicUsageVisitor struct {
	fset        *token.FileSet
	inv         *lintGoCommon.MagicInventory
	relFile     string
	insideConst bool
	isTestFile  bool
	violations  []magicUsageViolation
}

// Visit implements ast.Visitor.
func (v *magicUsageVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.GenDecl:
		if n.Tok == token.CONST {
			child := &magicUsageVisitor{
				fset:        v.fset,
				inv:         v.inv,
				relFile:     v.relFile,
				insideConst: true,
				isTestFile:  v.isTestFile,
			}

			for _, spec := range n.Specs {
				ast.Walk(child, spec)
			}

			v.violations = append(v.violations, child.violations...)

			return nil
		}

	case *ast.BasicLit:
		v.checkLiteral(n)
	}

	return v
}

// checkLiteral records a violation if the literal value matches a magic constant.
// Test-only magic constants are only checked when isTestFile is true.
func (v *magicUsageVisitor) checkLiteral(lit *ast.BasicLit) {
	if lintGoCommon.IsMagicTrivialLiteral(lit) {
		return
	}

	consts, ok := v.inv.ByValue[lit.Value]
	if !ok {
		return
	}

	// Prefer the first non-test constant as the canonical reference.
	var mc *lintGoCommon.MagicConstant

	for i := range consts {
		if !consts[i].IsTestConst {
			mc = &consts[i]

			break
		}
	}

	if mc == nil {
		// Only test constants match; skip unless scanning a test file.
		if !v.isTestFile {
			return
		}

		mc = &consts[0]
	}

	kind := magicUsageKindLiteral
	if v.insideConst {
		kind = magicUsageKindRedefine
	}

	pos := v.fset.Position(lit.Pos())

	v.violations = append(v.violations, magicUsageViolation{
		File:         v.relFile,
		Line:         pos.Line,
		Kind:         kind,
		LiteralValue: lit.Value,
		MagicName:    mc.Name,
	})
}
