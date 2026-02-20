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
