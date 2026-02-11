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

func TestWriteFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     any
		permissions os.FileMode
		wantErr     bool
		errContains string
	}{
		{
			name:        "Valid_string_content",
			content:     "test content",
			permissions: 0o600,
			wantErr:     false,
		},
		{
			name:        "Valid_byte_slice_content",
			content:     []byte("test content as bytes"),
			permissions: 0o600,
			wantErr:     false,
		},
		{
			name:        "Invalid_content_type",
			content:     123, // int, not string or []byte
			permissions: 0o600,
			wantErr:     true,
			errContains: "content must be string or []byte",
		},
		{
			name:        "Zero_permissions",
			content:     "test",
			permissions: 0,
			wantErr:     true,
			errContains: "missing file permissions",
		},
		{
			name:        "Empty_string_content",
			content:     "",
			permissions: 0o600,
			wantErr:     false,
		},
		{
			name:        "Empty_byte_slice_content",
			content:     []byte{},
			permissions: 0o600,
			wantErr:     false,
		},
		{
			name:        "Different_permissions",
			content:     "test",
			permissions: 0o644,
			wantErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory for test
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.txt")

			err := cryptoutilSharedUtilFiles.WriteFile(testFile, tc.content, tc.permissions)

			if tc.wantErr {
				require.Error(t, err, "WriteFile should return error")
				require.Contains(t, err.Error(), tc.errContains, "Error message should match")
			} else {
				require.NoError(t, err, "WriteFile should succeed")

				// Verify file exists
				_, statErr := os.Stat(testFile)
				require.NoError(t, statErr, "File should exist")

				// Verify content
				content, readErr := os.ReadFile(testFile)
				require.NoError(t, readErr, "Should read file")

				var expectedContent []byte

				switch v := tc.content.(type) {
				case string:
					expectedContent = []byte(v)
				case []byte:
					expectedContent = v
				}

				require.Equal(t, expectedContent, content, "File content should match")
				// Note: Skip permission verification on Windows as file permissions work differently
			}
		})
	}
}

func TestListAllFiles(t *testing.T) {
	t.Parallel()

	const expectedTxtFilesCount = 3

	tests := []struct {
		name               string
		setupFiles         []string // Files to create in temp dir
		setupDirs          []string // Subdirectories to create
		expectedExtensions []string // Extensions expected in result
		expectedTotalFiles int      // Total files expected across all extensions
		validateFunc       func(map[string][]string, string) bool
		wantErr            bool
		errContains        string
	}{
		{
			name:               "Single_txt_file_in_directory",
			setupFiles:         []string{"file1.txt"},
			expectedExtensions: []string{"txt"},
			expectedTotalFiles: 1,
			validateFunc: func(result map[string][]string, _ string) bool {
				txtFiles := result["txt"]

				return len(txtFiles) == 1
			},
			wantErr: false,
		},
		{
			name:               "Multiple_txt_files_in_directory",
			setupFiles:         []string{"file1.txt", "file2.txt", "file3.txt"},
			expectedExtensions: []string{"txt"},
			expectedTotalFiles: expectedTxtFilesCount,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result["txt"]) == expectedTxtFilesCount
			},
			wantErr: false,
		},
		{
			name: "Files_in_nested_directories",
			setupFiles: []string{
				"file1.txt",
				filepath.Join("subdir1", "file2.txt"),
				filepath.Join("subdir1", "subdir2", "file3.txt"),
			},
			setupDirs:          []string{"subdir1", filepath.Join("subdir1", "subdir2")},
			expectedExtensions: []string{"txt"},
			expectedTotalFiles: expectedTxtFilesCount,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result["txt"]) == expectedTxtFilesCount
			},
			wantErr: false,
		},
		{
			name:               "Empty_directory",
			setupFiles:         []string{},
			expectedExtensions: []string{},
			expectedTotalFiles: 0,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result) == 0
			},
			wantErr: false,
		},
		{
			name:               "Mixed_extensions",
			setupFiles:         []string{"file1.txt", "file2.go", "file3.yml"},
			expectedExtensions: []string{"txt", "go", "yml"},
			expectedTotalFiles: expectedTxtFilesCount,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result["txt"]) == 1 && len(result["go"]) == 1 && len(result["yml"]) == 1
			},
			wantErr: false,
		},
		{
			name:               "Dotfiles_like_gitignore",
			setupFiles:         []string{".gitignore"},
			expectedExtensions: []string{"gitignore"},
			expectedTotalFiles: 1,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result["gitignore"]) == 1
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory.
			tempDir := t.TempDir()

			// Create subdirectories.
			for _, dir := range tc.setupDirs {
				dirPath := filepath.Join(tempDir, dir)
				err := os.MkdirAll(dirPath, 0o755)
				require.NoError(t, err, "Should create subdirectory")
			}

			// Create files.
			for _, file := range tc.setupFiles {
				filePath := filepath.Join(tempDir, file)
				err := cryptoutilSharedUtilFiles.WriteFile(filePath, "test content", 0o600)
				require.NoError(t, err, "Should create test file")
			}

			// Call function under test with custom options to include all test file extensions.
			inclusions := []string{"txt", "go", "yml", "gitignore"}
			exclusions := []string{}

			result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tempDir, inclusions, exclusions)

			if tc.wantErr {
				require.Error(t, err, "ListAllFiles should return error")
				require.Contains(t, err.Error(), tc.errContains, "Error message should match")
			} else {
				require.NoError(t, err, "ListAllFiles should succeed")
				require.True(t, tc.validateFunc(result, tempDir), "Result validation failed")

				// Verify total file count.
				totalFiles := 0
				for _, fileList := range result {
					totalFiles += len(fileList)
				}

				require.Equal(t, tc.expectedTotalFiles, totalFiles, "Total file count should match")
			}
		})
	}
}

func TestListAllFilesWithOptions_DirectoryExclusions(t *testing.T) {
	t.Parallel()

	// Create temp directory.
	tempDir := t.TempDir()

	// Create directories including one to exclude.
	includedDir := filepath.Join(tempDir, "included")
	excludedDir := filepath.Join(tempDir, "excluded")

	err := os.MkdirAll(includedDir, 0o755)
	require.NoError(t, err, "Should create included directory")

	err = os.MkdirAll(excludedDir, 0o755)
	require.NoError(t, err, "Should create excluded directory")

	// Create files in both directories.
	err = cryptoutilSharedUtilFiles.WriteFile(filepath.Join(includedDir, "included.go"), "package included", 0o600)
	require.NoError(t, err, "Should create included file")

	err = cryptoutilSharedUtilFiles.WriteFile(filepath.Join(excludedDir, "excluded.go"), "package excluded", 0o600)
	require.NoError(t, err, "Should create excluded file")

	// Call function with exclusion using the normalized excluded directory path.
	inclusions := []string{"go"}
	excludedNormalized := filepath.ToSlash(excludedDir)
	exclusions := []string{excludedNormalized}

	result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tempDir, inclusions, exclusions)
	require.NoError(t, err, "ListAllFilesWithOptions should succeed")

	// Should only have 1 file (from included directory).
	require.Equal(t, 1, len(result["go"]), "Should have exactly 1 go file")
	require.Contains(t, result["go"][0], "included", "File should be from included directory")
}

func TestListAllFiles_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	result, err := cryptoutilSharedUtilFiles.ListAllFiles("/nonexistent/directory/that/does/not/exist")
	require.Error(t, err, "ListAllFiles should return error for non-existent directory")
	require.Contains(t, err.Error(), "failed to walk directory", "Error should mention directory walk failure")
	require.Nil(t, result, "Result should be nil on error")
}

// TestReadFileBytes tests reading a single file.
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
