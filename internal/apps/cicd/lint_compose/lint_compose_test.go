// Copyright (c) 2025 Justin Cranford

package lint_compose

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestLint_NoComposeFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := Lint(logger, filesByExtension)
	require.NoError(t, err, "lint should pass with no compose files")
}

func TestLint_ValidComposeFile(t *testing.T) {
	t.Parallel()

	// Create temp dir with valid compose file.
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "docker-compose.yml")

	validContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "8080:8080"  # Public API only
      # Admin API (9090) NOT exposed - internal 127.0.0.1 only
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
`
	err := os.WriteFile(composeFile, []byte(validContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "lint should pass with valid compose file")
}

func TestLint_AdminPortExposed(t *testing.T) {
	t.Parallel()

	// Create temp dir with violating compose file.
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "compose.yml")

	violatingContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "8080:8080"
      - "9090:9090"
`
	err := os.WriteFile(composeFile, []byte(violatingContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "lint should fail when admin port 9090 is exposed")
	require.Contains(t, err.Error(), "admin port exposure violations")
}

func TestLint_AdminPortMappedToDifferentHost(t *testing.T) {
	t.Parallel()

	// Create temp dir with violating compose file.
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "compose.yml")

	violatingContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "19090:9090"
`
	err := os.WriteFile(composeFile, []byte(violatingContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "lint should fail when admin port 9090 is mapped to different host port")
}

func TestLint_PortRangeToAdmin(t *testing.T) {
	t.Parallel()

	// Create temp dir with violating compose file.
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "compose.yml")

	violatingContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "9080-9089:9090"
`
	err := os.WriteFile(composeFile, []byte(violatingContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "lint should fail when port range maps to admin port 9090")
}

func TestLint_CommentedOutAdminPort(t *testing.T) {
	t.Parallel()

	// Create temp dir with valid compose file (commented out violation).
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "compose.yml")

	validContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "8080:8080"
      # - "9090:9090"  # Commented out - should not trigger
`
	err := os.WriteFile(composeFile, []byte(validContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "lint should pass when admin port is commented out")
}

func TestLint_QuotedPortMapping(t *testing.T) {
	t.Parallel()

	// Create temp dir with violating compose file using quoted port.
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "compose.yml")

	violatingContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "9090:9090"
`
	err := os.WriteFile(composeFile, []byte(violatingContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "lint should fail with quoted port mapping")
}

func TestFindComposeFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    map[string][]string
		expected int
	}{
		{
			name:     "no files",
			files:    map[string][]string{},
			expected: 0,
		},
		{
			name: "compose.yml",
			files: map[string][]string{
				"yml": {"compose.yml"},
			},
			expected: 1,
		},
		{
			name: "docker-compose.yml",
			files: map[string][]string{
				"yml": {"docker-compose.yml"},
			},
			expected: 1,
		},
		{
			name: "compose.yaml",
			files: map[string][]string{
				"yaml": {"compose.yaml"},
			},
			expected: 1,
		},
		{
			name: "multiple compose files",
			files: map[string][]string{
				"yml":  {"compose.yml", "docker-compose.yml", "compose.demo.yml"},
				"yaml": {"compose.yaml"},
			},
			expected: 4,
		},
		{
			name: "non-compose yml files",
			files: map[string][]string{
				"yml": {"config.yml", "settings.yml"},
			},
			expected: 0,
		},
		{
			name: "mixed files",
			files: map[string][]string{
				"yml": {"compose.yml", "config.yml"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := findComposeFiles(tt.files)
			require.Len(t, result, tt.expected)
		})
	}
}

func TestCheckComposeFile_InvalidFile(t *testing.T) {
	t.Parallel()

	violations, err := checkComposeFile("/nonexistent/compose.yml")
	require.Error(t, err, "should error on invalid file")
	require.Nil(t, violations)
}

func TestLint_MultipleViolations(t *testing.T) {
	t.Parallel()

	// Create temp dir with multiple violating compose files.
	tempDir := t.TempDir()
	composeFile1 := filepath.Join(tempDir, "compose.yml")
	composeFile2 := filepath.Join(tempDir, "docker-compose.yml")

	violating1 := `version: '3.8'
services:
  app1:
    ports:
      - "9090:9090"
`
	violating2 := `version: '3.8'
services:
  app2:
    ports:
      - "19090:9090"
`
	err := os.WriteFile(composeFile1, []byte(violating1), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(composeFile2, []byte(violating2), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile1, composeFile2},
	}

	err = Lint(logger, filesByExtension)
	require.Error(t, err, "lint should fail with multiple violations")
	require.Contains(t, err.Error(), "2 admin port exposure violations")
}

// TestCheckComposeFile_FileOpenError tests the error path when compose file cannot be opened.
func TestCheckComposeFile_FileOpenError(t *testing.T) {
	t.Parallel()

	violations, err := checkComposeFile("/nonexistent/path/to/compose.yml")
	require.Error(t, err, "should return error for non-existent file")
	require.Nil(t, violations, "should return nil violations on error")
	require.Contains(t, err.Error(), "failed to open file", "error should indicate file open failure")
}

// TestLint_FileOpenErrorContinues tests that Lint continues processing when one file cannot be opened.
func TestLint_FileOpenErrorContinues(t *testing.T) {
	t.Parallel()

	// Create temp dir with one valid compose file.
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "compose.yml")

	validContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "8080:8080"
`
	err := os.WriteFile(composeFile, []byte(validContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	// Include both a valid file and a non-existent file.
	filesByExtension := map[string][]string{
		"yml": {composeFile, "/nonexistent/compose.yml"},
	}

	// Lint should succeed (warning for non-existent file, but no violations).
	err = Lint(logger, filesByExtension)
	require.NoError(t, err, "lint should pass even with one file error")
}
