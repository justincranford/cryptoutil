// Copyright (c) 2025 Justin Cranford

package utf8

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

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

			result := FilterTextFiles(tc.input)
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

			issues := CheckFileEncoding(testFile)

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
	issues := CheckFileEncoding("/nonexistent/path/to/file.txt")

	require.Len(t, issues, 1, "Should return one issue")
	require.Contains(t, issues[0], "failed to open file", "Issue should indicate file open failure")
}

func TestCheck_EmptyFilesByExtension(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestCheck_WithValidUTF8Files(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err := Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestCheck_WithBOMFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	bomFile := filepath.Join(tmpDir, "utf8bom.go")

	// Write file with UTF-8 BOM (EF BB BF prefix).
	contentWithBOM := "\xef\xbb\xbfpackage main\n"
	require.NoError(t, os.WriteFile(bomFile, []byte(contentWithBOM), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {bomFile},
	}

	err := Check(logger, filesByExtension)
	require.Error(t, err, "Should detect UTF-8 BOM violation")
	require.Contains(t, err.Error(), "encoding violations")
}

func TestFlattenFileMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string][]string
		expected int
	}{
		{
			name:     "empty map",
			input:    map[string][]string{},
			expected: 0,
		},
		{
			name: "single extension with files",
			input: map[string][]string{
				"go": {"file1.go", "file2.go"},
			},
			expected: 2,
		},
		{
			name: "multiple extensions",
			input: map[string][]string{
				"go":  {"file1.go"},
				"yml": {"config.yml", "compose.yml"},
				"md":  {"README.md"},
			},
			expected: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := flattenFileMap(tc.input)
			require.Len(t, result, tc.expected)
		})
	}
}

func TestCheckFilesEncoding_MultipleFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	validFile := filepath.Join(tmpDir, "valid.go")
	require.NoError(t, os.WriteFile(validFile, []byte("package main\n"), 0o600))

	bomFile := filepath.Join(tmpDir, "bom.go")
	require.NoError(t, os.WriteFile(bomFile, []byte("\xef\xbb\xbfpackage main\n"), 0o600))

	violations := checkFilesEncoding([]string{validFile, bomFile})
	require.Len(t, violations, 1, "Only BOM file should be a violation")
}
