// Copyright (c) 2025-2026 Justin Cranford.
package cgo_free_sqlite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Test constants for repeated string literals.
const (
	testCleanGoFile  = "clean.go"
	testCleanContent = "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\") }\n"
	testMainContent  = "package main\n\nfunc main() {}\n"
)

// Sequential: changes the process working directory with os.Chdir.
func TestCheckGoFilesForCGO_WithTempDir(t *testing.T) {
	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory with test files.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create clean Go file.
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Test with clean file - should have no violations.
	violations, err := CheckGoFilesForCGO()
	require.NoError(t, err)
	require.Empty(t, violations)
}

// Sequential: changes the process working directory with os.Chdir.
func TestCheckGoFilesForCGO_WithBannedImport(t *testing.T) {
	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory with test files.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Build banned import string dynamically to avoid self-flagging.
	var banned strings.Builder
	banned.WriteString("github.com/")
	banned.WriteString("mattn/go-sqlite3")

	// Create file with banned CGO import.
	bannedFile := testBannedGoFilename
	bannedContent := "package main\n\nimport _ \"" + banned.String() + "\"\n\nfunc main() {}\n"

	require.NoError(t, os.WriteFile(bannedFile, []byte(bannedContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Test - should find the violation.
	violations, err := CheckGoFilesForCGO()
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], testBannedGoFilename)
}

// Sequential: changes the process working directory with os.Chdir.
func TestCheckGoFilesForCGO_SkipsExcludedDirectories(t *testing.T) {
	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	var banned strings.Builder
	banned.WriteString("github.com/")
	banned.WriteString("mattn/go-sqlite3")

	for _, excludedDir := range cryptoutilSharedMagic.DirectoryNameExclusions {
		bannedFile := filepath.Join(excludedDir, testCleanGoFile)
		require.NoError(t, os.MkdirAll(filepath.Dir(bannedFile), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		bannedContent := "package main\n\nimport _ \"" + banned.String() + "\"\n\nfunc init() {}\n"
		require.NoError(t, os.WriteFile(bannedFile, []byte(bannedContent), cryptoutilSharedMagic.CacheFilePermissions))
	}

	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testMainContent), cryptoutilSharedMagic.CacheFilePermissions))

	violations, err := CheckGoFilesForCGO()
	require.NoError(t, err)
	require.Empty(t, violations, "excluded directories should be skipped")
}

func TestCheckGoFileForCGO_ScannerError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a Go file with a line longer than bufio.MaxScanTokenSize (64KB) to trigger scanner.Err().
	longLine := "// " + strings.Repeat("x", 70000) + "\n"
	goFile := filepath.Join(tempDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"+longLine), cryptoutilSharedMagic.CacheFilePermissions))

	_, err := CheckGoFileForCGO(goFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error reading")
}

func TestCheckGoFilesForCGO_ListError(t *testing.T) {
	t.Parallel()

	_, err := checkGoFilesForCGO(func(string) (map[string][]string, error) {
		return nil, os.ErrPermission
	})
	require.ErrorIs(t, err, os.ErrPermission)
	require.ErrorContains(t, err, "error listing project files")
}

func TestCheckGoFilesForCGO_FileError(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(t.TempDir(), "missing.go")
	_, err := checkGoFilesForCGO(func(string) (map[string][]string, error) {
		return map[string][]string{"go": {missingPath}}, nil
	})
	require.ErrorContains(t, err, "error checking")
	require.ErrorContains(t, err, "failed to open")
}

func TestCheckGoFileForCGO_OpenError(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(t.TempDir(), "missing.go")
	_, err := CheckGoFileForCGO(missingPath)
	require.ErrorContains(t, err, "failed to open")
}
