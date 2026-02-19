// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// UsageKind classifies how a magic value appears outside the magic package.
type UsageKind string

const (
	// UsageKindRedefine means the value appears as the right-hand side of a const
	// declaration outside the magic package (should reference magic.XXX instead).
	UsageKindRedefine UsageKind = "const-redefine"

	// UsageKindLiteral means the value appears as a bare literal in non-const
	// code (should reference magic.XXX instead of repeating the literal).
	UsageKindLiteral UsageKind = "literal-use"
)

// UsageViolation records a single occurrence of a magic value used outside
// the magic package.
type UsageViolation struct {
	// File is the path of the offending file, relative to rootDir.
	File string

	// Line is the 1-based line number.
	Line int

	// Kind distinguishes const-redefine from literal-use.
	Kind UsageKind

	// LiteralValue is the raw Go literal that triggered the match, e.g. `"https"`.
	LiteralValue string

	// MagicName is the name of the matching magic constant, e.g. "ProtocolHTTPS".
	MagicName string
}

// UsageResult holds the outcome of out-of-package magic value detection.
type UsageResult struct {
	// Valid is false when at least one violation was found.
	Valid bool

	// Violations lists every occurrence of a magic value used outside the
	// magic package.
	Violations []UsageViolation

	// Errors lists file-system or parsing errors encountered during validation.
	Errors []string
}

// ValidateUsage builds a complete inventory of the magic package and then
// walks every eligible Go source file under rootDir, searching for any
// BasicLit whose value matches a magic constant.  Matches are classified as
// either a const-redefine (the literal is the value of a const declaration) or
// a literal-use (the literal appears in an expression context).
//
// Files that are excluded from scanning:
//   - the magic package directory itself
//   - files ending in .gen.go (code-generated)
//   - paths containing vendor, test-output, or workflow-reports
func ValidateUsage(magicDir, rootDir string) (*UsageResult, error) {
	result := &UsageResult{Valid: true}

	inv, err := parseMagicPackage(magicDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("parse error: %v", err))
		result.Valid = false

		return result, nil
	}

	if len(inv.Constants) == 0 {
		return result, nil
	}

	// Resolve absolute paths for exclusion comparison.
	absMagicDir, err := filepath.Abs(magicDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("cannot resolve magic dir: %v", err))
		result.Valid = false

		return result, nil
	}

	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("cannot resolve root dir: %v", err))
		result.Valid = false

		return result, nil
	}

	walkErr := filepath.Walk(absRootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("walk error at %s: %v", path, err))

			return nil
		}

		if info.IsDir() {
			// Skip the magic package directory entirely.
			if path == absMagicDir {
				return filepath.SkipDir
			}

			absPath, _ := filepath.Abs(path)
			if shouldSkipPath(absPath) {
				return filepath.SkipDir
			}

			return nil
		}

		// Only scan .go files, skip generated files.
		if !strings.HasSuffix(path, ".go") || isGeneratedFile(filepath.Base(path)) {
			return nil
		}

		relPath, _ := filepath.Rel(absRootDir, path)
		if shouldSkipPath(relPath) {
			return nil
		}

		isTestFile := strings.HasSuffix(path, "_test.go")

		violations := scanFile(path, relPath, inv, isTestFile)
		if len(violations) > 0 {
			result.Violations = append(result.Violations, violations...)
			result.Valid = false
		}

		return nil
	})
	if walkErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("walk failed: %v", walkErr))
		result.Valid = false
	}

	return result, nil
}

// scanFile parses a single Go source file and returns all violations found.
// isTestFile controls whether test-only magic constants are checked.
func scanFile(absPath, relPath string, inv *MagicInventory, isTestFile bool) []UsageViolation {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		return nil // skip unparseable files silently.
	}

	v := &usageVisitor{
		fset:        fset,
		inv:         inv,
		relFile:     relPath,
		insideConst: false,
		isTestFile:  isTestFile,
	}

	ast.Walk(v, file)

	return v.violations
}

// usageVisitor is an ast.Visitor that records BasicLit nodes whose value
// matches a magic constant.  It tracks whether it is currently visiting the
// specs inside a const declaration so that violations can be classified
// correctly.  isTestFile controls whether test-only magic constants are matched.
type usageVisitor struct {
	fset        *token.FileSet
	inv         *MagicInventory
	relFile     string
	insideConst bool
	isTestFile  bool
	violations  []UsageViolation
}

// Visit implements ast.Visitor.
func (v *usageVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.GenDecl:
		if n.Tok == token.CONST {
			// Walk the const specs with a child visitor that has insideConst=true,
			// then return nil to prevent the default traversal from visiting them again.
			child := &usageVisitor{
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

// checkLiteral tests whether a BasicLit matches a magic constant and records
// a violation if so.  Test-only magic constants are only checked when the
// target file is itself a test file.
func (v *usageVisitor) checkLiteral(lit *ast.BasicLit) {
	if isTrivialLiteral(lit) {
		return
	}

	consts, ok := v.inv.ByValue[lit.Value]
	if !ok {
		return
	}

	// Find the first non-test (or any) constant as the canonical reference.
	var mc *MagicConstant

	for i := range consts {
		if !consts[i].IsTestConst {
			mc = &consts[i]

			break
		}
	}

	if mc == nil {
		// Only test constants match this value; skip unless we're in a test file.
		if !v.isTestFile {
			return
		}

		mc = &consts[0]
	}

	kind := UsageKindLiteral
	if v.insideConst {
		kind = UsageKindRedefine
	}

	pos := v.fset.Position(lit.Pos())

	v.violations = append(v.violations, UsageViolation{
		File:         v.relFile,
		Line:         pos.Line,
		Kind:         kind,
		LiteralValue: lit.Value,
		MagicName:    mc.Name,
	})
}

// FormatUsageResult formats the usage result as a human-readable CI/CD report.
func FormatUsageResult(result *UsageResult) string {
	var sb strings.Builder

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(&sb, "ERROR: %s\n", e)
		}
	}

	if len(result.Violations) == 0 && len(result.Errors) == 0 {
		fmt.Fprint(&sb, "validate-usage: OK (no magic values used as literals outside the magic package)\n")

		return sb.String()
	}

	fmt.Fprintf(&sb, "validate-usage: FAIL (%d violation(s) found)\n\n", len(result.Violations))

	for _, v := range result.Violations {
		fmt.Fprintf(&sb, "  %s:%d  [%s]  literal %s  →  use magic.%s\n",
			v.File, v.Line, v.Kind, v.LiteralValue, v.MagicName)
	}

	fmt.Fprint(&sb, "\n")

	return sb.String()
}
