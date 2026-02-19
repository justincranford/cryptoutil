// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// magicDefaultDir is the conventional path to the shared magic package, relative to project root.
const magicDefaultDir = "internal/shared/magic"

// magicMinStringLen is the minimum unquoted length of a string literal to flag as a magic-value issue.
// Strings shorter than this are too common (e.g. "", ".", "/") to report reliably.
const magicMinStringLen = 3

// magicTrivialInts is the set of integer literal strings too common to be meaningful.
var magicTrivialInts = map[string]bool{
	"0": true, "1": true, "2": true, "3": true, "4": true, "-1": true,
}

// magicConstant represents a single constant defined in the magic package
// whose value is a basic literal (string, integer, float, or rune).
// Constants that reference other identifiers (e.g. DefaultProfile = EmptyString)
// are skipped because their resolved value requires full type-checking.
type magicConstant struct {
	// Name is the constant identifier, e.g. "ProtocolHTTPS".
	Name string

	// Value is the raw Go literal as it appears in source, e.g. `"https"` or `443`.
	Value string

	// File is the base filename within the magic package.
	File string

	// Line is the 1-based source line number.
	Line int

	// IsTestConst is true when the constant is defined in magic_testing.go or
	// has a name starting with "Test".  Test constants are only matched against
	// test files (_test.go) to avoid false positives in production code.
	IsTestConst bool
}

// magicInventory holds all BasicLit-valued constants parsed from the magic package.
type magicInventory struct {
	// Constants is the full sorted list.
	Constants []magicConstant

	// ByValue maps a raw literal value to all constants that share it.
	ByValue map[string][]magicConstant

	// ByName maps a constant name to its definition.
	ByName map[string]magicConstant
}

// parseMagicDir parses all non-test .go files in magicDir and returns a
// magicInventory of every constant whose value is a basic literal.
func parseMagicDir(magicDir string) (*magicInventory, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, magicDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse magic package at %s: %w", magicDir, err)
	}

	inv := &magicInventory{
		ByValue: make(map[string][]magicConstant),
		ByName:  make(map[string]magicConstant),
	}

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			relFile := filepath.Base(filename)

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
						if i >= len(vspec.Values) {
							continue
						}

						lit, ok := vspec.Values[i].(*ast.BasicLit)
						if !ok {
							continue // derived constant, skip.
						}

						pos := fset.Position(lit.Pos())
						isTest := relFile == "magic_testing.go" || strings.HasPrefix(nameIdent.Name, "Test")
						mc := magicConstant{
							Name:        nameIdent.Name,
							Value:       lit.Value,
							File:        relFile,
							Line:        pos.Line,
							IsTestConst: isTest,
						}

						inv.Constants = append(inv.Constants, mc)
						inv.ByValue[lit.Value] = append(inv.ByValue[lit.Value], mc)
						inv.ByName[nameIdent.Name] = mc
					}
				}

				return true
			})
		}
	}

	sort.Slice(inv.Constants, func(i, j int) bool {
		if inv.Constants[i].File != inv.Constants[j].File {
			return inv.Constants[i].File < inv.Constants[j].File
		}

		return inv.Constants[i].Line < inv.Constants[j].Line
	})

	return inv, nil
}

// isMagicTrivialLiteral returns true for literals too common to flag as magic-value violations.
func isMagicTrivialLiteral(lit *ast.BasicLit) bool {
	switch lit.Kind {
	case token.STRING:
		// Strip surrounding quotes to measure actual content length.
		raw := lit.Value
		if len(raw) >= 2 {
			raw = raw[1 : len(raw)-1]
		}

		return len(raw) < magicMinStringLen

	case token.INT:
		return magicTrivialInts[lit.Value]

	default:
		return false
	}
}

// magicGeneratedAPIDirs lists api/ subdirectories containing only generated files.
// Matches the exclusion list used by golangci-lint in .golangci.yml.
var magicGeneratedAPIDirs = map[string]bool{
	"client": true, "model": true, "server": true, "idp": true, "authz": true,
}

const (
	magicExcludeDirTestOutput       = "test-output"
	magicExcludeDirWorkflowReports  = "workflow-reports"
)

// magicShouldSkipPath returns true if the given relative path should be excluded from magic scanning.
func magicShouldSkipPath(path string) bool {
	slashed := filepath.ToSlash(path)
	parts := strings.Split(slashed, "/")

	for i, part := range parts {
		switch part {
		case excludeDirVendor, magicExcludeDirTestOutput, magicExcludeDirWorkflowReports:
			return true

		case "api":
			if i+1 < len(parts) && magicGeneratedAPIDirs[parts[i+1]] {
				return true
			}
		}
	}

	return false
}

// isMagicGeneratedFile returns true when the file name indicates code-generator output.
func isMagicGeneratedFile(name string) bool {
	return strings.HasSuffix(name, ".gen.go") || strings.Contains(name, "_gen_")
}
