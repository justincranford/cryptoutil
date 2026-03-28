// Copyright (c) 2025 Justin Cranford

package infra_tool_naming_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessInfraToolNaming "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/infra_tool_naming"
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

func TestFindViolationsInDir_EmptyCmdDir_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd"))

	violations, err := lintFitnessInfraToolNaming.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_NonExistentCmdDir_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	violations, err := lintFitnessInfraToolNaming.FindViolationsInDir(tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read cmd/ directory")
	assert.Nil(t, violations)
}

func TestFindViolationsInDir_ValidCICDPrefix_NoViolations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", "cicd-newtool"))
	mkdir(t, filepath.Join(tmpDir, cryptoutilSharedMagic.CICDInfraToolInternalDir, "cicd_newtool"))

	violations, err := lintFitnessInfraToolNaming.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestFindViolationsInDir_MissingCICDPrefix_Violation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", "release-tool"))

	violations, err := lintFitnessInfraToolNaming.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0], "release-tool")
	assert.Contains(t, violations[0], "MUST be prefixed")
}

func TestFindViolationsInDir_MissingCounterpart_Violation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", "cicd-orphan"))

	violations, err := lintFitnessInfraToolNaming.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0], "cicd-orphan")
	assert.Contains(t, violations[0], "missing counterpart")
}

func TestFindViolationsInDir_RegistryEntries_Ignored(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Registry entries should be skipped (not flagged as infra tools).
	mkdir(t, filepath.Join(tmpDir, "cmd", cryptoutilSharedMagic.DefaultOTLPServiceDefault))
	mkdir(t, filepath.Join(tmpDir, "cmd", cryptoutilSharedMagic.JoseProductName))
	mkdir(t, filepath.Join(tmpDir, "cmd", cryptoutilSharedMagic.OTLPServiceJoseJA))

	violations, err := lintFitnessInfraToolNaming.FindViolationsInDir(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestCheckInDir_Violations_ReturnsError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd", "bad-tool"))

	err := lintFitnessInfraToolNaming.CheckInDir(newTestLogger(), tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "infra-tool-naming")
}

func TestCheckInDir_Clean_NoError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mkdir(t, filepath.Join(tmpDir, "cmd"))

	err := lintFitnessInfraToolNaming.CheckInDir(newTestLogger(), tmpDir)

	require.NoError(t, err)
}

func TestFindViolationsInDir_Integration(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)

	violations, err := lintFitnessInfraToolNaming.FindViolationsInDir(root)

	require.NoError(t, err)
	assert.Empty(t, violations, "project cmd/ should follow infra-tool naming conventions: %v", violations)
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
