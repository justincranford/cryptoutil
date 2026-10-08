// Copyright (c) 2025-2026 Justin Cranford.
package cgo_free_sqlite

import (
	"os"
	"path/filepath"
	"runtime"
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

// Sequential: changes the process working directory with os.Chdir.
func TestCheckGoFilesForCGO_ErrorPath(t *testing.T) {
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create an unreadable Go file.
	require.NoError(t, os.WriteFile("unreadable.go", []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod("unreadable.go", 0o000))

	defer func() {
		_ = os.Chmod(filepath.Join(tempDir, "unreadable.go"), cryptoutilSharedMagic.CacheFilePermissions)
	}()

	// Test - should get error.
	_, err = CheckGoFilesForCGO()
	require.Error(t, err)
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
