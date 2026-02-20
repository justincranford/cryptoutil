// Copyright (c) 2025 Justin Cranford

package no_unaliased_cryptoutil_imports

import (
	"testing"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

// Test constants for repeated string literals.
const (
	osWindows       = "windows"
	testCleanGoFile = "clean.go"
	testCleanContent = "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\") }\n"
)

func TestCheckGoFileForUnaliasedCryptoutilImports_Clean(t *testing.T) {
	t.Parallel()

	// Create temp file with properly aliased imports.
	tmpDir := t.TempDir()
	cleanFile := filepath.Join(tmpDir, "clean.go")

	content := `package main

import (
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilCommon "cryptoutil/internal/apps/cicd/common"
)

func main() {
	_ = cryptoutilMagic.TestValue
	_ = cryptoutilCommon.NewLogger("test")
}
`

	err := os.WriteFile(cleanFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoFileForUnaliasedCryptoutilImports(cleanFile)
	require.NoError(t, err)
	require.Empty(t, violations, "Properly aliased imports should have no violations")
}

func TestCheckGoFileForUnaliasedCryptoutilImports_Unaliased(t *testing.T) {
	t.Parallel()

	// Create temp file with unaliased cryptoutil import.
	// Using raw string builder to avoid linter flagging this test file.
	tmpDir := t.TempDir()
	unaliasedFile := filepath.Join(tmpDir, "unaliased.go")

	// Build content dynamically to avoid false positive from import checker.
	var content strings.Builder
	content.WriteString("package main\n\nimport (\n\t\"")
	content.WriteString("cryptoutil/internal/shared/magic")
	content.WriteString("\"\n)\n\nfunc main() {\n\t_ = magic.TestValue\n}\n")

	err := os.WriteFile(unaliasedFile, []byte(content.String()), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoFileForUnaliasedCryptoutilImports(unaliasedFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Unaliased cryptoutil import should be detected")
	require.Contains(t, strings.Join(violations, "\n"), "unaliased cryptoutil import detected")
}

func TestCheckGoFileForUnaliasedCryptoutilImports_SingleLineImport(t *testing.T) {
	t.Parallel()

	// Create temp file with single-line unaliased import.
	// Using raw string builder to avoid linter flagging this test file.
	tmpDir := t.TempDir()
	singleLineFile := filepath.Join(tmpDir, "singleline.go")

	// Build content dynamically to avoid false positive from import checker.
	var content strings.Builder
	content.WriteString("package main\n\nimport \"")
	content.WriteString("cryptoutil/internal/shared/magic")
	content.WriteString("\"\n\nfunc main() {\n}\n")

	err := os.WriteFile(singleLineFile, []byte(content.String()), 0o600)
	require.NoError(t, err)

	violations, err := CheckGoFileForUnaliasedCryptoutilImports(singleLineFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations, "Single-line unaliased import should be detected")
}

func TestCheckGoFileForUnaliasedCryptoutilImports_FileNotFound(t *testing.T) {
	t.Parallel()

	violations, err := CheckGoFileForUnaliasedCryptoutilImports("/nonexistent/path/file.go")
	require.Error(t, err)
	require.Nil(t, violations)
	require.Contains(t, err.Error(), "failed to open")
}

func TestPrintCryptoutilImportViolations(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test redirects os.Stderr which is global.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	violations := []string{
		"file1.go:5: unaliased cryptoutil import detected",
		"file2.go:10: unaliased cryptoutil import detected",
	}

	PrintCryptoutilImportViolations(violations)

	_ = w.Close()
	os.Stderr = oldStderr

	output, _ := io.ReadAll(r)

	require.Contains(t, string(output), "Unaliased cryptoutil imports found")
	require.Contains(t, string(output), "file1.go")
	require.Contains(t, string(output), "file2.go")
	require.Contains(t, string(output), "golangci-lint run --fix")
}

func TestFindUnaliasedCryptoutilImports_WithTempDir(t *testing.T) {
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

	// Create clean Go file with no cryptoutil imports.
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), 0o600))

	// Test - should have no violations.
	violations, err := FindUnaliasedCryptoutilImports()
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckNoUnaliasedCryptoutilImports_WithTempDir(t *testing.T) {
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

	// Create clean Go file with no cryptoutil imports.
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with no violations.
	err = Check(logger)
	require.NoError(t, err)
}

func TestCheckNoUnaliasedCryptoutilImports_WithViolations(t *testing.T) {
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

	// Build cryptoutil import path dynamically to avoid self-flagging.
	var importPath strings.Builder
	importPath.WriteString("cryptoutil/")
	importPath.WriteString("internal/shared/magic")

	// Create Go file with unaliased cryptoutil import.
	badContent := "package main\n\nimport \"" + importPath.String() + "\"\n\nfunc main() {}\n"
	require.NoError(t, os.WriteFile("bad.go", []byte(badContent), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail with violations.
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unaliased cryptoutil imports")
}

func TestFindUnaliasedCryptoutilImports_ErrorPath(t *testing.T) {
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

	// Create a file (not a directory) that will be treated as a directory during walk.
	// This will cause an error in filepath.Walk.
	require.NoError(t, os.WriteFile("main.go", []byte("package main\n"), 0o600))

	// Make main.go unreadable to trigger error.
	require.NoError(t, os.Chmod("main.go", 0o000))

	defer func() {
		// Restore permissions for cleanup.
		_ = os.Chmod(filepath.Join(tempDir, "main.go"), 0o600)
	}()

	// Test - should get error from reading file.
	_, err = FindUnaliasedCryptoutilImports()
	require.Error(t, err)
}
