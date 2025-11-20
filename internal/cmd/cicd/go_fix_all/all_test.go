package go_fix_all

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
	logger := cryptoutilCmd.NewLogger("test-all-empty")

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_AllFixTypes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-all-combined")

	// Create file with staticcheck error string issues.
	staticcheckFile := filepath.Join(tmpDir, "errors.go")
	staticcheckContent := `package test

import "errors"

var ErrFailed = errors.New("Failed to process")
`
	require.NoError(t, os.WriteFile(staticcheckFile, []byte(staticcheckContent), 0o600))

	// Create file with copyloopvar issues.
	copyloopvarFile := filepath.Join(tmpDir, "loop.go")
	copyloopvarContent := `package test

func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	require.NoError(t, os.WriteFile(copyloopvarFile, []byte(copyloopvarContent), 0o600))

	// Create test file with thelper issues.
	thelperFile := filepath.Join(tmpDir, "helpers_test.go")
	thelperContent := `package test

import "testing"

func setupTest(t *testing.T) {
	t.Log("setup")
}
`
	require.NoError(t, os.WriteFile(thelperFile, []byte(thelperContent), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 5, processed) // staticcheck: 2 (errors.go + helpers_test.go), copyloopvar: 2 (errors.go + loop.go), thelper: 1 (helpers_test.go).
	require.Equal(t, 3, modified)  // All 3 files should be modified.
	require.Equal(t, 3, issuesFixed)

	// Verify staticcheck fix.
	staticcheckFixed, err := os.ReadFile(staticcheckFile)
	require.NoError(t, err)
	require.Contains(t, string(staticcheckFixed), `errors.New("failed to process")`)

	// Verify copyloopvar fix.
	copyloopvarFixed, err := os.ReadFile(copyloopvarFile)
	require.NoError(t, err)
	require.NotContains(t, string(copyloopvarFixed), "item := item")

	// Verify thelper fix.
	thelperFixed, err := os.ReadFile(thelperFile)
	require.NoError(t, err)
	require.Contains(t, string(thelperFixed), "t.Helper()")
}

func TestFix_InvalidDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-all-invalid")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path", "1.25.4")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}
