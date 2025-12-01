// Copyright (c) 2025 Justin Cranford

package format_gotest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilCmdCicdFormatGotest "cryptoutil/internal/cmd/cicd/format_gotest"
)

func TestFormat_NoTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a non-test Go file.
	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte("package main\n\nfunc main() {}\n"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed with no test files")
}

func TestFormat_WithHelperNeedingFix(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with helper function missing t.Helper().
	testFile := filepath.Join(tmpDir, "helper_test.go")
	content := `package example

import "testing"

func setupTest(t *testing.T) {
	doSomething()
}

func doSomething() {}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed")

	// Verify the file was modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), ".Helper()", "File should contain .Helper()")
}

func TestFormat_AlreadyHasHelper(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with helper function that already has t.Helper().
	testFile := filepath.Join(tmpDir, "helper_test.go")
	content := `package example

import "testing"

func setupTest(t *testing.T) {
	t.Helper()
	// setup code
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	originalContent := content

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed")

	// Verify the file was not modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, originalContent, string(modifiedContent), "File should not be changed")
}
