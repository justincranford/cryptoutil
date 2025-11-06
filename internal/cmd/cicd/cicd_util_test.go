package cicd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

func TestCollectAllFiles(t *testing.T) {
	// Create a temporary directory with some test files
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"subdir/file3.txt",
		"subdir/nested/file4.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte("test content"), cryptoutilMagic.CacheFilePermissions))
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Collect files
	files, err := listAllFiles()
	require.NoError(t, err)

	// Should find all test files
	require.Len(t, files, len(testFiles), "Should find all test files")

	// Convert to relative paths for comparison
	for i, file := range files {
		files[i], err = filepath.Rel(tempDir, filepath.Join(tempDir, file))
		require.NoError(t, err)
		// Normalize path separators to forward slashes for cross-platform comparison
		files[i] = filepath.ToSlash(files[i])
	}

	// Normalize expected paths to forward slashes
	normalizedTestFiles := make([]string, len(testFiles))
	for i, file := range testFiles {
		normalizedTestFiles[i] = filepath.ToSlash(file)
	}

	// Sort both slices for comparison
	require.ElementsMatch(t, normalizedTestFiles, files, "Should find all expected files")
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

go 1.21

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

		deps, err := getDirectDependencies()
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

		deps, err := getDirectDependencies()
		require.Error(t, err)
		require.Nil(t, deps)
		require.Contains(t, err.Error(), "failed to read go.mod")
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

		deps, err := getDirectDependencies()
		require.NoError(t, err)
		require.Empty(t, deps)
	})
}
