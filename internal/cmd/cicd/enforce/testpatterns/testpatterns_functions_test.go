// Copyright (c) 2025 Justin Cranford

package testpatterns_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/enforce/testpatterns"
)

func TestCheckTestFile_ValidTestFile(t *testing.T) {
	t.Parallel()

	// Create a valid test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "valid_test.go")
	content := `package example_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	googleUuid "github.com/google/uuid"
)

func TestExample(t *testing.T) {
	id := googleUuid.NewV7()
	require.NotNil(t, id)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	// Act
	issues := testpatterns.CheckTestFile(testFile)

	// Assert
	require.Empty(t, issues, "Valid test file should have no issues")
}

func TestCheckTestFile_UUIDNew(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "uuid_new_test.go")
	content := `package example_test

import (
	"testing"
	"github.com/google/uuid"
)

func TestExample(t *testing.T) {
	id := uuid.New()
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	// Act
	issues := testpatterns.CheckTestFile(testFile)

	// Assert
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "uuid.New()")
	require.Contains(t, issues[0], "uuid.NewV7()")
}

func TestCheckTestFile_HardcodedUUID(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "hardcoded_uuid_test.go")
	content := `package example_test

import "testing"

func TestExample(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	// Act
	issues := testpatterns.CheckTestFile(testFile)

	// Assert
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "hardcoded UUID")
	require.Contains(t, issues[0], "uuid.NewV7()")
}

func TestCheckTestFile_TErrorf(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "t_errorf_test.go")
	content := `package example_test

import "testing"

func TestExample(t *testing.T) {
	t.Errorf("expected %d, got %d", 1, 2)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	// Act
	issues := testpatterns.CheckTestFile(testFile)

	// Assert
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "t.Errorf()")
	require.Contains(t, issues[0], "require.Errorf()")
}

func TestCheckTestFile_TFatalf(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "t_fatalf_test.go")
	content := `package example_test

import "testing"

func TestExample(t *testing.T) {
	t.Fatalf("test failed: %v", err)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	// Act
	issues := testpatterns.CheckTestFile(testFile)

	// Assert
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "t.Fatalf()")
	require.Contains(t, issues[0], "require.Fatalf()")
}

func TestCheckTestFile_TestifyUsageWithoutImport(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "missing_import_test.go")
	content := `package example_test

import "testing"

func TestExample(t *testing.T) {
	require.Equal(t, 1, 1)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	// Act
	issues := testpatterns.CheckTestFile(testFile)

	// Assert
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "testify assertions")
	require.Contains(t, issues[0], "doesn't import testify")
}

func TestCheckTestFile_MultipleIssues(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "multiple_issues_test.go")
	content := `package example_test

import (
	"testing"
	"github.com/google/uuid"
)

func TestExample(t *testing.T) {
	id1 := uuid.New()
	id2 := "550e8400-e29b-41d4-a716-446655440000"
	t.Errorf("error")
	t.Fatalf("fatal")
	require.Equal(t, 1, 1)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	// Act
	issues := testpatterns.CheckTestFile(testFile)

	// Assert: Should have multiple issues
	require.Greater(t, len(issues), 1, "Should detect multiple issues")
	// Check for specific issues
	hasUUIDNew := false
	hasHardcodedUUID := false
	hasTErrorf := false
	hasTFatalf := false
	hasMissingImport := false

	for _, issue := range issues {
		if contains(issue, "uuid.New()") {
			hasUUIDNew = true
		}

		if contains(issue, "hardcoded UUID") {
			hasHardcodedUUID = true
		}

		if contains(issue, "t.Errorf()") {
			hasTErrorf = true
		}

		if contains(issue, "t.Fatalf()") {
			hasTFatalf = true
		}

		if contains(issue, "doesn't import testify") {
			hasMissingImport = true
		}
	}

	require.True(t, hasUUIDNew, "Should detect uuid.New() usage")
	require.True(t, hasHardcodedUUID, "Should detect hardcoded UUID")
	require.True(t, hasTErrorf, "Should detect t.Errorf() usage")
	require.True(t, hasTFatalf, "Should detect t.Fatalf() usage")
	require.True(t, hasMissingImport, "Should detect missing testify import")
}

func TestCheckTestFile_NonExistentFile(t *testing.T) {
	t.Parallel()

	// Act
	issues := testpatterns.CheckTestFile("/nonexistent/path/test.go")

	// Assert
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error reading file")
}

func TestEnforce_NoTestFiles(t *testing.T) {
	t.Parallel()

	logger := common.NewLogger("test-enforce-no-files")
	allFiles := []string{
		"main.go",
		"util.go",
		"config.go",
	}

	// Act
	err := testpatterns.Enforce(logger, allFiles)

	// Assert
	require.NoError(t, err, "Should not error when no test files present")
}

func TestEnforce_ValidTestFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "valid_test.go")
	content := `package example_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	googleUuid "github.com/google/uuid"
)

func TestExample(t *testing.T) {
	id := googleUuid.NewV7()
	require.NotNil(t, id)
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := common.NewLogger("test-enforce-valid")
	allFiles := []string{testFile}

	// Act
	err = testpatterns.Enforce(logger, allFiles)

	// Assert
	require.NoError(t, err, "Should not error when all test files are valid")
}

func TestEnforce_InvalidTestFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid_test.go")
	content := `package example_test

import (
	"testing"
	googleUuid "github.com/google/uuid"
)

func TestExample(t *testing.T) {
	id := googleUuid.New()
	t.Errorf("error")
}
`
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := common.NewLogger("test-enforce-invalid")
	allFiles := []string{testFile}

	// Act
	err = testpatterns.Enforce(logger, allFiles)

	// Assert
	require.Error(t, err, "Should error when test files have violations")
	require.Contains(t, err.Error(), "test pattern violations")
}

func TestEnforce_ExcludedFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create cicd_test.go with deliberate violations (should be excluded)
	cicdTestFile := filepath.Join(tempDir, "cicd_test.go")
	cicdContent := `package cicd_test

import "testing"

func TestExample(t *testing.T) {
	t.Errorf("deliberate violation")
}
`
	err := os.WriteFile(cicdTestFile, []byte(cicdContent), 0o600)
	require.NoError(t, err)

	logger := common.NewLogger("test-enforce-excluded")
	allFiles := []string{cicdTestFile}

	// Act
	err = testpatterns.Enforce(logger, allFiles)

	// Assert: Should not error because cicd_test.go is excluded
	require.NoError(t, err, "Should not check excluded files like cicd_test.go")
}

func TestEnforce_MultipleFilesWithIssues(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create first invalid file
	file1 := filepath.Join(tempDir, "test1_test.go")
	content1 := `package example_test

import (
	"testing"
	"github.com/google/uuid"
)

func TestExample1(t *testing.T) {
	id := uuid.New()
}
`
	err := os.WriteFile(file1, []byte(content1), 0o600)
	require.NoError(t, err)

	// Create second invalid file
	file2 := filepath.Join(tempDir, "test2_test.go")
	content2 := `package test2_test

import "testing"

func TestExample2(t *testing.T) {
	t.Errorf("error")
}
`
	err = os.WriteFile(file2, []byte(content2), 0o600)
	require.NoError(t, err)

	logger := common.NewLogger("test-enforce-multiple")
	allFiles := []string{file1, file2}

	// Act
	err = testpatterns.Enforce(logger, allFiles)

	// Assert
	require.Error(t, err, "Should error when multiple test files have violations")
	require.Contains(t, err.Error(), "test pattern violations")
	require.Contains(t, err.Error(), "2 files")
}

// Helper function to check if a string contains a substring (case-insensitive check not needed here).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
