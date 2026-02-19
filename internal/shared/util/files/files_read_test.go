// Copyright (c) 2025 Justin Cranford

package files_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

func TestReadFileBytes(t *testing.T) {
	t.Parallel()

	t.Run("read existing file", func(t *testing.T) {
		t.Parallel()

		// Create temp file.
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := []byte("test content")
		err := os.WriteFile(testFile, testContent, 0o600)
		require.NoError(t, err, "Failed to create test file")

		// Read file.
		content, err := cryptoutilSharedUtilFiles.ReadFileBytes(testFile)
		require.NoError(t, err, "Failed to read file")
		require.Equal(t, testContent, content, "Content should match")
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()

		content, err := cryptoutilSharedUtilFiles.ReadFileBytes("/nonexistent/path/file.txt")
		require.Error(t, err, "Should return error for missing file")
		require.Nil(t, content, "Content should be nil on error")
		require.Contains(t, err.Error(), "failed to read file", "Error should indicate read failure")
	})
}

// TestReadFilesBytes tests reading multiple files.
func TestReadFilesBytes(t *testing.T) {
	t.Parallel()

	t.Run("read multiple files", func(t *testing.T) {
		t.Parallel()

		// Create temp files.
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		content1 := []byte("content 1")
		content2 := []byte("content 2")

		require.NoError(t, os.WriteFile(file1, content1, 0o600))
		require.NoError(t, os.WriteFile(file2, content2, 0o600))

		// Read files.
		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes([]string{file1, file2})
		require.NoError(t, err, "Failed to read files")
		require.Len(t, contents, 2, "Should have 2 file contents")
		require.Equal(t, content1, contents[0], "First file content should match")
		require.Equal(t, content2, contents[1], "Second file content should match")
	})

	t.Run("no files specified", func(t *testing.T) {
		t.Parallel()

		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes([]string{})
		require.Error(t, err, "Should return error for empty file list")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "no files specified", "Error should indicate no files")
	})

	t.Run("nil file list", func(t *testing.T) {
		t.Parallel()

		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes(nil)
		require.Error(t, err, "Should return error for nil file list")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "no files specified", "Error should indicate no files")
	})

	t.Run("empty file path in list", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		require.NoError(t, os.WriteFile(file1, []byte("content"), 0o600))

		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes([]string{file1, ""})
		require.Error(t, err, "Should return error for empty path")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "empty file path", "Error should indicate empty path")
	})

	t.Run("whitespace file path", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		require.NoError(t, os.WriteFile(file1, []byte("content"), 0o600))

		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes([]string{file1, "   "})
		require.Error(t, err, "Should return error for whitespace path")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "empty file path", "Error should indicate empty path")
	})

	t.Run("file not found in list", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		require.NoError(t, os.WriteFile(file1, []byte("content"), 0o600))

		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes([]string{file1, "/nonexistent/file.txt"})
		require.Error(t, err, "Should return error for missing file")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "failed to read file", "Error should indicate read failure")
	})
}

// TestReadFileBytesLimit tests reading files with size limits.
func TestReadFileBytesLimit(t *testing.T) {
	t.Parallel()

	t.Run("read within limit", func(t *testing.T) {
		t.Parallel()

		// Create temp file with known content.
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := []byte("1234567890")
		require.NoError(t, os.WriteFile(testFile, testContent, 0o600))

		// Read with limit larger than file size.
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, 100)
		require.NoError(t, err, "Failed to read file")
		require.Equal(t, testContent, content, "Should read entire file")
	})

	t.Run("read with exact limit", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := []byte("1234567890")
		require.NoError(t, os.WriteFile(testFile, testContent, 0o600))

		// Read with limit equal to file size.
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, int64(len(testContent)))
		require.NoError(t, err, "Failed to read file")
		require.Equal(t, testContent, content, "Should read entire file")
	})

	t.Run("read partial content", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := []byte("1234567890")
		require.NoError(t, os.WriteFile(testFile, testContent, 0o600))

		// Read first 5 bytes.
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, 5)
		require.NoError(t, err, "Failed to read file")
		require.Equal(t, []byte("12345"), content, "Should read first 5 bytes")
		require.Len(t, content, 5, "Should read exactly 5 bytes")
	})

	t.Run("no limit (zero)", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := []byte("1234567890")
		require.NoError(t, os.WriteFile(testFile, testContent, 0o600))

		// Read with limit 0 (should read entire file).
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, 0)
		require.NoError(t, err, "Failed to read file")
		require.Equal(t, testContent, content, "Should read entire file with limit 0")
	})

	t.Run("no limit (negative)", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		testContent := []byte("1234567890")
		require.NoError(t, os.WriteFile(testFile, testContent, 0o600))

		// Read with negative limit (should read entire file).
		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, -1)
		require.NoError(t, err, "Failed to read file")
		require.Equal(t, testContent, content, "Should read entire file with negative limit")
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()

		content, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit("/nonexistent/file.txt", 100)
		require.Error(t, err, "Should return error for missing file")
		require.Nil(t, content, "Content should be nil on error")
		require.Contains(t, err.Error(), "failed to open file", "Error should indicate open failure")
	})
}

// TestReadFilesBytesLimit tests reading multiple files with size limits.
func TestReadFilesBytesLimit(t *testing.T) {
	t.Parallel()

	t.Run("read multiple files within limits", func(t *testing.T) {
		t.Parallel()

		// Create temp files.
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		content1 := []byte("content 1")
		content2 := []byte("content 2")

		require.NoError(t, os.WriteFile(file1, content1, 0o600))
		require.NoError(t, os.WriteFile(file2, content2, 0o600))

		// Read files with high limits.
		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit([]string{file1, file2}, 10, 100)
		require.NoError(t, err, "Failed to read files")
		require.Len(t, contents, 2, "Should have 2 file contents")
		require.Equal(t, content1, contents[0], "First file content should match")
		require.Equal(t, content2, contents[1], "Second file content should match")
	})

	t.Run("too many files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		file3 := filepath.Join(tmpDir, "file3.txt")

		require.NoError(t, os.WriteFile(file1, []byte("1"), 0o600))
		require.NoError(t, os.WriteFile(file2, []byte("2"), 0o600))
		require.NoError(t, os.WriteFile(file3, []byte("3"), 0o600))

		// Read with maxFiles=2 but provide 3 files.
		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit([]string{file1, file2, file3}, 2, 100)
		require.Error(t, err, "Should return error for too many files")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "too many files specified", "Error should indicate too many files")
	})

	t.Run("file too large", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "large.txt")
		largeContent := []byte(strings.Repeat("A", 100))
		require.NoError(t, os.WriteFile(file1, largeContent, 0o600))

		// Read with maxBytesPerFile=10 (file has 100 bytes).
		// Should succeed but only read first 10 bytes.
		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit([]string{file1}, 10, 10)
		require.NoError(t, err, "Should read partial content")
		require.Len(t, contents, 1, "Should have 1 file content")
		require.Len(t, contents[0], 10, "Should read only 10 bytes")
		require.Equal(t, []byte("AAAAAAAAAA"), contents[0], "Should read first 10 bytes")
	})

	t.Run("no files specified", func(t *testing.T) {
		t.Parallel()

		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit([]string{}, 10, 100)
		require.Error(t, err, "Should return error for empty file list")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "no files specified", "Error should indicate no files")
	})

	t.Run("empty file path", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		require.NoError(t, os.WriteFile(file1, []byte("content"), 0o600))

		contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit([]string{file1, ""}, 10, 100)
		require.Error(t, err, "Should return error for empty path")
		require.Nil(t, contents, "Contents should be nil on error")
		require.Contains(t, err.Error(), "empty file path", "Error should indicate empty path")
	})
}
