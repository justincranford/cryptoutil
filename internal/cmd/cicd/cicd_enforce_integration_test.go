// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
)

// TestAllEnforceUtf8_ValidFiles tests wrapper with valid UTF-8 files.
func TestAllEnforceUtf8_ValidFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-utf8-valid")

	// Create valid UTF-8 file.
	validFile := filepath.Join(tmpDir, "valid.txt")
	require.NoError(t, os.WriteFile(validFile, []byte("Hello, World! 你好世界"), 0o600))

	// Call wrapper.
	files := []string{validFile}
	err := allEnforceUtf8(logger, files)
	require.NoError(t, err)
}

// TestAllEnforceUtf8_NoFiles tests wrapper with empty file list.
func TestAllEnforceUtf8_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-utf8-no-files")

	// Call wrapper with empty file list.
	err := allEnforceUtf8(logger, []string{})
	require.NoError(t, err)
}

// TestAllEnforceUtf8_InvalidUtf8 tests wrapper with invalid UTF-8 file.
func TestAllEnforceUtf8_InvalidUtf8(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-utf8-invalid")

	// Create file with invalid UTF-8 bytes.
	invalidFile := filepath.Join(tmpDir, "invalid.txt")
	invalidBytes := []byte{0xFF, 0xFE, 0xFD} // Invalid UTF-8 sequence.
	require.NoError(t, os.WriteFile(invalidFile, invalidBytes, 0o600))

	// Call wrapper.
	files := []string{invalidFile}
	err := allEnforceUtf8(logger, files)
	require.Error(t, err)
	require.Contains(t, err.Error(), "file encoding violations found")
}

// TestGoEnforceAny_ValidFiles tests wrapper with files using 'any' type.
func TestGoEnforceAny_ValidFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-any-valid")

	// Create Go file using 'any' type (valid).
	validFile := filepath.Join(tmpDir, "valid.go")
	validContent := `package test

func Process(data any) {
	_ = data
}
`
	require.NoError(t, os.WriteFile(validFile, []byte(validContent), 0o600))

	// Call wrapper.
	files := []string{validFile}
	err := goEnforceAny(logger, files)
	require.NoError(t, err)
}

// TestGoEnforceAny_NoFiles tests wrapper with empty file list.
func TestGoEnforceAny_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-any-no-files")

	// Call wrapper with empty file list.
	err := goEnforceAny(logger, []string{})
	require.NoError(t, err)
}

// TestGoEnforceAny_InterfaceEmpty tests wrapper with interface{} usage.
func TestGoEnforceAny_InterfaceEmpty(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-any-interface-empty")

	// Create Go file using interface{} (invalid - should use 'any').
	invalidFile := filepath.Join(tmpDir, "invalid.go")
	invalidContent := `package test

func Process(data interface{}) {
	_ = data
}
`
	require.NoError(t, os.WriteFile(invalidFile, []byte(invalidContent), 0o600))

	// Call wrapper.
	files := []string{invalidFile}
	err := goEnforceAny(logger, files)
	require.Error(t, err)
	require.Contains(t, err.Error(), "modified 1 files with 1 total replacements")
}
