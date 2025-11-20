// Copyright (c) 2025 Justin Cranford

package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/common/util/files"
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory for test
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.txt")

			err := files.WriteFile(testFile, tc.content, tc.permissions)

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

	tests := []struct {
		name         string
		setupFiles   []string        // Files to create in temp dir
		setupDirs    []string        // Subdirectories to create
		expectedFunc func([]string, string) bool // Custom validation function
		wantErr      bool
		errContains  string
	}{
		{
			name: "Single_file_in_directory",
			setupFiles: []string{"file1.txt"},
			expectedFunc: func(result []string, baseDir string) bool {
				return len(result) == 1 && filepath.Base(result[0]) == "file1.txt"
			},
			wantErr: false,
		},
		{
			name: "Multiple_files_in_directory",
			setupFiles: []string{"file1.txt", "file2.txt", "file3.txt"},
			expectedFunc: func(result []string, baseDir string) bool {
				return len(result) == 3
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
			setupDirs: []string{"subdir1", filepath.Join("subdir1", "subdir2")},
			expectedFunc: func(result []string, baseDir string) bool {
				return len(result) == 3
			},
			wantErr: false,
		},
		{
			name:       "Empty_directory",
			setupFiles: []string{},
			expectedFunc: func(result []string, baseDir string) bool {
				return len(result) == 0
			},
			wantErr: false,
		},
		{
			name:       "Mixed_files_and_empty_subdirectories",
			setupFiles: []string{"file1.txt"},
			setupDirs:  []string{"emptydir"},
			expectedFunc: func(result []string, baseDir string) bool {
				return len(result) == 1 // Only file1.txt, not the empty directory
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory
			tempDir := t.TempDir()

			// Create subdirectories
			for _, dir := range tc.setupDirs {
				dirPath := filepath.Join(tempDir, dir)
				err := os.MkdirAll(dirPath, 0o755)
				require.NoError(t, err, "Should create subdirectory")
			}

			// Create files
			for _, file := range tc.setupFiles {
				filePath := filepath.Join(tempDir, file)
				err := files.WriteFile(filePath, "test content", 0o600)
				require.NoError(t, err, "Should create test file")
			}

			// Call function under test
			result, err := files.ListAllFiles(tempDir)

			if tc.wantErr {
				require.Error(t, err, "ListAllFiles should return error")
				require.Contains(t, err.Error(), tc.errContains, "Error message should match")
			} else {
				require.NoError(t, err, "ListAllFiles should succeed")
				require.True(t, tc.expectedFunc(result, tempDir), "Result validation failed")
			}
		})
	}
}

func TestListAllFiles_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	result, err := files.ListAllFiles("/nonexistent/directory/that/does/not/exist")
	require.Error(t, err, "ListAllFiles should return error for non-existent directory")
	require.Contains(t, err.Error(), "failed to walk directory", "Error should mention directory walk failure")
	require.Nil(t, result, "Result should be nil on error")
}
