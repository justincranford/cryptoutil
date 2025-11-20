// Copyright (c) 2025 Justin Cranford

package go_check_identity_imports

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// TestCheck_NoViolations tests Check with valid identity module (no forbidden imports).
func TestCheck_NoViolations(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, false) // false = no violations

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check
	logger := common.NewLogger("test-check-no-violations")
	err = Check(logger)

	testify.NoError(t, err, "Check should succeed with no violations")
}

// TestCheck_WithViolations tests Check detecting forbidden imports.
func TestCheck_WithViolations(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, true) // true = with violations

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check
	logger := common.NewLogger("test-check-with-violations")
	err = Check(logger)

	testify.Error(t, err, "Check should fail with violations")
	testify.Contains(t, err.Error(), "forbidden imports detected", "Error should mention forbidden imports")
}

// TestCheck_CacheHit_NoViolations tests Check using cached results (cache hit, no violations).
func TestCheck_CacheHit_NoViolations(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, false)

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Get file mod times
	goModStat, err := os.Stat(goModPath)
	testify.NoError(t, err, "Stat go.mod should succeed")

	identityDir := filepath.Join(tempDir, "internal", "identity")
	identityModTime, err := getLatestModTime(identityDir)
	testify.NoError(t, err, "Get identity mod time should succeed")

	// Create valid cache (recent, no violations)
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "identity-imports-cache.json")
	cache := Cache{
		LastCheck:           time.Now().UTC(),
		GoModModTime:        goModStat.ModTime(),
		IdentityModTime:     identityModTime,
		HasForbiddenImports: false,
		ForbiddenImports:    []string{},
	}

	err = saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should use cache
	logger := common.NewLogger("test-check-cache-hit")
	err = Check(logger)

	testify.NoError(t, err, "Check should succeed using cached results")
}

// TestCheck_CacheHit_WithViolations tests Check using cached results with violations.
func TestCheck_CacheHit_WithViolations(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, true) // true = with violations

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Get file mod times
	goModStat, err := os.Stat(goModPath)
	testify.NoError(t, err, "Stat go.mod should succeed")

	identityDir := filepath.Join(tempDir, "internal", "identity")
	identityModTime, err := getLatestModTime(identityDir)
	testify.NoError(t, err, "Get identity mod time should succeed")

	// Create cache with violations
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "identity-imports-cache.json")
	cache := Cache{
		LastCheck:           time.Now().UTC(),
		GoModModTime:        goModStat.ModTime(),
		IdentityModTime:     identityModTime,
		HasForbiddenImports: true,
		ForbiddenImports:    []string{"test.go:5: forbidden import \"cryptoutil/internal/server\" (KMS server domain)"},
	}

	err = saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should use cache and fail
	logger := common.NewLogger("test-check-cache-hit-violations")
	err = Check(logger)

	testify.Error(t, err, "Check should fail using cached violations")
	testify.Contains(t, err.Error(), "forbidden imports detected", "Error should mention forbidden imports")
}

// TestCheck_CacheExpired tests Check when cache is expired.
func TestCheck_CacheExpired(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, false)

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Get file mod times
	goModStat, err := os.Stat(goModPath)
	testify.NoError(t, err, "Stat go.mod should succeed")

	identityDir := filepath.Join(tempDir, "internal", "identity")
	identityModTime, err := getLatestModTime(identityDir)
	testify.NoError(t, err, "Get identity mod time should succeed")

	// Create expired cache (1 hour old)
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "identity-imports-cache.json")
	cache := Cache{
		LastCheck:           time.Now().UTC().Add(-1 * time.Hour),
		GoModModTime:        goModStat.ModTime(),
		IdentityModTime:     identityModTime,
		HasForbiddenImports: false,
		ForbiddenImports:    []string{},
	}

	err = saveCache(cacheFile, cache)
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
	setupTestEnvironment(t, tempDir, false)

	// Create go.mod with old timestamp
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	oldTime := time.Now().Add(-2 * time.Hour)
	err = os.Chtimes(goModPath, oldTime, oldTime)
	testify.NoError(t, err, "Set go.mod time should succeed")

	// Get identity mod time
	identityDir := filepath.Join(tempDir, "internal", "identity")
	identityModTime, err := getLatestModTime(identityDir)
	testify.NoError(t, err, "Get identity mod time should succeed")

	// Create cache with old go.mod time
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "identity-imports-cache.json")
	cache := Cache{
		LastCheck:           time.Now().UTC().Add(-1 * time.Hour),
		GoModModTime:        oldTime,
		IdentityModTime:     identityModTime,
		HasForbiddenImports: false,
		ForbiddenImports:    []string{},
	}

	err = saveCache(cacheFile, cache)
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

// TestCheck_IdentityModuleChanged tests Check when identity module files were modified.
func TestCheck_IdentityModuleChanged(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, false)

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Get go.mod mod time
	goModStat, err := os.Stat(goModPath)
	testify.NoError(t, err, "Stat go.mod should succeed")

	// Get old identity mod time
	identityDir := filepath.Join(tempDir, "internal", "identity")
	oldIdentityModTime, err := getLatestModTime(identityDir)
	testify.NoError(t, err, "Get identity mod time should succeed")

	// Create cache with old identity time
	cacheDir := filepath.Join(tempDir, ".cicd")
	err = os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create cache directory should succeed")

	cacheFile := filepath.Join(cacheDir, "identity-imports-cache.json")
	cache := Cache{
		LastCheck:           time.Now().UTC().Add(-1 * time.Minute),
		GoModModTime:        goModStat.ModTime(),
		IdentityModTime:     oldIdentityModTime,
		HasForbiddenImports: false,
		ForbiddenImports:    []string{},
	}

	err = saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save cache should succeed")

	// Modify identity module file
	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

	testFile := filepath.Join(identityDir, "test.go")
	newContent := `package identity

import (
	"fmt"
	"encoding/json"
	"time"
)

func Test() {
	fmt.Println("updated")
}
`
	err = os.WriteFile(testFile, []byte(newContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Update test file should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should invalidate cache due to identity module change
	logger := common.NewLogger("test-check-identity-changed")
	err = Check(logger)

	testify.NoError(t, err, "Check should succeed after identity module change")
}

// TestCheck_CacheSaveError tests Check when cache save fails (non-writable directory).
func TestCheck_CacheSaveError(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, false)

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - cache save may fail but check should still succeed
	logger := common.NewLogger("test-check-cache-save-error")
	err = Check(logger)

	// Check itself should succeed even if cache save fails
	testify.NoError(t, err, "Check should succeed even if cache save fails")
}

// TestCheck_MissingGoMod tests Check when go.mod doesn't exist.
func TestCheck_MissingGoMod(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()
	setupTestEnvironment(t, tempDir, false)

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

// TestCheck_MissingIdentityModule tests Check when internal/identity directory doesn't exist.
func TestCheck_MissingIdentityModule(t *testing.T) {
	// Note: Cannot use t.Parallel() because test changes working directory
	tempDir := t.TempDir()

	// Don't create identity module

	// Create go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module cryptoutil\n\ngo 1.25\n"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Create go.mod should succeed")

	// Change to temp directory
	origDir := changeToTempDir(t, tempDir)
	defer restoreDir(t, origDir)

	// Run check - should fail due to missing identity module
	logger := common.NewLogger("test-check-missing-identity")
	err = Check(logger)

	testify.Error(t, err, "Check should fail when identity module is missing")
}

// Helper: setupTestEnvironment creates test identity module with or without violations.
func setupTestEnvironment(t *testing.T, tempDir string, withViolations bool) {
	t.Helper()

	identityDir := filepath.Join(tempDir, "internal", "identity")
	err := os.MkdirAll(identityDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create identity directory should succeed")

	var content string
	if withViolations {
		content = `package identity

import (
	"fmt"
	"cryptoutil/internal/server"
)

func Test() {
	fmt.Println("test")
}`
	} else {
		content = `package identity

import (
	"fmt"
	"encoding/json"
)

func Test() {
	fmt.Println("test")
}`
	}

	testFile := filepath.Join(identityDir, "test.go")
	err = os.WriteFile(testFile, []byte(content), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write test file should succeed")
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
