// Copyright (c) 2025-2026 Justin Cranford.
package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

func TestWriteFile_ReportsFilesystemError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "file.txt")
	err := cryptoutilSharedUtilFiles.WriteFile(path, "content", 0o600)
	require.ErrorContains(t, err, "failed to write file")
}

func TestListAllFilesWithOptions_IncludesDotfilesByName(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dotfile := filepath.Join(root, ".gitignore")
	require.NoError(t, os.WriteFile(dotfile, []byte("ignored"), 0o600))

	files, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(root, []string{"gitignore"}, nil)
	require.NoError(t, err)
	require.Equal(t, []string{filepath.ToSlash(dotfile)}, files["gitignore"])
}
