package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"cryptoutil/internal/cmd/cicd/common"
)

func TestGoFixCopyLoopVar_NoGoFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_NoGoFiles")
	files := []string{"README.md", "Dockerfile", "config.yml"}

	err := goFixCopyLoopVar(logger, files)
	require.NoError(t, err, "Should succeed with no Go files")
}

func TestGoFixCopyLoopVar_NoLoopVarCopies(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_NoLoopVarCopies")
	tempDir := t.TempDir()

	// Create a Go file with no loop variable copies
	testFile := filepath.Join(tempDir, "test.go")
	content := `package main

func main() {
	items := []int{1, 2, 3}
	for i, item := range items {
		println(i, item)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixCopyLoopVar(logger, []string{testFile})
	require.NoError(t, err, "Should succeed with no copies to fix")
}

func TestGoFixCopyLoopVar_BasicFix(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_BasicFix")
	tempDir := t.TempDir()

	// Create a Go file with loop variable copy
	testFile := filepath.Join(tempDir, "test.go")
	content := `package main

func main() {
	items := []string{"a", "b", "c"}
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixCopyLoopVar(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 1 loop variable copies")

	// Verify the file was modified
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), "item := item", "Loop variable copy should be removed")
}

func TestGoFixCopyLoopVar_MultipleVars(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_MultipleVars")
	tempDir := t.TempDir()

	// Create a Go file with multiple loop variable copies
	testFile := filepath.Join(tempDir, "test.go")
	content := `package main

func main() {
	items := []string{"a", "b", "c"}
	for i, item := range items {
		i := i
		item := item
		println(i, item)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixCopyLoopVar(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 2 loop variable copies")

	// Verify both copies were removed
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), "i := i", "Loop variable copy should be removed")
	require.NotContains(t, string(modifiedContent), "item := item", "Loop variable copy should be removed")
}

func TestGoFixCopyLoopVar_PreserveOtherAssignments(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_PreserveOtherAssignments")
	tempDir := t.TempDir()

	// Create a Go file with loop variable copy AND other assignments
	testFile := filepath.Join(tempDir, "test.go")
	content := `package main

func main() {
	items := []string{"a", "b", "c"}
	for _, item := range items {
		item := item
		value := item + "_suffix"
		println(value)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixCopyLoopVar(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 1 loop variable copies")

	// Verify only the copy was removed, not the other assignment
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), "item := item", "Loop variable copy should be removed")
	require.Contains(t, string(modifiedContent), "value := item + \"_suffix\"", "Other assignment should be preserved")
}

func TestGoFixCopyLoopVar_MultipleLoops(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_MultipleLoops")
	tempDir := t.TempDir()

	// Create a Go file with multiple loops
	testFile := filepath.Join(tempDir, "test.go")
	content := `package main

func main() {
	items := []string{"a", "b", "c"}
	for _, item := range items {
		item := item
		println(item)
	}

	numbers := []int{1, 2, 3}
	for _, num := range numbers {
		num := num
		println(num)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixCopyLoopVar(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 2 loop variable copies")

	// Verify both loop variable copies were removed
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NotContains(t, string(modifiedContent), "item := item", "Loop variable copy should be removed")
	require.NotContains(t, string(modifiedContent), "num := num", "Loop variable copy should be removed")
}

func TestGoFixCopyLoopVar_ExcludesGenerated(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_ExcludesGenerated")
	tempDir := t.TempDir()

	// Create generated Go files
	genFile1 := filepath.Join(tempDir, "openapi_gen.go")
	genFile2 := filepath.Join(tempDir, "proto.pb.go")

	content := `package main

func main() {
	items := []string{"a", "b", "c"}
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	err := os.WriteFile(genFile1, []byte(content), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(genFile2, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixCopyLoopVar(logger, []string{genFile1, genFile2})
	require.NoError(t, err, "Should skip generated files")
}

func TestGoFixCopyLoopVar_MultipleFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixCopyLoopVar_MultipleFiles")
	tempDir := t.TempDir()

	// Create multiple Go files
	file1 := filepath.Join(tempDir, "file1.go")
	file2 := filepath.Join(tempDir, "file2.go")
	file3 := filepath.Join(tempDir, "file3.go") // No copies

	content1 := `package main

func test1() {
	items := []string{"a", "b"}
	for _, item := range items {
		item := item
		println(item)
	}
}
`
	content2 := `package main

func test2() {
	numbers := []int{1, 2}
	for _, num := range numbers {
		num := num
		println(num)
	}
}
`
	content3 := `package main

func test3() {
	items := []string{"x", "y"}
	for _, item := range items {
		println(item)
	}
}
`
	err := os.WriteFile(file1, []byte(content1), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content2), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(file3, []byte(content3), 0o600)
	require.NoError(t, err)

	err = goFixCopyLoopVar(logger, []string{file1, file2, file3})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "fixed 2 loop variable copies")
}

func TestIsGo125OrHigher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version  string
		expected bool
		name     string
	}{
		{"1.25.4", true, "Go 1.25.4"},
		{"1.25.0", true, "Go 1.25.0"},
		{"1.25", true, "Go 1.25"},
		{"1.26", true, "Go 1.26"},
		{"2.0", true, "Go 2.0"},
		{"1.24", false, "Go 1.24"},
		{"1.22.5", false, "Go 1.22.5"},
		{"1.20", false, "Go 1.20"},
		{"invalid", false, "Invalid version"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := isGo125OrHigher(tc.version)
			require.Equal(t, tc.expected, result, "Version check mismatch for %s", tc.version)
		})
	}
}
