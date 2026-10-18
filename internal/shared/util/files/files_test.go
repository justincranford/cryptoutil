// Copyright (c) 2025-2026 Justin Cranford.
package files_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
			content:     testTest,
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
			content:     testTest,
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
			setupFiles:         []string{testFile1Txt},
			expectedExtensions: []string{testTxt},
			expectedTotalFiles: 1,
			validateFunc: func(result map[string][]string, _ string) bool {
				txtFiles := result[testTxt]

				return len(txtFiles) == 1
			},
			wantErr: false,
		},
		{
			name:               "Multiple_txt_files_in_directory",
			setupFiles:         []string{testFile1Txt, "file2.txt", "file3.txt"},
			expectedExtensions: []string{testTxt},
			expectedTotalFiles: expectedTxtFilesCount,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result[testTxt]) == expectedTxtFilesCount
			},
			wantErr: false,
		},
		{
			name: "Files_in_nested_directories",
			setupFiles: []string{
				testFile1Txt,
				filepath.Join("subdir1", "file2.txt"),
				filepath.Join("subdir1", "subdir2", "file3.txt"),
			},
			setupDirs:          []string{"subdir1", filepath.Join("subdir1", "subdir2")},
			expectedExtensions: []string{testTxt},
			expectedTotalFiles: expectedTxtFilesCount,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result[testTxt]) == expectedTxtFilesCount
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
			setupFiles:         []string{testFile1Txt, "file2.go", "file3.yml"},
			expectedExtensions: []string{testTxt, testGo, testYml},
			expectedTotalFiles: expectedTxtFilesCount,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result[testTxt]) == 1 && len(result[testGo]) == 1 && len(result[testYml]) == 1
			},
			wantErr: false,
		},
		{
			name:               "Dotfiles_like_gitignore",
			setupFiles:         []string{testGitignore},
			expectedExtensions: []string{testGitignoreValueA1AA6596},
			expectedTotalFiles: 1,
			validateFunc: func(result map[string][]string, _ string) bool {
				return len(result[testGitignoreValueA1AA6596]) == 1
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
			inclusions := []string{testTxt, testGo, testYml, testGitignoreValueA1AA6596}
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
	inclusions := []string{testGo}
	excludedNormalized := filepath.ToSlash(excludedDir)
	exclusions := []string{excludedNormalized}

	result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tempDir, inclusions, exclusions)
	require.NoError(t, err, "ListAllFilesWithOptions should succeed")

	// Should only have 1 file (from included directory).
	require.Equal(t, 1, len(result[testGo]), "Should have exactly 1 go file")
	require.Contains(t, result[testGo][0], "included", "File should be from included directory")
}

func TestListAllFiles_ExcludesVenv(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	venvDir := filepath.Join(tempDir, ".venv")
	require.NoError(t, os.MkdirAll(venvDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(filepath.Join(tempDir, "included.go"), "package included", cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(filepath.Join(venvDir, "excluded.go"), "package excluded", cryptoutilSharedMagic.CacheFilePermissions))

	filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFiles(tempDir)
	require.NoError(t, err)
	require.Len(t, filesByExtension[testGo], 1)
	require.Contains(t, filesByExtension[testGo][0], "included.go")
}

func TestListAllFiles_ExcludesGatlingOutputOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gatlingOutput := filepath.Join(root, "test", cryptoutilSharedMagic.WorkflowNameLoad, "target", "gatling", "simulation.log")
	otherTargetOutput := filepath.Join(root, "other", "target", "keep.log")
	otherLoadOutput := filepath.Join(root, "test", cryptoutilSharedMagic.WorkflowNameLoad, "reports", "keep.log")

	for _, path := range []string{gatlingOutput, otherTargetOutput, otherLoadOutput} {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(path, []byte{0, 1}, cryptoutilSharedMagic.CacheFilePermissions))
	}

	filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFiles(root)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{
		filepath.ToSlash(otherTargetOutput),
		filepath.ToSlash(otherLoadOutput),
	}, filesByExtension[testLog])
}

func TestListAllFiles_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	result, err := cryptoutilSharedUtilFiles.ListAllFiles("/nonexistent/directory/that/does/not/exist")
	require.Error(t, err, "ListAllFiles should return error for non-existent directory")
	require.Contains(t, err.Error(), "failed to walk directory", "Error should mention directory walk failure")
	require.Nil(t, result, "Result should be nil on error")
}

// TestReadFileBytes tests reading a single file.
