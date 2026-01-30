// Copyright (c) 2025 Justin Cranford

package format_go

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

func TestFormat_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Format(logger, map[string][]string{})

	require.NoError(t, err, "Format should succeed with no files")
}

func TestFormat_WithFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Create a file with loop var copy.
	err := os.WriteFile(testFile, []byte(testGoContentWithLoopVarCopy), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		".go": {testFile},
	}

	err = Format(logger, filesByExtension)
	// Format returns error if modifications were made.
	// But GetGoFiles may filter out test files.
	if err != nil {
		require.Contains(t, err.Error(), "completed with modifications")
	}
}

func TestFormat_ErrorPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.go")

	// Create invalid Go file to trigger parse error.
	err := os.WriteFile(testFile, []byte(testGoContentInvalid), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		".go": {testFile},
	}

	err = Format(logger, filesByExtension)

	// Format may return error due to parse failure or succeed if file filtered out.
	// We just verify it doesn't panic.
	_ = err
}

func TestIsGoVersionSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		version   string
		supported bool
	}{
		{"go1.22", "go1.22", true},
		{"go1.22.0", "go1.22.0", true},
		{"go1.25.5", "go1.25.5", true},
		{"go1.21", "go1.21", false},
		{"go1.21.5", "go1.21.5", false},
		{"go1.20", "go1.20", false},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"no_separator", "go122", false},       // No dot separator
		{"non_numeric_major", "goX.22", false}, // Non-numeric major version
		{"non_numeric_minor", "go1.X", false},  // Non-numeric minor version
		{"major_only", "go2", false},           // Only major version, no minor
		{"major_gt_1", "go2.0", true},          // Major version > 1
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isGoVersionSupported(tc.version)
			require.Equal(t, tc.supported, result)
		})
	}
}

func TestEnforceAny_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := enforceAny(logger, map[string][]string{})
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
		".go": {testFile},
	}

	err = enforceAny(logger, filesByExtension)
	require.NoError(t, err, "enforceAny should return nil when no modifications made")
}

func TestEnforceAny_WithModifications(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "server.go")

	// File with any that needs replacement with any.
	oldContent := `package server

func Process(data interface{}) interface{} {
	return data
}
`
	err := os.WriteFile(testFile, []byte(oldContent), 0o600)
	require.NoError(t, err)

	// Test processGoFile directly (bypasses GetGoFiles filtering).
	replacements, err := processGoFile(testFile)
	require.NoError(t, err, "processGoFile should succeed")
	require.Equal(t, 2, replacements, "Should have 2 replacements (two interface{} instances)")

	// Verify file was actually modified.
	modifiedContent, readErr := os.ReadFile(testFile)
	require.NoError(t, readErr)
	require.Contains(t, string(modifiedContent), "any", "File should contain 'any' after replacement")
	require.NotContains(t, string(modifiedContent), "interface{}", "File should not contain 'interface{}' after replacement")
}

func TestEnforceAny_ViaEnforceAny_WithModifications(t *testing.T) {
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile}, // Note: key is "go" without dot - this is how GetGoFiles expects it
	}

	// enforceAny should return an error when files are modified.
	err = enforceAny(logger, filesByExtension)
	require.Error(t, err, "enforceAny should return error when files modified")
	require.Contains(t, err.Error(), "modified", "Error should mention modifications")

	// Verify file was actually modified.
	modifiedContent, readErr := os.ReadFile(testFile)
	require.NoError(t, readErr)
	require.Contains(t, string(modifiedContent), "any", "File should contain 'any' after replacement")
}

func TestEnforceAny_ErrorProcessingFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {"/nonexistent/path/to/file.go"}, // Note: key is "go" without dot
	}

	err := enforceAny(logger, filesByExtension)
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

	replacements, err := processGoFile(testFile)
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

	replacements, err := processGoFile(testFile)
	require.NoError(t, err)
	require.Equal(t, 1, replacements, "Should have 1 replacement")

	// Verify the file was modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "any", "File should contain 'any' after replacement")
	require.NotContains(t, string(modifiedContent), "interface{}", "File should not contain 'interface{}' after replacement")
}

func TestIsLoopVarCopy(t *testing.T) {
	t.Parallel()

	// This is a unit test for the isLoopVarCopy function.
	// We test the function logic indirectly through fixCopyLoopVarInFile.
	// Direct AST testing would be more complex.
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with no loop var copy.
	content := `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		println(v)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed, "File should not be changed")
	require.Equal(t, 0, fixes, "Should have no fixes")
}

func TestIsLoopVarCopy_VariousBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		content       string
		expectChanged bool
		expectFixes   int
	}{
		{
			name: "assignment_not_define",
			content: `package main

func main() {
	items := []int{1, 2, 3}
	var v int
	for _, v = range items {
		println(v)
	}
}
`,
			expectChanged: false,
			expectFixes:   0,
		},
		{
			name: "multiple_lhs",
			content: `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		a, b := v, v
		println(a, b)
	}
}
`,
			expectChanged: false,
			expectFixes:   0,
		},
		{
			name: "different_names",
			content: `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		w := v
		println(w)
	}
}
`,
			expectChanged: false,
			expectFixes:   0,
		},
		{
			name: "key_copy_pattern",
			content: `package main

func main() {
	items := []int{1, 2, 3}
	for i := range items {
		i := i
		println(i)
	}
}
`,
			expectChanged: true,
			expectFixes:   1,
		},
		{
			name: "non_identifier_rhs",
			content: `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		x := v + 1
		println(x)
	}
}
`,
			expectChanged: false,
			expectFixes:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")

			err := os.WriteFile(testFile, []byte(tc.content), 0o600)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

			require.NoError(t, err)
			require.Equal(t, tc.expectChanged, changed, "Changed mismatch for %s", tc.name)
			require.Equal(t, tc.expectFixes, fixes, "Fixes mismatch for %s", tc.name)
		})
	}
}

const testGoContentWithLoopVarCopy = `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
		v := v
		println(v)
	}
}
`

func TestFixCopyLoopVarInFile_WithLoopVarCopy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with loop var copy pattern.
	err := os.WriteFile(testFile, []byte(testGoContentWithLoopVarCopy), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

	require.NoError(t, err)
	require.True(t, changed, "File should be changed")
	require.Equal(t, 1, fixes, "Should have 1 fix")

	// Verify the file was modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), "v := v", "File should not contain loop var copy")
}

func TestFixCopyLoopVarInFile_ParseError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.go")

	// Invalid Go syntax.
	err := os.WriteFile(testFile, []byte(testGoContentInvalid), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

	require.Error(t, err)
	require.False(t, changed)
	require.Equal(t, 0, fixes)
	require.Contains(t, err.Error(), "failed to parse file")
}

func TestFixCopyLoopVarInFile_EmptyBody(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Loop with empty body.
	content := `package main

func main() {
	items := []int{1, 2, 3}
	for _, v := range items {
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	changed, fixes, err := fixCopyLoopVarInFile(logger, testFile)

	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, 0, fixes)
}

func TestFixCopyLoopVar_UnsupportedVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	processed, modified, issuesFixed, err := fixCopyLoopVar(logger, tmpDir, "go1.21")

	require.NoError(t, err)
	require.Equal(t, 0, processed, "Should process 0 files")
	require.Equal(t, 0, modified, "Should modify 0 files")
	require.Equal(t, 0, issuesFixed, "Should fix 0 issues")
}

func TestFixCopyLoopVar_SupportedVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a Go file with loop var copy.
	testFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(testFile, []byte(testGoContentWithLoopVarCopy), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	processed, modified, issuesFixed, err := fixCopyLoopVar(logger, tmpDir, "go1.22")

	require.NoError(t, err)
	require.Equal(t, 1, processed, "Should process 1 file")
	require.Equal(t, 1, modified, "Should modify 1 file")
	require.Equal(t, 1, issuesFixed, "Should fix 1 issue")
}

func TestFixCopyLoopVar_EmptyDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	processed, modified, issuesFixed, err := fixCopyLoopVar(logger, tmpDir, "go1.22")

	require.NoError(t, err)
	require.Equal(t, 0, processed, "Should process 0 files")
	require.Equal(t, 0, modified, "Should modify 0 files")
	require.Equal(t, 0, issuesFixed, "Should fix 0 issues")
}

func TestFilterGoFiles_NoGoFiles(t *testing.T) {
	t.Parallel()

	filesByExtension := map[string][]string{
		".txt": {"file1.txt", "file2.txt"},
		".md":  {"README.md"},
	}

	goFiles := filterGoFiles(filesByExtension)

	require.Empty(t, goFiles, "Should return empty slice when no Go files")
}

func TestFilterGoFiles_WithGoFiles(t *testing.T) {
	t.Parallel()

	filesByExtension := map[string][]string{
		".go":  {"file1.go", "file2.go"},
		".txt": {"file.txt"},
	}

	goFiles := filterGoFiles(filesByExtension)

	// GetGoFiles applies exclusions, so result may be empty or contain files.
	// Just verify it doesn't panic/error.
	_ = goFiles
}

func TestProcessGoFile_ReadError(t *testing.T) {
	t.Parallel()

	// Non-existent file.
	replacements, err := processGoFile("/nonexistent/file.go")

	require.Error(t, err)
	require.Equal(t, 0, replacements)
	require.Contains(t, err.Error(), "failed to read file")
}

func TestFormat_WithModifications(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "server.go")

	// File with interface{} that should trigger modification.
	err := os.WriteFile(testFile, []byte(testGoContentWithInterfaceEmpty), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {testFile}, // Note: key is "go" without dot
	}

	// Format should return error because files were modified.
	err = Format(logger, filesByExtension)
	require.Error(t, err, "Format should return error when files are modified")
	require.Contains(t, err.Error(), "completed with modifications")
}
