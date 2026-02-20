// Copyright (c) 2025 Justin Cranford

package non_fips_algorithms

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// Test constants for repeated string literals.
const (
	osWindows         = "windows"
	testPackageMainDef = "package main\n"
)

func TestFindGoFiles_WithTempDir(t *testing.T) {
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

	// Create Go files.
	require.NoError(t, os.WriteFile("main.go", []byte(testPackageMainDef), 0o600))
	require.NoError(t, os.WriteFile("util.go", []byte(testPackageMainDef), 0o600))
	require.NoError(t, os.WriteFile("main_test.go", []byte(testPackageMainDef), 0o600))

	// Create excluded directories.
	require.NoError(t, os.MkdirAll("vendor", 0o755))
	require.NoError(t, os.WriteFile("vendor/vendored.go", []byte("package vendor\n"), 0o600))

	// Test - should find main.go and util.go, but NOT test files, vendor files.
	files, err := FindGoFiles()
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestCheckNonFIPS_WithTempDir(t *testing.T) {
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

	// Create clean Go file without banned algorithms.
	cleanContent := "package main\n\nimport (\n\t\"crypto/sha256\"\n)\n\nfunc main() { sha256.New() }\n"
	require.NoError(t, os.WriteFile("main.go", []byte(cleanContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with FIPS-compliant code.
	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckNonFIPS_WithViolations(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory and redirects stderr.

	// Save current directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	// Create temp directory.
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	// Create Go file with banned algorithm (bcrypt).
	badContent := "package main\n\nimport \"golang.org/x/crypto/bcrypt\"\n\nfunc main() { bcrypt.GenerateFromPassword(nil, 0) }\n"
	require.NoError(t, os.WriteFile("bad.go", []byte(badContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail with violations.
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-FIPS algorithm violations")
}

func TestFindGoFiles_ErrorPath(t *testing.T) {
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

	// Create a subdirectory that will trigger walk error.
	subDir := "subdir"
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "file.go"), []byte("package main\n"), 0o600))

	// Make subdirectory unreadable.
	require.NoError(t, os.Chmod(subDir, 0o000))

	defer func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(filepath.Join(tempDir, subDir), 0o755)
	}()

	// Test - should get error from walking directory.
	_, err = FindGoFiles()
	require.Error(t, err)
}
