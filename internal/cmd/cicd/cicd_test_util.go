// Package cicd provides test utilities for CI/CD quality control checks.
package cicd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// writeTempFile is a helper function for creating temporary test files.
func writeTempFile(t *testing.T, tempDir, filename, content string) string {
	t.Helper()

	filePath := filepath.Join(tempDir, filename)
	writeTestFile(t, filePath, content)

	return filePath
}

// writeTestFile is a helper function for creating test files with content.
func writeTestFile(t *testing.T, filePath, content string) {
	t.Helper()

	require.NoError(t, writeFile(filePath, content, cryptoutilMagic.CacheFilePermissions))
}
