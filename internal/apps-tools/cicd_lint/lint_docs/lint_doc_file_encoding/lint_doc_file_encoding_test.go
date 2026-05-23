// Copyright (c) 2025-2026 Justin Cranford.
package lint_doc_file_encoding

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     []byte
		expectIssue bool
	}{
		{name: "valid utf8 lf", content: []byte("line1\nline2\n"), expectIssue: false},
		{name: "utf8 bom", content: []byte{0xEF, 0xBB, 0xBF, 'a', '\n'}, expectIssue: true},
		{name: "utf16 bom", content: []byte{0xFF, 0xFE, 'a', 0x00}, expectIssue: true},
		{name: "crlf", content: []byte("line1\r\nline2\n"), expectIssue: true},
		{name: "nul byte", content: []byte{'a', 0x00, 'b', '\n'}, expectIssue: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "file.md")
			require.NoError(t, os.WriteFile(filePath, tc.content, cryptoutilSharedMagic.KeyFilePermissions))

			issues, err := checkFile(filePath)
			require.NoError(t, err)

			if tc.expectIssue {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

func TestCheckInDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirDocs), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "instructions"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".claude"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".claude", "agents"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "skill-a"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".claude", "skills", "skill-a"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	valid := []byte("ok\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirDocs, "ENG-HANDBOOK.md"), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDCopilotInstructionsFile), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, claudeFile), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, claudeLocal), []byte("{}\n"), cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, skillsReadme), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "instructions", "a.instructions.md"), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents", "a.agent.md"), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".claude", "agents", "a.md"), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "skill-a", cryptoutilSharedMagic.CICDSkillFileName), valid, cryptoutilSharedMagic.KeyFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".claude", "skills", "skill-a", cryptoutilSharedMagic.CICDSkillFileName), valid, cryptoutilSharedMagic.KeyFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	require.NoError(t, CheckInDir(logger, tmpDir))

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".claude", "skills", "skill-a", cryptoutilSharedMagic.CICDSkillFileName), []byte("bad\r\n"), cryptoutilSharedMagic.KeyFilePermissions))
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CRLF")
}
