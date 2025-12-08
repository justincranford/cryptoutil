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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isGoVersionSupported(tc.version)
			require.Equal(t, tc.supported, result)
		})
	}
}

func TestProcessGoFile_NoChanges(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with no interface{}.
	content := `package main

func main() {
	var x any = 42
	println(x)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	replacements, err := processGoFile(testFile)
	require.NoError(t, err)
	require.Equal(t, 0, replacements, "Should have no replacements")
}

func TestProcessGoFile_WithChanges(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// File with interface{} that should be replaced.
	// Using a special marker to avoid self-modification during linting.
	content := "package main\n\nfunc main() {\n\tvar x interface{} = 42\n\tprintln(x)\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	replacements, err := processGoFile(testFile)
	require.NoError(t, err)
	require.Equal(t, 1, replacements, "Should have 1 replacement")

	// Verify the file was modified.
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "any", "File should contain 'any'")
	require.NotContains(t, string(modifiedContent), "interface{}", "File should not contain 'interface{}'")
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
