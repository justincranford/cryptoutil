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

// TestFormat tests the Format function which wraps FormatDir with current directory.
func TestFormat(t *testing.T) {
	// Note: Not parallel - Format uses current working directory.
	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := cryptoutilCmdCicdFormatGotest.Format(logger)

	// Format runs on current directory which has test files that already have t.Helper().
	require.NoError(t, err, "Format should succeed on already-formatted files")
}

// TestFormat_HelperWithoutTestingT tests that helper functions without testing.T are skipped.
func TestFormat_HelperWithoutTestingT(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with helper function that doesn't have *testing.T parameter.
	testFile := filepath.Join(tmpDir, "helper_test.go")
	content := `package example

import "testing"

// setupData is a helper function without testing.T parameter.
func setupData() string {
	return "data"
}

func TestExample(t *testing.T) {
	data := setupData()
	_ = data
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed")

	// Verify the file was not modified (no t.Helper() added since no testing.T param).
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), ".Helper()", "File should NOT contain .Helper()")
}

// TestFormat_InvalidGoFile tests that invalid Go files cause parse errors.
func TestFormat_InvalidGoFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with invalid Go syntax.
	testFile := filepath.Join(tmpDir, "invalid_test.go")
	content := `package example

func invalidSyntax( {
	// missing closing paren
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.Error(t, err, "FormatDir should fail on invalid Go file")
	require.Contains(t, err.Error(), "format-go-test failed", "Error should indicate format failure")
}

// TestFormat_BenchmarkHelper tests that helper functions with *testing.B are also handled.
func TestFormat_BenchmarkHelper(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with benchmark helper function missing b.Helper().
	testFile := filepath.Join(tmpDir, "bench_test.go")
	content := `package example

import "testing"

func setupBenchmark(b *testing.B) {
	doSomething()
}

func doSomething() {}

func BenchmarkExample(b *testing.B) {
	setupBenchmark(b)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed")

	// Verify b.Helper() was added.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "b.Helper", "File should contain b.Helper call")
}

// TestFormat_WalkError tests that walk errors are handled.
func TestFormat_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := cryptoutilCmdCicdFormatGotest.FormatDir(logger, "/nonexistent/directory/path")

	require.Error(t, err, "FormatDir should fail on nonexistent directory")
	require.Contains(t, err.Error(), "format-go-test failed", "Error should indicate format failure")
}

// TestFormat_FuncWithoutBody tests helper function declarations without body (interfaces).
func TestFormat_FuncWithoutBody(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with helper function interface (no body).
	testFile := filepath.Join(tmpDir, "interface_test.go")
	content := `package example

import "testing"

// TestInterface defines test helper methods.
type TestInterface interface {
	Setup(t *testing.T)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed with interface methods")
}

// TestFormat_PointerReceiver tests helper function with pointer receiver.
func TestFormat_PointerReceiver(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with helper function with pointer receiver.
	testFile := filepath.Join(tmpDir, "method_test.go")
	content := `package example

import "testing"

type TestSuite struct{}

func (s *TestSuite) setupHelper(t *testing.T) {
	doSomething()
}

func doSomething() {}

func TestExample(t *testing.T) {
	s := &TestSuite{}
	s.setupHelper(t)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed")

	// Verify t.Helper() was added.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "t.Helper", "File should contain t.Helper call")
}

// TestFormat_MixedStatements tests helper function with various statement types.
func TestFormat_MixedStatements(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with helper function containing various statement types.
	testFile := filepath.Join(tmpDir, "mixed_test.go")
	content := `package example

import "testing"

func setupHelper(t *testing.T) {
	// Assignment statement (not expression statement)
	x := 1
	_ = x
	// Function call (expression statement with non-selector call)
	doSomething()
	// Method call on t that is NOT Helper
	t.Log("setup")
}

func doSomething() {}

func TestExample(t *testing.T) {
	setupHelper(t)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed")

	// Verify t.Helper() was added.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "t.Helper", "File should contain t.Helper call")
}

// TestFormat_NonPointerTestingT tests helper with non-pointer testing.T (not matched).
func TestFormat_NonPointerTestingT(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with helper function that has testing.T by value (non-pointer).
	testFile := filepath.Join(tmpDir, "nonptr_test.go")
	content := `package example

import "testing"

// Helper function that takes testing.T by value (unusual but valid).
func setupByValue(t testing.T) {
	doSomething()
}

func doSomething() {}

func TestExample(t *testing.T) {
	setupByValue(*t)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = cryptoutilCmdCicdFormatGotest.FormatDir(logger, tmpDir)

	require.NoError(t, err, "FormatDir should succeed")

	// Verify t.Helper() was NOT added (non-pointer param).
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), ".Helper", "File should NOT contain .Helper call")
}
