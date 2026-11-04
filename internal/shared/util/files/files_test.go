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

func TestListRepositoryFilesWithOptions_SelectedWorktreeRootsRemainScannable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: cryptoutilSharedMagic.CICDExcludeDirDotWorktree},
		{name: cryptoutilSharedMagic.CICDExcludeDirWorktree},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := filepath.Join(t.TempDir(), tc.name)
			require.NoError(t, os.MkdirAll(rootDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
				filepath.Join(rootDir, "included.txt"),
				"selected root contents",
				cryptoutilSharedMagic.CacheFilePermissions,
			))

			result, err := cryptoutilSharedUtilFiles.ListRepositoryFilesWithOptions(
				rootDir,
				[]string{testTxt},
				[]string{filepath.ToSlash(rootDir), tc.name},
			)

			require.NoError(t, err)
			require.Len(t, result.FilesByExtension[testTxt], 1)
			require.Contains(t, result.FilesByExtension[testTxt][0], "included.txt")
		})
	}
}

func TestListAllFilesWithOptions_PreservesGenericNestedCheckoutContents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		markerIsDir bool
	}{
		{name: testGitMarkerFile},
		{name: testGitMarkerDirectory, markerIsDir: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			checkoutDir := filepath.Join(rootDir, "ordinary-named-checkout")
			require.NoError(t, os.MkdirAll(checkoutDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
				filepath.Join(rootDir, "included.txt"),
				"selected root contents",
				cryptoutilSharedMagic.CacheFilePermissions,
			))
			require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
				filepath.Join(checkoutDir, "excluded.txt"),
				"nested checkout contents",
				cryptoutilSharedMagic.CacheFilePermissions,
			))

			gitMarkerPath := filepath.Join(checkoutDir, cryptoutilSharedMagic.CICDExcludeDirGit)
			if tc.markerIsDir {
				require.NoError(t, os.Mkdir(gitMarkerPath, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			} else {
				require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
					gitMarkerPath,
					"gitdir: ../.git/worktrees/ordinary-named-checkout",
					cryptoutilSharedMagic.CacheFilePermissions,
				))
			}

			result, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(rootDir, []string{testTxt}, nil)

			require.NoError(t, err)
			require.Len(t, result[testTxt], 2)
			require.ElementsMatch(t, []string{
				filepath.ToSlash(filepath.Join(rootDir, "included.txt")),
				filepath.ToSlash(filepath.Join(checkoutDir, "excluded.txt")),
			}, result[testTxt])
		})
	}
}

func TestListAllFilesWithOptions_PreservesRelativePathExclusions(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, "source"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, testGeneratedDirectory, testNestedDirectory), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
		filepath.Join(rootDir, "source", "included.go"),
		"package source",
		cryptoutilSharedMagic.CacheFilePermissions,
	))
	require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
		filepath.Join(rootDir, testGeneratedDirectory, testNestedDirectory, "excluded.go"),
		"package generated",
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(
		rootDir,
		[]string{testGo},
		[]string{testGeneratedDirectory},
	)

	require.NoError(t, err)
	require.Equal(t, []string{filepath.ToSlash(filepath.Join(rootDir, "source", "included.go"))}, filesByExtension[testGo])
}

func TestListAllFiles_ExcludesVenv(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	venvDir := filepath.Join(tempDir, cryptoutilSharedMagic.CICDExcludeDirVenv)
	require.NoError(t, os.MkdirAll(venvDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(filepath.Join(tempDir, "included.go"), "package included", cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(filepath.Join(venvDir, "excluded.go"), "package excluded", cryptoutilSharedMagic.CacheFilePermissions))

	filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFiles(tempDir)
	require.NoError(t, err)
	require.Len(t, filesByExtension[testGo], 1)
	require.Contains(t, filesByExtension[testGo][0], "included.go")
}

func TestListAllFiles_ExcludesMachineOwnedDirectories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dirName string
	}{
		{name: "git metadata", dirName: cryptoutilSharedMagic.CICDExcludeDirGit},
		{name: "hidden worktree container", dirName: cryptoutilSharedMagic.CICDExcludeDirDotWorktree},
		{name: "worktree container", dirName: cryptoutilSharedMagic.CICDExcludeDirWorktree},
		{name: "DAST reports", dirName: cryptoutilSharedMagic.CICDExcludeDirDastReports},
		{name: "E2E reports", dirName: cryptoutilSharedMagic.CICDExcludeDirE2EReports},
		{name: "load reports", dirName: cryptoutilSharedMagic.CICDExcludeDirLoadReports},
		{name: "test results", dirName: cryptoutilSharedMagic.CICDExcludeDirTestResults},
		{name: "temporary output", dirName: cryptoutilSharedMagic.CICDExcludeDirTemp},
		{name: "legacy CI cache", dirName: cryptoutilSharedMagic.CICDExcludeDirDotCICD},
		{name: "CI lint cache", dirName: cryptoutilSharedMagic.CICDExcludeDirDotCICDLint},
		{name: "build output", dirName: cryptoutilSharedMagic.CICDExcludeDirBuild},
		{name: "distribution output", dirName: cryptoutilSharedMagic.CICDExcludeDirDist},
		{name: "generated output", dirName: cryptoutilSharedMagic.CICDExcludeDirOutput},
		{name: "legacy Python virtual environment", dirName: cryptoutilSharedMagic.CICDExcludeDirVenvLegacy},
		{name: "Python environment", dirName: cryptoutilSharedMagic.CICDExcludeDirEnv},
		{name: "hidden environment", dirName: cryptoutilSharedMagic.CICDExcludeDirDotEnv},
		{name: "cache", dirName: cryptoutilSharedMagic.CICDExcludeDirCache},
		{name: "Gatling Maven target output", dirName: cryptoutilSharedMagic.DirectoryNameExclusionTestLoadTarget},
		{name: "Python bytecode cache", dirName: cryptoutilSharedMagic.CICDExcludeDirPycache},
		{name: "pytest cache", dirName: cryptoutilSharedMagic.CICDExcludeDirPytestCache},
		{name: "tox environment", dirName: cryptoutilSharedMagic.CICDExcludeDirTox},
		{name: "nox environment", dirName: cryptoutilSharedMagic.CICDExcludeDirNox},
		{name: "Hypothesis cache", dirName: cryptoutilSharedMagic.CICDExcludeDirHypothesis},
		{name: "HTML coverage output", dirName: cryptoutilSharedMagic.CICDExcludeDirHTMLCov},
		{name: "Ruff cache", dirName: cryptoutilSharedMagic.CICDExcludeDirRuffCache},
		{name: "mypy cache", dirName: cryptoutilSharedMagic.CICDExcludeDirMypyCache},
		{name: "mutation test output", dirName: cryptoutilSharedMagic.CICDExcludeDirGremlins},
		{name: "vendored dependencies", dirName: cryptoutilSharedMagic.CICDExcludeDirVendor},
		{name: "generated test output", dirName: cryptoutilSharedMagic.CICDExcludeDirTestOutput},
		{name: "Python virtual environment", dirName: cryptoutilSharedMagic.CICDExcludeDirVenv},
		{name: "Node.js dependencies", dirName: cryptoutilSharedMagic.CICDExcludeDirNodeModules},
		{name: "workflow reports", dirName: cryptoutilSharedMagic.CICDExcludeDirWorkflowReports},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			excludedDir := filepath.Join(rootDir, tc.dirName)
			require.NoError(t, os.MkdirAll(excludedDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
				filepath.Join(rootDir, "included.go"),
				"package included",
				cryptoutilSharedMagic.CacheFilePermissions,
			))
			require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
				filepath.Join(excludedDir, "excluded.go"),
				"package excluded",
				cryptoutilSharedMagic.CacheFilePermissions,
			))

			filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFiles(rootDir)

			require.NoError(t, err)
			require.Len(t, filesByExtension[testGo], 1)
			require.Contains(t, filesByExtension[testGo][0], "included.go")
		})
	}
}

func TestListAllFiles_ExcludesNestedMachineOwnedOutput(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	nestedOutputDirs := []string{
		filepath.Join(testGeneratedDirectory, cryptoutilSharedMagic.CICDExcludeDirTemp),
		filepath.Join("reports", cryptoutilSharedMagic.CICDExcludeDirDastReports),
		filepath.Join("artifacts", cryptoutilSharedMagic.CICDExcludeDirBuild),
	}

	require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
		filepath.Join(rootDir, "included.go"),
		"package included",
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	for _, directory := range nestedOutputDirs {
		outputDir := filepath.Join(rootDir, directory)
		require.NoError(t, os.MkdirAll(outputDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, cryptoutilSharedUtilFiles.WriteFile(
			filepath.Join(outputDir, "generated.go"),
			"package generated",
			cryptoutilSharedMagic.CacheFilePermissions,
		))
	}

	filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFiles(rootDir)

	require.NoError(t, err)
	require.Len(t, filesByExtension[testGo], 1)
	require.Contains(t, filesByExtension[testGo][0], "included.go")
}

func TestListAllFiles_ExcludesGatlingOutputOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gatlingOutput := filepath.Join(root, "test", cryptoutilSharedMagic.WorkflowNameLoad, cryptoutilSharedMagic.CICDExcludeDirTarget, "gatling", "simulation.log")
	otherTargetOutput := filepath.Join(root, "other", cryptoutilSharedMagic.CICDExcludeDirTarget, "keep.log")
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
