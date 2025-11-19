// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilTestutil "cryptoutil/internal/common/testutil"
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

go 1.25.4

require (
	github.com/example/direct1 v1.0.0
	github.com/example/direct2 v2.0.0
	github.com/example/indirect v1.0.0 // indirect
)

require (
	github.com/example/direct3 v3.0.0
)
`
		cryptoutilTestutil.WriteTempFile(t, tempDir, "go.mod", goModContent)

		// Read the file content to pass to getDirectDependencies
		goModBytes := cryptoutilTestutil.ReadTestFile(t, "go.mod")

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

		cryptoutilTestutil.WriteTempFile(t, tempDir, "go.mod", "module example.com/test\ngo 1.21\n")

		// Read the file content to pass to getDirectDependencies
		goModBytes := cryptoutilTestutil.ReadTestFile(t, "go.mod")

		deps, err := getDirectDependencies(goModBytes)
		require.NoError(t, err)
		require.Empty(t, deps)
	})
}

func TestLoadDepCache(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid cache file", func(t *testing.T) {
		cacheContent := `{
			"last_check": "2025-01-01T00:00:00Z",
			"go_mod_mod_time": "2025-01-01T00:00:00Z",
			"go_sum_mod_time": "2025-01-01T00:00:00Z",
			"outdated_deps": ["github.com/example/old"],
			"mode": "direct"
		}`
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

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
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", "invalid json")
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
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

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
	content := cryptoutilTestutil.ReadTestFile(t, cacheFile)

	var loadedCache cryptoutilMagic.DepCache

	require.NoError(t, json.Unmarshal(content, &loadedCache))
	require.Equal(t, cache, loadedCache)

	// Check file permissions (should be 0o600 on Unix, but may differ on Windows)
	info, err := os.Stat(cacheFile)
	require.NoError(t, err)
	// On Windows, permissions might be different, so we just check that the file exists and is readable
	require.True(t, info.Mode().IsRegular(), "Cache file should be a regular file")
}

func TestCheckAndUseDepCache(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock file stats
	goModStat := &mockFileInfo{name: "go.mod", modTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
	goSumStat := &mockFileInfo{name: "go.sum", modTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}

	logger := NewLogUtil("test")

	t.Run("cache hit - valid cache with no outdated deps", func(t *testing.T) {
		// Create a valid cache file with recent timestamp
		recentTime := time.Now().UTC().Add(-cryptoutilMagic.TestCacheValidMinutes * time.Minute) // 30 minutes ago
		cacheContent := fmt.Sprintf(`{
			"last_check": "%s",
			"go_mod_mod_time": "2025-01-01T12:00:00Z",
			"go_sum_mod_time": "2025-01-01T12:00:00Z",
			"outdated_deps": [],
			"mode": "direct"
		}`, recentTime.Format(time.RFC3339))
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.True(t, cacheUsed)
		require.Equal(t, cacheHitState, cacheState)
		require.NoError(t, err)
	})

	t.Run("cache hit - valid cache with outdated deps", func(t *testing.T) {
		// Create a valid cache file with recent timestamp and outdated deps
		recentTime := time.Now().UTC().Add(-cryptoutilMagic.TestCacheValidMinutes * time.Minute) // 30 minutes ago
		cacheContent := fmt.Sprintf(`{
			"last_check": "%s",
			"go_mod_mod_time": "2025-01-01T12:00:00Z",
			"go_sum_mod_time": "2025-01-01T12:00:00Z",
			"outdated_deps": ["github.com/example/dep v1.0.0 [v1.1.0]"],
			"mode": "direct"
		}`, recentTime.Format(time.RFC3339))
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.True(t, cacheUsed)
		require.Equal(t, cacheHitState, cacheState)
		require.Error(t, err)
		require.Contains(t, err.Error(), "outdated dependencies found in cache")
	})

	t.Run("cache expired - time based", func(t *testing.T) {
		// Create cache older than 1 hour
		oldTime := time.Now().UTC().Add(-cryptoutilMagic.TestCacheExpiredHours * time.Hour)
		cacheContent := fmt.Sprintf(`{
			"last_check": "%s",
			"go_mod_mod_time": "2025-01-01T12:00:00Z",
			"go_sum_mod_time": "2025-01-01T12:00:00Z",
			"outdated_deps": [],
			"mode": "direct"
		}`, oldTime.Format(time.RFC3339))
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.False(t, cacheUsed)
		require.Contains(t, cacheState, "cache_expired_time")
		require.Contains(t, cacheState, "age:")
		require.NoError(t, err)
	})

	t.Run("cache expired - go.mod modified", func(t *testing.T) {
		// Create cache with old go.mod modtime
		cacheContent := `{
			"last_check": "2025-01-01T13:00:00Z",
			"go_mod_mod_time": "2025-01-01T11:00:00Z",
			"go_sum_mod_time": "2025-01-01T12:00:00Z",
			"outdated_deps": [],
			"mode": "direct"
		}`
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.False(t, cacheUsed)
		require.Equal(t, "cache_expired_files (go.mod modified)", cacheState)
		require.NoError(t, err)
	})

	t.Run("cache expired - go.sum modified", func(t *testing.T) {
		// Create cache with old go.sum modtime
		cacheContent := `{
			"last_check": "2025-01-01T13:00:00Z",
			"go_mod_mod_time": "2025-01-01T12:00:00Z",
			"go_sum_mod_time": "2025-01-01T11:00:00Z",
			"outdated_deps": [],
			"mode": "direct"
		}`
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.False(t, cacheUsed)
		require.Equal(t, "cache_expired_files (go.sum modified)", cacheState)
		require.NoError(t, err)
	})

	t.Run("cache expired - both files modified", func(t *testing.T) {
		// Create cache with old modtimes for both files
		cacheContent := `{
			"last_check": "2025-01-01T13:00:00Z",
			"go_mod_mod_time": "2025-01-01T11:00:00Z",
			"go_sum_mod_time": "2025-01-01T11:00:00Z",
			"outdated_deps": [],
			"mode": "direct"
		}`
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.False(t, cacheUsed)
		require.Equal(t, "cache_expired_files (go.mod and go.sum modified)", cacheState)
		require.NoError(t, err)
	})

	t.Run("cache mode mismatch", func(t *testing.T) {
		// Create cache with different mode
		cacheContent := `{
			"last_check": "2025-01-01T13:00:00Z",
			"go_mod_mod_time": "2025-01-01T12:00:00Z",
			"go_sum_mod_time": "2025-01-01T12:00:00Z",
			"outdated_deps": [],
			"mode": "all"
		}`
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", cacheContent)

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.False(t, cacheUsed)
		require.Equal(t, "cache_mode_mismatch", cacheState)
		require.NoError(t, err)
	})

	t.Run("cache invalid - malformed JSON", func(t *testing.T) {
		// Create invalid JSON cache file
		cacheFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test_cache.json", "invalid json content")

		cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, "direct", goModStat, goSumStat, logger)
		require.False(t, cacheUsed)
		require.Equal(t, "cache_invalid", cacheState)
		require.NoError(t, err)
	})

	t.Run("cache not exists", func(t *testing.T) {
		// Use non-existent cache file
		nonExistentCache := filepath.Join(tempDir, "nonexistent.json")

		cacheUsed, cacheState, err := checkAndUseDepCache(nonExistentCache, "direct", goModStat, goSumStat, logger)
		require.False(t, cacheUsed)
		require.Equal(t, "cache_not_exists", cacheState)
		require.NoError(t, err)
	})
}

// mockFileInfo implements os.FileInfo for testing.
type mockFileInfo struct {
	name    string
	modTime time.Time
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return cryptoutilMagic.CacheFilePermissions }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() any           { return nil }

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
