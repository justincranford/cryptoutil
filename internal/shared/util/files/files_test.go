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
			permissions: cryptoutilSharedMagic.CacheFilePermissions,
			wantErr:     false,
		},
		{
			name:        "Valid_byte_slice_content",
			content:     []byte("test content as bytes"),
			permissions: cryptoutilSharedMagic.CacheFilePermissions,
			wantErr:     false,
		},
		{
			name:        "Invalid_content_type",
			content:     123, // int, not string or []byte
			permissions: cryptoutilSharedMagic.CacheFilePermissions,
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
			permissions: cryptoutilSharedMagic.CacheFilePermissions,
			wantErr:     false,
		},
		{
			name:        "Empty_byte_slice_content",
			content:     []byte{},
			permissions: cryptoutilSharedMagic.CacheFilePermissions,
			wantErr:     false,
		},
		{
			name:        "Different_permissions",
			content:     "test",
			permissions: cryptoutilSharedMagic.CICDOutputFilePermissions,
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
				err := os.MkdirAll(dirPath, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
				require.NoError(t, err, "Should create subdirectory")
			}

			// Create files.
			for _, file := range tc.setupFiles {
				filePath := filepath.Join(tempDir, file)
				err := cryptoutilSharedUtilFiles.WriteFile(filePath, "test content", cryptoutilSharedMagic.CacheFilePermissions)
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

	err := os.MkdirAll(includedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err, "Should create included directory")

	err = os.MkdirAll(excludedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	require.NoError(t, err, "Should create excluded directory")

	// Create files in both directories.
	err = cryptoutilSharedUtilFiles.WriteFile(filepath.Join(includedDir, "included.go"), "package included", cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err, "Should create included file")

	err = cryptoutilSharedUtilFiles.WriteFile(filepath.Join(excludedDir, "excluded.go"), "package excluded", cryptoutilSharedMagic.CacheFilePermissions)
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
