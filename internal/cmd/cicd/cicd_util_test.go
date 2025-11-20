// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilTestutil "cryptoutil/internal/common/testutil"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

func TestListAllFiles(t *testing.T) {
	// Create a temporary directory with some test files
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"subdir/file3.txt",
		"subdir/nested/file4.go",
	}

	for _, file := range testFiles {
		dir := filepath.Dir(file)
		if dir != "." {
			require.NoError(t, os.MkdirAll(filepath.Join(tempDir, dir), cryptoutilMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		}

		cryptoutilTestutil.WriteTempFile(t, tempDir, file, "test content")
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()

	require.NoError(t, os.Chdir(tempDir))

	// Collect files
	files, err := cryptoutilFiles.ListAllFiles(".")
	require.NoError(t, err)

	// Should find all test files
	require.Len(t, files, len(testFiles), "Should find all test files")

	// Convert to relative paths for comparison
	for i, file := range files {
		files[i], err = filepath.Rel(tempDir, filepath.Join(tempDir, file))
		require.NoError(t, err)
		// Normalize path separators to forward slashes for cross-platform comparison
		files[i] = filepath.ToSlash(files[i])
	}

	// Normalize expected paths to forward slashes
	normalizedTestFiles := make([]string, len(testFiles))
	for i, file := range testFiles {
		normalizedTestFiles[i] = filepath.ToSlash(file)
	}

	// Sort both slices for comparison
	require.ElementsMatch(t, normalizedTestFiles, files, "Should find all expected files")
}
