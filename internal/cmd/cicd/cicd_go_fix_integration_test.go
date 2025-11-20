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
