// Copyright (c) 2025 Justin Cranford

package outdated_deps

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheckOutdatedDeps_NoGoMod(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read go.mod")
}

// TestCheckOutdatedDeps_NoGoSum tests Check when go.sum is missing.
func TestCheckOutdatedDeps_NoGoSum(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod only (no go.sum).
	err = os.WriteFile("go.mod", []byte("module test\ngo 1.25.7\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read go.sum")
}

// TestCheckOutdatedDeps_CacheUsed tests Check when valid cache exists.
func TestCheckOutdatedDeps_CacheUsed(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod and go.sum.
	err = os.WriteFile("go.mod", []byte("module test\ngo 1.25.7\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Get file stats for cache.
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)
	goSumStat, err := os.Stat("go.sum")
	require.NoError(t, err)

	// Create a valid cache file with no outdated deps.
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilSharedMagic.ModeNameDirect,
	}
	err = saveDepCache(cryptoutilSharedMagic.DepCacheFileName, cache)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

// TestCheckOutdatedDeps_CacheWithError tests Check when cache has outdated deps.
func TestCheckOutdatedDeps_CacheWithError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod and go.sum.
	err = os.WriteFile("go.mod", []byte("module test\ngo 1.25.7\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Get file stats for cache.
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)
	goSumStat, err := os.Stat("go.sum")
	require.NoError(t, err)

	// Create a valid cache file WITH outdated deps.
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{"example.com/dep v1.0.0 [v1.1.0]"},
		Mode:         cryptoutilSharedMagic.ModeNameDirect,
	}
	err = saveDepCache(cryptoutilSharedMagic.DepCacheFileName, cache)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cached dependency check failed")
}

// TestCheckOutdatedDeps_GoListError tests Check when go list command fails.
func TestCheckOutdatedDeps_GoListError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a malformed go.mod file that will make go list fail.
	// Using invalid syntax to force a parsing error.
	err = os.WriteFile("go.mod", []byte("invalid go.mod content without module directive\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check dependencies")
}

// TestCheckOutdatedDeps_NoOutdatedDeps tests Check with up-to-date deps (fresh check).
func TestCheckOutdatedDeps_NoOutdatedDeps(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid go module that go list can process.
	// Using "example.com/test" which won't have any real dependencies.
	err = os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.25.7\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	// This should succeed because there are no dependencies.
	require.NoError(t, err)
}

// TestSaveDepCache_WriteError tests saveDepCache when directory creation fails.
func TestSaveDepCache_WriteError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	// Create a read-only directory structure.
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	err := os.Mkdir(readOnlyDir, 0o500)
	require.NoError(t, err)

	// Try to write to a nested path inside the read-only directory.
	cacheFile := filepath.Join(readOnlyDir, "nested", "cache.json")

	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: time.Now().UTC(),
		GoSumModTime: time.Now().UTC(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilSharedMagic.ModeNameDirect,
	}

	err = saveDepCache(cacheFile, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create output directory")
}

// TestLint_WithLinterError tests the Lint function when a linter returns an error.
func TestLint_WithLinterError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No go.mod file, so the linter will fail.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read go.mod")
}

// TestLint_Success tests the Lint function when all linters pass.
func TestLint_Success(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid go module setup.
	err = os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.25.7\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

// TestCheckOutdatedDeps_WithOutdatedDeps tests Check finding outdated deps (fresh check).
func TestCheckOutdatedDeps_WithOutdatedDeps(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a go module that has an outdated dependency.
	// We need a real dependency that might have updates.
	goModContent := `module example.com/test

go 1.25.7

require github.com/pkg/errors v0.8.0
`
	err = os.WriteFile("go.mod", []byte(goModContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte("github.com/pkg/errors v0.8.0 h1:WdK/asTD0HN+q6hsWO3/vpuAkAr+tw6aNJNDFFf0+qw=\ngithub.com/pkg/errors v0.8.0/go.mod h1:bwawxfHBFNV+L2hUp1rHADufV3IMtnDRdf1r5NINEl0=\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	// This should fail because pkg/errors v0.8.0 has updates available.
	require.Error(t, err)
	require.Contains(t, err.Error(), "outdated dependencies found")
}

// TestCheckOutdatedDeps_SaveCacheError tests warning when cache save fails.
func TestCheckOutdatedDeps_SaveCacheError(t *testing.T) {
	// This test cannot be parallel because it changes working directory.
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(origDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid go module.
	err = os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.25.7\n"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Create a file at the cache directory location to prevent directory creation.
	cacheDir := filepath.Dir(cryptoutilSharedMagic.DepCacheFileName)
	err = os.WriteFile(cacheDir, []byte("blocking file"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// This should succeed but with a warning about cache save failure.
	err = Check(logger)
	// The check itself should still pass (no outdated deps).
	require.NoError(t, err)
}

// TestSaveDepCache_WriteFileError tests saveDepCache when file write fails.
func TestSaveDepCache_WriteFileError(t *testing.T) {
	t.Parallel()

	// Create a directory at the cache file location to make write fail.
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "cache.json")

	// Create a directory with the same name as the intended file.
	err := os.Mkdir(cacheFile, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err)

	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: time.Now().UTC(),
		GoSumModTime: time.Now().UTC(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilSharedMagic.ModeNameDirect,
	}

	err = saveDepCache(cacheFile, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write cache file")
}
