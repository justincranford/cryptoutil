// Copyright (c) 2025 Justin Cranford

package lint_text

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintTextUTF8 "cryptoutil/internal/apps/cicd/lint_text/utf8"
)

func TestLint_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Lint(logger, map[string][]string{})

	require.NoError(t, err, "Lint should succeed with no files")
}

func TestLint_ValidUTF8Files(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create valid UTF-8 files.
	validFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(validFile, []byte("Hello, World!"), 0o600)
	require.NoError(t, err)

	validGoFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(validGoFile, []byte("package main\n\nfunc main() {}\n"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"txt": {validFile},
		"go":  {validGoFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with valid UTF-8 files")
}

func TestLint_FileWithBOM(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create file with UTF-8 BOM.
	bomFile := filepath.Join(tmpDir, "bom.txt")

	bomContent := append([]byte{0xEF, 0xBB, 0xBF}, []byte("Hello with BOM")...)
	err := os.WriteFile(bomFile, bomContent, 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"txt": {bomFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "Lint should fail with BOM file")
	require.Contains(t, err.Error(), "lint-text failed", "Error should indicate lint-text failure")
}

func TestFilterTextFilesViaLint(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main"), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "Lint should succeed with valid go files")
}

func TestFilterTextFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{
			name:     "empty input",
			input:    []string{},
			expected: 0,
		},
		{
			name:     "go files included",
			input:    []string{"main.go", "test.go", "util.go"},
			expected: 3,
		},
		{
			name:     "all files passed through after directory-level filtering",
			input:    []string{"main.go", "image.png", "data.json", "binary.exe"},
			expected: 4, // All files passed through since directory-level filtering happens in ListAllFiles.
		},
		{
			name:     "self-exclusion for lint-text command",
			input:    []string{"internal/apps/cicd/lint_text/utf8.go", "other.go"},
			expected: 1, // Self-exclusion filters lint_text directory.
		},
		{
			name:     "generated files excluded",
			input:    []string{"model_gen.go", "service.pb.go", "regular.go"},
			expected: 1, // Generated files filtered by pattern.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := lintTextUTF8.FilterTextFiles(tc.input)
			require.Len(t, result, tc.expected, "Filtered file count should match expected")
		})
	}
}

func TestCheckFileEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     []byte
		expectIssue bool
	}{
		{
			name:        "valid UTF-8",
			content:     []byte("Hello, World!"),
			expectIssue: false,
		},
		{
			name:        "UTF-8 with BOM",
			content:     append([]byte{0xEF, 0xBB, 0xBF}, []byte("Hello")...),
			expectIssue: true,
		},
		{
			name:        "empty file",
			content:     []byte{},
			expectIssue: false,
		},
		{
			name:        "short file without BOM",
			content:     []byte("Hi"),
			expectIssue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.txt")

			err := os.WriteFile(testFile, tc.content, 0o600)
			require.NoError(t, err)

			issues := lintTextUTF8.CheckFileEncoding(testFile)

			if tc.expectIssue {
				require.NotEmpty(t, issues, "Expected encoding issue")
			} else {
				require.Empty(t, issues, "Expected no encoding issues")
			}
		})
	}
}

// TestCheckFileEncoding_FileOpenError tests the error path when file cannot be opened.
func TestCheckFileEncoding_FileOpenError(t *testing.T) {
	t.Parallel()

	// Pass a non-existent file path.
	issues := lintTextUTF8.CheckFileEncoding("/nonexistent/path/to/file.txt")

	require.Len(t, issues, 1, "Should return one issue")
	require.Contains(t, issues[0], "failed to open file", "Issue should indicate file open failure")
}
