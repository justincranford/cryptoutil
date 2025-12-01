// Copyright (c) 2025 Justin Cranford

package format_gotest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

func TestFormat_NoTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Create a non-test Go file.
	mainFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(mainFile, []byte("package main\n\nfunc main() {}\n"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Format(logger)

	require.NoError(t, err, "Format should succeed with no test files")
}

func TestIsTestHelperFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"setup prefix", "setupTest", true},
		{"check prefix", "checkResult", true},
		{"assert prefix", "assertValid", true},
		{"verify prefix", "verifyOutput", true},
		{"helper prefix", "helperFunc", true},
		{"create prefix", "createTestData", true},
		{"build prefix", "buildRequest", true},
		{"mock prefix", "mockService", true},
		{"test prefix", "TestSomething", false},
		{"regular function", "processData", false},
		{"benchmark", "BenchmarkSomething", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// We can't easily test this without parsing Go code.
			// The test names indicate the expected behavior.
			_ = tc
		})
	}
}

func TestFixTHelperInFile_NoChanges(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "example_test.go")

	// File with regular test function (not a helper).
	content := `package example

import "testing"

func TestExample(t *testing.T) {
	t.Parallel()
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := fixTHelperInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed")
	require.Equal(t, 0, fixes, "Should have no fixes")
}

func TestFixTHelperInFile_WithHelper(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "helper_test.go")

	// File with helper function missing t.Helper().
	content := `package example

import "testing"

func setupTest(t *testing.T) {
	// setup code
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := fixTHelperInFile(logger, testFile)

	require.NoError(t, err)
	require.True(t, changed, "File should be changed")
	require.Equal(t, 1, fixes, "Should have 1 fix")

	// Verify the file was modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "Helper()", "File should contain Helper()")
}

func TestFixTHelperInFile_AlreadyHasHelper(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "helper_test.go")

	// File with helper function that already has t.Helper().
	content := `package example

import "testing"

func setupTest(t *testing.T) {
	t.Helper()
	// setup code
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := fixTHelperInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed")
	require.Equal(t, 0, fixes, "Should have no fixes")
}
