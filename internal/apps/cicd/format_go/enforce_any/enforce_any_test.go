// Copyright (c) 2025 Justin Cranford

package enforce_any

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

func TestEnforceAny_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Enforce(logger, map[string][]string{})
	require.NoError(t, err, "enforceAny should succeed with no files")
}

func TestEnforceAny_NoModifications(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "clean.go")

	// File already using 'any' (no modifications needed).
	// Using the same constant as TestProcessGoFile_NoChanges.
	err := os.WriteFile(testFile, []byte(testGoContentClean), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = Enforce(logger, filesByExtension)
	require.NoError(t, err, "enforceAny should return nil when no modifications made")
}

func TestEnforceAny_WithModifications(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "server.go")

	// File with interface{} that needs replacement with any.
	oldContent := `package server

func Process(data interface{}) interface{} {
	return data
}
`
	err := os.WriteFile(testFile, []byte(oldContent), 0o600)
	require.NoError(t, err)

	// Test processGoFile directly (bypasses GetGoFiles filtering).
	replacements, err := ProcessGoFile(testFile)
	require.NoError(t, err, "processGoFile should succeed")
	require.Equal(t, 2, replacements, "Should have 2 replacements (two interface{} instances)")

	// Verify file was actually modified.
	modifiedContent, readErr := os.ReadFile(testFile)
	require.NoError(t, readErr)
	require.Contains(t, string(modifiedContent), "any", "File should contain 'any' after replacement")
	require.NotContains(t, string(modifiedContent), "interface{}", "File should not contain 'interface{}' after replacement")
}

func TestEnforceAny_ErrorProcessingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {"/nonexistent/path/to/file.go"},
	}

	err := Enforce(logger, filesByExtension)
	// Should continue processing and return nil (errors are logged, not returned).
	require.NoError(t, err, "enforceAny should continue after file errors")
}

func TestProcessGoFile_NoChanges(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with no any.
	err := os.WriteFile(testFile, []byte(testGoContentClean), 0o600)
	require.NoError(t, err)

	replacements, err := ProcessGoFile(testFile)
	require.NoError(t, err)
	require.Equal(t, 0, replacements, "Should have no replacements")
}

const (
	testGoContentClean              = "package main\n\nfunc main() {\n\tvar x any = 42\n\tprintln(x)\n}\n"
	testGoContentWithInterfaceEmpty = "package main\n\nfunc main() {\n\tvar x interface{} = 42\n\tprintln(x)\n}\n"
	testGoContentInvalid            = "package main\n\nfunc main() {\n\tthis is not valid go code\n}\n"
)

func TestProcessGoFile_WithChanges(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with any that should be replaced.
	// Using testGoContentWithInterfaceEmpty constant to avoid self-modification during linting.
	err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceEmpty), 0o600)
	require.NoError(t, err)

	replacements, err := ProcessGoFile(testFile)
	require.NoError(t, err)
	require.Equal(t, 1, replacements, "Should have 1 replacement")

	// Verify the file was modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "any", "File should contain 'any' after replacement")
	require.NotContains(t, string(modifiedContent), "interface{}", "File should not contain 'interface{}' after replacement")
}

func TestFilterGoFiles_NoGoFiles(t *testing.T) {
	t.Parallel()

	filesByExtension := map[string][]string{
		".txt": {"file1.txt", "file2.txt"},
		"md":   {"README.md"},
	}

	goFiles := FilterGoFiles(filesByExtension)

	require.Empty(t, goFiles, "Should return empty slice when no Go files")
}

func TestFilterGoFiles_WithGoFiles(t *testing.T) {
	t.Parallel()

	filesByExtension := map[string][]string{
		"go":   {"file1.go", "file2.go"},
		".txt": {"file.txt"},
	}

	goFiles := FilterGoFiles(filesByExtension)

	// GetGoFiles applies exclusions, so result may be empty or contain files.
	// Just verify it doesn't panic/error.
	_ = goFiles
}

func TestProcessGoFile_ReadError(t *testing.T) {
	t.Parallel()

	// Non-existent file.
	replacements, err := ProcessGoFile("/nonexistent/file.go")

	require.Error(t, err)
	require.Equal(t, 0, replacements)
	require.Contains(t, err.Error(), "failed to read file")
}

// testGoContentWithInterfaceBraces is Go content with interface{} in a temp file
// that will NOT match the format-go self-exclusion pattern (not in format_go dir).
// CRITICAL: Uses interface{} in string literal intentionally - this is test data, not production code.
const testGoContentWithInterfaceBraces = "package server\n\nfunc Handle(data interface{}) interface{} {\n\treturn data\n}\n"

// TestEnforceAny_WithModificationsViaEnforceAny calls enforceAny directly with a file
// containing interface{} to cover the filesModified > 0 code path.
func TestEnforceAny_WithModificationsViaEnforceAny(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "server.go")

	err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceBraces), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile},
	}

	err = Enforce(logger, filesByExtension)
	require.Error(t, err, "enforceAny should return error when files were modified")
	require.Contains(t, err.Error(), "modified")
}

// TestProcessGoFile_WriteError tests that processGoFile returns an error when
// the file has interface{} but cannot be written (read-only file).
func TestProcessGoFile_WriteError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "readonly.go")

	// Create file with interface{} that needs replacement.
	err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceBraces), 0o600)
	require.NoError(t, err)

	// Make file read-only so WriteFile will fail.
	err = os.Chmod(testFile, 0o400)
	require.NoError(t, err)

	replacements, err := ProcessGoFile(testFile)
	require.Error(t, err, "processGoFile should fail when file is read-only")
	require.Equal(t, 0, replacements, "Should return 0 replacements on error")
	require.Contains(t, err.Error(), "failed to write file")

	// Restore write permission for cleanup.
	_ = os.Chmod(testFile, 0o600)
}
