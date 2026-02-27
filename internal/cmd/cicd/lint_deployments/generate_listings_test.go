package lint_deployments

import (
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestClassifyFileType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{name: "compose yaml", filename: "compose.yml", want: fileTypeCompose},
		{name: "compose yaml variant", filename: "compose-cryptoutil.yml", want: fileTypeCompose},
		{name: "compose yaml extension", filename: "compose.yaml", want: fileTypeCompose},
		{name: "secret file", filename: "postgres_password.secret", want: fileTypeSecret},
		{name: "never file", filename: "postgres_url.secret.never", want: fileTypeSecret},
		{name: "dockerfile", filename: "Dockerfile", want: fileTypeDocker},
		{name: "config yaml", filename: "sm-kms-app-common.yml", want: fileTypeConfig},
		{name: "config yaml alt", filename: "ca-config-schema.yaml", want: fileTypeConfig},
		{name: "sql file", filename: "init-db.sql", want: fileTypeSQL},
		{name: "markdown", filename: "README.md", want: fileTypeDoc},
		{name: "json file", filename: "listings.json", want: fileTypeJSON},
		{name: "other file", filename: "something.txt", want: fileTypeOther},
		{name: "no extension", filename: "Makefile", want: fileTypeOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := classifyFileType(tc.filename)
			require.Equal(t, tc.want, got, "classifyFileType(%q)", tc.filename)
		})
	}
}

func TestClassifyFileStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{name: "compose file", relPath: "service/compose.yml", want: RequiredFileStatus},
		{name: "secret file", relPath: "service/secrets/db.secret", want: RequiredFileStatus},
		{name: "dockerfile", relPath: "service/Dockerfile", want: RequiredFileStatus},
		{name: "never file", relPath: "service/secrets/db.secret.never", want: RequiredFileStatus},
		{name: "markdown", relPath: "service/README.md", want: OptionalFileStatus},
		{name: "config file", relPath: "service/config/app.yml", want: OptionalFileStatus},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := classifyFileStatus(tc.relPath)
			require.Equal(t, tc.want, got, "classifyFileStatus(%q)", tc.relPath)
		})
	}
}

func TestGenerateDirectoryListing(t *testing.T) {
	t.Parallel()

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateDirectoryListing("/nonexistent/path")
		require.Error(t, err, "expected error for nonexistent directory")
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		listing, err := GenerateDirectoryListing(tmpDir)
		require.NoError(t, err)

		require.Empty(t, listing)
	})

	t.Run("directory with files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		// Create test files.
		createTestFile(t, tmpDir, "compose.yml", "")
		createTestFile(t, tmpDir, "Dockerfile", "")
		createTestDir(t, tmpDir, "secrets")
		createTestFile(t, tmpDir, "secrets/db.secret", "")
		createTestDir(t, tmpDir, "config")
		createTestFile(t, tmpDir, "config/app.yml", "")

		listing, err := GenerateDirectoryListing(tmpDir)
		require.NoError(t, err)

		require.Len(t, listing, 4)

		// Verify compose.yml classification.
		entry, ok := listing["compose.yml"]
		if !ok {
			t.Error("missing compose.yml in listing")
		} else {
			require.Equal(t, fileTypeCompose, entry.Type)

			require.Equal(t, RequiredFileStatus, entry.Status)
		}

		// Verify Dockerfile classification.
		entry, ok = listing["Dockerfile"]
		if !ok {
			t.Error("missing Dockerfile in listing")
		} else {
			require.Equal(t, fileTypeDocker, entry.Type)
		}

		// Verify secret classification.
		entry, ok = listing["secrets/db.secret"]
		if !ok {
			t.Error("missing secrets/db.secret in listing")
		} else {
			require.Equal(t, fileTypeSecret, entry.Type)
		}
	})

	t.Run("skips generated listing files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestFile(t, tmpDir, "compose.yml", "")
		createTestFile(t, tmpDir, "deployments-all-files.json", "{}")

		listing, err := GenerateDirectoryListing(tmpDir)
		require.NoError(t, err)

		if _, ok := listing["deployments-all-files.json"]; ok {
			t.Error("generated listing file should be skipped")
		}

		require.Len(t, listing, 1)
	})
}

func TestGenerateDeploymentsListing(t *testing.T) {
	t.Parallel()

	t.Run("valid directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestFile(t, tmpDir, "compose.yml", "")
		createTestDir(t, tmpDir, "secrets")
		createTestFile(t, tmpDir, "secrets/db.secret", "")

		data, err := GenerateDeploymentsListing(tmpDir)
		require.NoError(t, err)

		// Verify valid JSON.
		var result map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &result), "invalid JSON output")

		require.Len(t, result, 2)
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateDeploymentsListing("/nonexistent")
		require.Error(t, err, "expected error for nonexistent directory")
	})
}

func TestGenerateConfigsListing(t *testing.T) {
	t.Parallel()

	t.Run("valid directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestFile(t, tmpDir, "app.yml", "")

		data, err := GenerateConfigsListing(tmpDir)
		require.NoError(t, err)

		var result map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &result), "invalid JSON output")
	})
}

func TestMarshalListing(t *testing.T) {
	t.Parallel()

	t.Run("empty listing", func(t *testing.T) {
		t.Parallel()

		listing := make(DirectoryListing)

		data, err := marshalListing(listing)
		require.NoError(t, err)

		// Verify empty JSON object.
		var result map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &result), "invalid JSON")

		require.Empty(t, result)
	})

	t.Run("sorted output", func(t *testing.T) {
		t.Parallel()

		listing := DirectoryListing{
			"z/file.yml":    {Type: "config", Status: OptionalFileStatus},
			"a/file.yml":    {Type: "config", Status: OptionalFileStatus},
			"m/compose.yml": {Type: "compose", Status: RequiredFileStatus},
		}

		data, err := marshalListing(listing)
		require.NoError(t, err)

		// Verify "a/" appears before "m/" and "m/" before "z/".
		output := string(data)
		aIdx := len(output)
		mIdx := len(output)
		zIdx := len(output)

		for i, c := range output {
			if c == 'a' && i > 0 && output[i-1] == '"' {
				aIdx = i
			}

			if c == 'm' && i > 0 && output[i-1] == '"' {
				mIdx = i
			}

			if c == 'z' && i > 0 && output[i-1] == '"' {
				zIdx = i
			}
		}

		require.True(t, aIdx < mIdx && mIdx < zIdx, "keys not sorted: a=%d m=%d z=%d", aIdx, mIdx, zIdx)
	})
}

func TestWriteListingFile(t *testing.T) {
	t.Parallel()

	t.Run("writes valid JSON file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestFile(t, tmpDir, "compose.yml", "")

		outputPath := filepath.Join(tmpDir, "listing.json")

		err := WriteListingFile(tmpDir, outputPath)
		require.NoError(t, err)

		// Read and verify.
		data, readErr := os.ReadFile(outputPath)
		require.NoError(t, readErr, "failed to read output file")

		var result map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &result), "invalid JSON")
	})

	t.Run("nonexistent base directory", func(t *testing.T) {
		t.Parallel()

		err := WriteListingFile("/nonexistent", "/tmp/output.json")
		require.Error(t, err, "expected error for nonexistent directory")
	})
}

// createTestFile creates a file with content in the given dir.
func createTestFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	path := filepath.Join(dir, name)

	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute); err != nil {
		require.NoError(t, err, "failed to create directory %s", dirPath)
	}

	if err := os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions); err != nil {
		require.NoError(t, err, "failed to create file %s", path)
	}
}

// createTestDir creates a directory in the given dir.
func createTestDir(t *testing.T, dir string, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(path, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute); err != nil {
		require.NoError(t, err, "failed to create directory %s", path)
	}
}

// TestGenerateConfigsListing_Error tests error propagation from GenerateConfigsListing.
func TestGenerateConfigsListing_Error(t *testing.T) {
	t.Parallel()

	_, err := GenerateConfigsListing("/nonexistent-configs-dir-xyz")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate configs listing")
}

// TestWriteListingFile_WriteError tests write failure in WriteListingFile.
func TestWriteListingFile_WriteError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid input dir with a file.
	createTestFile(t, tmpDir, "test.yml", "content")

	// Use a directory path that doesn't exist for output.
	badOutput := filepath.Join("/nonexistent-dir-xyz", "output.json")
	err := WriteListingFile(tmpDir, badOutput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write listing file")
}

// TestGenerateDirectoryListing_WalkError tests walk error in GenerateDirectoryListing.
func TestGenerateDirectoryListing_WalkError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a subdirectory then remove read permission.
	subDir := filepath.Join(tmpDir, "restricted")
	require.NoError(t, os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	createTestFile(t, subDir, "file.txt", "data")

	// Remove read permission on subdirectory.
	require.NoError(t, os.Chmod(subDir, 0o000))

	t.Cleanup(func() {
		// Restore permission for cleanup.
		_ = os.Chmod(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	_, err := GenerateDirectoryListing(tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk directory")
}
