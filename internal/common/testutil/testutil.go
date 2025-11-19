// Copyright (c) 2025 Justin Cranford
//
//

package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// WriteTempFile is a helper function for creating temporary test files.
func WriteTempFile(t *testing.T, tempDir, filename, content string) string {
	t.Helper()

	filePath := filepath.Join(tempDir, filename)
	WriteTestFile(t, filePath, content)

	return filePath
}

// WriteTestFile is a helper function for creating test files with content.
func WriteTestFile(t *testing.T, filePath, content string) {
	t.Helper()

	err := os.WriteFile(filePath, []byte(content), cryptoutilMagic.CacheFilePermissions)
	require.NoError(t, err)
}

// ReadTestFile is a helper function for reading test files with content.
func ReadTestFile(t *testing.T, filePath string) []byte {
	t.Helper()

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	return content
}
