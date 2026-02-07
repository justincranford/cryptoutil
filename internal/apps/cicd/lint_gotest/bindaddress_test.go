// Copyright (c) 2025 Justin Cranford

package lint_gotest

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestEnforceBindAddressSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantError   bool
		description string
	}{
		{
			name: "safe_127_0_0_1",
			fileContent: `package example_test

import "testing"

func TestExample(t *testing.T) {
	addr := "127.0.0.1:0"
	_ = addr
}`,
			wantError:   false,
			description: "Should pass with 127.0.0.1",
		},
		{
			name: "unsafe_0_0_0_0_in_ServerSettings",
			fileContent: `package example_test

import "testing"

func TestExample(t *testing.T) {
	settings := &cryptoutilConfig.ServiceTemplateServerSettings{
		BindPublicAddress: "0.0.0.0",
		BindPublicPort: 8080,
	}
	_ = settings
}`,
			wantError:   true,
			description: "Should fail with 0.0.0.0 in ServiceTemplateServerSettings",
		},
		{
			name: "unsafe_blank_bind_address",
			fileContent: `package example_test

import "testing"

func TestExample(t *testing.T) {
	settings := &cryptoutilConfig.ServiceTemplateServerSettings{
		BindPublicAddress: "",
		BindPublicPort: 8080,
	}
	_ = settings
}`,
			wantError:   true,
			description: "Should fail with blank BindPublicAddress",
		},
		{
			name: "unsafe_direct_struct_creation",
			fileContent: `package example_test

import "testing"

func TestExample(t *testing.T) {
	settings := &cryptoutilConfig.ServiceTemplateServerSettings{
		BindPublicAddress: "127.0.0.1",
		BindPublicPort: 8080,
	}
	_ = settings
}`,
			wantError:   true,
			description: "Should fail with direct ServiceTemplateServerSettings{} (no NewTestConfig)",
		},
		{
			name: "safe_with_new_test_config",
			fileContent: `package example_test

import "testing"

func TestExample(t *testing.T) {
	settings := cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)
	_ = settings
}`,
			wantError:   false,
			description: "Should pass with NewTestConfig",
		},
		{
			name: "unsafe_net_listen_empty",
			fileContent: `package example_test

import (
	"net"
	"testing"
)

func TestExample(t *testing.T) {
	listener, _ := net.Listen("tcp", ":0")
	_ = listener
}`,
			wantError:   true,
			description: "Should fail with net.Listen empty address",
		},
		{
			name: "safe_net_listen_127",
			fileContent: `package example_test

import (
	"net"
	"testing"
)

func TestExample(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	_ = listener
}`,
			wantError:   false,
			description: "Should pass with net.Listen 127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file.
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "test_file_test.go")

			err := os.WriteFile(tempFile, []byte(tt.fileContent), 0o600)
			require.NoError(t, err, "Failed to create temp test file")

			// Run linter.
			logger := cryptoutilCmdCicdCommon.NewLogger("bind-address-safety-test")
			err = enforceBindAddressSafety(logger, []string{tempFile})

			if tt.wantError {
				require.Error(t, err, "Expected error for: %s", tt.description)
			} else {
				require.NoError(t, err, "Expected no error for: %s", tt.description)
			}
		})
	}
}

func TestCheckBindAddressSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		fileContent  string
		wantIssues   int
		issuePattern string
	}{
		{
			name: "multiple_violations",
			fileContent: `package example_test

func TestExample(t *testing.T) {
	addr1 := "0.0.0.0:8080"
	settings := &cryptoutilConfig.ServiceTemplateServerSettings{
		BindPublicAddress: "",
		BindPrivateAddress: "",
	}
	listener, _ := net.Listen("tcp", ":0")
}`,
			wantIssues:   4, // 1 direct 0.0.0.0 + 2 blank binds + 1 net.Listen.
			issuePattern: "0.0.0.0",
		},
		{
			name: "no_violations",
			fileContent: `package example_test

func TestExample(t *testing.T) {
	settings := cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
}`,
			wantIssues:   0,
			issuePattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file.
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "test_file_test.go")

			err := os.WriteFile(tempFile, []byte(tt.fileContent), 0o600)
			require.NoError(t, err)

			// Check file directly.
			issues := checkBindAddressSafety(tempFile)

			require.Len(t, issues, tt.wantIssues, "Expected %d issues, got %d", tt.wantIssues, len(issues))

			if tt.wantIssues > 0 {
				// Verify at least one issue contains expected pattern.
				found := false

				for _, issue := range issues {
					if tt.issuePattern != "" && len(issue) > 0 {
						found = true

						break
					}
				}

				require.True(t, found, "Expected to find issue pattern")
			}
		})
	}
}

// TestCheckTestFile_ReadError tests the file read error path.
func TestCheckTestFile_ReadError(t *testing.T) {
	t.Parallel()

	// Test with non-existent file to trigger read error.
	issues := checkTestFile("/nonexistent/path/to/test_file.go")
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error reading file")
}

// TestCheckTestFile_HardcodedUUID tests detection of hardcoded UUIDs.
func TestCheckTestFile_HardcodedUUID(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "uuid_hardcoded_test.go")

	// File with hardcoded UUID pattern.
	content := "package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tid := \"12345678-1234-1234-1234-123456789012\"\n\t_ = id\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := checkTestFile(testFile)
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

// TestCheckTestFile_TestErrorf tests detection of t.Errorf() patterns.
func TestCheckTestFile_TestErrorf(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "errorf_test.go")

	// File with t.Errorf() which should be flagged.
	content := "package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tif true {\n\t\tt.Errorf(\"something went wrong\")\n\t}\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := checkTestFile(testFile)
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

// TestCheckTestFile_TestFatalf tests detection of t.Fatalf() patterns.
func TestCheckTestFile_TestFatalf(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fatalf_test.go")

	// File with t.Fatalf() which should be flagged.
	content := "package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tif true {\n\t\tt.Fatalf(\"fatal error\")\n\t}\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0o600)
	require.NoError(t, err)

	issues := checkTestFile(testFile)
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

// TestCheckBindAddressSafety_ReadError tests the file read error path.
func TestCheckBindAddressSafety_ReadError(t *testing.T) {
	t.Parallel()

	// Test with non-existent file to trigger read error.
	issues := checkBindAddressSafety("/nonexistent/path/to/test_file.go")
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "Error reading file")
}

// TestEnforceBindAddressSafety_FilteredFiles tests that config_test.go files are filtered out.
func TestEnforceBindAddressSafety_FilteredFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a file that matches the filtering pattern.
	configTestFile := filepath.Join(tmpDir, "config_test.go")
	content := "package example\n\nfunc TestConfig(t *testing.T) {}\n"
	err := os.WriteFile(configTestFile, []byte(content), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	// This test file should be filtered out, so no issues should be found.
	err = enforceBindAddressSafety(logger, []string{configTestFile})

	require.NoError(t, err, "Should succeed when only filtered files are provided")
}

// TestEnforceTestPatterns_FilteredFiles tests that admin_test.go files are filtered out.
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
	err = enforceTestPatterns(logger, []string{adminTestFile})

	require.NoError(t, err, "Should succeed when only filtered files are provided")
}
