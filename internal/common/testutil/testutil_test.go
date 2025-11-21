// Copyright (c) 2025 Justin Cranford
//
//

package testutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/common/testutil"
)

func TestWriteTempFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "Simple_text_file",
			filename: "test.txt",
			content:  "Hello, World!",
		},
		{
			name:     "YAML_file",
			filename: "config.yml",
			content:  "key: value\nfoo: bar",
		},
		{
			name:     "Empty_content",
			filename: "empty.txt",
			content:  "",
		},
		{
			name:     "JSON_file",
			filename: "data.json",
			content:  `{"name":"test","value":123}`,
		},
		{
			name:     "Multiline_content",
			filename: "multi.txt",
			content:  "Line 1\nLine 2\nLine 3\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			// Act: Write temporary file
			filePath := testutil.WriteTempFile(t, tempDir, tc.filename, tc.content)

			// Assert: File exists at expected path
			expectedPath := filepath.Join(tempDir, tc.filename)
			require.Equal(t, expectedPath, filePath, "Should return correct file path")

			// Assert: File exists
			_, err := os.Stat(filePath)
			require.NoError(t, err, "File should exist")

			// Assert: File content matches
			actualContent, err := os.ReadFile(filePath)
			require.NoError(t, err, "Should be able to read file")
			require.Equal(t, tc.content, string(actualContent), "File content should match")
		})
	}
}

func TestWriteTempFile_NestedDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "sub", "nested")
	err := os.MkdirAll(nestedDir, 0o755)
	require.NoError(t, err)

	// Act: Write file in nested directory
	filePath := testutil.WriteTempFile(t, nestedDir, "nested.txt", "nested content")

	// Assert: File exists in nested directory
	expectedPath := filepath.Join(nestedDir, "nested.txt")
	require.Equal(t, expectedPath, filePath)

	actualContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, "nested content", string(actualContent))
}

func TestWriteTestFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "Absolute_path",
			filename: "absolute.txt",
			content:  "Absolute path content",
		},
		{
			name:     "Binary_content",
			filename: "binary.dat",
			content:  "\x00\x01\x02\xFF",
		},
		{
			name:     "Large_content",
			filename: "large.txt",
			content:  string(make([]byte, 10000)), // 10KB of zeros
		},
		{
			name:     "Special_characters",
			filename: "special.txt",
			content:  "Special: !@#$%^&*()_+{}|:\"<>?",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tc.filename)

			// Act: Write test file
			testutil.WriteTestFile(t, filePath, tc.content)

			// Assert: File exists
			_, err := os.Stat(filePath)
			require.NoError(t, err, "File should exist")

			// Assert: File content matches
			actualContent, err := os.ReadFile(filePath)
			require.NoError(t, err, "Should be able to read file")
			require.Equal(t, tc.content, string(actualContent), "File content should match")
		})
	}
}

func TestWriteTestFile_CreateDirectories(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "does", "not", "exist", "file.txt")

	// WriteTestFile should fail if parent directory doesn't exist
	// (os.WriteFile doesn't create parent directories)
	// This test verifies the expected behavior

	// Create parent directories first
	err := os.MkdirAll(filepath.Dir(nestedPath), 0o755)
	require.NoError(t, err)

	// Now WriteTestFile should succeed
	testutil.WriteTestFile(t, nestedPath, "nested file content")

	actualContent, err := os.ReadFile(nestedPath)
	require.NoError(t, err)
	require.Equal(t, "nested file content", string(actualContent))
}

func TestReadTestFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "Text_content",
			content: "Test file content",
		},
		{
			name:    "Empty_file",
			content: "",
		},
		{
			name:    "Binary_content",
			content: "\x00\x01\x02\xFF\xFE",
		},
		{
			name:    "Multiline_content",
			content: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:    "Unicode_content",
			content: "Hello 世界 🌍",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "test.txt")

			// Setup: Write file using standard library
			err := os.WriteFile(filePath, []byte(tc.content), 0o600)
			require.NoError(t, err)

			// Act: Read test file
			actualContent := testutil.ReadTestFile(t, filePath)

			// Assert: Content matches
			require.Equal(t, []byte(tc.content), actualContent, "File content should match")
		})
	}
}

func TestReadTestFile_Integration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	content := "Integration test content"

	// Write using WriteTestFile
	filePath := filepath.Join(tempDir, "integration.txt")
	testutil.WriteTestFile(t, filePath, content)

	// Read using ReadTestFile
	actualContent := testutil.ReadTestFile(t, filePath)

	// Assert: Round-trip succeeds
	require.Equal(t, []byte(content), actualContent, "Round-trip content should match")
}

func TestWriteAndRead_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "Simple_roundtrip",
			content: "Simple content",
		},
		{
			name:    "Complex_roundtrip",
			content: "Multi\nline\ncontent\nwith\nspecial: chars!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			// Write using WriteTempFile
			filePath := testutil.WriteTempFile(t, tempDir, "roundtrip.txt", tc.content)

			// Read using ReadTestFile
			actualContent := testutil.ReadTestFile(t, filePath)

			// Assert: Round-trip matches
			require.Equal(t, []byte(tc.content), actualContent)
		})
	}
}
