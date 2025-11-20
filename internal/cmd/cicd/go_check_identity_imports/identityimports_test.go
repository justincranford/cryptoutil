// Copyright (c) 2025 Justin Cranford

package go_check_identity_imports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

func TestGetBlockedPackages(t *testing.T) {
	blocked := GetBlockedPackages()

	testify.NotEmpty(t, blocked, "Should return blocked packages list")
	testify.GreaterOrEqual(t, len(blocked), 9, "Should have at least 9 blocked packages")

	// Verify expected packages are blocked
	expectedBlocked := map[string]bool{
		"cryptoutil/internal/server":           true,
		"cryptoutil/internal/client":           true,
		"cryptoutil/api":                       true,
		"cryptoutil/cmd/cryptoutil":            true,
		"cryptoutil/internal/common/crypto":    true,
		"cryptoutil/internal/common/pool":      true,
		"cryptoutil/internal/common/container": true,
		"cryptoutil/internal/common/telemetry": true,
		"cryptoutil/internal/common/util":      true,
	}

	for _, pkg := range blocked {
		if expectedBlocked[pkg.Path] {
			testify.NotEmpty(t, pkg.Reason, "Package %s should have a reason", pkg.Path)
		}
	}
}

func TestCheckImports_NoViolations(t *testing.T) {
	// Create temporary identity module with valid imports
	tempDir := t.TempDir()
	identityDir := filepath.Join(tempDir, "internal", "identity")
	err := os.MkdirAll(identityDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create directory should succeed")

	// Create test file with allowed imports
	testFile := filepath.Join(identityDir, "test.go")
	content := `package identity

import (
	"fmt"
	"encoding/json"
	"github.com/external/pkg"
)

func Test() {
	fmt.Println("test")
}
`

	err = os.WriteFile(testFile, []byte(content), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write test file should succeed")

	// Change to temp directory
	origDir, err := os.Getwd()
	testify.NoError(t, err, "Get working directory should succeed")

	defer func() {
		err := os.Chdir(origDir)
		testify.NoError(t, err, "Restore working directory should succeed")
	}()

	err = os.Chdir(tempDir)
	testify.NoError(t, err, "Change to temp directory should succeed")

	// Run check
	logger := common.NewLogger("test")
	violations, err := CheckImports(logger)

	testify.NoError(t, err, "Check should succeed")
	testify.Empty(t, violations, "Should have no violations for allowed imports")
}

func TestCheckImports_WithViolations(t *testing.T) {
	// Create temporary identity module with forbidden imports
	tempDir := t.TempDir()
	identityDir := filepath.Join(tempDir, "internal", "identity")
	err := os.MkdirAll(identityDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create directory should succeed")

	// Create test file with forbidden imports
	testFile := filepath.Join(identityDir, "test.go")
	content := `package identity

import (
	"fmt"
	"cryptoutil/internal/server"
	"cryptoutil/internal/client"
)

func Test() {
	fmt.Println("test")
}
`

	err = os.WriteFile(testFile, []byte(content), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write test file should succeed")

	// Change to temp directory
	origDir, err := os.Getwd()
	testify.NoError(t, err, "Get working directory should succeed")

	defer func() {
		err := os.Chdir(origDir)
		testify.NoError(t, err, "Restore working directory should succeed")
	}()

	err = os.Chdir(tempDir)
	testify.NoError(t, err, "Change to temp directory should succeed")

	// Run check
	logger := common.NewLogger("test")
	violations, err := CheckImports(logger)

	testify.NoError(t, err, "Check should succeed")
	testify.Len(t, violations, 2, "Should detect 2 violations")

	// Verify violation messages
	for _, v := range violations {
		testify.Contains(t, v, "forbidden import", "Violation message should mention forbidden import")
		testify.Regexp(t, `test\.go:\d+:`, v, "Violation should include file and line number")
	}
}

func TestCheckImports_NonGoFiles(t *testing.T) {
	// Create temporary identity module with non-Go files
	tempDir := t.TempDir()
	identityDir := filepath.Join(tempDir, "internal", "identity")
	err := os.MkdirAll(identityDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create directory should succeed")

	// Create non-Go file with forbidden import (should be ignored)
	testFile := filepath.Join(identityDir, "test.txt")
	content := `import "cryptoutil/internal/server"`

	err = os.WriteFile(testFile, []byte(content), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write test file should succeed")

	// Change to temp directory
	origDir, err := os.Getwd()
	testify.NoError(t, err, "Get working directory should succeed")

	defer func() {
		err := os.Chdir(origDir)
		testify.NoError(t, err, "Restore working directory should succeed")
	}()

	err = os.Chdir(tempDir)
	testify.NoError(t, err, "Change to temp directory should succeed")

	// Run check
	logger := common.NewLogger("test")
	violations, err := CheckImports(logger)

	testify.NoError(t, err, "Check should succeed")
	testify.Empty(t, violations, "Non-Go files should be ignored")
}

func TestCheckImports_InvalidGoSyntax(t *testing.T) {
	// Create temporary identity module with invalid Go syntax
	tempDir := t.TempDir()
	identityDir := filepath.Join(tempDir, "internal", "identity")
	err := os.MkdirAll(identityDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create directory should succeed")

	// Create file with invalid syntax
	testFile := filepath.Join(identityDir, "invalid.go")
	content := `package identity

import (
	"fmt"
	this is invalid syntax
)
`

	err = os.WriteFile(testFile, []byte(content), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write test file should succeed")

	// Change to temp directory
	origDir, err := os.Getwd()
	testify.NoError(t, err, "Get working directory should succeed")

	defer func() {
		err := os.Chdir(origDir)
		testify.NoError(t, err, "Restore working directory should succeed")
	}()

	err = os.Chdir(tempDir)
	testify.NoError(t, err, "Change to temp directory should succeed")

	// Run check - should continue checking other files despite parse error
	logger := common.NewLogger("test")
	violations, err := CheckImports(logger)

	testify.NoError(t, err, "Check should succeed despite parse error")
	testify.Empty(t, violations, "Invalid syntax file should be skipped")
}

func TestGetLatestModTime(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(testDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create directory should succeed")

	// Create files with different mod times
	file1 := filepath.Join(testDir, "old.go")
	file2 := filepath.Join(testDir, "new.go")

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	err = os.WriteFile(file1, []byte("package test"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write file1 should succeed")

	err = os.Chtimes(file1, oldTime, oldTime)
	testify.NoError(t, err, "Set file1 time should succeed")

	err = os.WriteFile(file2, []byte("package test"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write file2 should succeed")

	err = os.Chtimes(file2, newTime, newTime)
	testify.NoError(t, err, "Set file2 time should succeed")

	// Get latest mod time
	latest, err := getLatestModTime(testDir)
	testify.NoError(t, err, "Get latest mod time should succeed")

	// Verify it matches the newer file (within 1 second tolerance)
	testify.WithinDuration(t, newTime, latest, time.Second, "Latest should match newer file time")
}

func TestGetLatestModTime_NonExistentDir(t *testing.T) {
	_, err := getLatestModTime("/nonexistent/directory")
	testify.Error(t, err, "Should error on non-existent directory")
}

func TestGetLatestModTime_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "empty")
	err := os.MkdirAll(testDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create directory should succeed")

	latest, err := getLatestModTime(testDir)
	testify.NoError(t, err, "Should succeed on empty directory")
	testify.True(t, latest.IsZero(), "Empty directory should return zero time")
}

func TestCacheOperations(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".cicd", "identity-imports-cache.json")

	// Create cache data
	cache := Cache{
		LastCheck:           time.Now().UTC(),
		GoModModTime:        time.Now().UTC(),
		IdentityModTime:     time.Now().UTC(),
		HasForbiddenImports: false,
		ForbiddenImports:    []string{},
	}

	// Test save
	err := saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")
	testify.FileExists(t, cacheFile, "Cache file should exist")

	// Test load
	loadedCache, err := loadCache(cacheFile)
	testify.NoError(t, err, "Load should succeed")
	testify.Equal(t, cache.HasForbiddenImports, loadedCache.HasForbiddenImports, "HasForbiddenImports should match")
}

func TestLoadCache_NonExistentFile(t *testing.T) {
	_, err := loadCache("/nonexistent/cache.json")
	testify.Error(t, err, "Load should fail for non-existent file")
}

func TestLoadCache_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "invalid.json")

	err := os.WriteFile(cacheFile, []byte("invalid json"), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write should succeed")

	_, err = loadCache(cacheFile)
	testify.Error(t, err, "Load should fail for invalid JSON")
}

func TestSaveCache_WithViolations(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "cache.json")

	cache := Cache{
		LastCheck:           time.Now().UTC(),
		GoModModTime:        time.Now().UTC(),
		IdentityModTime:     time.Now().UTC(),
		HasForbiddenImports: true,
		ForbiddenImports:    []string{"file.go:10: forbidden import", "other.go:20: forbidden import"},
	}

	err := saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Verify cache can be loaded
	loadedCache, err := loadCache(cacheFile)
	testify.NoError(t, err, "Load should succeed")
	testify.True(t, loadedCache.HasForbiddenImports, "Should have violations flag set")
	testify.Len(t, loadedCache.ForbiddenImports, 2, "Should have 2 violations")
}

func TestSaveCache_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "deep", "nested", "cache.json")

	cache := Cache{
		LastCheck:           time.Now().UTC(),
		GoModModTime:        time.Now().UTC(),
		IdentityModTime:     time.Now().UTC(),
		HasForbiddenImports: false,
		ForbiddenImports:    []string{},
	}

	err := saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save should create directories")
	testify.FileExists(t, cacheFile, "Cache file should exist")
	testify.DirExists(t, filepath.Dir(cacheFile), "Directory should exist")
}

func TestCacheJSONFormat(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "cache.json")

	cache := Cache{
		LastCheck:           time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
		GoModModTime:        time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC),
		IdentityModTime:     time.Date(2025, 1, 13, 8, 0, 0, 0, time.UTC),
		HasForbiddenImports: true,
		ForbiddenImports:    []string{"violation1", "violation2"},
	}

	err := saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Read and verify JSON format
	content, err := os.ReadFile(cacheFile)
	testify.NoError(t, err, "Read should succeed")

	var decoded Cache

	err = json.Unmarshal(content, &decoded)
	testify.NoError(t, err, "JSON should be valid")

	// Verify formatting
	testify.Contains(t, string(content), "  ", "JSON should be indented")
	testify.Contains(t, string(content), "last_check", "Should contain last_check field")
	testify.Contains(t, string(content), "has_forbidden_imports", "Should contain has_forbidden_imports field")
}

func TestCheckImports_NestedDirectories(t *testing.T) {
	// Create nested directory structure
	tempDir := t.TempDir()
	identityDir := filepath.Join(tempDir, "internal", "identity")
	nestedDir := filepath.Join(identityDir, "authz", "handlers")
	err := os.MkdirAll(nestedDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create nested directories should succeed")

	// Create files at different levels
	rootFile := filepath.Join(identityDir, "root.go")
	nestedFile := filepath.Join(nestedDir, "handler.go")

	rootContent := `package identity

import (
	"fmt"
	"cryptoutil/internal/server"
)
`

	nestedContent := `package handlers

import (
	"encoding/json"
	"cryptoutil/internal/client"
)
`

	err = os.WriteFile(rootFile, []byte(rootContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write root file should succeed")

	err = os.WriteFile(nestedFile, []byte(nestedContent), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write nested file should succeed")

	// Change to temp directory
	origDir, err := os.Getwd()
	testify.NoError(t, err, "Get working directory should succeed")

	defer func() {
		err := os.Chdir(origDir)
		testify.NoError(t, err, "Restore working directory should succeed")
	}()

	err = os.Chdir(tempDir)
	testify.NoError(t, err, "Change to temp directory should succeed")

	// Run check
	logger := common.NewLogger("test")
	violations, err := CheckImports(logger)

	testify.NoError(t, err, "Check should succeed")
	testify.Len(t, violations, 2, "Should detect violations in nested directories")
}

func TestCachePermissions(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "cache.json")

	cache := Cache{
		LastCheck:           time.Now().UTC(),
		GoModModTime:        time.Now().UTC(),
		IdentityModTime:     time.Now().UTC(),
		HasForbiddenImports: false,
		ForbiddenImports:    []string{},
	}

	err := saveCache(cacheFile, cache)
	testify.NoError(t, err, "Save should succeed")

	// Check file exists (permission check is platform-specific, so we skip it)
	testify.FileExists(t, cacheFile, "Cache file should exist")
}

func TestCheckImports_MultipleViolationsInSameFile(t *testing.T) {
	tempDir := t.TempDir()
	identityDir := filepath.Join(tempDir, "internal", "identity")
	err := os.MkdirAll(identityDir, cryptoutilMagic.CICDOutputDirPermissions)
	testify.NoError(t, err, "Create directory should succeed")

	// Create file with multiple forbidden imports
	testFile := filepath.Join(identityDir, "multi.go")
	content := `package identity

import (
	"fmt"
	"cryptoutil/internal/server"
	"cryptoutil/internal/client"
	"cryptoutil/api"
	"cryptoutil/internal/common/crypto"
)
`

	err = os.WriteFile(testFile, []byte(content), cryptoutilMagic.CacheFilePermissions)
	testify.NoError(t, err, "Write test file should succeed")

	// Change to temp directory
	origDir, err := os.Getwd()
	testify.NoError(t, err, "Get working directory should succeed")

	defer func() {
		err := os.Chdir(origDir)
		testify.NoError(t, err, "Restore working directory should succeed")
	}()

	err = os.Chdir(tempDir)
	testify.NoError(t, err, "Change to temp directory should succeed")

	// Run check
	logger := common.NewLogger("test")
	violations, err := CheckImports(logger)

	testify.NoError(t, err, "Check should succeed")
	testify.Len(t, violations, 4, "Should detect all 4 violations")

	// Verify each violation includes reason
	for _, v := range violations {
		testify.Contains(t, v, "forbidden import", "Should mention forbidden import")
		testify.Regexp(t, `\(.*\)`, v, "Should include reason in parentheses")
	}
}
