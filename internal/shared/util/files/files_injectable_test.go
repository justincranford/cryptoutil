// Copyright (c) 2025-2026 Justin Cranford.
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestReadFileBytesLimit_StatError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

	_, err := readFileBytesLimit(testFile, cryptoutilSharedMagic.DefaultLogsBatchSize, os.Open,
		func(_ *os.File) (os.FileInfo, error) { return nil, fmt.Errorf("injected stat error") },
		(*os.File).Read, (*os.File).Close)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get file stats")
}

func TestReadFileBytesLimit_ReadError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

	_, err := readFileBytesLimit(testFile, cryptoutilSharedMagic.DefaultLogsBatchSize, os.Open,
		(*os.File).Stat,
		func(_ *os.File, _ []byte) (int, error) { return 0, fmt.Errorf("injected read error") },
		(*os.File).Close)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read bytes from file")
}

func TestReadFileBytesLimit_CloseError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

	// Close error only prints a warning, no error returned.
	data, err := readFileBytesLimit(testFile, cryptoutilSharedMagic.DefaultLogsBatchSize, os.Open,
		(*os.File).Stat, (*os.File).Read,
		func(f *os.File) error {
			_ = f.Close() // Close the OS handle to prevent Windows file-lock on TempDir cleanup.

			return fmt.Errorf("injected close error")
		})
	require.NoError(t, err)
	require.Equal(t, []byte("content"), data)
}
