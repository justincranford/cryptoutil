// Copyright (c) 2025 Justin Cranford

// Package magic_duplicates verifies that magic constants have no duplicate values across magic files.
package magic_duplicates

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
	magicDuplicatesWalkFn = filepath.Walk
	magicDuplicatesAbsFn  = filepath.Abs
)

// crossFileConstant records one const declaration in a non-magic Go file.
type crossFileConstant struct {
	File  string
	Line  int
	Name  string
	Value string
}

// Check is a LinterFunc that scans the magic package for constants that share
// the same literal value under multiple names, and also scans all other Go
// files for constants whose values are duplicated across two or more files
// outside the magic package.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	if err := CheckMagicDuplicatesInDir(logger, lintGoCommon.MagicDefaultDir); err != nil {
		return err
	}

	return CheckCrossFileDuplicatesInDir(logger, lintGoCommon.MagicDefaultDir, ".")
}

// CheckMagicDuplicatesInDir is the testable implementation that accepts an explicit magicDir.
func CheckMagicDuplicatesInDir(logger *cryptoutilCmdCicdCommon.Logger, magicDir string) error {
	logger.Log("Checking magic package for duplicate constant values...")

	inv, err := lintGoCommon.ParseMagicDir(magicDir)
	if err != nil {
		return fmt.Errorf("failed to parse magic package: %w", err)
	}

	type dupGroup struct {
		value  string
		consts []lintGoCommon.MagicConstant
	}

	var groups []dupGroup

	for value, consts := range inv.ByValue {
		if len(consts) < 2 {
			continue
		}

		group := dupGroup{value: value, consts: append([]lintGoCommon.MagicConstant(nil), consts...)}
		sort.Slice(group.consts, func(i, j int) bool {
			if group.consts[i].File != group.consts[j].File {
				return group.consts[i].File < group.consts[j].File
			}

			return group.consts[i].Name < group.consts[j].Name
		})

		groups = append(groups, group)
	}

	if len(groups) == 0 {
		logger.Log("✅ magic-duplicates: no duplicate values found")

		return nil
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].value < groups[j].value })

	var sb strings.Builder

	fmt.Fprintf(&sb, "magic-duplicates: %d duplicate value group(s) found (informational — fix incrementally)\n\n", len(groups))

	for _, g := range groups {
		fmt.Fprintf(&sb, "  value %s defined %d times:\n", g.value, len(g.consts))

		for _, mc := range g.consts {
			fmt.Fprintf(&sb, "    %s:%d  %s\n", mc.File, mc.Line, mc.Name)
		}

		fmt.Fprint(&sb, "\n")
	}

	// magic-duplicates is informational: it logs violations but does not block CI.
	// The magic package has accumulated duplicate values that need incremental cleanup.
	// Run 'cicd lint-go' regularly to measure progress.
	logger.Log(sb.String())

	return nil
}

// CheckCrossFileDuplicatesInDir walks rootDir for Go files outside the magic
// package and reports constants whose literal string values are shared across
// two or more distinct files. These are candidates for consolidation into the
// magic package.
func CheckCrossFileDuplicatesInDir(logger *cryptoutilCmdCicdCommon.Logger, magicDir, rootDir string) error {
	logger.Log("Checking for constant values duplicated across multiple files outside the magic package...")

	absMagicDir, err := magicDuplicatesAbsFn(magicDir)
	if err != nil {
		return fmt.Errorf("cannot resolve magic dir: %w", err)
	}

	absRootDir, err := magicDuplicatesAbsFn(rootDir)
	if err != nil {
		return fmt.Errorf("cannot resolve root dir: %w", err)
	}

	byValue := make(map[string][]crossFileConstant)

	var walkErrors []string

	walkErr := magicDuplicatesWalkFn(absRootDir, func(path string, info os.FileInfo, walkFileErr error) error {
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
		if lintGoCommon.MagicShouldSkipPath(relPath) {
			return nil
		}

		collectConstsFromFile(path, relPath, byValue)

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("directory walk failed: %w", walkErr)
	}

	if len(walkErrors) > 0 {
		return fmt.Errorf("walk errors: %s", strings.Join(walkErrors, "; "))
	}

	type crossDupGroup struct {
		value  string
		consts []crossFileConstant
	}

	var groups []crossDupGroup

	for value, consts := range byValue {
		if countDistinctFiles(consts) < 2 {
			continue
		}

		group := crossDupGroup{value: value, consts: append([]crossFileConstant(nil), consts...)}
		sort.Slice(group.consts, func(i, j int) bool {
			if group.consts[i].File != group.consts[j].File {
				return group.consts[i].File < group.consts[j].File
			}

			return group.consts[i].Line < group.consts[j].Line
		})

		groups = append(groups, group)
	}

	if len(groups) == 0 {
		logger.Log("✅ magic-cross-duplicates: no cross-file duplicate constant values found")

		return nil
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].value < groups[j].value })

	var sb strings.Builder

	fmt.Fprintf(&sb, "magic-cross-duplicates: %d cross-file duplicate group(s) found (informational — consolidate into magic package)\n\n", len(groups))

	for _, g := range groups {
		fileCount := countDistinctFiles(g.consts)
		fmt.Fprintf(&sb, "  value %s shared as const in %d files:\n", g.value, fileCount)

		for _, c := range g.consts {
			fmt.Fprintf(&sb, "    %s:%d  %s\n", c.File, c.Line, c.Name)
		}

		fmt.Fprint(&sb, "\n")
	}

	// magic-cross-duplicates is informational: it logs violations but does not block CI.
	// Run 'cicd lint-go' to measure progress as cross-file duplicates are consolidated.
	logger.Log(sb.String())

	return nil
}

// collectConstsFromFile parses a Go source file and adds all const string
// literal declarations to the byValue map.
func collectConstsFromFile(absPath, relPath string, byValue map[string][]crossFileConstant) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		return // skip unparseable files silently.
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, value := range valSpec.Values {
				basicLit, ok := value.(*ast.BasicLit)
				if !ok || basicLit.Kind != token.STRING {
					continue
				}

				if i >= len(valSpec.Names) {
					continue
				}

				name := valSpec.Names[i].Name
				line := fset.Position(basicLit.Pos()).Line
				byValue[basicLit.Value] = append(byValue[basicLit.Value], crossFileConstant{
					File:  relPath,
					Line:  line,
					Name:  name,
					Value: basicLit.Value,
				})
			}
		}
	}
}

// countDistinctFiles returns how many distinct File values appear in consts.
func countDistinctFiles(consts []crossFileConstant) int {
	files := make(map[string]struct{})

	for _, c := range consts {
		files[c.File] = struct{}{}
	}

	return len(files)
}
