// Copyright (c) 2025 Justin Cranford

package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFileBytesLimit_StatError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := filesStatFn
	defer func() { filesStatFn = originalFn }()

	filesStatFn = func(_ *os.File) (os.FileInfo, error) {
		return nil, fmt.Errorf("injected stat error")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0o600))

	_, err := ReadFileBytesLimit(testFile, 1024)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get file stats")
}

func TestReadFileBytesLimit_ReadError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := filesReadFn
	defer func() { filesReadFn = originalFn }()

	filesReadFn = func(_ *os.File, _ []byte) (int, error) {
		return 0, fmt.Errorf("injected read error")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0o600))

	_, err := ReadFileBytesLimit(testFile, 1024)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read bytes from file")
}

func TestReadFileBytesLimit_CloseError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := filesCloseFn
	defer func() { filesCloseFn = originalFn }()

	filesCloseFn = func(_ *os.File) error {
		return fmt.Errorf("injected close error")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0o600))

	// Close error only prints a warning, no error returned.
	data, err := ReadFileBytesLimit(testFile, 1024)
	require.NoError(t, err)
	require.Equal(t, []byte("content"), data)
}
