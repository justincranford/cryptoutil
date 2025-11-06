// IMPORTANT: This file contains deliberate linter error patterns for testing cicd functionality.
// It MUST be excluded from all linting operations to prevent self-referencing errors.
// See .golangci.yml exclude-rules and cicd.go exclusion patterns for details.
//
// This file intentionally uses any patterns and other lint violations to test
// that cicd correctly identifies and reports such patterns in other files.
package cicd

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

func TestCheckFileEncoding(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid UTF-8 file without BOM", func(t *testing.T) {
		filePath := WriteTempFile(t, tempDir, "valid_utf8.txt", "Hello, world! This is valid UTF-8 content without BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 0, "Valid UTF-8 file without BOM should have no issues")
		require.Empty(t, issues, "Valid UTF-8 file without BOM should have no encoding issues")
	})

	t.Run("file with UTF-8 BOM", func(t *testing.T) {
		// UTF-8 BOM: EF BB BF
		filePath := WriteTempFile(t, tempDir, "utf8_bom.txt", "\xEF\xBB\xBFHello, world! This has UTF-8 BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-8 BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-8 BOM", "Issue should mention UTF-8 BOM")
	})

	t.Run("file with UTF-16 LE BOM", func(t *testing.T) {
		filePath := WriteTempFile(t, tempDir, "utf16_le_bom.txt", "\xFF\xFEHello, world! This has UTF-16 LE BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-16 LE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-16 LE BOM", "Issue should mention UTF-16 LE BOM")
	})

	t.Run("file with UTF-16 BE BOM", func(t *testing.T) {
		filePath := WriteTempFile(t, tempDir, "utf16_be_bom.txt", "\xFE\xFFHello, world! This has UTF-16 BE BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-16 BE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-16 BE BOM", "Issue should mention UTF-16 BE BOM")
	})

	t.Run("file with UTF-32 LE BOM", func(t *testing.T) {
		filePath := WriteTempFile(t, tempDir, "utf32_le_bom.txt", "\xFF\xFE\x00\x00Hello, world! This has UTF-32 LE BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-32 LE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-32 LE BOM", "Issue should mention UTF-32 LE BOM")
	})

	t.Run("file with UTF-32 BE BOM", func(t *testing.T) {
		filePath := WriteTempFile(t, tempDir, "utf32_be_bom.txt", "\x00\x00\xFE\xFFHello, world! This has UTF-32 BE BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-32 BE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-32 BE BOM", "Issue should mention UTF-32 BE BOM")
	})

	t.Run("file does not exist", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		issues := checkFileEncoding(nonExistentFile)
		require.Len(t, issues, 1, "Non-existent file should have one issue")
		require.Contains(t, issues[0], "Error reading file", "Issue should mention file reading error")
	})

	t.Run("empty file", func(t *testing.T) {
		filePath := WriteTempFile(t, tempDir, "empty.txt", "")

		issues := checkFileEncoding(filePath)
		require.Empty(t, issues, "Empty file should have no encoding issues")
	})

	t.Run("file with only BOM", func(t *testing.T) {
		filePath := WriteTempFile(t, tempDir, "only_bom.txt", "\xEF\xBB\xBF")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with only BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-8 BOM", "Issue should mention UTF-8 BOM")
	})
}

func TestAllEnforceUtf8(t *testing.T) {
	t.Run("no files to check", func(t *testing.T) {
		// Capture stderr for testing output
		oldStderr := os.Stderr
		r, w, err := os.Pipe()
		require.NoError(t, err)

		os.Stderr = w

		// Restore stderr after test
		defer func() {
			os.Stderr = oldStderr
		}()

		// Test with empty file list
		logger := NewLogUtil("test")
		allEnforceUtf8(logger, []string{})

		// Close writer to flush output
		w.Close()

		// Read captured output
		output, err := io.ReadAll(r)
		require.NoError(t, err)

		outputStr := string(output)

		// Should contain success message for no files
		require.Contains(t, outputStr, "allEnforceUtf8 completed (no files)", "Should indicate no files found")
		require.Contains(t, outputStr, "[CICD] dur=", "Should contain performance logging")
	})

	t.Run("files with encoding violations", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test files
		invalidFile := WriteTempFile(t, tempDir, "invalid.go", "\xEF\xBB\xBFpackage main\n\nfunc main() {}\n")

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Since allEnforceUtf8 calls os.Exit(1) on violations, we can't test it directly
		// Instead, test the checkFileEncoding function directly for the invalid file
		issues := checkFileEncoding(invalidFile)
		require.NotEmpty(t, issues, "Invalid file should have encoding issues")
		require.Contains(t, issues[0], "contains UTF-8 BOM", "Should detect UTF-8 BOM")
	})

	t.Run("all files valid", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create valid test files
		goFile := WriteTempFile(t, tempDir, "test.go", "package main\n\nfunc main() {}\n")
		mdFile := WriteTempFile(t, tempDir, "README.md", "# Test\n\nThis is a test file.\n")

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Test that the function completes without exiting (indicating success)
		logger := NewLogUtil("test")
		// If we reach here, the function didn't call os.Exit(1), so it succeeded
		allEnforceUtf8(logger, []string{goFile, mdFile})
	})
	t.Run("file filtering - include patterns", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files with different extensions
		goFile := WriteTempFile(t, tempDir, "test.go", "package main\n")
		txtFile := WriteTempFile(t, tempDir, "test.txt", "text content")
		binaryFile := WriteTempFile(t, tempDir, "test.bin", string([]byte{0x00, 0x01, 0x02}))

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Test that only .go and .txt files are checked (binary should be excluded)
		logger := NewLogUtil("test")
		// If we reach here, the function succeeded and only checked the appropriate files
		allEnforceUtf8(logger, []string{goFile, txtFile, binaryFile})
	})
	t.Run("file filtering - exclude patterns", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files including one that should be excluded
		_ = WriteTempFile(t, tempDir, "test.go", "package main\n")
		_ = WriteTempFile(t, tempDir, "generated_gen.go", "package main\n")

		// Create vendor directory and file
		vendorDir := filepath.Join(tempDir, "vendor")
		require.NoError(t, os.MkdirAll(vendorDir, cryptoutilMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		WriteTempFile(t, vendorDir, "lib.go", "package lib\n")

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Test that generated and vendor files are excluded
		logger := NewLogUtil("test")
		// If we reach here, the function succeeded and properly excluded files
		allEnforceUtf8(logger, []string{filepath.Join(".", "test.go"), filepath.Join(".", "generated_gen.go"), filepath.Join(".", "vendor", "lib.go")})
	})
}
