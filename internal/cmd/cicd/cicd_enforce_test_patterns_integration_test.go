// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/enforce/testpatterns"
)

func TestGoEnforceTestPatterns_NoFiles(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceTestPatterns_NoFiles")
	allFiles := []string{}

	err := goEnforceTestPatterns(logger, allFiles)
	require.NoError(t, err)
}

func TestGoEnforceTestPatterns_NoTestFiles(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceTestPatterns_NoTestFiles")
	allFiles := []string{"main.go", "util.go", "README.md"}

	err := goEnforceTestPatterns(logger, allFiles)
	require.NoError(t, err)
}

func TestGoEnforceTestPatterns_ValidTestFile(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceTestPatterns_ValidTestFile")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "example_test.go")

	validContent := `package example

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

	err := os.WriteFile(testFile, []byte(validContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{testFile}
	err = goEnforceTestPatterns(logger, allFiles)
	require.NoError(t, err)
}

func TestGoEnforceTestPatterns_InvalidTestFile_UUIDNew(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceTestPatterns_InvalidTestFile_UUIDNew")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "example_test.go")

	invalidContent := `package example

import (
	"testing"
	"github.com/google/uuid"
)

func TestExample(t *testing.T) {
	id := uuid.New() // Should use NewV7
	_ = id
}
`

	err := os.WriteFile(testFile, []byte(invalidContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{testFile}
	err = goEnforceTestPatterns(logger, allFiles)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test pattern violations")
}

func TestGoEnforceTestPatterns_InvalidTestFile_TErrorf(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceTestPatterns_InvalidTestFile_TErrorf")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "example_test.go")

	invalidContent := `package example

import "testing"

func TestExample(t *testing.T) {
	t.Errorf("this should use require.Error")
}
`

	err := os.WriteFile(testFile, []byte(invalidContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{testFile}
	err = goEnforceTestPatterns(logger, allFiles)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test pattern violations")
}

func TestGoEnforceTestPatterns_InvalidTestFile_TFatalf(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceTestPatterns_InvalidTestFile_TFatalf")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "example_test.go")

	invalidContent := `package example

import "testing"

func TestExample(t *testing.T) {
	t.Fatalf("this should use require.FailNow")
}
`

	err := os.WriteFile(testFile, []byte(invalidContent), 0o600)
	require.NoError(t, err)

	allFiles := []string{testFile}
	err = goEnforceTestPatterns(logger, allFiles)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test pattern violations")
}

func TestGoEnforceTestPatterns_ExcludesCICDFiles(t *testing.T) {
	logger := common.NewLogger("TestGoEnforceTestPatterns_ExcludesCICDFiles")

	// These files should be excluded even if they have violations
	allFiles := []string{
		"internal/cmd/cicd/cicd_test.go",
		"internal/cmd/cicd/cicd.go",
		"internal/cmd/cicd/cicd_enforce_test_patterns_test.go",
	}

	err := goEnforceTestPatterns(logger, allFiles)
	require.NoError(t, err, "CICD files should be excluded from pattern checks")
}

func TestCheckTestFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "valid_test.go")

	validContent := `package example

import (
	"testing"
	"github.com/stretchr/testify/require"
	googleUuid "github.com/google/uuid"
)

func TestValidExample(t *testing.T) {
	id := googleUuid.NewV7()
	require.NotNil(t, id)
}
`

	err := os.WriteFile(testFile, []byte(validContent), 0o600)
	require.NoError(t, err)

	issues := testpatterns.CheckTestFile(testFile)
	require.Empty(t, issues, "Valid file should have no issues")
}

func TestCheckTestFile_AllViolations(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid_test.go")

	invalidContent := `package example

import "testing"

func TestExample(t *testing.T) {
	id := uuid.New() // Violation: should use NewV7
	hardcoded := "550e8400-e29b-41d4-a716-446655440000" // Violation: hardcoded UUID
	t.Errorf("error: %v", id) // Violation: should use require
	t.Fatalf("fatal: %v", hardcoded) // Violation: should use require
}
`

	err := os.WriteFile(testFile, []byte(invalidContent), 0o600)
	require.NoError(t, err)

	issues := testpatterns.CheckTestFile(testFile)
	require.NotEmpty(t, issues, "Invalid file should have issues")
	require.GreaterOrEqual(t, len(issues), 4, "Should detect multiple violations")
}

func TestCheckTestFile_FileReadError(t *testing.T) {
	issues := testpatterns.CheckTestFile("/nonexistent/file/that/does/not/exist.go")
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error reading file")
}

func TestCheckTestFile_HardcodedUUID(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "uuid_test.go")

	content := `package example

import "testing"

func TestUUID(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	_ = id
}
`

	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := testpatterns.CheckTestFile(testFile)
	require.NotEmpty(t, issues)

	foundUUIDIssue := false

	for _, issue := range issues {
		if contains(issue, "hardcoded UUID") {
			foundUUIDIssue = true

			break
		}
	}

	require.True(t, foundUUIDIssue, "Should detect hardcoded UUID")
}

func TestCheckTestFile_TestifyUsageWithoutImport(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "testify_test.go")

	content := `package example

import "testing"

func TestExample(t *testing.T) {
	require.NotNil(t, someValue)
}
`

	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := testpatterns.CheckTestFile(testFile)
	require.NotEmpty(t, issues)

	foundImportIssue := false

	for _, issue := range issues {
		if contains(issue, "testify") && contains(issue, "import") {
			foundImportIssue = true

			break
		}
	}

	require.True(t, foundImportIssue, "Should detect missing testify import")
}

// Helper function.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
