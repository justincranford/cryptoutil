package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
)

// TestGoFixStaticcheckErrorStrings_Integration tests the wrapper calling fix/staticcheck package.
func TestGoFixStaticcheckErrorStrings_Integration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-staticcheck-wrapper")

	// Create file with staticcheck error string issue.
	errorFile := filepath.Join(tmpDir, "errors.go")
	errorContent := `package test

import "errors"

var ErrFailed = errors.New("Failed to process")
`
	require.NoError(t, os.WriteFile(errorFile, []byte(errorContent), 0o600))

	// Call wrapper.
	err := goFixStaticcheckErrorStrings(logger, tmpDir)
	require.NoError(t, err)

	// Verify fix applied.
	fixed, err := os.ReadFile(errorFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), `errors.New("failed to process")`)
}

// TestGoFixStaticcheckErrorStrings_NoIssues tests wrapper when no issues are found.
func TestGoFixStaticcheckErrorStrings_NoIssues(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-staticcheck-no-issues")

	// Create file WITHOUT staticcheck error string issue.
	cleanFile := filepath.Join(tmpDir, "clean.go")
	cleanContent := `package test

import "errors"

var ErrFailed = errors.New("failed to process")
`
	require.NoError(t, os.WriteFile(cleanFile, []byte(cleanContent), 0o600))

	// Call wrapper.
	err := goFixStaticcheckErrorStrings(logger, tmpDir)
	require.NoError(t, err)
}

// TestGoFixCopyLoopVar_Integration tests the wrapper calling fix/copyloopvar package.
func TestGoFixCopyLoopVar_Integration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-wrapper")

	// Create file with copyloopvar issue.
	loopFile := filepath.Join(tmpDir, "loop.go")
	loopContent := `package test

func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	require.NoError(t, os.WriteFile(loopFile, []byte(loopContent), 0o600))

	// Call wrapper with file list.
	files := []string{loopFile}
	err := goFixCopyLoopVar(logger, files)
	require.Error(t, err) // Should return error indicating fixes made.
	require.Contains(t, err.Error(), "fixed 1 loop variable copies")

	// Verify fix applied.
	fixed, err := os.ReadFile(loopFile)
	require.NoError(t, err)
	require.NotContains(t, string(fixed), "item := item")
}

// TestGoFixCopyLoopVar_NoFiles tests wrapper with empty file list.
func TestGoFixCopyLoopVar_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-copyloopvar-no-files")

	// Call wrapper with empty file list.
	err := goFixCopyLoopVar(logger, []string{})
	require.NoError(t, err)
}

// TestGoFixCopyLoopVar_NoIssues tests wrapper when no issues are found.
func TestGoFixCopyLoopVar_NoIssues(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-no-issues")

	// Create file WITHOUT copyloopvar issue.
	cleanFile := filepath.Join(tmpDir, "clean.go")
	cleanContent := `package test

func Process(items []int) {
	for _, item := range items {
		println(item)
	}
}
`
	require.NoError(t, os.WriteFile(cleanFile, []byte(cleanContent), 0o600))

	// Call wrapper with file list.
	files := []string{cleanFile}
	err := goFixCopyLoopVar(logger, files)
	require.NoError(t, err) // Should return nil when no issues found.
}

// TestGoFixCopyLoopVar_ErrorPropagation tests wrapper error handling.
func TestGoFixCopyLoopVar_ErrorPropagation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-error")

	// Create invalid Go file that will cause parsing errors.
	invalidFile := filepath.Join(tmpDir, "invalid.go")
	invalidContent := `package test

func broken( {
	// Syntax error
}
`
	require.NoError(t, os.WriteFile(invalidFile, []byte(invalidContent), 0o600))

	// Call wrapper with file list.
	files := []string{invalidFile}
	err := goFixCopyLoopVar(logger, files)
	require.Error(t, err) // Should propagate parsing error.
	require.Contains(t, err.Error(), "copyloopvar fix failed")
}

// TestGoFixTHelper_Integration tests the wrapper calling fix/thelper package.
func TestGoFixTHelper_Integration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-wrapper")

	// Create test file with thelper issue.
	testFile := filepath.Join(tmpDir, "helpers_test.go")
	testContent := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0o600))

	// Call wrapper with file list.
	files := []string{testFile}
	err := goFixTHelper(logger, files)
	require.Error(t, err) // Should return error indicating fixes made.
	require.Contains(t, err.Error(), "added t.Helper() to 1 test helper functions")

	// Verify fix applied.
	fixed, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(fixed), "t.Helper()")
}

// TestGoFixTHelper_NoFiles tests wrapper with empty file list.
func TestGoFixTHelper_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-thelper-no-files")

	// Call wrapper with empty file list.
	err := goFixTHelper(logger, []string{})
	require.NoError(t, err)
}

// TestGoFixTHelper_NoIssues tests wrapper when no issues are found.
func TestGoFixTHelper_NoIssues(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-no-issues")

	// Create test file WITHOUT thelper issue (not a helper function).
	testFile := filepath.Join(tmpDir, "test_test.go")
	testContent := `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Log("test")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0o600))

	// Call wrapper with file list.
	files := []string{testFile}
	err := goFixTHelper(logger, files)
	require.NoError(t, err) // Should return nil when no issues found.
}

// TestGoFixTHelper_ErrorPropagation tests wrapper error handling.
func TestGoFixTHelper_ErrorPropagation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-thelper-error")

	// Create invalid Go file that will cause parsing errors.
	invalidFile := filepath.Join(tmpDir, "invalid_test.go")
	invalidContent := `package test

import "testing"

func broken( t *testing.T {
	// Syntax error
}
`
	require.NoError(t, os.WriteFile(invalidFile, []byte(invalidContent), 0o600))

	// Call wrapper with file list.
	files := []string{invalidFile}
	err := goFixTHelper(logger, files)
	require.Error(t, err) // Should propagate parsing error.
	require.Contains(t, err.Error(), "thelper fix failed")
}

// TestGoFixAll_Integration tests the orchestrator calling all fix commands.
func TestGoFixAll_Integration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-fix-all-wrapper")

	// Create files with all three types of issues.
	errorFile := filepath.Join(tmpDir, "errors.go")
	errorContent := `package test

import "errors"

var ErrFailed = errors.New("Failed to process")
`
	require.NoError(t, os.WriteFile(errorFile, []byte(errorContent), 0o600))

	loopFile := filepath.Join(tmpDir, "loop.go")
	loopContent := `package test

func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	require.NoError(t, os.WriteFile(loopFile, []byte(loopContent), 0o600))

	testFile := filepath.Join(tmpDir, "helpers_test.go")
	testContent := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0o600))

	// Call wrapper.
	err := goFixAll(logger, tmpDir)
	require.NoError(t, err)

	// Verify all fixes applied.
	errorFixed, err := os.ReadFile(errorFile)
	require.NoError(t, err)
	require.Contains(t, string(errorFixed), `errors.New("failed to process")`)

	loopFixed, err := os.ReadFile(loopFile)
	require.NoError(t, err)
	require.NotContains(t, string(loopFixed), "item := item")

	testFixed, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(testFixed), "t.Helper()")
}
