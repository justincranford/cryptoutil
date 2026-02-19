// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

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

