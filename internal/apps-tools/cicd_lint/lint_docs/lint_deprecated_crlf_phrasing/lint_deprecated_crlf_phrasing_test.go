// Copyright (c) 2025-2026 Justin Cranford.
package lint_deprecated_crlf_phrasing

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckWithFS_NoViolations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDGithubInstructionsDir)), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDGithubAgentsDir)), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDClaudeAgentsDir)), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	safeContent := []byte("line-ending policy is LF\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDGithubInstructionsDir), "sample.instructions.md"), safeContent, cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDGithubAgentsDir), "sample.agent.md"), safeContent, cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDClaudeAgentsDir), "sample.md"), safeContent, cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := checkWithFS(logger, func() (string, error) { return root, nil }, filepath.WalkDir, os.Open)
	require.NoError(t, err)
}

func TestCheckWithFS_FindsViolations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDGithubInstructionsDir)), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDGithubAgentsDir)), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDClaudeAgentsDir)), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	bad := []byte("fix(tooling): renormalize CRLF files to LF\n")
	filePath := filepath.Join(root, filepath.FromSlash(cryptoutilSharedMagic.CICDGithubInstructionsDir), "sample.instructions.md")
	require.NoError(t, os.WriteFile(filePath, bad, cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := checkWithFS(logger, func() (string, error) { return root, nil }, filepath.WalkDir, os.Open)
	require.Error(t, err)
	require.Contains(t, err.Error(), "violation")
}
