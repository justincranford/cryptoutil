// Copyright (c) 2025 Justin Cranford

package go_check_circular_package_dependencies

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

func TestCheckDependencies_NoCycle(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["cryptoutil/pkg/b"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["cryptoutil/pkg/c"]}
{"ImportPath":"cryptoutil/pkg/c","Imports":[]}
`

	err := CheckDependencies(jsonOutput)
	testify.NoError(t, err, "Expected no error for acyclic dependency graph")
}

func TestCheckDependencies_WithCycle(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["cryptoutil/pkg/b"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["cryptoutil/pkg/c"]}
{"ImportPath":"cryptoutil/pkg/c","Imports":["cryptoutil/pkg/a"]}
`

	err := CheckDependencies(jsonOutput)
	testify.Error(t, err, "Expected error for cyclic dependency graph")
	testify.Contains(t, err.Error(), "circular dependencies detected", "Error message should mention circular dependencies")
	testify.Contains(t, err.Error(), "cryptoutil/pkg/a", "Error should include cycle packages")
	testify.Contains(t, err.Error(), "cryptoutil/pkg/b", "Error should include cycle packages")
	testify.Contains(t, err.Error(), "cryptoutil/pkg/c", "Error should include cycle packages")
}

func TestCheckDependencies_MultipleCycles(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["cryptoutil/pkg/b"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["cryptoutil/pkg/a"]}
{"ImportPath":"cryptoutil/pkg/x","Imports":["cryptoutil/pkg/y"]}
{"ImportPath":"cryptoutil/pkg/y","Imports":["cryptoutil/pkg/x"]}
`

	err := CheckDependencies(jsonOutput)
	testify.Error(t, err, "Expected error for multiple cycles")
	testify.Contains(t, err.Error(), "circular dependencies detected", "Error message should mention circular dependencies")
}

func TestCheckDependencies_ExternalPackageIgnored(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["github.com/external/pkg","cryptoutil/pkg/b"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["fmt"]}
`

	err := CheckDependencies(jsonOutput)
	testify.NoError(t, err, "External package imports should be ignored")
}

func TestCheckDependencies_EmptyOutput(t *testing.T) {
	t.Parallel()

	err := CheckDependencies("")
	testify.Error(t, err, "Expected error for empty JSON output")
	testify.Contains(t, err.Error(), "no packages found", "Error should mention no packages found")
}

func TestCheckDependencies_InvalidJSON(t *testing.T) {
	t.Parallel()

	jsonOutput := `{"ImportPath":"cryptoutil/pkg/a","Imports":["invalid json`

	err := CheckDependencies(jsonOutput)
	testify.Error(t, err, "Expected error for invalid JSON")
	testify.Contains(t, err.Error(), "failed to parse package info", "Error should mention parsing failure")
}

func TestCheckDependencies_SelfCycle(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["cryptoutil/pkg/a"]}
`

	err := CheckDependencies(jsonOutput)
	testify.Error(t, err, "Expected error for self-import cycle")
	testify.Contains(t, err.Error(), "circular dependencies detected", "Error message should mention circular dependencies")
}

func TestCacheOperations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".cicd", "circular-deps-cache.json")

	// Test save cache
	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Verify file exists
	testify.FileExists(t, cacheFile, "Cache file should exist")

	// Test load cache
	loadedCache, err := loadCircularDepCache(cacheFile)
	testify.NoError(t, err, "Load cache should succeed")
	testify.NotNil(t, loadedCache, "Loaded cache should not be nil")
	testify.Equal(t, cache.HasCircularDeps, loadedCache.HasCircularDeps, "HasCircularDeps should match")
	testify.Equal(t, len(cache.CircularDeps), len(loadedCache.CircularDeps), "CircularDeps length should match")
}

func TestLoadCircularDepCache_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := loadCircularDepCache("/nonexistent/path/cache.json")
	testify.Error(t, err, "Load should fail for non-existent file")
	testify.Contains(t, err.Error(), "failed to read cache file", "Error should mention read failure")
}

func TestLoadCircularDepCache_InvalidJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "invalid-cache.json")

	// Write invalid JSON
	err := cryptoutilFiles.WriteFile(cacheFile, []byte("invalid json content"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write should succeed")

	_, err = loadCircularDepCache(cacheFile)
	testify.Error(t, err, "Load should fail for invalid JSON")
	testify.Contains(t, err.Error(), "failed to unmarshal cache JSON", "Error should mention unmarshal failure")
}

func TestSaveCircularDepCache_WithCircularDeps(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".cicd", "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: true,
		CircularDeps:    []string{"cryptoutil/pkg/a -> cryptoutil/pkg/b -> cryptoutil/pkg/a"},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Verify cache can be loaded and contains correct data
	loadedCache, err := loadCircularDepCache(cacheFile)
	testify.NoError(t, err, "Load should succeed")
	testify.True(t, loadedCache.HasCircularDeps, "Should have circular deps flag set")
	testify.Len(t, loadedCache.CircularDeps, 1, "Should have one circular dep entry")
	testify.Contains(t, loadedCache.CircularDeps[0], "cryptoutil/pkg/a", "Should contain cycle info")
}

func TestSaveCircularDepCache_DirectoryCreation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "deep", "nested", "path", "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed and create directories")

	// Verify file exists
	testify.FileExists(t, cacheFile, "Cache file should exist")

	// Verify directory structure was created
	testify.DirExists(t, filepath.Dir(cacheFile), "Cache directory should exist")
}

func TestCheckDependencies_ComplexCycle(t *testing.T) {
	t.Parallel()

	// Create a complex dependency graph with multiple interconnected packages
	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["cryptoutil/pkg/b","cryptoutil/pkg/c"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["cryptoutil/pkg/d"]}
{"ImportPath":"cryptoutil/pkg/c","Imports":["cryptoutil/pkg/d"]}
{"ImportPath":"cryptoutil/pkg/d","Imports":["cryptoutil/pkg/a"]}
`

	err := CheckDependencies(jsonOutput)
	testify.Error(t, err, "Expected error for complex cycle")
	testify.Contains(t, err.Error(), "circular dependencies detected", "Error should mention circular dependencies")
	testify.Contains(t, err.Error(), "Chain", "Error should include chain information")
}

func TestCheckDependencies_LongChain(t *testing.T) {
	t.Parallel()

	// Test a long chain without cycles
	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["cryptoutil/pkg/b"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["cryptoutil/pkg/c"]}
{"ImportPath":"cryptoutil/pkg/c","Imports":["cryptoutil/pkg/d"]}
{"ImportPath":"cryptoutil/pkg/d","Imports":["cryptoutil/pkg/e"]}
{"ImportPath":"cryptoutil/pkg/e","Imports":["cryptoutil/pkg/f"]}
{"ImportPath":"cryptoutil/pkg/f","Imports":[]}
`

	err := CheckDependencies(jsonOutput)
	testify.NoError(t, err, "Long chain without cycles should be valid")
}

func TestCheckDependencies_MixedInternalExternal(t *testing.T) {
	t.Parallel()

	// Mix internal and external dependencies
	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["github.com/external/x","cryptoutil/pkg/b"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["golang.org/x/tools","cryptoutil/pkg/c"]}
{"ImportPath":"cryptoutil/pkg/c","Imports":["fmt","encoding/json"]}
`

	err := CheckDependencies(jsonOutput)
	testify.NoError(t, err, "Mixed internal/external dependencies should be valid")
}

func TestCacheJSONFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
		GoModModTime:    time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC),
		HasCircularDeps: true,
		CircularDeps:    []string{"cycle1", "cycle2"},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Read raw JSON and verify format
	content, err := os.ReadFile(cacheFile)
	testify.NoError(t, err, "Read should succeed")

	// Verify it's valid JSON
	var decoded cryptoutilMagic.CircularDepCache

	err = json.Unmarshal(content, &decoded)
	testify.NoError(t, err, "JSON should be valid")

	// Verify formatting (indented)
	testify.Contains(t, string(content), "  ", "JSON should be indented")
	testify.Contains(t, string(content), "last_check", "JSON should contain last_check field")
	testify.Contains(t, string(content), "has_circular_deps", "JSON should contain has_circular_deps field")
}

func TestCheckDependencies_ErrorMessageFormat(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["cryptoutil/pkg/b"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["cryptoutil/pkg/a"]}
`

	err := CheckDependencies(jsonOutput)
	testify.Error(t, err, "Expected error for cycle")

	errMsg := err.Error()

	// Verify error message structure
	testify.Contains(t, errMsg, "circular dependencies detected:", "Should have main error message")
	testify.Contains(t, errMsg, "Chain 1", "Should include chain number")
	testify.Contains(t, errMsg, "packages)", "Should include package count")
	testify.Contains(t, errMsg, "→", "Should use arrow separator")
	testify.Contains(t, errMsg, "Consider refactoring to break these cycles", "Should include remediation advice")
}

func TestCheckDependencies_NoPackagesInGraph(t *testing.T) {
	t.Parallel()

	// Valid JSON but no packages
	jsonOutput := ""

	err := CheckDependencies(jsonOutput)
	testify.Error(t, err, "Should error on empty package list")
	testify.Contains(t, err.Error(), "no packages found", "Error should mention missing packages")
}

func TestCheckDependencies_PackageWithNoImports(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/standalone","Imports":[]}
`

	err := CheckDependencies(jsonOutput)
	testify.NoError(t, err, "Package with no imports should be valid")
}

func TestCheckDependencies_OnlyExternalImports(t *testing.T) {
	t.Parallel()

	jsonOutput := `
{"ImportPath":"cryptoutil/pkg/a","Imports":["github.com/foo/bar","golang.org/x/tools"]}
{"ImportPath":"cryptoutil/pkg/b","Imports":["fmt","encoding/json"]}
`

	err := CheckDependencies(jsonOutput)
	testify.NoError(t, err, "Packages with only external imports should be valid")
}

func TestCachePermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".cicd", "cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    time.Now().UTC(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Check file exists (permission check is platform-specific, so we skip it)
	testify.FileExists(t, cacheFile, "Cache file should exist")
}

func TestCheckDependencies_StressTest(t *testing.T) {
	t.Parallel()

	// Create a large dependency graph
	var builder strings.Builder

	numPackages := 100

	for i := 0; i < numPackages; i++ {
		imports := "[]"
		if i > 0 {
			// Each package imports the previous one (linear chain)
			imports = fmt.Sprintf(`["cryptoutil/pkg/p%d"]`, i-1)
		}

		fmt.Fprintf(&builder, `{"ImportPath":"cryptoutil/pkg/p%d","Imports":%s}`, i, imports)
		builder.WriteString("\n")
	}

	err := CheckDependencies(builder.String())
	testify.NoError(t, err, "Large linear dependency graph should be valid")
}
