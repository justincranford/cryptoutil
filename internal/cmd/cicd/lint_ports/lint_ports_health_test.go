// Copyright (c) 2025 Justin Cranford

package lint_ports

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestIsLikelyHealthPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "health lowercase", path: "/health", want: true},
		{name: "healthz", path: "/healthz", want: true},
		{name: "livez", path: "/livez", want: true},
		{name: "readyz", path: "/readyz", want: true},
		{name: "alive", path: "/alive", want: true},
		{name: "ready", path: "/ready", want: true},
		{name: "standard path", path: "/admin/api/v1/livez", want: true},
		{name: "api endpoint", path: "/api/v1/users", want: false},
		{name: "root", path: "/", want: false},
		{name: "metrics", path: "/metrics", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isLikelyHealthPath(tt.path)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCheckHealthPathsInDockerfile_ValidPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez || exit 1
`), 0o600)
	require.NoError(t, err)

	violations := checkHealthPathsInDockerfile(dockerfile)
	require.Empty(t, violations)
}

func TestCheckHealthPathsInDockerfile_InvalidPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1
`), 0o600)
	require.NoError(t, err)

	violations := checkHealthPathsInDockerfile(dockerfile)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0].Reason, "Non-standard health path")
}

func TestCheckHealthPathsInCompose_ValidPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  app:
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
      interval: 30s
      timeout: 5s
`), 0o600)
	require.NoError(t, err)

	violations := checkHealthPathsInCompose(composeFile)
	require.Empty(t, violations)
}

func TestCheckHealthPathsInCompose_InvalidPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  app:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
`), 0o600)
	require.NoError(t, err)

	violations := checkHealthPathsInCompose(composeFile)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0].Reason, "Non-standard health path")
}

func TestLintHealthPaths_NoViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD wget -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"dockerfile": {dockerfile},
	}

	err = lintHealthPaths(logger, filesByExtension)
	require.NoError(t, err)
}

func TestLintHealthPaths_WithViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD curl -f http://localhost:8080/health
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"dockerfile": {dockerfile},
	}

	err = lintHealthPaths(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "health path violations")
}

// =============================================================================
// Additional Coverage Tests
// =============================================================================

func TestCheckHealthPathsInDockerfile_FileNotExists(t *testing.T) {
	t.Parallel()

	violations := checkHealthPathsInDockerfile("/nonexistent/path/Dockerfile")
	require.Empty(t, violations)
}

func TestCheckHealthPathsInCompose_FileNotExists(t *testing.T) {
	t.Parallel()

	violations := checkHealthPathsInCompose("/nonexistent/path/compose.yml")
	require.Empty(t, violations)
}

func TestCheckHealthPathsInCompose_NoHealthcheck(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  app:
    image: nginx
    ports:
      - "8080:80"
`), 0o600)
	require.NoError(t, err)

	violations := checkHealthPathsInCompose(composeFile)
	require.Empty(t, violations)
}

func TestCheckHealthPathsInDockerfile_CorrectPortDifferentPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// A Dockerfile with correct admin port but different health path.
	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD wget -q -O /dev/null http://127.0.0.1:9090/healthz
`), 0o600)
	require.NoError(t, err)

	violations := checkHealthPathsInDockerfile(dockerfile)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0].Reason, "Non-standard health path")
}

func TestCheckHealthPathsInCompose_CorrectPathStandardPort(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  app:
    healthcheck:
      test: curl -f http://127.0.0.1:9090/admin/api/v1/livez
      interval: 30s
`), 0o600)
	require.NoError(t, err)

	violations := checkHealthPathsInCompose(composeFile)
	require.Empty(t, violations)
}
