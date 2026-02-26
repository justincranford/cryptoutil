package files_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

// TestReadFilesBytesLimit_BoundaryMaxFiles tests the boundary condition where
// len(filePaths) == maxFiles (should pass) vs len(filePaths) == maxFiles+1 (should fail).
// This kills the CONDITIONALS_BOUNDARY mutant on `len(filePaths) > int(maxFiles)`.
func TestReadFilesBytesLimit_BoundaryMaxFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fileCount int
		maxFiles  int64
		wantErr   bool
	}{
		{
			name:      "exactly at max files limit",
			fileCount: 2,
			maxFiles:  2,
			wantErr:   false,
		},
		{
			name:      "one over max files limit",
			fileCount: 3,
			maxFiles:  2,
			wantErr:   true,
		},
		{
			name:      "single file with max 1",
			fileCount: 1,
			maxFiles:  1,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			filePaths := make([]string, tc.fileCount)

			for i := range tc.fileCount {
				filePaths[i] = filepath.Join(tmpDir, "file"+string(rune('a'+i))+".txt")
				require.NoError(t, os.WriteFile(filePaths[i], []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))
			}

			contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit(filePaths, tc.maxFiles, cryptoutilSharedMagic.JoseJAMaxMaterials)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, contents)
				require.Contains(t, err.Error(), "too many files")
			} else {
				require.NoError(t, err)
				require.Len(t, contents, tc.fileCount)
			}
		})
	}
}

// TestReadFileBytesLimit_BoundaryFileSize tests the boundary condition where
// fileInfo.Size() == maxBytes (should read exact) vs fileInfo.Size() < maxBytes.
// This kills the CONDITIONALS_BOUNDARY mutant on `fileInfo.Size() < maxBytes`.
func TestReadFileBytesLimit_BoundaryFileSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		contentSize  int
		maxBytes     int64
		expectedSize int
	}{
		{
			name:         "file size equals maxBytes",
			contentSize:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			maxBytes:     cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			expectedSize: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		},
		{
			name:         "file size one less than maxBytes",
			contentSize:  9,
			maxBytes:     cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			expectedSize: 9,
		},
		{
			name:         "file size one more than maxBytes",
			contentSize:  11,
			maxBytes:     cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			expectedSize: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.txt")
			content := make([]byte, tc.contentSize)

			for i := range tc.contentSize {
				content[i] = byte('A' + (i % 26))
			}

			require.NoError(t, os.WriteFile(testFile, content, cryptoutilSharedMagic.CacheFilePermissions))

			result, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, tc.maxBytes)
			require.NoError(t, err)
			require.Len(t, result, tc.expectedSize)
			require.Equal(t, content[:tc.expectedSize], result)
		})
	}
}

// TestReadFileBytesLimit_MaxBytesZeroBoundary tests the boundary between
// maxBytes <= 0 (read entire file) vs maxBytes > 0 (read with limit).
// This kills the CONDITIONALS_NEGATION mutant on `maxBytes <= 0`.
func TestReadFileBytesLimit_MaxBytesZeroBoundary(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("ABCDEFGHIJ")
	require.NoError(t, os.WriteFile(testFile, testContent, cryptoutilSharedMagic.CacheFilePermissions))

	tests := []struct {
		name         string
		maxBytes     int64
		expectedSize int
	}{
		{
			name:         "maxBytes 0 reads entire file",
			maxBytes:     0,
			expectedSize: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		},
		{
			name:         "maxBytes -1 reads entire file",
			maxBytes:     -1,
			expectedSize: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		},
		{
			name:         "maxBytes 1 reads one byte",
			maxBytes:     1,
			expectedSize: 1,
		},
		{
			name:         "maxBytes 5 reads five bytes",
			maxBytes:     cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			expectedSize: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := cryptoutilSharedUtilFiles.ReadFileBytesLimit(testFile, tc.maxBytes)
			require.NoError(t, err)
			require.Len(t, result, tc.expectedSize)
			require.Equal(t, testContent[:tc.expectedSize], result)
		})
	}
}

// TestReadFilesBytesLimit_EmptyPathIndex verifies error message contains correct
// 1-based index. This kills ARITHMETIC_BASE mutants on `i+1` in error messages.
func TestReadFilesBytesLimit_EmptyPathIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		filePaths         []string
		expectedErrIndex  string
		expectedErrLength string
	}{
		{
			name:              "empty path at index 1 of 1",
			filePaths:         []string{"  "},
			expectedErrIndex:  "1 of 1",
			expectedErrLength: "1 of 1",
		},
		{
			name:              "empty path at index 2 of 3",
			filePaths:         []string{"/valid/path.txt", "  ", "/also/valid.txt"},
			expectedErrIndex:  "2 of 3",
			expectedErrLength: "2 of 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit(tc.filePaths, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.JoseJAMaxMaterials)
			require.Error(t, err)
			require.Nil(t, contents)
			require.Contains(t, err.Error(), tc.expectedErrIndex)
		})
	}
}

// TestListAllFilesWithOptions_ExtensionFiltering verifies that files are correctly
// included or excluded based on extension matching. This kills CONDITIONALS_NEGATION
// mutants on the extension checking logic.
func TestListAllFilesWithOptions_ExtensionFiltering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		files      map[string]string // filename -> content
		inclusions []string
		wantKeys   []string
		wantAbsent []string
	}{
		{
			name:       "include only matching extensions",
			files:      map[string]string{"file.go": "go code", "file.txt": "text", "file.md": "markdown"},
			inclusions: []string{"go", "md"},
			wantKeys:   []string{"go", "md"},
			wantAbsent: []string{"txt"},
		},
		{
			name:       "dotfile without extension matched by name",
			files:      map[string]string{".gitignore": "*.o", ".dockerignore": "node_modules"},
			inclusions: []string{"gitignore"},
			wantKeys:   []string{"gitignore"},
			wantAbsent: []string{"dockerignore"},
		},
		{
			name:       "regular file without extension not matched",
			files:      map[string]string{"Makefile": "all:", "README": "readme"},
			inclusions: []string{"go"},
			wantKeys:   []string{},
			wantAbsent: []string{"go"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()

			for name, content := range tc.files {
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			matches, err := cryptoutilSharedUtilFiles.ListAllFilesWithOptions(tmpDir, tc.inclusions, nil)
			require.NoError(t, err)

			for _, key := range tc.wantKeys {
				require.Contains(t, matches, key, "should include extension: %s", key)
				require.NotEmpty(t, matches[key], "should have files for extension: %s", key)
			}

			for _, key := range tc.wantAbsent {
				if files, ok := matches[key]; ok {
					require.Empty(t, files, "should not include extension: %s", key)
				}
			}
		})
	}
}

// TestWriteFile_ByteSliceContent verifies that []byte content is written correctly.
// This kills the CONDITIONALS_NEGATION mutant on the `case []byte:` type switch.
func TestWriteFile_ByteSliceContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "bytes.bin")
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}

	err := cryptoutilSharedUtilFiles.WriteFile(testFile, binaryContent, cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	readBack, readErr := os.ReadFile(testFile)
	require.NoError(t, readErr)
	require.Equal(t, binaryContent, readBack, "binary content should be written and read back correctly")
}

// TestReadFilesBytes_EmptyPathIndex verifies error message contains correct
// 1-based index for the non-limit variant. This kills ARITHMETIC_BASE mutants.
func TestReadFilesBytes_EmptyPathIndex(t *testing.T) {
	t.Parallel()

	contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes([]string{"  "})
	require.Error(t, err)
	require.Nil(t, contents)
	require.EqualError(t, err, "empty file path 1 of 1 in list")
}

// TestReadFilesBytes_ReadErrorIndex verifies error message contains correct
// 1-based index when a file fails to read (not validation phase).
// This kills the ARITHMETIC_BASE mutant on i+1 at the read loop (line 139).
func TestReadFilesBytes_ReadErrorIndex(t *testing.T) {
	t.Parallel()

	// Provide a non-existent file path that passes validation (non-empty)
	// but fails during ReadFileBytes.
	contents, err := cryptoutilSharedUtilFiles.ReadFilesBytes([]string{"/nonexistent/path/file.txt"})
	require.Error(t, err)
	require.Nil(t, contents)
	require.Contains(t, err.Error(), "failed to read file 1 of 1")
}

// TestReadFilesBytesLimit_ReadErrorIndex verifies error message contains correct
// 1-based index when a file fails to read in the limited variant.
// This kills the ARITHMETIC_BASE mutant on i+1 at the read loop (line 180).
func TestReadFilesBytesLimit_ReadErrorIndex(t *testing.T) {
	t.Parallel()

	// Provide a non-existent file path that passes validation but fails during read.
	contents, err := cryptoutilSharedUtilFiles.ReadFilesBytesLimit(
		[]string{"/nonexistent/path/file.txt"}, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.JoseJAMaxMaterials,
	)
	require.Error(t, err)
	require.Nil(t, contents)
	require.Contains(t, err.Error(), "failed to read file 1 of 1")
}
