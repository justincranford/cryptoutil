// Copyright (c) 2025 Justin Cranford

// Package import_alias_formula verifies that all Go source files use canonical import aliases
// for the packages listed in alias_map.yaml. The canonical aliases are defined once in the
// YAML and enforced here as a fitness linter - acting as the single source of truth for
// the project's import alias convention, which is also reflected in .golangci.yml.
//
// Violations are reported per file and line; the linter exits non-zero when any are found.
package import_alias_formula

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

// =========================================================================
// Types
// =========================================================================

// AliasEntry maps one import path to its required alias.
type AliasEntry struct {
	ImportPath string `yaml:"import_path"`
	Alias      string `yaml:"alias"`
}

// AliasMap is the top-level YAML structure for alias_map.yaml.
type AliasMap struct {
	ExternalAliases []AliasEntry `yaml:"external_aliases"`
	InternalAliases []AliasEntry `yaml:"internal_aliases"`
}

// =========================================================================
// Test seams - replaceable in tests to exercise OS-level error paths.
// See ARCHITECTURE.md Section 10.2.4 (Test Seam Injection Pattern).
// =========================================================================

var importAliasReadFileFn = os.ReadFile
var importAliasWalkDirFn = filepath.WalkDir
var importAliasGetwdFn = os.Getwd
var findImportAliasProjectRootFn = findImportAliasProjectRoot

// =========================================================================
// Public API
// =========================================================================

// ExcludedDirs lists directory names that are never walked.
var ExcludedDirs = map[string]bool{
	cryptoutilSharedMagic.CICDExcludeDirVendor: true,
	cryptoutilSharedMagic.CICDExcludeDirGit:    true,
}

// codeGeneratedCheckBytes is the number of bytes to scan for the "Code generated" marker.
const codeGeneratedCheckBytes = 512

// LoadAliasMap reads and parses the alias_map.yaml from rootDir.
func LoadAliasMap(rootDir string) (*AliasMap, error) {
	yamlPath := filepath.Join(rootDir, filepath.FromSlash(cryptoutilSharedMagic.CICDImportAliasMapFile))

	data, err := importAliasReadFileFn(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("import-alias-formula: failed to read %s: %w", yamlPath, err)
	}

	var m AliasMap

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err = dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("import-alias-formula: failed to parse %s: %w", yamlPath, err)
	}

	return &m, nil
}

// AllEntries returns the combined list of external + internal alias entries.
func AllEntries(m *AliasMap) []AliasEntry {
	combined := make([]AliasEntry, 0, len(m.ExternalAliases)+len(m.InternalAliases))
	combined = append(combined, m.ExternalAliases...)
	combined = append(combined, m.InternalAliases...)

	return combined
}

// Check runs the linter from the current working directory (project root).
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	rootDir, err := findImportAliasProjectRootFn()
	if err != nil {
		return fmt.Errorf("import-alias-formula: %w", err)
	}

	return CheckInDir(logger, rootDir)
}

// CheckInDir runs the linter from rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Enforcing import alias formula...")

	aliasMap, err := LoadAliasMap(rootDir)
	if err != nil {
		return fmt.Errorf("import-alias-formula: %w", err)
	}

	entries := AllEntries(aliasMap)
	if len(entries) == 0 {
		logger.Log("import-alias-formula: alias map is empty - skipping")

		return nil
	}

	// Build a path->alias lookup.
	required := make(map[string]string, len(entries))

	for _, e := range entries {
		required[e.ImportPath] = e.Alias
	}

	var violations []string

	var goFiles []string

	walkErr := importAliasWalkDirFn(rootDir, func(path string, d fs.DirEntry, walkFileErr error) error {
		if walkFileErr != nil {
			return walkFileErr
		}

		if d.IsDir() {
			if ExcludedDirs[d.Name()] {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") && !isGeneratedGoFile(path) {
			goFiles = append(goFiles, path)
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("import-alias-formula: failed to walk %s: %w", rootDir, walkErr)
	}

	for _, goFile := range goFiles {
		fileViolations, checkErr := checkFile(goFile, required)
		if checkErr != nil {
			return checkErr
		}

		violations = append(violations, fileViolations...)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			logger.Log(v)
		}

		return fmt.Errorf("import-alias-formula: %d violation(s) found", len(violations))
	}

	logger.Log(fmt.Sprintf("import-alias-formula: all %d Go files pass alias rules", len(goFiles)))

	return nil
}

// =========================================================================
// Internal helpers
// =========================================================================

// isGeneratedGoFile returns true when the first 512 bytes of a file contain the
// standard "Code generated" marker. Generated files are excluded from alias checks.
func isGeneratedGoFile(path string) bool {
	data, err := importAliasReadFileFn(path)
	if err != nil {
		return false
	}

	limit := len(data)
	if limit > codeGeneratedCheckBytes {
		limit = codeGeneratedCheckBytes
	}

	return bytes.Contains(data[:limit], []byte("Code generated"))
}

// checkFile uses the Go AST parser to scan import declarations for alias violations.
// Using the AST parser avoids false positives from raw string literals.
func checkFile(path string, required map[string]string) ([]string, error) {
	data, err := importAliasReadFileFn(path)
	if err != nil {
		return nil, fmt.Errorf("import-alias-formula: failed to read %s: %w", path, err)
	}

	fset := token.NewFileSet()

	f, parseErr := parser.ParseFile(fset, path, data, parser.ImportsOnly)
	if parseErr != nil {
		// If the file does not parse (e.g. it is a test fixture with intentionally
		// malformed imports), skip it silently.
		return nil, nil //nolint:nilerr // deliberate: non-parseable files are skipped
	}

	var violations []string

	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		expectedAlias, ok := required[importPath]
		if !ok {
			continue
		}

		actualAlias := ""
		if imp.Name != nil {
			actualAlias = imp.Name.Name
		}

		// Blank identifier and dot imports are always permitted.
		if actualAlias == "_" || actualAlias == "." {
			continue
		}

		if actualAlias == expectedAlias {
			continue
		}

		pos := fset.Position(imp.Path.Pos())

		if actualAlias == "" {
			violations = append(violations, fmt.Sprintf("%s:%d  %q imported without alias (want: %s)", path, pos.Line, importPath, expectedAlias))
		} else {
			violations = append(violations, fmt.Sprintf("%s:%d  %q aliased as %q (want: %s)", path, pos.Line, importPath, actualAlias, expectedAlias))
		}
	}

	return violations, nil
}

// findImportAliasProjectRoot walks up from cwd until go.mod is found.
func findImportAliasProjectRoot() (string, error) {
	cwd, err := importAliasGetwdFn()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	dir := cwd

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", cwd)
		}

		dir = parent
	}
}
