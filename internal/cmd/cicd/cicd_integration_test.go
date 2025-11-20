// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

const (
	testGoModMinimal2 = `module example.com/test

go 1.25.4
`
	testMainContent2 = `package main

func main() {}
`
)

func TestGoUpdateDeps_MissingGoMod(t *testing.T) {
	logger := common.NewLogger("TestGoUpdateDeps_MissingGoMod")

	// Change to temp directory without go.mod
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Best effort to restore directory
		_ = os.Chdir(originalDir)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
	require.Error(t, err)
	require.Contains(t, err.Error(), "go.mod")
}

func TestGoUpdateDeps_MissingGoSum(t *testing.T) {
	logger := common.NewLogger("TestGoUpdateDeps_MissingGoSum")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Best effort to restore directory
		_ = os.Chdir(originalDir)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod but not go.sum
	goModContent := `module example.com/test

go 1.25.4

require (
	github.com/stretchr/testify v1.8.0
)
`
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
	require.Error(t, err)
	require.Contains(t, err.Error(), "go.sum")
}

func TestGoCheckCircularPackageDeps_MissingGoMod(t *testing.T) {
	logger := common.NewLogger("TestGoCheckCircularPackageDeps_MissingGoMod")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Best effort to restore directory
		_ = os.Chdir(originalDir)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = goCheckCircularPackageDeps(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "go.mod")
}

func TestGoCheckCircularPackageDeps_WithValidProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := common.NewLogger("TestGoCheckCircularPackageDeps_WithValidProject")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Best effort to restore directory
		_ = os.Chdir(originalDir)
		// Clean up cache file
		cacheFile := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir, "circular-dep-cache.json")
		_ = os.Remove(cacheFile) //nolint:errcheck // Test cleanup - file may not exist
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a minimal valid Go project
	goModContent := testGoModMinimal2
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	// Create a simple Go file
	mainContent := `package main

func main() {
	println("Hello")
}
`
	err = os.WriteFile("main.go", []byte(mainContent), 0o600)
	require.NoError(t, err)

	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err, "Simple project should have no circular dependencies")
}

func TestGoCheckCircularPackageDeps_CacheHit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := common.NewLogger("TestGoCheckCircularPackageDeps_CacheHit")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Best effort to restore directory
		_ = os.Chdir(originalDir)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod
	goModContent := testGoModMinimal2
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	// Create main.go
	mainContent := testMainContent2
	err = os.WriteFile("main.go", []byte(mainContent), 0o600)
	require.NoError(t, err)

	// First run - creates cache
	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err)

	// Second run - should use cache
	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err)
}

func TestGoCheckCircularPackageDeps_ExpiredCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := common.NewLogger("TestGoCheckCircularPackageDeps_ExpiredCache")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		//nolint:errcheck // Best effort to restore directory
		_ = os.Chdir(originalDir)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod
	goModContent := testGoModMinimal2
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	// Create main.go
	mainContent := testMainContent2
	err = os.WriteFile("main.go", []byte(mainContent), 0o600)
	require.NoError(t, err)

	// Create .cicd directory
	cicdDir := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir)
	err = os.MkdirAll(cicdDir, 0o755)
	require.NoError(t, err)

	// Create expired cache
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)

	expiredCache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().Add(-2 * time.Hour),
		GoModModTime:    goModStat.ModTime(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	cacheFile := filepath.Join(cicdDir, "circular-dep-cache.json")
	err = saveCircularDepCache(cacheFile, expiredCache)
	require.NoError(t, err)

	// Should regenerate cache
	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err)
}

func TestSaveCircularDepCache_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create .cicd directory - let saveCircularDepCache do it
	cacheFile := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir, "test-cache.json")

	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now(),
		GoModModTime:    time.Now(),
		HasCircularDeps: false,
		CircularDeps:    []string{},
	}

	err := saveCircularDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Verify directory was created
	cicdDir := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir)
	info, err := os.Stat(cicdDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())

	// Verify file was created
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)
}

func TestSaveDepCache_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create .cicd directory - let saveDepCache do it
	cacheFile := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir, "test-dep-cache.json")

	cache := cryptoutilMagic.DepCache{
		LastCheck:    time.Now(),
		GoModModTime: time.Now(),
		GoSumModTime: time.Now(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilMagic.ModeNameDirect,
	}

	err := saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Verify directory was created
	cicdDir := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir)
	info, err := os.Stat(cicdDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())

	// Verify file was created
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)
}
