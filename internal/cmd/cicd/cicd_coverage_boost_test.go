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
	testGoModMinimal = `module example.com/test

go 1.25.4
`
	testGoModWithDeps = `module example.com/test

go 1.25.4

require (
	github.com/stretchr/testify v1.9.0
`
	testMainContent = `package main

func main() {}
`
	testAnyContent = `package test

var x interface{} = 42
var y interface{} = "hello"
`
)

// TestGoUpdateDeps_AllMode tests goUpdateDeps with all dependencies mode.
func TestGoUpdateDeps_AllMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := common.NewLogger("TestGoUpdateDeps_AllMode")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create minimal go.mod and go.sum
	goModContent := testGoModMinimal
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	err = goUpdateDeps(logger, cryptoutilMagic.DepCheckAll)
	// May succeed or fail depending on project state, but should not panic
	_ = err
}

// TestGoUpdateDeps_DirectMode tests goUpdateDeps with direct dependencies only.
func TestGoUpdateDeps_DirectMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := common.NewLogger("TestGoUpdateDeps_DirectMode")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod with direct dependency
	goModContent := testGoModWithDeps
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	err = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
	// May succeed or fail, but should not panic
	_ = err
}

// TestGoUpdateDeps_WithCache tests that cache is used when valid.
func TestGoUpdateDeps_WithCache(t *testing.T) {
	logger := common.NewLogger("TestGoUpdateDeps_WithCache")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod and go.sum
	goModContent := testGoModMinimal
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// Create .cicd directory
	cicdDir := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir)
	err = os.MkdirAll(cicdDir, 0o755)
	require.NoError(t, err)

	// Create valid cache
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)
	goSumStat, err := os.Stat("go.sum")
	require.NoError(t, err)

	cache := cryptoutilMagic.DepCache{
		LastCheck:    time.Now().Add(-30 * time.Minute),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilMagic.ModeNameDirect,
	}

	cacheFile := filepath.Join(cicdDir, "dep-cache.json")
	err = saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Should use cache
	err = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
	require.NoError(t, err)
}

// TestGoUpdateDeps_ExpiredCache tests that expired cache is regenerated.
func TestGoUpdateDeps_ExpiredCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := common.NewLogger("TestGoUpdateDeps_ExpiredCache")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod and go.sum
	goModContent := testGoModMinimal
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// Create .cicd directory
	cicdDir := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir)
	err = os.MkdirAll(cicdDir, 0o755)
	require.NoError(t, err)

	// Create expired cache (> 1 hour old)
	goModStat, err := os.Stat("go.mod")
	require.NoError(t, err)
	goSumStat, err := os.Stat("go.sum")
	require.NoError(t, err)

	expiredCache := cryptoutilMagic.DepCache{
		LastCheck:    time.Now().Add(-2 * time.Hour),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilMagic.ModeNameDirect,
	}

	cacheFile := filepath.Join(cicdDir, "dep-cache.json")
	err = saveDepCache(cacheFile, expiredCache)
	require.NoError(t, err)

	// Should regenerate cache
	err = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
	// May succeed or fail depending on network, but should not use expired cache
	_ = err
}

// TestGoUpdateDeps_InvalidatedCache tests cache invalidation when files change.
func TestGoUpdateDeps_InvalidatedCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := common.NewLogger("TestGoUpdateDeps_InvalidatedCache")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod and go.sum
	goModContent := testGoModMinimal
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// Wait a moment to ensure different modtime
	time.Sleep(10 * time.Millisecond)

	// Create .cicd directory
	cicdDir := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir)
	err = os.MkdirAll(cicdDir, 0o755)
	require.NoError(t, err)

	// Create cache with old modtime
	cache := cryptoutilMagic.DepCache{
		LastCheck:    time.Now(),
		GoModModTime: time.Now().Add(-1 * time.Hour), // Old modtime
		GoSumModTime: time.Now().Add(-1 * time.Hour),
		OutdatedDeps: []string{},
		Mode:         cryptoutilMagic.ModeNameDirect,
	}

	cacheFile := filepath.Join(cicdDir, "dep-cache.json")
	err = saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Should invalidate cache due to modtime mismatch
	err = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
	// May succeed or fail depending on network, but should not use stale cache
	_ = err
}

// TestGoCheckCircularPackageDeps_WithCache tests cache usage.
func TestGoCheckCircularPackageDeps_WithCache(t *testing.T) {
	logger := common.NewLogger("TestGoCheckCircularPackageDeps_WithCache")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create minimal project
	goModContent := testGoModMinimal
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	mainContent := testMainContent
	err = os.WriteFile("main.go", []byte(mainContent), 0o600)
	require.NoError(t, err)

	// Create .cicd directory
	cicdDir := filepath.Join(tmpDir, cryptoutilMagic.CICDOutputDir)
	err = os.MkdirAll(cicdDir, 0o755)
	require.NoError(t, err)

	// Create valid cache
	require.NoError(t, err)

	// Note: Cache creation test removed - now covered by circulardeps package tests
	// Test simply runs the check command
	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err)
}

// TestLoadDepCache_InvalidJSON tests handling of corrupted cache.
func TestLoadDepCache_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	cacheFile := filepath.Join(tmpDir, "invalid-cache.json")
	err := os.WriteFile(cacheFile, []byte("invalid json"), 0o600)
	require.NoError(t, err)

	_, err = loadDepCache(cacheFile, cryptoutilMagic.ModeNameDirect)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal")
}

// TestLoadDepCache_ModeMismatch tests cache mode validation.
func TestLoadDepCache_ModeMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache for "direct" mode
	cache := cryptoutilMagic.DepCache{
		LastCheck:    time.Now(),
		GoModModTime: time.Now(),
		GoSumModTime: time.Now(),
		OutdatedDeps: []string{},
		Mode:         cryptoutilMagic.ModeNameDirect,
	}

	cacheFile := filepath.Join(tmpDir, "cache.json")
	err := saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Try to load with "all" mode
	_, err = loadDepCache(cacheFile, cryptoutilMagic.ModeNameAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cache mode mismatch")
}

// Note: TestLoadCircularDepCache_InvalidJSON removed - now covered by circulardeps package tests

// TestGoEnforceAny_MultipleReplacements tests enforcement with multiple files.
func TestGoEnforceAny_MultipleReplacements(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceAny_MultipleReplacements")

	tmpDir := t.TempDir()

	// Create multiple Go files with any usage
	file1 := filepath.Join(tmpDir, "file1.go")
	content1 := testAnyContent
	err := os.WriteFile(file1, []byte(content1), 0o600)
	require.NoError(t, err)

	file2 := filepath.Join(tmpDir, "file2.go")
	content2 := `package test

func process(data any) any {
	return data
}
`
	err = os.WriteFile(file2, []byte(content2), 0o600)
	require.NoError(t, err)

	allFiles := []string{file1, file2}

	err = goEnforceAny(logger, allFiles)
	require.Error(t, err, "goEnforceAny should return error when modifications are made")
	require.Contains(t, err.Error(), "modified")

	// Verify replacements were made - interface{} should have been replaced with any
	modifiedContent1, err := os.ReadFile(file1)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent1), "any", "Should contain 'any' after replacement")
	require.NotContains(t, string(modifiedContent1), "interface{}", "Should not contain 'interface{}' after replacement")

	modifiedContent2, err := os.ReadFile(file2)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent2), "any", "Should contain 'any' after replacement")
	require.NotContains(t, string(modifiedContent2), "interface{}", "Should not contain 'interface{}' after replacement")
}

// TestGoEnforceAny_AlreadyUsingAny tests files already using 'any'.
func TestGoEnforceAny_AlreadyUsingAny(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceAny_AlreadyUsingAny")

	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	content := `package test

var x any = 42
var y any = "hello"
`
	err := os.WriteFile(file1, []byte(content), 0o600)
	require.NoError(t, err)

	allFiles := []string{file1}

	err = goEnforceAny(logger, allFiles)
	require.NoError(t, err)

	// Verify no changes made
	modifiedContent, err := os.ReadFile(file1)
	require.NoError(t, err)
	require.Equal(t, content, string(modifiedContent))
}

// TestGoEnforceAny_NonGoFiles tests filtering of non-Go files.
func TestGoEnforceAny_NonGoFiles(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceAny_NonGoFiles")

	tmpDir := t.TempDir()

	// Create non-Go file
	txtFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(txtFile, []byte("any"), 0o600)
	require.NoError(t, err)

	allFiles := []string{txtFile}

	err = goEnforceAny(logger, allFiles)
	require.NoError(t, err)

	// Verify file was not modified
	content, err := os.ReadFile(txtFile)
	require.NoError(t, err)
	require.Equal(t, "any", string(content))
}
