// Copyright (c) 2025-2026 Justin Cranford.

package testmain_orchestration_policy

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestFileContainsLine_ReadError(t *testing.T) {
	t.Parallel()

	contains, err := fileContainsLine("ignored", requiredImportSubstring, func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	})
	require.False(t, contains)
	require.Error(t, err)
	require.ErrorContains(t, err, "injected read failure")
}

func TestCheckTestMainFile_ReadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, testmainFileName)
	require.NoError(t, os.WriteFile(filePath, []byte("package server\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := checkTestMainFile(filePath, cryptoutilSharedMagic.OTLPServiceSMKMS, "server", func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	})
	require.Nil(t, violations)
	require.Error(t, err)
	require.ErrorContains(t, err, "injected read failure")
}

func TestCheckTestMainFile_Table(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	missingPath := filepath.Join(tempDir, "missing", testmainFileName)
	dirPath := filepath.Join(tempDir, "dir", testmainFileName)
	require.NoError(t, os.MkdirAll(dirPath, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	badFilePath := filepath.Join(tempDir, "bad", testmainFileName)
	require.NoError(t, os.MkdirAll(filepath.Dir(badFilePath), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(badFilePath, []byte("package server\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	goodFilePath := filepath.Join(tempDir, "good", testmainFileName)
	require.NoError(t, os.MkdirAll(filepath.Dir(goodFilePath), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(goodFilePath, []byte("package server\nimport \"cryptoutil/internal/apps-framework/service/test_orch_integration\"\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	tests := []struct {
		name        string
		filePath    string
		wantCount   int
		wantErrText string
	}{
		{name: "missing file violation", filePath: missingPath, wantCount: 1},
		{name: "directory violation", filePath: dirPath, wantCount: 1},
		{name: "missing import violation", filePath: badFilePath, wantCount: 1},
		{name: "compliant file", filePath: goodFilePath, wantCount: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations, err := checkTestMainFile(tc.filePath, cryptoutilSharedMagic.OTLPServiceSMKMS, "server", os.ReadFile)
			if tc.wantErrText != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErrText)

				return
			}

			require.NoError(t, err)
			require.Len(t, violations, tc.wantCount)
		})
	}
}

func TestFindViolationsWithReader_ReadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serverDir := filepath.Join(tempDir, "internal", "apps", ps.PSID, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, testmainFileName), []byte("package server\n"), cryptoutilSharedMagic.FilePermissionsDefault))
	}

	violations, err := findViolationsWithReader(tempDir, func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	})
	require.Nil(t, violations)
	require.Error(t, err)
	require.ErrorContains(t, err, "injected read failure")
}

func TestFindViolationsWithReader_ClientBranchReadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		serverDir := filepath.Join(tempDir, "internal", "apps", ps.PSID, "server")
		clientDir := filepath.Join(tempDir, "internal", "apps", ps.PSID, "client")

		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(clientDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, testmainFileName), []byte("package server\nimport \"service/test_orch_integration\"\n"), cryptoutilSharedMagic.FilePermissionsDefault))
		require.NoError(t, os.WriteFile(filepath.Join(clientDir, testmainFileName), []byte("package client\n"), cryptoutilSharedMagic.FilePermissionsDefault))
	}

	violations, err := findViolationsWithReader(tempDir, func(path string) ([]byte, error) {
		if strings.Contains(path, string(filepath.Separator)+"client"+string(filepath.Separator)) {
			return nil, errors.New("injected client read failure")
		}

		return []byte("package server\nimport \"service/test_orch_integration\"\n"), nil
	})
	require.Nil(t, violations)
	require.Error(t, err)
	require.ErrorContains(t, err, "injected client read failure")
}

func TestCheckTestMainFile_StatError(t *testing.T) {
	t.Parallel()

	violations, err := checkTestMainFile("bad\x00path", cryptoutilSharedMagic.OTLPServiceSMKMS, "server", os.ReadFile)
	require.Nil(t, violations)
	require.Error(t, err)
	require.ErrorContains(t, err, "stat")
}
