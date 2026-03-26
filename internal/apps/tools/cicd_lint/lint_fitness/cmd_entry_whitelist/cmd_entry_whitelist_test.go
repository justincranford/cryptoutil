// Copyright (c) 2025 Justin Cranford

package cmd_entry_whitelist_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessCmdEntryWhitelist "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/cmd_entry_whitelist"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func mkdir(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(path, cryptoutilSharedMagic.DirPermissions))
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("skipping integration test: cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

func TestFindViolationsInDir_EmptyCmdDir_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd"))

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_NonExistentCmdDir_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read cmd/ directory")
	assert.Nil(t, violations)
}

func TestFindViolationsInDir_PSIDDirs_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		mkdir(t, filepath.Join(tmpDir, "cmd", ps.PSID))
	}

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_ProductDirs_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, p := range lintFitnessRegistry.AllProducts() {
		mkdir(t, filepath.Join(tmpDir, "cmd", p.ID))
	}

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_SuiteDir_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	for _, s := range lintFitnessRegistry.AllSuites() {
		mkdir(t, filepath.Join(tmpDir, "cmd", s.ID))
	}

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_InfraToolDirs_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", cryptoutilSharedMagic.CICDCmdDirCicdLint))
	mkdir(t, filepath.Join(tmpDir, "cmd", cryptoutilSharedMagic.CICDCmdDirWorkflow))

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_UnknownDir_Violation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", "identity-compose"))

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Equal(t, "identity-compose", violations[0])
}

func TestFindViolationsInDir_MultipleUnknownDirs_AllViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", "identity-compose"))
	mkdir(t, filepath.Join(tmpDir, "cmd", "sm-run"))
	mkdir(t, filepath.Join(tmpDir, "cmd", "jose-legacy"))

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Len(t, violations, 3)
}

func TestFindViolationsInDir_FilesInCmdRoot_Ignored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd"))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "cmd", "README.md"),
		[]byte("readme"),
		cryptoutilSharedMagic.FilePermissions,
	))

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_MixedAllowedAndUnknown(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", cryptoutilSharedMagic.OTLPServiceSMKMS))
	mkdir(t, filepath.Join(tmpDir, "cmd", "sm-run"))

	violations, err := lintFitnessCmdEntryWhitelist.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Equal(t, "sm-run", violations[0])
}

func TestCheckInDir_ValidStructure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", cryptoutilSharedMagic.OTLPServiceSMKMS))

	err := lintFitnessCmdEntryWhitelist.CheckInDir(newTestLogger(), tmpDir)

	require.NoError(t, err)
}

func TestCheckInDir_InvalidStructure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", "identity-compose"))

	err := lintFitnessCmdEntryWhitelist.CheckInDir(newTestLogger(), tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cmd-entry-whitelist: found")
}

func TestCheckInDir_NonExistentRoot_ReturnsError(t *testing.T) {
	t.Parallel()

	err := lintFitnessCmdEntryWhitelist.CheckInDir(newTestLogger(), "/nonexistent/dir/that/does/not/exist")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check cmd/ entry whitelist")
}

func TestAllowedCount_ReturnsExpectedCount(t *testing.T) {
	t.Parallel()

	// 1 suite + 5 products + 10 PS-IDs + 2 infra tools = 18.
	count := lintFitnessCmdEntryWhitelist.AllowedCount()

	assert.Equal(t, 18, count)
}

func TestAllowedEntries_ReturnsSortedList(t *testing.T) {
	t.Parallel()

	entries := lintFitnessCmdEntryWhitelist.AllowedEntries()

	require.NotEmpty(t, entries)

	for i := 1; i < len(entries); i++ {
		assert.LessOrEqual(t, entries[i-1], entries[i], "AllowedEntries() must be sorted")
	}
}

func TestAllowedEntrySet_ContainsKnownEntries(t *testing.T) {
	t.Parallel()

	set := lintFitnessCmdEntryWhitelist.AllowedEntrySet()

	assert.True(t, set[cryptoutilSharedMagic.CICDCmdDirCicdLint])
	assert.True(t, set[cryptoutilSharedMagic.CICDCmdDirWorkflow])
	assert.True(t, set[cryptoutilSharedMagic.OTLPServiceSMKMS])
	assert.False(t, set["identity-compose"])
}

// TestCheck_Integration runs the linter against the real workspace.
func TestCheck_Integration(t *testing.T) {
	root := findProjectRoot(t)

	err := lintFitnessCmdEntryWhitelist.CheckInDir(newTestLogger(), root)

	require.NoError(t, err)
}

// TestCheck_FromWorkspaceRoot verifies Check() (no rootDir) works from project root.
// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_FromWorkspaceRoot(t *testing.T) {
	root := findProjectRoot(t)

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	err = lintFitnessCmdEntryWhitelist.Check(newTestLogger())
	require.NoError(t, err)
}
