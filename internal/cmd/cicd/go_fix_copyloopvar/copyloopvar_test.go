package go_fix_copyloopvar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
)

func TestFix_EmptyDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-empty")

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_OldGoVersion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-old-version")

	// Create a Go file with loop variable copy.
	goFile := filepath.Join(tmpDir, "loop.go")
	content := `package test

func Process(items []int) {
	for _, item := range items {
		item := item
		_ = item
	}
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	// Test with Go 1.21 (below minimum).
	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.21.0")
	require.NoError(t, err)
	require.Equal(t, 0, processed) // Should skip processing.
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_NoLoopVariableCopies(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-no-copies")

	goFile := filepath.Join(tmpDir, "clean.go")
	content := `package test

func Process(items []int) {
	for _, item := range items {
		println(item)
	}
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_SingleLoopVariableCopy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-single")

	goFile := filepath.Join(tmpDir, "loop.go")
	content := `package test

func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, issuesFixed)

	// Verify the fix.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.NotContains(t, string(fixed), "item := item")
	require.Contains(t, string(fixed), "println(item)")
}

func TestFix_MultipleLoopVariableCopies(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-multiple")

	goFile := filepath.Join(tmpDir, "loops.go")
	content := `package test

func Process(items []int, names []string) {
	for _, item := range items {
		item := item
		println(item)
	}

	for _, name := range names {
		name := name
		println(name)
	}
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 2, issuesFixed)

	// Verify the fix.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.NotContains(t, string(fixed), "item := item")
	require.NotContains(t, string(fixed), "name := name")
}

func TestFix_KeyAndValueCopies(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-key-value")

	goFile := filepath.Join(tmpDir, "map_loop.go")
	content := `package test

func Process(data map[string]int) {
	for key, val := range data {
		key := key
		val := val
		println(key, val)
	}
}
`
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, issuesFixed) // Only the first copy (key := key) removed.

	// Verify the fix.
	fixed, err := os.ReadFile(goFile)
	require.NoError(t, err)
	require.NotContains(t, string(fixed), "key := key")
}

func TestFix_TestFilesSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-skip-test")

	testFile := filepath.Join(tmpDir, "loop_test.go")
	content := `package test

func TestLoop(t *testing.T) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			println(tc.name)
		})
	}
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 0, processed) // Test files should be skipped.
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_GeneratedFilesSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-skip-gen")

	genFile := filepath.Join(tmpDir, "openapi_gen_model.go")
	content := `package model

func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	require.NoError(t, os.WriteFile(genFile, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 0, processed) // Generated files should be skipped.
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestFix_NestedDirectories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := cryptoutilCmd.NewLogger("test-copyloopvar-nested")

	// Create nested directory structure.
	subDir := filepath.Join(tmpDir, "sub", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	content := `package test
func Process(items []int) {
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	file1 := filepath.Join(tmpDir, "loop1.go")
	file2 := filepath.Join(tmpDir, "sub", "loop2.go")
	file3 := filepath.Join(subDir, "loop3.go")

	require.NoError(t, os.WriteFile(file1, []byte(content), 0o600))
	require.NoError(t, os.WriteFile(file2, []byte(content), 0o600))
	require.NoError(t, os.WriteFile(file3, []byte(content), 0o600))

	processed, modified, issuesFixed, err := Fix(logger, tmpDir, "1.25.4")
	require.NoError(t, err)
	require.Equal(t, 3, processed)
	require.Equal(t, 3, modified)
	require.Equal(t, 3, issuesFixed)
}

func TestFix_InvalidDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmd.NewLogger("test-copyloopvar-invalid")

	processed, modified, issuesFixed, err := Fix(logger, "/nonexistent/path", "1.25.4")
	require.Error(t, err)
	require.Equal(t, 0, processed)
	require.Equal(t, 0, modified)
	require.Equal(t, 0, issuesFixed)
}

func TestIsGoVersionSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version  string
		expected bool
	}{
		{"1.21.0", false},
		{"1.22.0", true},
		{"1.22.5", true},
		{"1.23.0", true},
		{"1.25.4", true},
		{"2.0.0", true},
		{"invalid", false},
		{"1.2", false}, // Edge case: 1.2 < 1.22.
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			result := isGoVersionSupported(tc.version)
			require.Equal(t, tc.expected, result)
		})
	}
}
