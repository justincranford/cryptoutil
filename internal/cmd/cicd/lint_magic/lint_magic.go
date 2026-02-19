// Copyright (c) 2025 Justin Cranford

// Package lint_magic provides validators for the internal/shared/magic package.
// It implements two validators:
//  1. ValidateDuplicates: finds constants with the same literal value defined
//     under multiple names within the magic package (Pet Peeve #1).
//  2. ValidateUsage: finds magic values used as bare literals or redefined as
//     constants outside the magic package, which fall through the built-in mnd
//     and goconst linters due to their per-file occurrence thresholds (Pet Peeve #2).
package lint_magic

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

// defaultMagicDir is the conventional path to the shared magic package, relative to project root.
const defaultMagicDir = "internal/shared/magic"

// defaultRootDir is the project root directory used when scanning for usages.
const defaultRootDir = "."

// minStringLen is the minimum unquoted length of a string literal that is flagged
// as a magic-value usage. Strings shorter than this are considered too common
// (e.g. "", ".", "/") to report reliably.
const minStringLen = 3

// trivialInts is the set of integer literal strings too common to be meaningful.
var trivialInts = map[string]bool{
"0": true, "1": true, "2": true, "3": true, "4": true, "-1": true,
}

// MagicConstant represents a single constant defined in the magic package
// whose value is a basic literal (string, integer, float, rune, or char).
// Constants that reference other identifiers (e.g. DefaultProfile = EmptyString)
// are not included because their value cannot be determined without full evaluation.
type MagicConstant struct {
// Name is the constant identifier, e.g. "ProtocolHTTPS".
Name string

// Value is the raw Go literal as it appears in source, e.g. `"https"` or `443`.
Value string

// File is the base filename within the magic package.
File string

// Line is the 1-based source line number of the constant value.
Line int

// IsTestConst is true when the constant is defined in magic_testing.go or
// has a name starting with "Test".  Test constants are only matched against
// test files (_test.go) to prevent false positives in production code.
IsTestConst bool
}

// MagicInventory holds all BasicLit-valued constants parsed from the magic package.
type MagicInventory struct {
// Constants is the full list, sorted by file then line.
Constants []MagicConstant

// ByValue maps a raw literal value to all constants that share it.
ByValue map[string][]MagicConstant

// ByName maps a constant name to its definition.
ByName map[string]MagicConstant
}

// parseMagicPackage parses all non-test .go files in magicDir and returns a
// MagicInventory containing every constant whose value is a basic literal.
func parseMagicPackage(magicDir string) (*MagicInventory, error) {
fset := token.NewFileSet()

pkgs, err := parser.ParseDir(fset, magicDir, func(fi os.FileInfo) bool {
return !strings.HasSuffix(fi.Name(), "_test.go")
}, 0)
if err != nil {
return nil, fmt.Errorf("failed to parse magic package at %s: %w", magicDir, err)
}

inv := &MagicInventory{
ByValue: make(map[string][]MagicConstant),
ByName:  make(map[string]MagicConstant),
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
continue // derived constant (e.g. Foo = OtherConst), skip.
}

pos := fset.Position(lit.Pos())
isTest := relFile == "magic_testing.go" || strings.HasPrefix(nameIdent.Name, "Test")
mc := MagicConstant{
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

// isTrivialLiteral returns true for literals that are too common to be useful
// to flag as magic-value violations.
func isTrivialLiteral(lit *ast.BasicLit) bool {
switch lit.Kind {
case token.STRING:
// Strip surrounding quotes to measure the actual content length.
raw := lit.Value
if len(raw) >= 2 {
raw = raw[1 : len(raw)-1]
}

return len(raw) < minStringLen

case token.INT:
return trivialInts[lit.Value]

default:
return false
}
}

// generatedAPIDirs lists subdirectories of api/ that contain only oapi-codegen
// generated files.  These are excluded from usage scanning to avoid noise,
// matching the same exclusion list used by golangci-lint.
var generatedAPIDirs = map[string]bool{
"client": true, "model": true, "server": true, "idp": true, "authz": true,
}

// shouldSkipPath returns true if the given path should be excluded from usage scanning.
func shouldSkipPath(path string) bool {
slashed := filepath.ToSlash(path)
parts := strings.Split(slashed, "/")

for i, part := range parts {
switch part {
case "vendor", "test-output", "workflow-reports":
return true

case "api":
// Skip known generated api subdirectories.
if i+1 < len(parts) && generatedAPIDirs[parts[i+1]] {
return true
}
}
}

return false
}

// isGeneratedFile returns true if the file name indicates it was produced by a
// code generator (oapi-codegen, protoc, etc.).
func isGeneratedFile(name string) bool {
return strings.HasSuffix(name, ".gen.go") || strings.Contains(name, "_gen_")
}
