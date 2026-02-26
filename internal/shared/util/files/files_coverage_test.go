// Copyright (c) 2025 Justin Cranford

package files_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

// TestListAllFilesWithOptions_ErrorPath tests error path in filepath.Walk callback.
func TestListAllFilesWithOptions_ErrorPath(t *testing.T) {
	t.Parallel()

	// Create temp directory
	tempDir := t.TempDir()

	// Create a file to trigger walk
	testFile := filepath.Join(tempDir, "test.txt")
	err := cryptoutilSharedUtilFiles.WriteFile(testFile, "content", cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err)

	// Create file in subdirectory
	subFile := filepath.Join(subDir, "subfile.txt")
	err = cryptoutilSharedUtilFiles.WriteFile(subFile, "subcontent", cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Test with directory exclusion using prefix match
	inclusions := []string{"txt"}
	exclusions := []string{filepath.ToSlash(subDir)}

	result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tempDir, inclusions, exclusions)
	require.NoError(t, err)

	// Should only have root file, not subdirectory file
	require.Len(t, result["txt"], 1)
	require.Contains(t, result["txt"][0], "test.txt")
	require.NotContains(t, result["txt"][0], "subfile.txt")
}

// TestListAllFilesWithOptions_NoExtension tests handling of files without extension.
func TestListAllFilesWithOptions_NoExtension(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create file without extension
	noExtFile := filepath.Join(tempDir, "Makefile")
	err := cryptoutilSharedUtilFiles.WriteFile(noExtFile, "content", cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	// Try to find it without matching extension
	inclusions := []string{"txt"}
	exclusions := []string{}

	result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tempDir, inclusions, exclusions)
	require.NoError(t, err)

	// Should not find Makefile since "txt" not matched
	require.Empty(t, result["txt"])

	// Now try with Makefile in inclusions (no dot prefix)
	inclusions = []string{"Makefile"}
	result, err = cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tempDir, inclusions, exclusions)
	require.NoError(t, err)

	// Still won't match because file has no extension and isn't a dotfile
	// This tests the empty extension branch with no dot prefix
	require.Empty(t, result)
}

// TestReadFileBytesLimit_ErrorPaths tests error paths in ReadFileBytesLimit.
func TestReadFileBytesLimit_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("Non-existent file", func(t *testing.T) {
		t.Parallel()

		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit("/nonexistent/file.txt", cryptoutilSharedMagic.JoseJAMaxMaterials)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to open file")
		require.Nil(t, content)
	})

	t.Run("Zero limit calls ReadFileBytes", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := []byte("test content")
		err := os.WriteFile(testFile, testContent, cryptoutilSharedMagic.CacheFilePermissions)
		require.NoError(t, err)

		// Zero limit should read entire file via ReadFileBytes
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, 0)
		require.NoError(t, err)
		require.Equal(t, testContent, content)
	})

	t.Run("Negative limit calls ReadFileBytes", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := []byte("test content")
		err := os.WriteFile(testFile, testContent, cryptoutilSharedMagic.CacheFilePermissions)
		require.NoError(t, err)

		// Negative limit should read entire file via ReadFileBytes
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, -1)
		require.NoError(t, err)
		require.Equal(t, testContent, content)
	})

	t.Run("File smaller than limit", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := []byte("small")
		err := os.WriteFile(testFile, testContent, cryptoutilSharedMagic.CacheFilePermissions)
		require.NoError(t, err)

		// Limit larger than file size - should read entire file
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, cryptoutilSharedMagic.JoseJADefaultListLimit)
		require.NoError(t, err)
		require.Equal(t, testContent, content)
	})
}
