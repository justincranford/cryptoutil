// Copyright (c) 2025 Justin Cranford

package configs_deployments_consistency_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessConfigsDeploymentsConsistency "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/configs_deployments_consistency"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("configs-deployments-consistency-test")
}

func setupDeployments(t *testing.T, tmpDir string, psIDs ...string) {
	t.Helper()

	deploymentsDir := filepath.Join(tmpDir, "deployments")
	require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	for _, psID := range psIDs {
		require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, psID), cryptoutilSharedMagic.CICDTempDirPermissions))
	}
}

func setupConfigs(t *testing.T, tmpDir string, psIDs ...string) {
	t.Helper()

	for _, psID := range psIDs {
		configsDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDConfigsDir, psID)
		require.NoError(t, os.MkdirAll(configsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	}
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

// TestFindViolationsInDir_NoDeploymentsDir verifies error when deployments/ is missing.
func TestFindViolationsInDir_NoDeploymentsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.Error(t, err)
	assert.Nil(t, violations)
	assert.Contains(t, err.Error(), "deployments/ directory not found")
}

// TestFindViolationsInDir_EmptyDeployments verifies no violations when deployments/ is empty.
func TestFindViolationsInDir_EmptyDeployments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir)

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_UnknownPSID verifies unknown PS-ID dirs are skipped.
func TestFindViolationsInDir_UnknownPSID(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir, "unknown-service")

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_MatchingConfigExists verifies no violation when configs dir exists.
func TestFindViolationsInDir_MatchingConfigExists(t *testing.T) {
	t.Parallel()

	ps := lintFitnessRegistry.AllProductServices()[0]

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir, ps.PSID)
	setupConfigs(t, tmpDir, ps.PSID)

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_MissingConfig verifies violation when configs dir is missing.
func TestFindViolationsInDir_MissingConfig(t *testing.T) {
	t.Parallel()

	ps := lintFitnessRegistry.AllProductServices()[0]

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir, ps.PSID)

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0], ps.PSID)
	assert.Contains(t, violations[0], "missing")
}

// TestFindViolationsInDir_AllRegisteredPSIDs verifies all registered PS-IDs pass when configs exist.
func TestFindViolationsInDir_AllRegisteredPSIDs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var psIDs []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		psIDs = append(psIDs, ps.PSID)
	}

	setupDeployments(t, tmpDir, psIDs...)
	setupConfigs(t, tmpDir, psIDs...)

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestFindViolationsInDir_MultipleViolations verifies multiple missing configs are all reported.
func TestFindViolationsInDir_MultipleViolations(t *testing.T) {
	t.Parallel()

	allPS := lintFitnessRegistry.AllProductServices()
	require.GreaterOrEqual(t, len(allPS), 2, "need at least 2 PS for this test")

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir, allPS[0].PSID, allPS[1].PSID)

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, violations, 2)
}

// TestFindViolationsInDir_FileInDeployments verifies files in deployments/ are ignored.
func TestFindViolationsInDir_FileInDeployments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "deployments", "README.md"), []byte("# Readme"), cryptoutilSharedMagic.CICDOutputFilePermissions))

	violations, err := lintFitnessConfigsDeploymentsConsistency.FindViolationsInDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

// TestCheckInDir_ValidStructure verifies CheckInDir succeeds with valid structure.
func TestCheckInDir_ValidStructure(t *testing.T) {
	t.Parallel()

	ps := lintFitnessRegistry.AllProductServices()[0]

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir, ps.PSID)
	setupConfigs(t, tmpDir, ps.PSID)

	err := lintFitnessConfigsDeploymentsConsistency.CheckInDir(newTestLogger(), tmpDir)
	require.NoError(t, err)
}

// TestCheckInDir_MissingConfig_ReturnsError verifies CheckInDir returns error on violation.
func TestCheckInDir_MissingConfig_ReturnsError(t *testing.T) {
	t.Parallel()

	ps := lintFitnessRegistry.AllProductServices()[0]

	tmpDir := t.TempDir()
	setupDeployments(t, tmpDir, ps.PSID)

	err := lintFitnessConfigsDeploymentsConsistency.CheckInDir(newTestLogger(), tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configs/deployments inconsistencies")
}

// TestCheckInDir_NoDeploymentsDir_ReturnsError verifies CheckInDir error on missing deployments/.
func TestCheckInDir_NoDeploymentsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	err := lintFitnessConfigsDeploymentsConsistency.CheckInDir(newTestLogger(), "/nonexistent/path/does/not/exist")
	require.Error(t, err)
}

// TestCheck_Integration runs the linter against the real workspace.
func TestCheck_Integration(t *testing.T) {
	root := findProjectRoot(t)

	err := lintFitnessConfigsDeploymentsConsistency.CheckInDir(newTestLogger(), root)
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

	err = lintFitnessConfigsDeploymentsConsistency.Check(newTestLogger())
	require.NoError(t, err)
}

// TestFormatViolations verifies FormatViolations joins violations.
func TestFormatViolations(t *testing.T) {
	t.Parallel()

	result := lintFitnessConfigsDeploymentsConsistency.FormatViolations([]string{"a", "b", "c"})
	assert.Equal(t, "a\nb\nc", result)
}
