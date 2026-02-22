// Copyright (c) 2025 Justin Cranford
//
//

package testutil_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

func TestWriteTempFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "Simple_text_file",
			filename: "test.txt",
			content:  "Hello, World!",
		},
		{
			name:     "YAML_file",
			filename: "config.yml",
			content:  "key: value\nfoo: bar",
		},
		{
			name:     "Empty_content",
			filename: "empty.txt",
			content:  "",
		},
		{
			name:     "JSON_file",
			filename: "data.json",
			content:  `{"name":"test","value":123}`,
		},
		{
			name:     "Multiline_content",
			filename: "multi.txt",
			content:  "Line 1\nLine 2\nLine 3\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			// Act: Write temporary file
			filePath := cryptoutilSharedTestutil.WriteTempFile(t, tempDir, tc.filename, tc.content)

			// Assert: File exists at expected path
			expectedPath := filepath.Join(tempDir, tc.filename)
			require.Equal(t, expectedPath, filePath, "Should return correct file path")

			// Assert: File exists
			_, err := os.Stat(filePath)
			require.NoError(t, err, "File should exist")

			// Assert: File content matches
			actualContent, err := os.ReadFile(filePath)
			require.NoError(t, err, "Should be able to read file")
			require.Equal(t, tc.content, string(actualContent), "File content should match")
		})
	}
}

func TestWriteTempFile_NestedDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "sub", "nested")
	err := os.MkdirAll(nestedDir, 0o755)
	require.NoError(t, err)

	// Act: Write file in nested directory
	filePath := cryptoutilSharedTestutil.WriteTempFile(t, nestedDir, "nested.txt", "nested content")

	// Assert: File exists in nested directory
	expectedPath := filepath.Join(nestedDir, "nested.txt")
	require.Equal(t, expectedPath, filePath)

	actualContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, "nested content", string(actualContent))
}

func TestWriteTestFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "Absolute_path",
			filename: "absolute.txt",
			content:  "Absolute path content",
		},
		{
			name:     "Binary_content",
			filename: "binary.dat",
			content:  "\x00\x01\x02\xFF",
		},
		{
			name:     "Large_content",
			filename: "large.txt",
			content:  string(make([]byte, 10000)), // 10KB of zeros
		},
		{
			name:     "Special_characters",
			filename: "special.txt",
			content:  "Special: !@#$%^&*()_+{}|:\"<>?",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tc.filename)

			// Act: Write test file
			cryptoutilSharedTestutil.WriteTestFile(t, filePath, tc.content)

			// Assert: File exists
			_, err := os.Stat(filePath)
			require.NoError(t, err, "File should exist")

			// Assert: File content matches
			actualContent, err := os.ReadFile(filePath)
			require.NoError(t, err, "Should be able to read file")
			require.Equal(t, tc.content, string(actualContent), "File content should match")
		})
	}
}

func TestWriteTestFile_CreateDirectories(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "does", "not", "exist", "file.txt")

	// WriteTestFile should fail if parent directory doesn't exist
	// (os.WriteFile doesn't create parent directories)
	// This test verifies the expected behavior

	// Create parent directories first
	err := os.MkdirAll(filepath.Dir(nestedPath), 0o755)
	require.NoError(t, err)

	// Now WriteTestFile should succeed
	cryptoutilSharedTestutil.WriteTestFile(t, nestedPath, "nested file content")

	actualContent, err := os.ReadFile(nestedPath)
	require.NoError(t, err)
	require.Equal(t, "nested file content", string(actualContent))
}

func TestReadTestFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "Text_content",
			content: "Test file content",
		},
		{
			name:    "Empty_file",
			content: "",
		},
		{
			name:    "Binary_content",
			content: "\x00\x01\x02\xFF\xFE",
		},
		{
			name:    "Multiline_content",
			content: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:    "Unicode_content",
			content: "Hello 世界 🌍",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "test.txt")

			// Setup: Write file using standard library
			err := os.WriteFile(filePath, []byte(tc.content), 0o600)
			require.NoError(t, err)

			// Act: Read test file
			actualContent := cryptoutilSharedTestutil.ReadTestFile(t, filePath)

			// Assert: Content matches
			require.Equal(t, []byte(tc.content), actualContent, "File content should match")
		})
	}
}

func TestReadTestFile_Integration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	content := "Integration test content"

	// Write using WriteTestFile
	filePath := filepath.Join(tempDir, "integration.txt")
	cryptoutilSharedTestutil.WriteTestFile(t, filePath, content)

	// Read using ReadTestFile
	actualContent := cryptoutilSharedTestutil.ReadTestFile(t, filePath)

	// Assert: Round-trip succeeds
	require.Equal(t, []byte(content), actualContent, "Round-trip content should match")
}

func TestWriteAndRead_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "Simple_roundtrip",
			content: "Simple content",
		},
		{
			name:    "Complex_roundtrip",
			content: "Multi\nline\ncontent\nwith\nspecial: chars!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			// Write using WriteTempFile
			filePath := cryptoutilSharedTestutil.WriteTempFile(t, tempDir, "roundtrip.txt", tc.content)

			// Read using ReadTestFile
			actualContent := cryptoutilSharedTestutil.ReadTestFile(t, filePath)

			// Assert: Round-trip matches
			require.Equal(t, []byte(tc.content), actualContent)
		})
	}
}

func TestIntegrationTimeout(t *testing.T) {
	t.Parallel()

	t.Run("default timeout", func(t *testing.T) {
		t.Parallel()

		timeout := cryptoutilSharedTestutil.IntegrationTimeout()
		require.Equal(t, cryptoutilSharedTestutil.DefaultIntegrationTimeout, timeout)
	})
}

// TestIntegrationTimeout_Override verifies that TestTimeoutOverride is used when set.
// Sequential: modifies package-level TestTimeoutOverride variable.
func TestIntegrationTimeout_Override(t *testing.T) {
	customTimeout := 42 * time.Second

	cryptoutilSharedTestutil.TestTimeoutOverride = customTimeout

	defer func() { cryptoutilSharedTestutil.TestTimeoutOverride = 0 }()

	timeout := cryptoutilSharedTestutil.IntegrationTimeout()
	require.Equal(t, customTimeout, timeout)
}

func TestIntegrationContext(t *testing.T) {
	t.Parallel()

	ctx := cryptoutilSharedTestutil.IntegrationContext(t)
	require.NotNil(t, ctx)

	// Context should have deadline.
	_, hasDeadline := ctx.Deadline()
	require.True(t, hasDeadline, "Context should have deadline")
}

func TestTestID(t *testing.T) {
	t.Parallel()

	t.Run("without prefix", func(t *testing.T) {
		t.Parallel()

		id := cryptoutilSharedTestutil.TestID("")
		require.NotEmpty(t, id)
		require.Len(t, id, 36) // UUID format: 8-4-4-4-12
	})

	t.Run("with prefix", func(t *testing.T) {
		t.Parallel()

		id := cryptoutilSharedTestutil.TestID("test")
		require.NotEmpty(t, id)
		require.Contains(t, id, "test-")
	})

	t.Run("unique IDs", func(t *testing.T) {
		t.Parallel()

		id1 := cryptoutilSharedTestutil.TestID("test")
		id2 := cryptoutilSharedTestutil.TestID("test")
		require.NotEqual(t, id1, id2, "IDs should be unique")
	})
}

func TestTestUserFactory(t *testing.T) {
	t.Parallel()

	factory := cryptoutilSharedTestutil.NewTestUserFactory("user-test")

	t.Run("creates user with unique ID", func(t *testing.T) {
		t.Parallel()

		user := factory.Create("admin")
		require.NotEmpty(t, user.ID)
		require.Contains(t, user.Username, "admin")
		require.Contains(t, user.Email, "@test.example.com")
		require.NotEmpty(t, user.Password)
		require.True(t, user.Enabled)
	})

	t.Run("creates unique users", func(t *testing.T) {
		t.Parallel()

		user1 := factory.Create("user")
		user2 := factory.Create("user")
		require.NotEqual(t, user1.ID, user2.ID, "IDs should be unique")
		require.NotEqual(t, user1.Username, user2.Username, "Usernames should be unique")
	})
}

func TestTestClientFactory(t *testing.T) {
	t.Parallel()

	factory := cryptoutilSharedTestutil.NewTestClientFactory("client-test")

	t.Run("creates confidential client", func(t *testing.T) {
		t.Parallel()

		client := factory.CreateConfidential("Test App")
		require.NotEmpty(t, client.ID)
		require.Contains(t, client.ClientID, "client-")
		require.NotEmpty(t, client.ClientSecret, "Confidential client should have secret")
		require.Equal(t, "Test App", client.Name)
		require.False(t, client.Public)
		require.NotEmpty(t, client.RedirectURIs)
		require.NotEmpty(t, client.Scopes)
	})

	t.Run("creates public client", func(t *testing.T) {
		t.Parallel()

		client := factory.CreatePublic("Public App")
		require.NotEmpty(t, client.ID)
		require.Contains(t, client.ClientID, "public-")
		require.Empty(t, client.ClientSecret, "Public client should not have secret")
		require.Equal(t, "Public App", client.Name)
		require.True(t, client.Public)
	})
}

func TestTestTenantFactory(t *testing.T) {
	t.Parallel()

	factory := cryptoutilSharedTestutil.NewTestTenantFactory("tenant-test")

	t.Run("creates tenant with UUIDv4", func(t *testing.T) {
		t.Parallel()

		tenant := factory.Create("ACME Corp")
		require.NotEmpty(t, tenant.ID)
		require.Len(t, tenant.ID, 36) // UUID format
		require.Contains(t, tenant.Name, "ACME Corp")
		require.Contains(t, tenant.Description, "Test tenant")
		require.Equal(t, "default", tenant.RealmID)
		require.True(t, tenant.Enabled)
	})

	t.Run("creates unique tenants", func(t *testing.T) {
		t.Parallel()

		tenant1 := factory.Create("Tenant")
		tenant2 := factory.Create("Tenant")
		require.NotEqual(t, tenant1.ID, tenant2.ID, "IDs should be unique")
	})
}

// TestCaptureOutput tests the CaptureOutput function.
