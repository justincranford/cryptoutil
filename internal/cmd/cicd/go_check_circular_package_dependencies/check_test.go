// Copyright (c) 2025 Justin Cranford

package go_check_circular_package_dependencies

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// TestCheck_NoCycle tests Check function with no circular dependencies.
func TestCheck_NoCycle(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := `module testproject

go 1.25
`
	err := os.WriteFile(goModPath, []byte(goModContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Create simple package structure (no cycles)
	pkgDir := filepath.Join(tempDir, "pkg")
	err = os.MkdirAll(pkgDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create pkg directory should succeed")

	pkgFile := filepath.Join(pkgDir, "test.go")
	pkgContent := `package pkg

import "fmt"

func Test() {
	fmt.Println("test")
}
`
	err = os.WriteFile(pkgFile, []byte(pkgContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create test.go should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check
	logger := common.NewLogger("test-check-nocycle")
	err = Check(logger)

	testify.NoError(t, err, "Check should succeed with no circular dependencies")
}

// TestCheck_CacheHit_NoCycle tests Check using cached results (no cycle).
func TestCheck_CacheHit_NoCycle(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := `module testproject

go 1.25
`
	err := os.WriteFile(goModPath, []byte(goModContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Create package
	pkgDir := filepath.Join(tempDir, "pkg")
	err = os.MkdirAll(pkgDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create pkg directory should succeed")

	pkgFile := filepath.Join(pkgDir, "test.go")
	pkgContent := `package pkg

import "fmt"

func Test() {
	fmt.Println("test")
}
`
	err = os.WriteFile(pkgFile, []byte(pkgContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create test.go should succeed")

	// Get go.mod stat
	goModStat, err := os.Stat(goModPath)
	testify.NoError(t, err, "Stat go.mod should succeed")

	// Create valid cache (recent, no cycles)
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "circular-deps-cache.json")
	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    goModStat.ModTime(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err = saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should use cache
	logger := common.NewLogger("test-check-cache-hit-nocycle")
	err = Check(logger)

	testify.NoError(t, err, "Check should succeed using cached results")
}

// TestCheck_CacheExpired tests Check when cache is expired.
func TestCheck_CacheExpired(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := `module testproject

go 1.25
`
	err := os.WriteFile(goModPath, []byte(goModContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Create package
	pkgDir := filepath.Join(tempDir, "pkg")
	err = os.MkdirAll(pkgDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create pkg directory should succeed")

	pkgFile := filepath.Join(pkgDir, "test.go")
	pkgContent := `package pkg

func Test() {}
`
	err = os.WriteFile(pkgFile, []byte(pkgContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create test.go should succeed")

	// Get go.mod stat
	goModStat, err := os.Stat(goModPath)
	testify.NoError(t, err, "Stat go.mod should succeed")

	// Create expired cache (1 hour old)
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "circular-deps-cache.json")
	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC().Add(-1 * time.Hour),
		GoModModTime:    goModStat.ModTime(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err = saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should ignore expired cache and re-check
	logger := common.NewLogger("test-check-cache-expired")
	err = Check(logger)

	testify.NoError(t, err, "Check should succeed after cache expiration")
}

// TestCheck_GoModChanged tests Check when go.mod was modified after cache.
func TestCheck_GoModChanged(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()

	// Create go.mod with old timestamp
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := `module testproject

go 1.25
`
	err := os.WriteFile(goModPath, []byte(goModContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	oldTime := time.Now().Add(-2 * time.Hour)
	err = os.Chtimes(goModPath, oldTime, oldTime)
	testify.NoError(t, err, "Set go.mod time should succeed")

	// Create package
	pkgDir := filepath.Join(tempDir, "pkg")
	err = os.MkdirAll(pkgDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create pkg directory should succeed")

	pkgFile := filepath.Join(pkgDir, "test.go")
	pkgContent := `package pkg

func Test() {}
`
	err = os.WriteFile(pkgFile, []byte(pkgContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create test.go should succeed")

	// Create cache with old go.mod time
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "circular-deps-cache.json")
	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC().Add(-1 * time.Hour),
		GoModModTime:    oldTime,
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err = saveCircularDepCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Update go.mod timestamp to be newer than cache
	newTime := time.Now()
	err = os.Chtimes(goModPath, newTime, newTime)
	testify.NoError(t, err, "Update go.mod time should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should invalidate cache due to go.mod change
	logger := common.NewLogger("test-check-gomod-changed")
	err = Check(logger)

	testify.NoError(t, err, "Check should succeed after go.mod change")
}

// TestCheck_MissingGoMod tests Check when go.mod doesn't exist.
func TestCheck_MissingGoMod(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()

	// Don't create go.mod

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should fail due to missing go.mod
	logger := common.NewLogger("test-check-missing-gomod")
	err := Check(logger)

	testify.Error(t, err, "Check should fail when go.mod is missing")
	testify.Contains(t, err.Error(), "go.mod", "Error should mention go.mod")
}

// Helper: changeToTempDir changes to temporary directory and returns original directory.
func changeToTempDir(t *testing.T, tempDir string) string {
	t.Helper()

	origDir, err := os.Getwd()
	testify.NoError(t, err, "Get working directory should succeed")

	err = os.Chdir(tempDir)
	testify.NoError(t, err, "Change to temp directory should succeed")

	return origDir
}

// Helper: restoreDir restores original working directory.
func restoreDir(t *testing.T, origDir string) {
	t.Helper()

	err := os.Chdir(origDir)
	testify.NoError(t, err, "Restore working directory should succeed")
}
