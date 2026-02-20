// Copyright (c) 2025 Justin Cranford

package test_patterns

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestCheckTestFile_ReadError(t *testing.T) {
	t.Parallel()

	// Test with non-existent file to trigger read error.
	issues := CheckTestFile("/nonexistent/path/to/test_file.go")
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error reading file")
}

func TestCheckTestFile_HardcodedUUID(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "uuid_hardcoded_test.go")

	// File with hardcoded UUID pattern.
	content := "package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tid := \"12345678-1234-1234-1234-123456789012\"\n\t_ = id\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := CheckTestFile(testFile)
	require.NotEmpty(t, issues, "Should find hardcoded UUID issue")

	foundUUIDIssue := false

	for _, issue := range issues {
		if issue == "Found hardcoded UUID - consider using uuid.NewV7() for test data" {
			foundUUIDIssue = true

			break
		}
	}

	require.True(t, foundUUIDIssue, "Should find hardcoded UUID pattern issue")
}

func TestCheckTestFile_TestErrorf(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "errorf_test.go")

	// File with t.Errorf() which should be flagged.
	content := "package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tif true {\n\t\tt.Errorf(\"something went wrong\")\n\t}\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := CheckTestFile(testFile)
	require.NotEmpty(t, issues, "Should find t.Errorf issue")

	foundErrorfIssue := false

	for _, issue := range issues {
		if issue == "Found 1 instances of t.Errorf() - should use require.Errorf() or assert.Errorf()" {
			foundErrorfIssue = true

			break
		}
	}

	require.True(t, foundErrorfIssue, "Should find t.Errorf() pattern issue")
}

func TestCheckTestFile_TestFatalf(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fatalf_test.go")

	// File with t.Fatalf() which should be flagged.
	content := "package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tif true {\n\t\tt.Fatalf(\"fatal error\")\n\t}\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := CheckTestFile(testFile)
	require.NotEmpty(t, issues, "Should find t.Fatalf issue")

	foundFatalfIssue := false

	for _, issue := range issues {
		if issue == "Found 1 instances of t.Fatalf() - should use require.Fatalf() or assert.Fatalf()" {
			foundFatalfIssue = true

			break
		}
	}

	require.True(t, foundFatalfIssue, "Should find t.Fatalf() pattern issue")
}

func TestEnforceTestPatterns_FilteredFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a file that matches the filtering pattern.
	adminTestFile := filepath.Join(tmpDir, "admin_test.go")
	content := "package example\n\nfunc TestAdmin(t *testing.T) {}\n"
	err := os.WriteFile(adminTestFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// This test file should be filtered out, so no issues should be found.
	err = Check(logger, []string{adminTestFile})

	require.NoError(t, err, "Should succeed when only filtered files are provided")
}

func TestCheck_EmptyTestFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := Check(logger, []string{})
	require.NoError(t, err, "Should succeed with empty test files list")
}

func TestCheck_WithValidTestFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "service_test.go")

	// Write a test file with proper UUIDv7 usage and testify assertions.
	content := `package service

import (
"testing"
googleUuid "github.com/google/uuid"
"github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
t.Parallel()
id := googleUuid.NewV7()
require.NotNil(t, id)
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := Check(logger, []string{testFile})
	require.NoError(t, err, "Should pass for valid test file")
}

func TestCheck_WithMultipleFilteredFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Files that should be filtered (excluded from checking).
	filteredFiles := []string{
		filepath.Join(tmpDir, "cicd_test.go"),
		filepath.Join(tmpDir, "testmain_test.go"),
		filepath.Join(tmpDir, "e2e_test.go"),
	}

	for _, f := range filteredFiles {
		require.NoError(t, os.WriteFile(f, []byte("package main\n"), 0o600))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// All files should be filtered out → Check() returns nil (no test files to check).
	err := Check(logger, filteredFiles)
	require.NoError(t, err, "Should succeed when all test files are filtered")
}

func TestCheck_WithViolatingTestFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "service_bad_test.go")

	// File with uuid.New() instead of NewV7() triggers a violation.
	content := `package service

import "testing"

func TestBadUUID(t *testing.T) {
	t.Helper()
	uuid := uuid.New()
	_ = uuid
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := Check(logger, []string{testFile})
	require.Error(t, err, "Should fail for uuid.New() violation")
	require.Contains(t, err.Error(), "test pattern violations")
}

func TestCheckTestFile_UUIDNewViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "uuid_test.go")

	content := "package foo\nfunc T() { _ = uuid.New() }\n"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	issues := CheckTestFile(testFile)
	require.NotEmpty(t, issues, "uuid.New() should be flagged")
	require.Contains(t, issues[0], "uuid.New()")
}

func TestCheckTestFile_TestifyWithoutImport(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "noImport_test.go")

	content := "package foo\nfunc T() { require.NoError(t, nil) }\n"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	issues := CheckTestFile(testFile)
	require.NotEmpty(t, issues, "testify usage without import should be flagged")
}

func TestCheckTestFile_ErrorfViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "errorf_test.go")

	content := "package foo\nfunc T(t *testing.T) { t.Errorf(\"fail\") }\n"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	issues := CheckTestFile(testFile)
	require.NotEmpty(t, issues, "t.Errorf() should be flagged")
}

func TestCheckTestFile_FatalfViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fatalf_test.go")

	content := "package foo\nfunc T(t *testing.T) { t.Fatalf(\"fail\") }\n"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0o600))

	issues := CheckTestFile(testFile)
	require.NotEmpty(t, issues, "t.Fatalf() should be flagged")
}
