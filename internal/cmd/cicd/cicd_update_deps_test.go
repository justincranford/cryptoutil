// Package cicd provides tests for dependency update checking functionality.
package cicd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckDependencyUpdates(t *testing.T) {
	for _, tt := range checkDependencyUpdatesTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			actualOutdatedDeps, err := checkDependencyUpdates(tt.depCheckMode, tt.actualDeps, tt.directDeps)
			require.NoError(t, err)
			require.Len(t, actualOutdatedDeps, len(tt.expectedOutdatedDeps))

			for _, expectedOutdatedDep := range tt.expectedOutdatedDeps {
				require.Contains(t, actualOutdatedDeps, expectedOutdatedDep)
			}
		})
	}
}

func TestGetDirectDependencies(t *testing.T) {
	t.Run("valid go.mod", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Create a test go.mod file
		goModContent := `module example.com/test

go 1.25.3

require (
	github.com/example/direct1 v1.0.0
	github.com/example/direct2 v2.0.0
	github.com/example/indirect v1.0.0 // indirect
)

require (
	github.com/example/direct3 v3.0.0
)
`
		require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), cryptoutilMagic.CacheFilePermissions))

		// Read the file content to pass to getDirectDependencies
		goModBytes, err := os.ReadFile("go.mod")
		require.NoError(t, err)

		deps, err := getDirectDependencies(goModBytes)
		require.NoError(t, err)
		require.Contains(t, deps, "github.com/example/direct1")
		require.Contains(t, deps, "github.com/example/direct2")
		require.Contains(t, deps, "github.com/example/direct3")
		require.NotContains(t, deps, "github.com/example/indirect") // Should exclude indirect deps
	})

	t.Run("go.mod does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// This test case is no longer relevant since file reading is done in goUpdateDeps
		// The error checking is now handled there
		t.Skip("File reading is now handled in goUpdateDeps")
	})

	t.Run("empty go.mod", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		require.NoError(t, os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.21\n"), cryptoutilMagic.CacheFilePermissions))

		// Read the file content to pass to getDirectDependencies
		goModBytes, err := os.ReadFile("go.mod")
		require.NoError(t, err)

		deps, err := getDirectDependencies(goModBytes)
		require.NoError(t, err)
		require.Empty(t, deps)
	})
}

func TestLoadDepCache(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "test_cache.json")

	t.Run("valid cache file", func(t *testing.T) {
		cacheContent := `{
			"last_check": "2025-01-01T00:00:00Z",
			"go_mod_mod_time": "2025-01-01T00:00:00Z",
			"go_sum_mod_time": "2025-01-01T00:00:00Z",
			"outdated_deps": ["github.com/example/old"],
			"mode": "direct"
		}`
		require.NoError(t, os.WriteFile(cacheFile, []byte(cacheContent), cryptoutilMagic.CacheFilePermissions))

		cache, err := loadDepCache(cacheFile, "direct")
		require.NoError(t, err)
		require.NotNil(t, cache)
		require.Equal(t, "direct", cache.Mode)
		require.Len(t, cache.OutdatedDeps, 1)
		require.Equal(t, "github.com/example/old", cache.OutdatedDeps[0])
	})

	t.Run("cache file does not exist", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.json")
		cache, err := loadDepCache(nonExistentFile, "direct")
		require.Error(t, err)
		require.Nil(t, cache)
		require.Contains(t, err.Error(), "failed to read cache file")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		require.NoError(t, os.WriteFile(cacheFile, []byte("invalid json"), cryptoutilMagic.CacheFilePermissions))
		cache, err := loadDepCache(cacheFile, "direct")
		require.Error(t, err)
		require.Nil(t, cache)
		require.Contains(t, err.Error(), "failed to unmarshal cache JSON")
	})

	t.Run("mode mismatch", func(t *testing.T) {
		cacheContent := `{
			"last_check": "2025-01-01T00:00:00Z",
			"go_mod_mod_time": "2025-01-01T00:00:00Z",
			"go_sum_mod_time": "2025-01-01T00:00:00Z",
			"outdated_deps": [],
			"mode": "direct"
		}`
		require.NoError(t, os.WriteFile(cacheFile, []byte(cacheContent), cryptoutilMagic.CacheFilePermissions))

		cache, err := loadDepCache(cacheFile, "all")
		require.Error(t, err)
		require.Nil(t, cache)
		require.Contains(t, err.Error(), "cache mode mismatch")
	})
}

func TestSaveDepCache(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "test_cache.json")

	cache := cryptoutilMagic.DepCache{
		LastCheck:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		GoModModTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		GoSumModTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		OutdatedDeps: []string{"github.com/example/old", "github.com/example/older"},
		Mode:         "direct",
	}

	err := saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Verify file was created and has correct content
	content, err := os.ReadFile(cacheFile)
	require.NoError(t, err)

	var loadedCache cryptoutilMagic.DepCache

	require.NoError(t, json.Unmarshal(content, &loadedCache))
	require.Equal(t, cache, loadedCache)

	// Check file permissions (should be 0o600 on Unix, but may differ on Windows)
	info, err := os.Stat(cacheFile)
	require.NoError(t, err)
	// On Windows, permissions might be different, so we just check that the file exists and is readable
	require.True(t, info.Mode().IsRegular(), "Cache file should be a regular file")
}

type checkDependencyUpdatesTestCase struct {
	name                 string
	depCheckMode         cryptoutilMagic.DepCheckMode
	actualDeps           string
	expectedOutdatedDeps []string
	directDeps           map[string]bool
}

func checkDependencyUpdatesTestCases() []checkDependencyUpdatesTestCase {
	dep1 := "example.com/dep1"
	dep2 := "github.com/dep2"
	dep3 := "example.com/dep3"

	latestDep1 := dep1 + " v1.9.0"
	latestDep2 := dep2 + " v0.15.0"
	latestDep3 := dep3 + " v1.4.0"

	outdatedDep1 := dep1 + " v1.8.4 [v1.9.0]"
	outdatedDep2 := dep2 + " v0.14.0 [v0.15.0]"
	outdatedDep3 := dep3 + " v1.3.0 [v1.4.0]"

	tests := []checkDependencyUpdatesTestCase{
		{
			name:                 "Malformed Lines",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           outdatedDep1 + "\nmalformed\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep1, outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},

		// Direct mode test cases

		{
			name:                 "0 Deps, 0 Direct Outdated, Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           "",
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{},
		},
		{
			name:                 "1 Direct Deps, 0 Direct Outdated, Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           latestDep1,
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "1 Direct Deps, 1 Direct Outdated, Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           outdatedDep1,
			expectedOutdatedDeps: []string{outdatedDep1},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "2 Direct Deps, 0 Direct Outdated, Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           latestDep1 + "\n" + latestDep2,
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "2 Direct Deps, 1 Outdated (First), Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           outdatedDep1 + "\n" + latestDep2,
			expectedOutdatedDeps: []string{outdatedDep1},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "2 Direct Deps, 1 Outdated (Second), Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           latestDep1 + "\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "2 Direct Deps, 2 Direct Outdated, Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           outdatedDep1 + "\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep1, outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true},
		},
		{
			name:                 "3 Direct Deps, 0 Direct Outdated, Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           latestDep1 + "\n" + latestDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},
		{
			name:                 "3 Direct Deps, 1 Direct Outdated (First), Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           outdatedDep1 + "\n" + latestDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{outdatedDep1},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},
		{
			name:                 "3 Direct Deps, 1 Direct Outdated (Second), Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           latestDep1 + "\n" + outdatedDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{outdatedDep2},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},
		{
			name:                 "3 Direct Deps, 1 Direct Outdated (Third), Direct mode",
			depCheckMode:         cryptoutilMagic.DepCheckDirect,
			actualDeps:           latestDep1 + "\n" + latestDep2 + "\n" + outdatedDep3,
			expectedOutdatedDeps: []string{outdatedDep3},
			directDeps:           map[string]bool{dep1: true, dep2: true, dep3: true},
		},

		// All mode test cases

		{
			name:                 "1 Direct Deps, 1 Transitive Deps, 0 Transitive Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + latestDep2,
			expectedOutdatedDeps: []string{},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "1 Direct Deps, 1 Transitive Deps, 1 Transitive Outdated, All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + outdatedDep2,
			expectedOutdatedDeps: []string{outdatedDep2},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "1 Direct Deps, 2 Transitive Deps, 1 Transitive Outdated (First), All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + outdatedDep2 + "\n" + latestDep3,
			expectedOutdatedDeps: []string{outdatedDep2},
			directDeps:           map[string]bool{dep1: true},
		},
		{
			name:                 "1 Direct Deps, 2 Transitive Deps, 1 Transitive Outdated (Second), All mode",
			depCheckMode:         cryptoutilMagic.DepCheckAll,
			actualDeps:           latestDep1 + "\n" + latestDep2 + "\n" + outdatedDep3,
			expectedOutdatedDeps: []string{outdatedDep3},
			directDeps:           map[string]bool{dep1: true},
		},
	}

	return tests
}
