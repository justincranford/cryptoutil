// Copyright (c) 2025-2026 Justin Cranford.
package files_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestListAllFilesWithOptions_DotfileNoExtension covers the dotfile branch where baseName starts with ".".
func TestListAllFilesWithOptions_DotfileNoExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dotFile := filepath.Join(tmpDir, testGitignore)
	require.NoError(t, os.WriteFile(dotFile, []byte("*.out"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tmpDir, []string{testGitignoreValueA1AA6596}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestReadFilesBytesLimit_InnerReadError covers the inner ReadFileBytesLimit error path.
func TestReadFilesBytesLimit_InnerReadError(t *testing.T) {
	t.Parallel()

	missingFile := filepath.Join(t.TempDir(), "missing", "file.txt")

	_, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit([]string{missingFile}, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.DefaultLogsBatchSize)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read file")
}

// TestReadFileBytesLimit_UnreadableFile covers the os.Open error path.
func TestReadFileBytesLimit_UnreadableFile(t *testing.T) {
	t.Parallel()

	missingFile := filepath.Join(t.TempDir(), "missing", "file.txt")

	_, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(missingFile, cryptoutilSharedMagic.DefaultLogsBatchSize)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open file")
}
