// Copyright (c) 2025 Justin Cranford

// Package duplicate_ps_id_test_func_names detects test function names that appear in
// multiple PS-ID server packages, indicating boilerplate that should be extracted into
// a parameterised framework helper instead of being copied across services.
//
// A test function is reported when it appears in ≥ DuplicateThreshold PS-ID server
// packages. Results are ranked by occurrence count (worst offenders first).
//
// This linter is informational by default (threshold = 3). Setting DuplicateThreshold
// to a lower value tightens the constraint. The output helps prioritise migration work.
package duplicate_ps_id_test_func_names

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DuplicateThreshold is the minimum number of PS-ID server packages a test function
// must appear in before it is reported as a duplication candidate.
const DuplicateThreshold = 3

// FuncOccurrence records which PS-IDs contain a given test function name.
type FuncOccurrence struct {
	FuncName string
	PSIDs    []string
}

// Check runs the linter from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir scans PS-ID server test packages under rootDir and reports duplicated test
// function names that appear in ≥ DuplicateThreshold packages.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for test function names duplicated across PS-ID server packages...")

	occurrences, err := FindDuplicates(rootDir)
	if err != nil {
		return fmt.Errorf("duplicate-ps-id-test-func-names: scan failed: %w", err)
	}

	if len(occurrences) == 0 {
		logger.LogWithPrefix("duplicate-ps-id-test-func-names", "✅ No duplicated test function names found at threshold ≥3")

		return nil
	}

	_, _ = fmt.Fprintf(os.Stdout,
		"duplicate-ps-id-test-func-names: %d test function(s) appear in ≥%d PS-ID server packages (ranked worst-first):\n",
		len(occurrences), DuplicateThreshold)

	for _, o := range occurrences {
		_, _ = fmt.Fprintf(os.Stdout, "  [%d PS-IDs] %s\n    in: %s\n",
			len(o.PSIDs), o.FuncName, strings.Join(o.PSIDs, ", "))
	}

	_, _ = fmt.Fprintln(os.Stdout, "  Suggestion: extract these into parameterised framework helpers in")
	_, _ = fmt.Fprintln(os.Stdout, "  internal/apps/framework/service/testing/ and call them from each PS-ID.")

	return nil
}

// FindDuplicates returns test function names (sorted by descending occurrence count)
// that appear in ≥ DuplicateThreshold PS-ID server packages under rootDir.
func FindDuplicates(rootDir string) ([]FuncOccurrence, error) {
	// funcToPSIDs maps test function name → set of PS-ID dirs that contain it.
	funcToPSIDs := make(map[string]map[string]bool)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch d.Name() {
			case cryptoutilSharedMagic.CICDExcludeDirVendor, cryptoutilSharedMagic.CICDExcludeDirGit:
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		normalized := filepath.ToSlash(path)
		psID, ok := extractPSIDFromServerTestPath(normalized)

		if !ok {
			return nil
		}

		names, err := collectTestFuncNames(path)
		if err != nil {
			return err
		}

		for _, name := range names {
			if funcToPSIDs[name] == nil {
				funcToPSIDs[name] = make(map[string]bool)
			}

			funcToPSIDs[name][psID] = true
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", rootDir, err)
	}

	var result []FuncOccurrence

	for funcName, psIDs := range funcToPSIDs {
		if len(psIDs) < DuplicateThreshold {
			continue
		}

		ids := make([]string, 0, len(psIDs))

		for id := range psIDs {
			ids = append(ids, id)
		}

		sort.Strings(ids)

		result = append(result, FuncOccurrence{FuncName: funcName, PSIDs: ids})
	}

	// Rank by descending occurrence count, then alphabetically by function name.
	sort.Slice(result, func(i, j int) bool {
		if len(result[i].PSIDs) != len(result[j].PSIDs) {
			return len(result[i].PSIDs) > len(result[j].PSIDs)
		}

		return result[i].FuncName < result[j].FuncName
	})

	return result, nil
}

// extractPSIDFromServerTestPath returns the PS-ID for test files under
// internal/apps/{ps-id}/server/ (non-framework, non-tools, non-template).
// Returns ("", false) for files that should be skipped.
func extractPSIDFromServerTestPath(normalized string) (string, bool) {
	const prefix = "internal/apps/"

	idx := strings.Index(normalized, prefix)
	if idx < 0 {
		return "", false
	}

	rest := normalized[idx+len(prefix):]
	slash := strings.Index(rest, "/")

	if slash < 0 {
		return "", false
	}

	psID := rest[:slash]

	switch psID {
	case "framework", "tools", "template":
		return "", false
	}

	// Only consider files directly under server/ (not nested sub-packages like server/handler/).
	// after == "server/<filename>" — there must be exactly one slash (after "server/").
	after := rest[slash+1:]
	if !strings.HasPrefix(after, "server/") {
		return "", false
	}

	afterServer := after[len("server/"):]
	if strings.Contains(afterServer, "/") {
		return "", false
	}

	// Skip integration and e2e test files — those test different concerns.
	base := filepath.Base(normalized)
	if strings.HasSuffix(base, "_integration_test.go") || strings.HasSuffix(base, "_e2e_test.go") {
		return "", false
	}

	return psID, true
}

// collectTestFuncNames parses a test file and returns all Test* top-level function names.
func collectTestFuncNames(filePath string) ([]string, error) {
	fset := token.NewFileSet()

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	f, err := parser.ParseFile(fset, filePath, src, 0)
	if err != nil {
		return nil, nil //nolint:nilerr // skip unparseable files silently
	}

	var names []string

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}

		name := fn.Name.Name
		if strings.HasPrefix(name, "Test") {
			names = append(names, name)
		}
	}

	return names, nil
}
