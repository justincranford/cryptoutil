// Copyright (c) 2025 Justin Cranford

package go_enforce_any

import (
	"os"
	"path/filepath"
	"testing"

	testify "github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilTestutil "cryptoutil/internal/common/testutil"
)

// TestEnforce_NoGoFiles tests Enforce with no Go files.
func TestEnforce_NoGoFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create non-Go files
	txtFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.txt", "some text")
	mdFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "README.md", "# Readme")

	logger := common.NewLogger("test-enforce-nogofiles")
	err := Enforce(logger, []string{txtFile, mdFile})

	testify.NoError(t, err, "Enforce should succeed with no Go files")
}

// TestEnforce_AllFilesAlreadyClean tests Enforce when no replacements needed.
func TestEnforce_AllFilesAlreadyClean(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create Go files with 'any' already used
	content1 := `package test

var x any = 42
`
	testFile1 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test1.go", content1)

	content2 := `package test

func process(data any) any {
	return data
}
`
	testFile2 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test2.go", content2)

	logger := common.NewLogger("test-enforce-clean")
	err := Enforce(logger, []string{testFile1, testFile2})

	testify.NoError(t, err, "Enforce should succeed with no replacements")
}

// TestEnforce_ExcludedFiles tests that excluded files are not processed.
func TestEnforce_ExcludedFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create cicd_enforce_any.go file (should be excluded by pattern)
	cicdDir := filepath.Join(tempDir, "internal", "cmd", "cicd")
	err := os.MkdirAll(cicdDir, 0o755)
	testify.NoError(t, err, "Create directory should succeed")

	excludedContent := `package cicd

func enforceFn() {
	var x any
}
`
	excludedFile := filepath.Join(cicdDir, "cicd_enforce_any.go")
	err = os.WriteFile(excludedFile, []byte(excludedContent), 0o600)
	testify.NoError(t, err, "Write excluded file should succeed")

	// Create non-excluded file
	normalContent := `package test

var x any
`
	normalFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "normal.go", normalContent)

	logger := common.NewLogger("test-enforce-excluded")
	err = Enforce(logger, []string{excludedFile, normalFile})

	testify.Error(t, err, "Should return error for modified normal file")
	testify.Contains(t, err.Error(), "modified 1 files", "Should only modify 1 file (not excluded)")

	// Verify excluded file was NOT modified
	excludedActual := cryptoutilTestutil.ReadTestFile(t, excludedFile)
	testify.Equal(t, excludedContent, string(excludedActual), "Excluded file should not be modified")

	// Verify normal file WAS modified
	normalActual := cryptoutilTestutil.ReadTestFile(t, normalFile)
	testify.Contains(t, string(normalActual), "var x any", "Normal file should be modified")
}

// TestEnforce_MixedFiles tests Enforce with mix of modified and clean files.
func TestEnforce_MixedFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// File 1: Needs modification
	content1 := `package test

var x any
`
	testFile1 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test1.go", content1)

	// File 2: Already clean
	content2 := `package test

var y any
`
	testFile2 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test2.go", content2)

	// File 3: Needs modification
	content3 := `package test

func process(data any) {
}
`
	testFile3 := cryptoutilTestutil.WriteTempFile(t, tempDir, "test3.go", content3)

	logger := common.NewLogger("test-enforce-mixed")
	err := Enforce(logger, []string{testFile1, testFile2, testFile3})

	testify.Error(t, err, "Should return error when some files modified")
	testify.Contains(t, err.Error(), "modified 2 files", "Error should mention 2 modified files")
}

// TestEnforce_NonGoFilesIgnored tests that non-Go files are ignored.
func TestEnforce_NonGoFilesIgnored(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create mix of Go and non-Go files
	goFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.go", "package test\nvar x any")
	txtFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.txt", "any")
	mdFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "README.md", "any")

	logger := common.NewLogger("test-enforce-nongoignored")
	err := Enforce(logger, []string{goFile, txtFile, mdFile})

	testify.Error(t, err, "Should return error for modified Go file")

	// Verify only Go file was modified
	goContent := cryptoutilTestutil.ReadTestFile(t, goFile)
	testify.Contains(t, string(goContent), "var x any", "Go file should be modified")

	txtContent := cryptoutilTestutil.ReadTestFile(t, txtFile)
	testify.Equal(t, "any", string(txtContent), "Text file should not be modified")

	mdContent := cryptoutilTestutil.ReadTestFile(t, mdFile)
	testify.Equal(t, "any", string(mdContent), "Markdown file should not be modified")
}

// TestEnforce_EmptyFileList tests Enforce with empty file list.
func TestEnforce_EmptyFileList(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test-enforce-empty")
	err := Enforce(logger, []string{})

	testify.NoError(t, err, "Enforce should succeed with empty file list")
}

// TestProcessGoFile_NonExistentFile tests processing non-existent file.
func TestProcessGoFile_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := processGoFile("/nonexistent/file.go")

	testify.Error(t, err, "Should fail for non-existent file")
	testify.Contains(t, err.Error(), "failed to read file", "Error should mention read failure")
}

// TestProcessGoFile_MultipleReplacements tests file with many any occurrences.
func TestProcessGoFile_MultipleReplacements(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	content := `package test

type A struct {
	Field1 any
	Field2 any
}

type B struct {
	Field3 any
}

func Fn1(a any) any {
	return a
}

func Fn2(b any, c any) (any, any) {
	return b, c
}

var (
	V1 any
	V2 any
	V3 any
)
`
	testFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "multi.go", content)

	replacements, err := processGoFile(testFile)

	testify.NoError(t, err, "processGoFile should succeed")
	testify.Equal(t, 12, replacements, "Should replace all 12 occurrences")

	// Verify all replaced
	modifiedContent := cryptoutilTestutil.ReadTestFile(t, testFile)
	testify.NotContains(t, string(modifiedContent), "any", "Should not contain any after replacement")
	testify.Contains(t, string(modifiedContent), "Field1 any", "Field1 should use any")
	testify.Contains(t, string(modifiedContent), "Field2 any", "Field2 should use any")
	testify.Contains(t, string(modifiedContent), "Field3 any", "Field3 should use any")
	testify.Contains(t, string(modifiedContent), "func Fn1(a any) any", "Fn1 should use any")
	testify.Contains(t, string(modifiedContent), "V1 any", "V1 should use any")
	testify.Contains(t, string(modifiedContent), "V2 any", "V2 should use any")
	testify.Contains(t, string(modifiedContent), "V3 any", "V3 should use any")
}
