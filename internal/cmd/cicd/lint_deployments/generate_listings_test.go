package lint_deployments

import (
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"
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
			if got != tc.want {
				t.Errorf("classifyFileType(%q) = %q, want %q", tc.filename, got, tc.want)
			}
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
			if got != tc.want {
				t.Errorf("classifyFileStatus(%q) = %q, want %q", tc.relPath, got, tc.want)
			}
		})
	}
}

func TestGenerateDirectoryListing(t *testing.T) {
	t.Parallel()

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateDirectoryListing("/nonexistent/path")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		listing, err := GenerateDirectoryListing(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(listing) != 0 {
			t.Errorf("expected empty listing, got %d entries", len(listing))
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(listing) != 4 {
			t.Errorf("expected 4 entries, got %d", len(listing))
		}

		// Verify compose.yml classification.
		entry, ok := listing["compose.yml"]
		if !ok {
			t.Error("missing compose.yml in listing")
		} else {
			if entry.Type != fileTypeCompose {
				t.Errorf("compose.yml type = %q, want %s", entry.Type, fileTypeCompose)
			}

			if entry.Status != RequiredFileStatus {
				t.Errorf("compose.yml status = %q, want %s", entry.Status, RequiredFileStatus)
			}
		}

		// Verify Dockerfile classification.
		entry, ok = listing["Dockerfile"]
		if !ok {
			t.Error("missing Dockerfile in listing")
		} else if entry.Type != fileTypeDocker {
			t.Errorf("Dockerfile type = %q, want %s", entry.Type, fileTypeDocker)
		}

		// Verify secret classification.
		entry, ok = listing["secrets/db.secret"]
		if !ok {
			t.Error("missing secrets/db.secret in listing")
		} else if entry.Type != fileTypeSecret {
			t.Errorf("secret type = %q, want %s", entry.Type, fileTypeSecret)
		}
	})

	t.Run("skips generated listing files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestFile(t, tmpDir, "compose.yml", "")
		createTestFile(t, tmpDir, "deployments_all_files.json", "{}")

		listing, err := GenerateDirectoryListing(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := listing["deployments_all_files.json"]; ok {
			t.Error("generated listing file should be skipped")
		}

		if len(listing) != 1 {
			t.Errorf("expected 1 entry, got %d", len(listing))
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify valid JSON.
		var result map[string]json.RawMessage
		if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
			t.Fatalf("invalid JSON output: %v", jsonErr)
		}

		if len(result) != 2 {
			t.Errorf("expected 2 entries, got %d", len(result))
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateDeploymentsListing("/nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})
}

func TestGenerateConfigsListing(t *testing.T) {
	t.Parallel()

	t.Run("valid directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		createTestFile(t, tmpDir, "app.yml", "")

		data, err := GenerateConfigsListing(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var result map[string]json.RawMessage
		if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
			t.Fatalf("invalid JSON output: %v", jsonErr)
		}
	})
}

func TestMarshalListing(t *testing.T) {
	t.Parallel()

	t.Run("empty listing", func(t *testing.T) {
		t.Parallel()

		listing := make(DirectoryListing)

		data, err := marshalListing(listing)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify empty JSON object.
		var result map[string]json.RawMessage
		if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
			t.Fatalf("invalid JSON: %v", jsonErr)
		}

		if len(result) != 0 {
			t.Errorf("expected empty map, got %d entries", len(result))
		}
	})

	t.Run("sorted output", func(t *testing.T) {
		t.Parallel()

		listing := DirectoryListing{
			"z/file.yml":    {Type: "config", Status: OptionalFileStatus},
			"a/file.yml":    {Type: "config", Status: OptionalFileStatus},
			"m/compose.yml": {Type: "compose", Status: RequiredFileStatus},
		}

		data, err := marshalListing(listing)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

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

		if aIdx >= mIdx || mIdx >= zIdx {
			t.Errorf("keys not sorted: a=%d m=%d z=%d", aIdx, mIdx, zIdx)
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Read and verify.
		data, readErr := os.ReadFile(outputPath)
		if readErr != nil {
			t.Fatalf("failed to read output file: %v", readErr)
		}

		var result map[string]json.RawMessage
		if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
			t.Fatalf("invalid JSON: %v", jsonErr)
		}
	})

	t.Run("nonexistent base directory", func(t *testing.T) {
		t.Parallel()

		err := WriteListingFile("/nonexistent", "/tmp/output.json")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})
}

// createTestFile creates a file with content in the given dir.
func createTestFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	path := filepath.Join(dir, name)

	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dirPath, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
}

// createTestDir creates a directory in the given dir.
func createTestDir(t *testing.T, dir string, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}
