// Copyright (c) 2025 Justin Cranford

package all_enforce_utf8

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilTestutil "cryptoutil/internal/common/testutil"
)

func TestCheckFileEncoding(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	tests := []struct {
		name         string
		fileName     string
		content      string
		createFile   bool
		wantIssues   int
		wantContains string
	}{
		{
			name:       "valid UTF-8 file without BOM",
			fileName:   "valid_utf8.txt",
			content:    "Hello, world! This is valid UTF-8 content without BOM.",
			createFile: true,
			wantIssues: 0,
		},
		{
			name:         "file with UTF-8 BOM",
			fileName:     "utf8_bom.txt",
			content:      "\xEF\xBB\xBFHello, world! This has UTF-8 BOM.",
			createFile:   true,
			wantIssues:   1,
			wantContains: "contains UTF-8 BOM",
		},
		{
			name:         "file with UTF-16 LE BOM",
			fileName:     "utf16_le_bom.txt",
			content:      "\xFF\xFEHello, world! This has UTF-16 LE BOM.",
			createFile:   true,
			wantIssues:   1,
			wantContains: "contains UTF-16 LE BOM",
		},
		{
			name:         "file with UTF-16 BE BOM",
			fileName:     "utf16_be_bom.txt",
			content:      "\xFE\xFFHello, world! This has UTF-16 BE BOM.",
			createFile:   true,
			wantIssues:   1,
			wantContains: "contains UTF-16 BE BOM",
		},
		{
			name:         "file with UTF-32 LE BOM",
			fileName:     "utf32_le_bom.txt",
			content:      "\xFF\xFE\x00\x00Hello, world! This has UTF-32 LE BOM.",
			createFile:   true,
			wantIssues:   1,
			wantContains: "contains UTF-32 LE BOM",
		},
		{
			name:         "file with UTF-32 BE BOM",
			fileName:     "utf32_be_bom.txt",
			content:      "\x00\x00\xFE\xFFHello, world! This has UTF-32 BE BOM.",
			createFile:   true,
			wantIssues:   1,
			wantContains: "contains UTF-32 BE BOM",
		},
		{
			name:         "file does not exist",
			fileName:     "nonexistent.txt",
			content:      "",
			createFile:   false,
			wantIssues:   1,
			wantContains: "Error reading file",
		},
		{
			name:       "empty file",
			fileName:   "empty.txt",
			content:    "",
			createFile: true,
			wantIssues: 0,
		},
		{
			name:         "file with only BOM",
			fileName:     "only_bom.txt",
			content:      "\xEF\xBB\xBF",
			createFile:   true,
			wantIssues:   1,
			wantContains: "contains UTF-8 BOM",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var filePath string
			if tc.createFile {
				filePath = cryptoutilTestutil.WriteTempFile(t, tempDir, tc.fileName, tc.content)
			} else {
				filePath = filepath.Join(tempDir, tc.fileName)
			}

			issues := checkFileEncoding(filePath)
			if tc.wantIssues == 0 {
				require.Empty(t, issues, "%s: expected no issues", tc.name)
			} else {
				require.Len(t, issues, tc.wantIssues, "%s: unexpected number of issues", tc.name)
				if tc.wantContains != "" {
					require.Contains(t, issues[0], tc.wantContains, "%s: issue message mismatch", tc.name)
				}
			}
		})
	}
}

func TestEnforce(t *testing.T) {
	t.Run("no files to check", func(t *testing.T) {
		// Test with empty file list
		logger := common.NewLogger("test")
		err := Enforce(logger, []string{})
		require.NoError(t, err, "Should not return error for no files")
	})

	t.Run("files with encoding violations", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test files
		invalidFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "invalid.go", "\xEF\xBB\xBFpackage main\n\nfunc main() {}\n")

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()

		require.NoError(t, os.Chdir(tempDir))

		// Test that Enforce returns an error for encoding violations
		logger := common.NewLogger("test")
		err = Enforce(logger, []string{invalidFile})
		require.Error(t, err, "Should return error for encoding violations")
		require.Contains(t, err.Error(), "file encoding violations found", "Error should mention encoding violations")
	})

	t.Run("all files valid", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create valid test files
		goFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.go", "package main\n\nfunc main() {}\n")
		mdFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "README.md", "# Test\n\nThis is a test file.\n")

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()

		require.NoError(t, os.Chdir(tempDir))

		// Test that the function completes without error for valid files
		logger := common.NewLogger("test")
		err = Enforce(logger, []string{goFile, mdFile})
		require.NoError(t, err, "Should not return error for valid files")
	})
	t.Run("file filtering - include patterns", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files with different extensions
		goFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.go", "package main\n")
		txtFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.txt", "text content")
		binaryFile := cryptoutilTestutil.WriteTempFile(t, tempDir, "test.bin", string([]byte{0x00, 0x01, 0x02}))

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()

		require.NoError(t, os.Chdir(tempDir))

		// Test that only .go and .txt files are checked (binary should be excluded)
		logger := common.NewLogger("test")
		err = Enforce(logger, []string{goFile, txtFile, binaryFile})
		require.NoError(t, err, "Should not return error when only valid files are checked")
	})
	t.Run("file filtering - exclude patterns", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files including one that should be excluded
		_ = cryptoutilTestutil.WriteTempFile(t, tempDir, "test.go", "package main\n")
		_ = cryptoutilTestutil.WriteTempFile(t, tempDir, "generated_gen.go", "package main\n")

		// Create vendor directory and file
		vendorDir := filepath.Join(tempDir, "vendor")
		require.NoError(t, os.MkdirAll(vendorDir, cryptoutilMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		cryptoutilTestutil.WriteTempFile(t, vendorDir, "lib.go", "package lib\n")

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()

		require.NoError(t, os.Chdir(tempDir))

		// Test that generated and vendor files are excluded
		logger := common.NewLogger("test")
		err = Enforce(logger, []string{filepath.Join(".", "test.go"), filepath.Join(".", "generated_gen.go"), filepath.Join(".", "vendor", "lib.go")})
		require.NoError(t, err, "Should not return error when excluded files are properly filtered")
	})
}

// TestFilterTextFiles_EdgeCases tests edge cases in filterTextFiles helper function.
func TestFilterTextFiles_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		wantCount int
	}{
		{
			name:      "empty file list",
			files:     []string{},
			wantCount: 0,
		},
		{
			name: "all excluded files",
			files: []string{
				"vendor/some/file.go",
				".git/config",
				"node_modules/package/index.js",
			},
			wantCount: 0,
		},
		{
			name: "mixed included and excluded",
			files: []string{
				"main.go",
				"vendor/dep.go",
				"README.md",
				".git/config",
			},
			wantCount: 2, // main.go and README.md
		},
		{
			name: "generated files excluded",
			files: []string{
				"normal.go",
				"generated_gen.go",
				"proto.pb.go",
			},
			wantCount: 1, // Only normal.go
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTextFiles(tt.files)
			require.Equal(t, tt.wantCount, len(result), "Unexpected number of filtered files")
		})
	}
}
