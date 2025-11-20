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
)

func TestCheckFileEncoding(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid UTF-8 file without BOM", func(t *testing.T) {
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "valid_utf8.txt", "Hello, world! This is valid UTF-8 content without BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 0, "Valid UTF-8 file without BOM should have no issues")
		require.Empty(t, issues, "Valid UTF-8 file without BOM should have no encoding issues")
	})

	t.Run("file with UTF-8 BOM", func(t *testing.T) {
		// UTF-8 BOM: EF BB BF
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "utf8_bom.txt", "\xEF\xBB\xBFHello, world! This has UTF-8 BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-8 BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-8 BOM", "Issue should mention UTF-8 BOM")
	})

	t.Run("file with UTF-16 LE BOM", func(t *testing.T) {
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "utf16_le_bom.txt", "\xFF\xFEHello, world! This has UTF-16 LE BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-16 LE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-16 LE BOM", "Issue should mention UTF-16 LE BOM")
	})

	t.Run("file with UTF-16 BE BOM", func(t *testing.T) {
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "utf16_be_bom.txt", "\xFE\xFFHello, world! This has UTF-16 BE BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-16 BE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-16 BE BOM", "Issue should mention UTF-16 BE BOM")
	})

	t.Run("file with UTF-32 LE BOM", func(t *testing.T) {
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "utf32_le_bom.txt", "\xFF\xFE\x00\x00Hello, world! This has UTF-32 LE BOM.")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with UTF-32 LE BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-32 LE BOM", "Issue should mention UTF-32 LE BOM")
	})

	t.Run("file with UTF-32 BE BOM", func(t *testing.T) {
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "utf32_be_bom.txt", "\x00\x00\xFE\xFFHello, world! This has UTF-32 BE BOM.")

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
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "empty.txt", "")

		issues := checkFileEncoding(filePath)
		require.Empty(t, issues, "Empty file should have no encoding issues")
	})

	t.Run("file with only BOM", func(t *testing.T) {
		filePath := cryptoutilTestutil.WriteTempFile(t, tempDir, "only_bom.txt", "\xEF\xBB\xBF")

		issues := checkFileEncoding(filePath)
		require.Len(t, issues, 1, "File with only BOM should have one issue")
		require.Contains(t, issues[0], "contains UTF-8 BOM", "Issue should mention UTF-8 BOM")
	})
}

func TestAllEnforceUtf8(t *testing.T) {
	t.Run("no files to check", func(t *testing.T) {
		// Test with empty file list
		logger := common.NewLogger("test")
		err := allEnforceUtf8(logger, []string{})
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

		// Test that allEnforceUtf8 returns an error for encoding violations
		logger := common.NewLogger("test")
		err = allEnforceUtf8(logger, []string{invalidFile})
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
		err = allEnforceUtf8(logger, []string{goFile, mdFile})
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
		err = allEnforceUtf8(logger, []string{goFile, txtFile, binaryFile})
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
		err = allEnforceUtf8(logger, []string{filepath.Join(".", "test.go"), filepath.Join(".", "generated_gen.go"), filepath.Join(".", "vendor", "lib.go")})
		require.NoError(t, err, "Should not return error when excluded files are properly filtered")
	})
}
