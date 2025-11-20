package go_fix_thelper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
)

func TestFix_EmptyDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-empty")

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_NoTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-no-test")

	// Create non-test Go file.
	goFile := filepath.Join(tmpDir, "main.go")
	content := `package main

func main() {}
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, processed) // Non-test files should be skipped.
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_NoHelperFunctions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-no-helpers")

	testFile := filepath.Join(tmpDir, "test_test.go")
	content := `package test

import "testing"

func TestExample(t *testing.T) {
	t.Log("test")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_HelperFunctionMissingTHelper(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-missing")

	testFile := filepath.Join(tmpDir, "helpers_test.go")
	content := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, issuesFixed)

	// Verify t.Helper() was added.
	fixed, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), "t.Helper()")
}

func TestFix_HelperFunctionWithTHelper(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-has-helper")

	testFile := filepath.Join(tmpDir, "helpers_test.go")
	content := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Helper()
	t.Log("setup")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 0, modified) // Already has t.Helper().
	require.Equal(t, 0, issuesFixed)
}

func TestFix_MultipleHelperFunctions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-multiple")

	testFile := filepath.Join(tmpDir, "helpers_test.go")
	content := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}

func checkResult(t *testing.T, expected int) {
	t.Log("checking")
}

func assertValid(t *testing.T) {
	t.Log("asserting")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 3, issuesFixed)

	// Verify t.Helper() was added to all functions.
	fixed, err := os.ReadFile(testFile)
	require.NoError(t, err)

	fixedStr := string(fixed)
	require.Contains(t, fixedStr, "t.Helper()")
}

func TestFix_HelperFunctionPatterns(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-patterns")

	testFile := filepath.Join(tmpDir, "patterns_test.go")
	content := `package test

import "testing"

func setupEnvironment(t *testing.T) {}
func checkData(t *testing.T) {}
func assertCondition(t *testing.T) {}
func verifyState(t *testing.T) {}
func helperFunction(t *testing.T) {}
func createMock(t *testing.T) {}
func buildFixture(t *testing.T) {}
func mockService(t *testing.T) {}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 8, issuesFixed) // All 8 helper functions.
}

func TestFix_NestedDirectories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-nested")

	// Create nested directory structure.
	subDir := filepath.Join(tmpDir, "sub", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	content := `package test
import "testing"
func setupTest(t *testing.T) {
	t.Log("setup")
}
`
	file1 := filepath.Join(tmpDir, "test1_test.go")
	file2 := filepath.Join(tmpDir, "sub", "test2_test.go")
	file3 := filepath.Join(subDir, "test3_test.go")

	require.NoError(t, os.WriteFile(file1, []byte(content), 0o600))
	require.NoError(t, os.WriteFile(file2, []byte(content), 0o600))
	require.NoError(t, os.WriteFile(file3, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 3, processed)
	require.Equal(t, 3, modified)
	require.Equal(t, 3, issuesFixed)
}

func TestFix_InvalidDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-thelper-invalid")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_HelperWithoutTestingParam(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-no-param")

	testFile := filepath.Join(tmpDir, "helpers_test.go")
	content := `package test

func setupGlobal() {
	// No testing.T parameter
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 0, modified) // No testing.T parameter, can't add t.Helper().
	require.Equal(t, 0, issuesFixed)
}
