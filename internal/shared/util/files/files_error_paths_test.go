// Copyright (c) 2025 Justin Cranford

package files_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"


	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)


// TestWriteFile_OSError covers the os.WriteFile error path.
func TestWriteFile_OSError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("chmod not supported on Windows.")
	}

	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0o500))

	t.Cleanup(func() { _ = os.Chmod(readOnlyDir, 0o700) })

	err := cryptoutilSharedUtilFiles.WriteFile(filepath.Join(readOnlyDir, "file.txt"), "data", cryptoutilSharedMagic.CacheFilePermissions)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write file")
}

// TestListAllFilesWithOptions_DotfileNoExtension covers the dotfile branch where baseName starts with ".".
func TestListAllFilesWithOptions_DotfileNoExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dotFile := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(dotFile, []byte("*.out"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tmpDir, []string{"gitignore"}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestReadFilesBytesLimit_InnerReadError covers the inner ReadFileBytesLimit error path.
func TestReadFilesBytesLimit_InnerReadError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("chmod not supported on Windows.")
	}

	tmpDir := t.TempDir()
	unreadableFile := filepath.Join(tmpDir, "unreadable.txt")
	require.NoError(t, os.WriteFile(unreadableFile, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(unreadableFile, 0o000))

	t.Cleanup(func() { _ = os.Chmod(unreadableFile, cryptoutilSharedMagic.CacheFilePermissions) })

	_, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit([]string{unreadableFile}, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.DefaultLogsBatchSize)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read file")
}

// TestReadFileBytesLimit_UnreadableFile covers the os.Open error path.
func TestReadFileBytesLimit_UnreadableFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("chmod not supported on Windows.")
	}

	tmpDir := t.TempDir()
	unreadableFile := filepath.Join(tmpDir, "unreadable.txt")
	require.NoError(t, os.WriteFile(unreadableFile, []byte("hello world"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(unreadableFile, 0o000))

	t.Cleanup(func() { _ = os.Chmod(unreadableFile, cryptoutilSharedMagic.CacheFilePermissions) })

	_, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(unreadableFile, cryptoutilSharedMagic.DefaultLogsBatchSize)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open file")
}
