// Copyright (c) 2025-2026 Justin Cranford.
package no_unaliased_cryptoutil_imports

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Test constants for repeated string literals.
const (
	testCleanGoFile  = "clean.go"
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
	cryptoutilCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
)

func main() {
	_ = cryptoutilMagic.TestValue
	_ = cryptoutilCommon.NewLogger("test")
}
`

	err := os.WriteFile(cleanFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
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

	err := os.WriteFile(unaliasedFile, []byte(content.String()), cryptoutilSharedMagic.CacheFilePermissions)
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

	err := os.WriteFile(singleLineFile, []byte(content.String()), cryptoutilSharedMagic.CacheFilePermissions)
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

func TestCheckGoFileForUnaliasedCryptoutilImports_ScannerError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a Go file with a line longer than bufio.MaxScanTokenSize (64KB) to trigger scanner.Err().
	longLine := "// " + strings.Repeat("x", 70000) + "\n"
	goFile := filepath.Join(tempDir, "long.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"+longLine), cryptoutilSharedMagic.CacheFilePermissions))

	_, err := CheckGoFileForUnaliasedCryptoutilImports(goFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error reading")
}

// Sequential: replaces process-global os.Stderr.
func TestPrintCryptoutilImportViolations(t *testing.T) {
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

// Sequential: changes the process working directory with os.Chdir.
func TestFindUnaliasedCryptoutilImports_WithTempDir(t *testing.T) {
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
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Test - should have no violations.
	violations, err := FindUnaliasedCryptoutilImports(filepath.Walk)
	require.NoError(t, err)
	require.Empty(t, violations)
}

// Sequential: changes the process working directory with os.Chdir.
func TestCheckNoUnaliasedCryptoutilImports_WithTempDir(t *testing.T) {
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
	require.NoError(t, os.WriteFile(testCleanGoFile, []byte(testCleanContent), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should pass with no violations.
	err = Check(logger)
	require.NoError(t, err)
}

// Sequential: changes the process working directory and replaces os.Stderr.
func TestCheckNoUnaliasedCryptoutilImports_WithViolations(t *testing.T) {
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
	require.NoError(t, os.WriteFile("bad.go", []byte(badContent), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Test - should fail with violations.
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unaliased cryptoutil imports")
}

// Sequential: changes the process working directory with os.Chdir.
func TestFindUnaliasedCryptoutilImports_VendorDirSkipped(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	// Create a vendor directory with a Go file that has unaliased cryptoutil imports.
	vendorDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirVendor)
	require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	vendorFile := filepath.Join(vendorDir, "pkg.go")
	require.NoError(t, os.WriteFile(vendorFile, []byte("package vendor\nimport \"cryptoutil/internal/apps/foo\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// FindUnaliasedCryptoutilImports should skip vendor directory.
	violations, findErr := FindUnaliasedCryptoutilImports(filepath.Walk)
	require.NoError(t, findErr)
	require.Empty(t, violations, "vendor directory should be skipped")
}

// Sequential: changes the process working directory with os.Chdir.
func TestCheck_GoFileScanError(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	longLine := "// " + strings.Repeat("x", 70000) + "\n"
	require.NoError(t, os.WriteFile("bad.go", []byte("package main\n"+longLine), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// Check() calls FindUnaliasedCryptoutilImports() which fails on unreadable file.
	// This covers the "failed to check cryptoutil imports" error branch in Check().
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check cryptoutil imports")
}

// TestFindUnaliasedCryptoutilImports_WalkSeamError tests the walk error path with fn-param injection.
func TestFindUnaliasedCryptoutilImports_WalkSeamError(t *testing.T) {
	t.Parallel()

	stubWalkFn := func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	_, err := FindUnaliasedCryptoutilImports(stubWalkFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error walking directory tree")
}
