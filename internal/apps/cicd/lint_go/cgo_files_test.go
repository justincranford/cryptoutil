// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"


	"github.com/stretchr/testify/require"
	lintGoCGOFreeSQLite "cryptoutil/internal/apps/cicd/lint_go/cgo_free_sqlite"
)

func TestCheckGoFilesForCGO_WithTempDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

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
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), 0o600))

	// Test with clean file - should have no violations.
	violations, err := lintGoCGOFreeSQLite.CheckGoFilesForCGO()
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckGoFilesForCGO_WithBannedImport(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

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
	bannedFile := "banned.go"
	bannedContent := "package main\n\nimport _ \"" + banned.String() + "\"\n\nfunc main() {}\n"

	require.NoError(t, os.WriteFile(bannedFile, []byte(bannedContent), 0o600))

	// Test - should find the violation.
	violations, err := lintGoCGOFreeSQLite.CheckGoFilesForCGO()
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0], "banned.go")
}

func TestCheckGoFilesForCGO_SkipsVendor(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create vendor directory with file that would be flagged.
	require.NoError(t, os.MkdirAll("vendor", 0o755))

	var banned strings.Builder
	banned.WriteString("github.com/")
	banned.WriteString("mattn/go-sqlite3")

	vendorFile := "vendor/dep.go"
	vendorContent := "package vendor\n\nimport _ \"" + banned.String() + "\"\n\nfunc init() {}\n"

	require.NoError(t, os.WriteFile(vendorFile, []byte(vendorContent), 0o600))

	// Create clean main file.
	mainFile := "main.go"

	require.NoError(t, os.WriteFile(mainFile, []byte(testMainContent), 0o600))

	// Test - vendor should be skipped, no violations.
	violations, err := lintGoCGOFreeSQLite.CheckGoFilesForCGO()
	require.NoError(t, err)
	require.Empty(t, violations, "vendor directory should be skipped")
}

func TestCheckGoFilesForCGO_ErrorPath(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	if runtime.GOOS == osWindows {
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
	require.NoError(t, os.WriteFile("unreadable.go", []byte("package main\n"), 0o600))
	require.NoError(t, os.Chmod("unreadable.go", 0o000))

	defer func() {
		_ = os.Chmod(filepath.Join(tempDir, "unreadable.go"), 0o600)
	}()

	// Test - should get error.
	_, err = lintGoCGOFreeSQLite.CheckGoFilesForCGO()
	require.Error(t, err)
}


