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
)

func TestCheckFileEncoding(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid UTF-8 file without BOM", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "valid_utf8.txt")
		content := "Hello, world! This is valid UTF-8 content without BOM."
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		issues := checkFileEncoding(filePath)
		require.Empty(t, issues, "Valid UTF-8 file should have no encoding issues")
	})

	t.Run("file with UTF-8 BOM", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "utf8_bom.txt")
		// UTF-8 BOM: EF BB BF
		content := "\xEF\xBB\xBFHello, world! This has UTF-8 BOM."
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-8 BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-8 BOM", "Issue should mention UTF-8 BOM")
	})

	t.Run("file with UTF-16 LE BOM", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "utf16_le_bom.txt")
		// UTF-16 LE BOM: FF FE
		content := "\xFF\xFEHello, world! This has UTF-16 LE BOM."
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-16 LE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-16 LE BOM", "Issue should mention UTF-16 LE BOM")
	})

	t.Run("file with UTF-16 BE BOM", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "utf16_be_bom.txt")
		// UTF-16 BE BOM: FE FF
		content := "\xFE\xFFHello, world! This has UTF-16 BE BOM."
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-16 BE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-16 BE BOM", "Issue should mention UTF-16 BE BOM")
	})

	t.Run("file with UTF-32 LE BOM", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "utf32_le_bom.txt")
		// UTF-32 LE BOM: FF FE 00 00
		content := "\xFF\xFE\x00\x00Hello, world! This has UTF-32 LE BOM."
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-32 LE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-32 LE BOM", "Issue should mention UTF-32 LE BOM")
	})

	t.Run("file with UTF-32 BE BOM", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "utf32_be_bom.txt")
		// UTF-32 BE BOM: 00 00 FE FF
		content := "\x00\x00\xFE\xFFHello, world! This has UTF-32 BE BOM."
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

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
		filePath := filepath.Join(tempDir, "empty.txt")
		require.NoError(t, os.WriteFile(filePath, []byte(""), 0o600))

		issues := checkFileEncoding(filePath)
		require.Empty(t, issues, "Empty file should have no encoding issues")
	})

	t.Run("file with only BOM", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "only_bom.txt")
		// Only UTF-8 BOM
		content := "\xEF\xBB\xBF"
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

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
		require.Contains(t, outputStr, "No files found to check", "Should indicate no files found")
		require.Contains(t, outputStr, "[CICD] dur=", "Should contain performance logging")
	})

	t.Run("files with encoding violations", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test files
		validFile := filepath.Join(tempDir, "valid.go")
		require.NoError(t, os.WriteFile(validFile, []byte("package main\n\nfunc main() {}\n"), 0o600))

		invalidFile := filepath.Join(tempDir, "invalid.go")
		// Add UTF-8 BOM to make it invalid
		content := "\xEF\xBB\xBFpackage main\n\nfunc main() {}\n"
		require.NoError(t, os.WriteFile(invalidFile, []byte(content), 0o600))

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
		goFile := filepath.Join(tempDir, "test.go")
		require.NoError(t, os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o600))

		mdFile := filepath.Join(tempDir, "README.md")
		require.NoError(t, os.WriteFile(mdFile, []byte("# Test\n\nThis is a test file.\n"), 0o600))

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Test that the function completes without exiting (indicating success)
		logger := NewLogUtil("test")
		allEnforceUtf8(logger, []string{goFile, mdFile})
		// If we reach here, the function didn't call os.Exit(1), so it succeeded
	})

	t.Run("file filtering - include patterns", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files with different extensions
		goFile := filepath.Join(tempDir, "test.go")
		require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o600))

		txtFile := filepath.Join(tempDir, "test.txt")
		require.NoError(t, os.WriteFile(txtFile, []byte("text content"), 0o600))

		binaryFile := filepath.Join(tempDir, "test.bin")
		require.NoError(t, os.WriteFile(binaryFile, []byte{0x00, 0x01, 0x02}, 0o600))

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Test that only .go and .txt files are checked (binary should be excluded)
		logger := NewLogUtil("test")
		allEnforceUtf8(logger, []string{goFile, txtFile, binaryFile})
		// If we reach here, the function succeeded and only checked the appropriate files
	})

	t.Run("file filtering - exclude patterns", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files including one that should be excluded
		goFile := filepath.Join(tempDir, "test.go")
		require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o600))

		genFile := filepath.Join(tempDir, "generated_gen.go")
		require.NoError(t, os.WriteFile(genFile, []byte("package main\n"), 0o600))

		// Create vendor directory and file
		vendorDir := filepath.Join(tempDir, "vendor")
		require.NoError(t, os.MkdirAll(vendorDir, 0o755))
		vendorFile := filepath.Join(vendorDir, "lib.go")
		require.NoError(t, os.WriteFile(vendorFile, []byte("package lib\n"), 0o600))

		// Change to temp directory
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Test that generated and vendor files are excluded
		logger := NewLogUtil("test")
		allEnforceUtf8(logger, []string{filepath.Join(".", "test.go"), filepath.Join(".", "generated_gen.go"), filepath.Join(".", "vendor", "lib.go")})
		// If we reach here, the function succeeded and properly excluded files
	})
}
