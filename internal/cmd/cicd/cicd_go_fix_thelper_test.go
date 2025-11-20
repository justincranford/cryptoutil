package cicd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"cryptoutil/internal/cmd/cicd/common"
)

func TestGoFixTHelper_NoTestFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_NoTestFiles")
	files := []string{"main.go", "handler.go", "util.go"}

	err := goFixTHelper(logger, files)
	require.NoError(t, err, "Should succeed with no test files")
}

func TestGoFixTHelper_NoHelperFunctions(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_NoHelperFunctions")
	tempDir := t.TempDir()

	// Create a test file with no helper functions
	testFile := filepath.Join(tempDir, "main_test.go")
	content := `package main

import "testing"

func TestMain(t *testing.T) {
	result := 1 + 1
	if result != 2 {
		t.Errorf("expected 2, got %d", result)
	}
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixTHelper(logger, []string{testFile})
	require.NoError(t, err, "Should succeed with no helper functions")
}

func TestGoFixTHelper_AddsTHelperToSetup(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_AddsTHelperToSetup")
	tempDir := t.TempDir()

	// Create a test file with setup function missing t.Helper()
	testFile := filepath.Join(tempDir, "setup_test.go")
	content := `package main

import "testing"

func setupDatabase(t *testing.T) *Database {
	db := &Database{Name: "test"}
	return db
}

type Database struct {
	Name string
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixTHelper(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "added t.Helper() to 1 test helper functions")

	// Verify t.Helper() was added
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Contains(t, string(modifiedContent), "t.Helper()", "t.Helper() should be added")
}

func TestGoFixTHelper_PreservesExistingTHelper(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_PreservesExistingTHelper")
	tempDir := t.TempDir()

	// Create a test file with helper function that already has t.Helper()
	testFile := filepath.Join(tempDir, "helper_test.go")
	content := `package main

import "testing"

func helperFunction(t *testing.T) string {
	t.Helper()
	return "test"
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixTHelper(logger, []string{testFile})
	require.NoError(t, err, "Should succeed when t.Helper() already exists")

	// Verify only one t.Helper() exists
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	helperCount := strings.Count(string(modifiedContent), "t.Helper()")
	require.Equal(t, 1, helperCount, "Should have exactly one t.Helper() call")
}

func TestGoFixTHelper_MultipleHelperFunctions(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_MultipleHelperFunctions")
	tempDir := t.TempDir()

	// Create a test file with multiple helper functions
	testFile := filepath.Join(tempDir, "helpers_test.go")
	content := `package main

import "testing"

func setupTest(t *testing.T) {
	// Setup code
}

func checkResult(t *testing.T, result int) {
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func verifyState(t *testing.T) {
	// Verification code
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixTHelper(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "added t.Helper() to 3 test helper functions")

	// Verify t.Helper() was added to all helpers
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	helperCount := strings.Count(string(modifiedContent), "t.Helper()")
	require.Equal(t, 3, helperCount, "Should have three t.Helper() calls")
}

func TestGoFixTHelper_OnlyHelperPrefixes(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_OnlyHelperPrefixes")
	tempDir := t.TempDir()

	// Create a test file with functions that don't match helper patterns
	testFile := filepath.Join(tempDir, "test_test.go")
	content := `package main

import "testing"

func randomFunction(t *testing.T) {
	// Should not be modified - doesn't match helper pattern
}

func TestSomething(t *testing.T) {
	// Should not be modified - it's a test function
}

func helperDoSomething(t *testing.T) {
	// Should get Helper() call - matches "helper" prefix
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixTHelper(logger, []string{testFile})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "added t.Helper() to 1 test helper functions")

	// Verify only the helper function got Helper() call added
	modifiedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	helperCount := strings.Count(string(modifiedContent), ".Helper()")
	require.Equal(t, 1, helperCount, "Should have exactly one Helper() call")
}

func TestGoFixTHelper_ExcludesNonTestFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_ExcludesNonTestFiles")
	tempDir := t.TempDir()

	// Create a regular Go file with helper-like function
	regularFile := filepath.Join(tempDir, "helper.go")
	content := `package main

func setupDatabase(db *Database) {
	// Non-test file, should not be modified
}

type Database struct {
	Name string
}
`
	err := os.WriteFile(regularFile, []byte(content), 0o600)
	require.NoError(t, err)

	err = goFixTHelper(logger, []string{regularFile})
	require.NoError(t, err, "Should skip non-test files")

	// Verify file was not modified - should contain original content
	modifiedContent, err := os.ReadFile(regularFile)
	require.NoError(t, err)
	require.Equal(t, content, string(modifiedContent), "File content should be unchanged")
	require.NotContains(t, string(modifiedContent), ".Helper()", "Should not add Helper() call to non-test files")
}

func TestGoFixTHelper_MultipleFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("TestGoFixTHelper_MultipleFiles")
	tempDir := t.TempDir()

	// Create multiple test files
	file1 := filepath.Join(tempDir, "file1_test.go")
	file2 := filepath.Join(tempDir, "file2_test.go")
	file3 := filepath.Join(tempDir, "file3_test.go") // No helpers

	content1 := `package main

import "testing"

func setupFile1(t *testing.T) {
	// Setup
}
`
	content2 := `package main

import "testing"

func checkFile2(t *testing.T) {
	// Check
}
`
	content3 := `package main

import "testing"

func TestFile3(t *testing.T) {
	// Test
}
`
	err := os.WriteFile(file1, []byte(content1), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content2), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(file3, []byte(content3), 0o600)
	require.NoError(t, err)

	err = goFixTHelper(logger, []string{file1, file2, file3})
	require.Error(t, err, "Should return error when fixes are made")
	require.Contains(t, err.Error(), "added t.Helper() to 2 test helper functions")
}

func TestFilterTestFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    []string
		expected []string
	}{
		{
			name:     "Mixed files",
			files:    []string{"main.go", "main_test.go", "util.go", "util_test.go"},
			expected: []string{"main_test.go", "util_test.go"},
		},
		{
			name:     "Only test files",
			files:    []string{"test1_test.go", "test2_test.go"},
			expected: []string{"test1_test.go", "test2_test.go"},
		},
		{
			name:     "No test files",
			files:    []string{"main.go", "util.go"},
			expected: []string{},
		},
		{
			name:     "Empty list",
			files:    []string{},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := filterTestFiles(tc.files)
			require.Equal(t, tc.expected, result)
		})
	}
}
