// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Test constants for repeated string literals.
const (
	osWindows          = "windows"
	testCleanGoFile    = "clean.go"
	testCleanContent   = "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\") }\n"
	testMainContent    = "package main\n\nfunc main() {}\n"
	testPackageMainDef = "package main\n"
)

func TestCheckDependencies_NoCycles(t *testing.T) {
	t.Parallel()

	// Simulate go list -json output with no cycles.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "Should not detect cycles in acyclic graph")
}

func TestCheckDependencies_WithCycle(t *testing.T) {
	t.Parallel()

	// Simulate go list -json output with a cycle: a -> b -> c -> a.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err, "Should detect cycle")
	require.Contains(t, err.Error(), "circular dependency", "Error should mention circular dependency")
}

func TestCheckDependencies_EmptyOutput(t *testing.T) {
	t.Parallel()

	err := CheckDependencies("")
	require.NoError(t, err, "Empty output should not cause error")
}

func TestCheckDependencies_SinglePackage(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "Single package with no imports should not cause error")
}

func TestCheckDependencies_ExternalDepsIgnored(t *testing.T) {
	t.Parallel()

	// External dependencies should be ignored.
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["github.com/external/pkg", "example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["fmt", "github.com/another/pkg"]}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "External dependencies should be ignored")
}

func TestGetModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages map[string][]string
		expected string
	}{
		{
			name:     "empty packages",
			packages: map[string][]string{},
			expected: "",
		},
		{
			name: "single package",
			packages: map[string][]string{
				"example.com/pkg/a": {},
			},
			expected: "example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := getModulePath(tc.packages)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestLoadCircularDepCache_FileNotFound tests loading cache when file doesn't exist.
func TestLoadCircularDepCache_FileNotFound(t *testing.T) {
	t.Parallel()

	cache, err := loadCircularDepCache("nonexistent-file-12345.json")
	require.Error(t, err)
	require.Nil(t, cache)
	require.Contains(t, err.Error(), "failed to read cache file")
}

// TestLoadCircularDepCache_InvalidJSON tests loading cache with malformed JSON.
func TestLoadCircularDepCache_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Create temp file with invalid JSON.
	tmpFile := filepath.Join(t.TempDir(), "invalid-cache.json")
	err := os.WriteFile(tmpFile, []byte("{invalid json}"), 0o600)
	require.NoError(t, err)

	cache, err := loadCircularDepCache(tmpFile)
	require.Error(t, err)
	require.Nil(t, cache)
	require.Contains(t, err.Error(), "failed to unmarshal cache JSON")
}

// TestSaveLoadCircularDepCache_RoundTrip tests save/load cycle.
func TestSaveLoadCircularDepCache_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "cache.json")

	// Create cache data.
	original := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC().Truncate(time.Second), // Truncate to avoid precision loss.
		GoModModTime:    time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second),
		HasCircularDeps: true,
		CircularDeps:    []string{"pkg/a -> pkg/b -> pkg/a"},
	}

	// Save cache.
	err := saveCircularDepCache(tmpFile, original)
	require.NoError(t, err)

	// Verify file exists.
	_, err = os.Stat(tmpFile)
	require.NoError(t, err)

	// Load cache.
	loaded, err := loadCircularDepCache(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify data matches.
	require.Equal(t, original.HasCircularDeps, loaded.HasCircularDeps)
	require.Equal(t, original.CircularDeps, loaded.CircularDeps)
	require.True(t, original.LastCheck.Equal(loaded.LastCheck))
	require.True(t, original.GoModModTime.Equal(loaded.GoModModTime))
}

// TestSaveCircularDepCache_CreateDirectory tests directory creation.
func TestSaveCircularDepCache_CreateDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "subdir", "cache.json")

	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Verify file and directory exist.
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)
}

func TestLint(t *testing.T) {
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger)

	require.Error(t, err, "Lint fails when go.mod not in current directory")
	require.Contains(t, err.Error(), "lint-go failed")
}

func TestCheckDependencies_MalformedJSON(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
invalid json line
{"ImportPath": "example.com/pkg/b", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode package info")
}

func TestCheckDependencies_ComplexCycle(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b", "example.com/pkg/c"]}
{"ImportPath": "example.com/pkg/b", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency")
}

func TestCheckDependencies_SelfReference(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/a"]}`

	err := CheckDependencies(goListOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency")
}

func TestCheckDependencies_MultipleDisconnectedGraphs(t *testing.T) {
	t.Parallel()

	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": []}
{"ImportPath": "example.com/pkg/c", "Imports": ["example.com/pkg/d"]}
{"ImportPath": "example.com/pkg/d", "Imports": []}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err)
}

func TestCheckDependencies_MixedModulePrefixes(t *testing.T) {
	t.Parallel()

	// Test with packages from different module prefixes.
	// Only the "example.com" packages should be checked for cycles.
	// The "other.org" package should be skipped (line 183-184).
	goListOutput := `{"ImportPath": "example.com/pkg/a", "Imports": ["example.com/pkg/b"]}
{"ImportPath": "example.com/pkg/b", "Imports": []}
{"ImportPath": "other.org/different/pkg", "Imports": ["other.org/different/another"]}`

	err := CheckDependencies(goListOutput)
	require.NoError(t, err, "Packages from different module prefixes should be handled")
}

func TestGetModulePath_MultiplePackages(t *testing.T) {
	t.Parallel()

	packages := map[string][]string{
		"example.com/pkg/a": {},
		"example.com/pkg/b": {},
		"example.com/pkg/c": {},
	}

	result := getModulePath(packages)
	require.Equal(t, "example.com", result)
}

func TestGetModulePath_DifferentPrefixes(t *testing.T) {
	t.Parallel()

	packages := map[string][]string{
		"github.com/user/repo/pkg/a": {},
		"github.com/user/repo/pkg/b": {},
	}

	result := getModulePath(packages)
	require.Equal(t, "github.com", result)
}

func TestCheckGoModForCGO_ValidFile(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file without banned modules.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	modernc.org/sqlite v1.29.0
	github.com/golang-migrate/migrate/v4 v4.17.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoModForCGO(goModFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Valid go.mod should have no violations")
}

func TestCheckGoModForCGO_BannedModule(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file with banned module (direct dependency).
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	github.com/mattn/go-sqlite3 v1.14.19
	github.com/golang-migrate/migrate/v4/database/sqlite3 v4.17.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoModForCGO(goModFile)
	require.NoError(t, err)
	require.Len(t, violations, 2, "Should detect 2 banned modules")
	require.Contains(t, strings.Join(violations, "\n"), "go-sqlite3", "Should detect banned CGO sqlite")
	require.Contains(t, strings.Join(violations, "\n"), "database/sqlite3", "Should detect banned CGO migrate")
}

func TestCheckGoModForCGO_IndirectModule(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file with banned module as indirect.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	github.com/mattn/go-sqlite3 v1.14.19 // indirect
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoModForCGO(goModFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Indirect dependencies should not be flagged")
}

func TestCheckGoModForCGO_FileNotFound(t *testing.T) {
	t.Parallel()

	violations, err := checkGoModForCGO("/nonexistent/path/go.mod")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to open go.mod")
}

func TestCheckRequiredCGOModule_Found(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file with required module.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	modernc.org/sqlite v1.29.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	found, err := checkRequiredCGOModule(goModFile)
	require.NoError(t, err)
	require.True(t, found, "Required module should be found")
}

func TestCheckRequiredCGOModule_NotFound(t *testing.T) {
	t.Parallel()

	// Create temp go.mod file without required module.
	tmpDir := t.TempDir()
	goModFile := filepath.Join(tmpDir, "go.mod")

	content := `module example.com/myproject

go 1.21

require (
	github.com/some/other/module v1.0.0
)
`

	err := os.WriteFile(goModFile, []byte(content), 0o600)
	require.NoError(t, err)

	found, err := checkRequiredCGOModule(goModFile)
	require.NoError(t, err)
	require.False(t, found, "Required module should not be found")
}

func TestCheckRequiredCGOModule_FileNotFound(t *testing.T) {
	t.Parallel()

	found, err := checkRequiredCGOModule("/nonexistent/path/go.mod")
	require.Error(t, err)
	require.False(t, found)
	require.Contains(t, err.Error(), "failed to open go.mod")
}

func TestCheckGoFileForCGO_Clean(t *testing.T) {
	t.Parallel()

	// Create temp file without banned imports.
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "clean.go")

	content := `package main

import (
	"modernc.org/sqlite"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
)

func main() {
	// Using CGO-free sqlite
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoFileForCGO(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Clean file should have no violations")
}

func TestCheckGoFileForCGO_BannedImport(t *testing.T) {
	t.Parallel()

	// Create temp file with banned import.
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "banned.go")

	content := `package main

import (
	_ "github.com/mattn/go-sqlite3"
)

func main() {
}
`

	err := os.WriteFile(bannedFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoFileForCGO(bannedFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Banned import should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "banned CGO import detected")
}

func TestCheckGoFileForCGO_BannedMigrateImport(t *testing.T) {
	t.Parallel()

	// Create temp file with banned migrate import.
	tmpDir := t.TempDir()
	bannedFile := filepath.Join(tmpDir, "banned_migrate.go")

	content := `package main

import (
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

func main() {
}
`

	err := os.WriteFile(bannedFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoFileForCGO(bannedFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Banned migrate import should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "banned CGO migrate import detected")
}

func TestCheckGoFileForCGO_LintGoSkipped(t *testing.T) {
	t.Parallel()

	// Create temp file in a lint_go directory (should be skipped).
	tmpDir := t.TempDir()
	lintGoDir := filepath.Join(tmpDir, "lint_go")
	require.NoError(t, os.MkdirAll(lintGoDir, 0o755))

	skippedFile := filepath.Join(lintGoDir, "lint_go.go")

	// Even with banned imports, should be skipped.
	content := `package main

import (
	_ "github.com/mattn/go-sqlite3"
)
`

	err := os.WriteFile(skippedFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoFileForCGO(skippedFile)
	require.NoError(t, err)
	require.Empty(t, violations, "lint_go files should be skipped")
}

func TestCheckGoFileForCGO_FileNotFound(t *testing.T) {
	t.Parallel()

	violations, err := checkGoFileForCGO("/nonexistent/path/file.go")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to open")
}

func TestPrintCGOViolations_AllTypes(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.

	// Capture stderr to verify output.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	goModViolations := []string{"go.mod:5: banned CGO module"}
	importViolations := []string{"file.go:10: banned CGO import"}
	hasRequired := false

	printCGOViolations(goModViolations, importViolations, hasRequired)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.Contains(t, string(output), "CGO validation failed")
	require.Contains(t, string(output), "go.mod violations")
	require.Contains(t, string(output), "Import violations")
	require.Contains(t, string(output), "Required module missing")
}

func TestPrintCGOViolations_GoModOnly(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	printCGOViolations([]string{"go.mod:5: banned module"}, nil, true)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.Contains(t, string(output), "go.mod violations")
	require.NotContains(t, string(output), "Import violations")
	require.NotContains(t, string(output), "Required module missing")
}

func TestPrintCGOViolations_ImportOnly(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	printCGOViolations(nil, []string{"file.go:10: banned import"}, true)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.NotContains(t, string(output), "go.mod violations")
	require.Contains(t, string(output), "Import violations")
	require.NotContains(t, string(output), "Required module missing")
}

func TestCheckGoFileForUnaliasedCryptoutilImports_Clean(t *testing.T) {
	t.Parallel()

	// Create temp file with properly aliased imports.
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "clean.go")

	content := `package main

import (
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilCommon "cryptoutil/internal/apps/cicd/common"
)

func main() {
	_ = cryptoutilMagic.TestValue
	_ = cryptoutilCommon.NewLogger("test")
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGoFileForUnaliasedCryptoutilImports(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Properly aliased imports should have no violations")
}

func TestCheckGoFileForUnaliasedCryptoutilImports_Unaliased(t *testing.T) {
	t.Parallel()

	// Create temp file with unaliased cryptoutil import.
	// Using raw string builder to avoid linter flagging this test file.
	tmpDir := t.TempDir()
	unaliasedFile := filepath.Join(tmpDir, "unaliased.go")

	// Build content dynamically to avoid false positive from import checker.
	var content strings.Builder
	content.WriteString("package main\n\nimport (\n\t\"")
	content.WriteString("cryptoutil/internal/shared/magic")
	content.WriteString("\"\n)\n\nfunc main() {\n\t_ = magic.TestValue\n}\n")

	err := os.WriteFile(unaliasedFile, []byte(content.String()), 0o600)
	require.NoError(t, err)

	violations, err := checkGoFileForUnaliasedCryptoutilImports(unaliasedFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Unaliased cryptoutil import should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "unaliased cryptoutil import detected")
}

func TestCheckGoFileForUnaliasedCryptoutilImports_SingleLineImport(t *testing.T) {
	t.Parallel()

	// Create temp file with single-line unaliased import.
	// Using raw string builder to avoid linter flagging this test file.
	tmpDir := t.TempDir()
	singleLineFile := filepath.Join(tmpDir, "singleline.go")

	// Build content dynamically to avoid false positive from import checker.
	var content strings.Builder
	content.WriteString("package main\n\nimport \"")
	content.WriteString("cryptoutil/internal/shared/magic")
	content.WriteString("\"\n\nfunc main() {\n}\n")

	err := os.WriteFile(singleLineFile, []byte(content.String()), 0o600)
	require.NoError(t, err)

	violations, err := checkGoFileForUnaliasedCryptoutilImports(singleLineFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Single-line unaliased import should be detected")
}

func TestCheckGoFileForUnaliasedCryptoutilImports_FileNotFound(t *testing.T) {
	t.Parallel()

	violations, err := checkGoFileForUnaliasedCryptoutilImports("/nonexistent/path/file.go")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to open")
}

func TestPrintCryptoutilImportViolations(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	violations := []string{
		"file1.go:5: unaliased cryptoutil import detected",
		"file2.go:10: unaliased cryptoutil import detected",
	}

	printCryptoutilImportViolations(violations)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.Contains(t, string(output), "Unaliased cryptoutil imports found")
	require.Contains(t, string(output), "file1.go")
	require.Contains(t, string(output), "file2.go")
	require.Contains(t, string(output), "golangci-lint run --fix")
}

// findProjectRoot finds the project root by looking for go.mod.
func findProjectRoot() (string, error) {
	// Start from current directory and walk up.
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

func TestCheckCircularDeps_Integration(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - modifies shared cache file and changes working directory.

	// Find and change to project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Remove cache file to force fresh check.
	cacheFile := cryptoutilSharedMagic.CircularDepCacheFileName
	_ = os.Remove(cacheFile)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// First call: cache miss - performs actual check.
	err = checkCircularDeps(logger)
	require.NoError(t, err, "Project should have no circular dependencies")

	// Second call: cache hit - uses cached result.
	err = checkCircularDeps(logger)
	require.NoError(t, err, "Cached result should indicate no circular dependencies")

	// Clean up cache file.
	_ = os.Remove(cacheFile)
}

func TestCheckCircularDeps_CachedWithCircularDeps(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - modifies shared cache file and changes working directory.

	// Find and change to project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	cacheFile := cryptoutilSharedMagic.CircularDepCacheFileName

	// Create cache indicating circular deps exist.
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)

	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    goModStat.ModTime(),
		HasCircularDeps: true,
		CircularDeps:    []string{"pkg/a -> pkg/b -> pkg/a"},
	}

	err = saveCircularDepCache(cacheFile, cache)
	require.NoError(t, err)

	defer func() {
		_ = os.Remove(cacheFile)
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// Call should use cached result and return error.
	err = checkCircularDeps(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependencies detected (cached)")
}

func TestCheckCircularDeps_ExpiredCache(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - modifies shared cache file and changes working directory.

	// Find and change to project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	cacheFile := cryptoutilSharedMagic.CircularDepCacheFileName

	// Create expired cache.
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)

	// Set LastCheck to be older than cache validity duration.
	expiredTime := time.Now().UTC().Add(-cryptoutilSharedMagic.CircularDepCacheValidDuration - time.Hour)
	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       expiredTime,
		GoModModTime:    goModStat.ModTime(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err = saveCircularDepCache(cacheFile, cache)
	require.NoError(t, err)

	defer func() {
		_ = os.Remove(cacheFile)
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// Call should detect expired cache and perform fresh check.
	err = checkCircularDeps(logger)
	require.NoError(t, err, "Fresh check should pass (project has no circular deps)")
}

func TestCheckCircularDeps_GoModChanged(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - modifies shared cache file and changes working directory.

	// Find and change to project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	cacheFile := cryptoutilSharedMagic.CircularDepCacheFileName

	// Create cache with old go.mod mod time.
	oldModTime := time.Now().UTC().Add(-time.Hour * 24 * 365) // 1 year ago
	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    oldModTime, // go.mod was "modified" since cache was created
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err = saveCircularDepCache(cacheFile, cache)
	require.NoError(t, err)

	defer func() {
		_ = os.Remove(cacheFile)
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// Call should detect go.mod change and perform fresh check.
	err = checkCircularDeps(logger)
	require.NoError(t, err, "Fresh check should pass (project has no circular deps)")
}

func TestCheckCircularDeps_GoListError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create a temp directory with a go.mod that references a nonexistent module.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create go.mod with a require for a nonexistent module.
	goModContent := `module testmodule

go 1.21

require nonexistent.example.com/fake/module v999.999.999
`
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

	// Create a Go file that imports the nonexistent module.
	goFileContent := `package main

import "nonexistent.example.com/fake/module"

func main() { module.Do() }
`
	require.NoError(t, os.WriteFile("main.go", []byte(goFileContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// Call should fail because go list will fail on missing module.
	err = checkCircularDeps(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to run go list")
}

func TestCheckCircularDeps_FreshCheckWithActualCircularDeps(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory with actual circular dependencies.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create go.mod.
	goModContent := "module testcircular\n\ngo 1.21\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

	// Create package a that imports package b.
	require.NoError(t, os.MkdirAll("internal/a", 0o755))

	pkgAContent := `package a

import "testcircular/internal/b"

func A() { b.B() }
`
	require.NoError(t, os.WriteFile("internal/a/a.go", []byte(pkgAContent), 0o600))

	// Create package b that imports package a (circular!).
	require.NoError(t, os.MkdirAll("internal/b", 0o755))

	pkgBContent := `package b

import "testcircular/internal/a"

func B() { a.A() }
`
	require.NoError(t, os.WriteFile("internal/b/b.go", []byte(pkgBContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-circulardeps")

	// Call should detect circular dependencies during fresh check.
	// Note: go list may fail with import cycle error before CheckDependencies runs,
	// so we accept either "failed to run go list" or "circular dependency".
	err = checkCircularDeps(logger)
	require.Error(t, err)
	// The Go toolchain detects import cycles at compile time, so go list fails.
	require.True(t, strings.Contains(err.Error(), "failed to run go list") ||
		strings.Contains(err.Error(), "circular"),
		"Expected error about go list failure or circular deps, got: %v", err)
}

func TestCheckGoFilesForCGO_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory with test files.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create clean Go file.
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), 0o600))

	// Test with clean file - should have no violations.
	violations, err := checkGoFilesForCGO()
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckGoFilesForCGO_WithBannedImport(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory with test files.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Build banned import string dynamically to avoid self-flagging.
	var banned strings.Builder
	banned.WriteString("github.com/")
	banned.WriteString("mattn/go-sqlite3")

	// Create file with banned CGO import.
	bannedFile := "banned.go"
	bannedContent := "package main\n\nimport _ \"" + banned.String() + "\"\n\nfunc main() {}\n"

	require.NoError(t, os.WriteFile(bannedFile, []byte(bannedContent), 0o600))

	// Test - should find the violation.
	violations, err := checkGoFilesForCGO()
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "banned.go")
}

func TestCheckGoFilesForCGO_SkipsVendor(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create vendor directory with file that would be flagged.
	require.NoError(t, os.MkdirAll("vendor", 0o755))

	var banned strings.Builder
	banned.WriteString("github.com/")
	banned.WriteString("mattn/go-sqlite3")

	vendorFile := "vendor/dep.go"
	vendorContent := "package vendor\n\nimport _ \"" + banned.String() + "\"\n\nfunc init() {}\n"

	require.NoError(t, os.WriteFile(vendorFile, []byte(vendorContent), 0o600))

	// Create clean main file.
	mainFile := "main.go"

	require.NoError(t, os.WriteFile(mainFile, []byte(testMainContent), 0o600))

	// Test - vendor should be skipped, no violations.
	violations, err := checkGoFilesForCGO()
	require.NoError(t, err)
	require.Empty(t, violations, "vendor directory should be skipped")
}

func TestFindUnaliasedCryptoutilImports_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create clean Go file with no cryptoutil imports.
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), 0o600))

	// Test - should have no violations.
	violations, err := findUnaliasedCryptoutilImports()
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindGoFiles_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create Go files.
	require.NoError(t, os.WriteFile("main.go", []byte(testPackageMainDef), 0o600))
	require.NoError(t, os.WriteFile("util.go", []byte(testPackageMainDef), 0o600))
	require.NoError(t, os.WriteFile("main_test.go", []byte(testPackageMainDef), 0o600))

	// Create excluded directories.
	require.NoError(t, os.MkdirAll("vendor", 0o755))
	require.NoError(t, os.WriteFile("vendor/vendored.go", []byte("package vendor\n"), 0o600))

	// Test - should find main.go and util.go, but NOT test files, vendor files.
	files, err := findGoFiles()
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestCheckNoUnaliasedCryptoutilImports_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create clean Go file with no cryptoutil imports.
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with no violations.
	err = checkNoUnaliasedCryptoutilImports(logger)
	require.NoError(t, err)
}

func TestCheckCGOFreeSQLite_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Build required module string dynamically.
	var required strings.Builder
	required.WriteString("modernc.org/")
	required.WriteString("sqlite")

	// Create go.mod with required CGO-free module.
	goModContent := "module testmod\n\ngo 1.21\n\nrequire (\n\t" + required.String() + " v1.30.0\n)\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

	// Create clean Go file.
	require.NoError(t, os.WriteFile("main.go", []byte(testMainContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with required module present.
	err = checkCGOFreeSQLite(logger)
	require.NoError(t, err)
}

func TestCheckCGOFreeSQLite_MissingRequired(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create go.mod WITHOUT required CGO-free module.
	goModContent := "module testmod\n\ngo 1.21\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

	// Create clean Go file.
	require.NoError(t, os.WriteFile("main.go", []byte(testMainContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail because required module is missing.
	err = checkCGOFreeSQLite(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CGO validation failed")
}

func TestCheckNonFIPS_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create clean Go file without banned algorithms.
	cleanContent := "package main\n\nimport (\n\t\"crypto/sha256\"\n)\n\nfunc main() { sha256.New() }\n"
	require.NoError(t, os.WriteFile("main.go", []byte(cleanContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with FIPS-compliant code.
	err = checkNonFIPS(logger)
	require.NoError(t, err)
}

func TestSaveCircularDepCache_DirectoryCreationError(t *testing.T) {
	t.Parallel()

	// Create a cache object.
	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    nil,
	}

	// Try to save to an invalid path that can't be created.
	// Using a path that would require creating a directory on a file.
	tempFile, err := os.CreateTemp("", "notadir")
	require.NoError(t, err)

	defer func() { _ = os.Remove(tempFile.Name()) }()

	_ = tempFile.Close()

	// Try to save cache to a path inside the file (impossible).
	invalidPath := tempFile.Name() + "/cache.json"
	err = saveCircularDepCache(invalidPath, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create output directory")
}

func TestSaveCircularDepCache_WriteFileError(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test uses filesystem permissions.

	// Create a cache object.
	cache := cryptoutilSharedMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    nil,
	}

	// Create a temp directory that we can make read-only.
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))

	cacheFile := filepath.Join(cacheDir, "cache.json")

	// Create an existing file to write to.
	require.NoError(t, os.WriteFile(cacheFile, []byte("existing"), 0o600))

	// Make the cache file read-only.
	require.NoError(t, os.Chmod(cacheFile, 0o000))

	defer func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(cacheFile, 0o600)
	}()

	// Try to save - MkdirAll succeeds (dir exists) but WriteFile should fail.
	err := saveCircularDepCache(cacheFile, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write cache file")
}

func TestCheckNoUnaliasedCryptoutilImports_WithViolations(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory and redirects stderr.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Build cryptoutil import path dynamically to avoid self-flagging.
	var importPath strings.Builder
	importPath.WriteString("cryptoutil/")
	importPath.WriteString("internal/shared/magic")

	// Create Go file with unaliased cryptoutil import.
	badContent := "package main\n\nimport \"" + importPath.String() + "\"\n\nfunc main() {}\n"
	require.NoError(t, os.WriteFile("bad.go", []byte(badContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail with violations.
	err = checkNoUnaliasedCryptoutilImports(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unaliased cryptoutil imports")
}

func TestFindUnaliasedCryptoutilImports_ErrorPath(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == osWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create a file (not a directory) that will be treated as a directory during walk.
	// This will cause an error in filepath.Walk.
	require.NoError(t, os.WriteFile("main.go", []byte("package main\n"), 0o600))

	// Make main.go unreadable to trigger error.
	require.NoError(t, os.Chmod("main.go", 0o000))

	defer func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(filepath.Join(tempDir, "main.go"), 0o600)
	}()

	// Test - should get error from reading file.
	_, err = findUnaliasedCryptoutilImports()
	require.Error(t, err)
}

func TestCheckNonFIPS_WithViolations(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory and redirects stderr.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create Go file with banned algorithm (bcrypt).
	badContent := "package main\n\nimport \"golang.org/x/crypto/bcrypt\"\n\nfunc main() { bcrypt.GenerateFromPassword(nil, 0) }\n"
	require.NoError(t, os.WriteFile("bad.go", []byte(badContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail with violations.
	err = checkNonFIPS(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-FIPS algorithm violations")
}

func TestFindGoFiles_ErrorPath(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == osWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create a subdirectory that will trigger walk error.
	subDir := "subdir"
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "file.go"), []byte("package main\n"), 0o600))

	// Make subdirectory unreadable.
	require.NoError(t, os.Chmod(subDir, 0o000))

	defer func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(filepath.Join(tempDir, subDir), 0o755)
	}()

	// Test - should get error from walking directory.
	_, err = findGoFiles()
	require.Error(t, err)
}

func TestCheckGoFilesForCGO_ErrorPath(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == osWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create an unreadable Go file.
	require.NoError(t, os.WriteFile("unreadable.go", []byte("package main\n"), 0o600))
	require.NoError(t, os.Chmod("unreadable.go", 0o000))

	defer func() {
		_ = os.Chmod(filepath.Join(tempDir, "unreadable.go"), 0o600)
	}()

	// Test - should get error.
	_, err = checkGoFilesForCGO()
	require.Error(t, err)
}

func TestLint_Integration(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Find and change to project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(projectRoot))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint")

	// The actual project should pass all lint checks.
	err = Lint(logger)
	require.NoError(t, err, "Project should pass all lint checks")
}
