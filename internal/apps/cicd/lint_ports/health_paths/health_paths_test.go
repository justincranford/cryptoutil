// Copyright (c) 2025 Justin Cranford

package health_paths

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

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
		{name: "livez", path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath, want: true},
		{name: "readyz", path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, want: true},
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

			got := IsLikelyHealthPath(tt.path)
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
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHealthPathsInDockerfile(dockerfile)
	require.Empty(t, violations)
}

func TestCheckHealthPathsInDockerfile_InvalidPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHealthPathsInDockerfile(dockerfile)
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
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHealthPathsInCompose(composeFile)
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
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHealthPathsInCompose(composeFile)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0].Reason, "Non-standard health path")
}

func TestLintHealthPaths_NoViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD wget -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"dockerfile": {dockerfile},
	}

	err = Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestLintHealthPaths_WithViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD curl -f http://localhost:8080/health
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"dockerfile": {dockerfile},
	}

	err = Check(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "health path violations")
}

func TestCheckHealthPathsInDockerfile_FileNotExists(t *testing.T) {
	t.Parallel()

	violations := CheckHealthPathsInDockerfile("/nonexistent/path/Dockerfile")
	require.Empty(t, violations)
}

func TestCheckHealthPathsInCompose_FileNotExists(t *testing.T) {
	t.Parallel()

	violations := CheckHealthPathsInCompose("/nonexistent/path/compose.yml")
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
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHealthPathsInCompose(composeFile)
	require.Empty(t, violations)
}

func TestCheckHealthPathsInDockerfile_CorrectPortDifferentPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// A Dockerfile with correct admin port but different health path.
	dockerfile := filepath.Join(tempDir, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`FROM alpine:3.19
HEALTHCHECK CMD wget -q -O /dev/null http://127.0.0.1:9090/healthz
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHealthPathsInDockerfile(dockerfile)
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
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHealthPathsInCompose(composeFile)
	require.Empty(t, violations)
}

func TestCheck_WithYamlFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Non-compose yml file (should be skipped, not a compose file).
	configFile := filepath.Join(tmpDir, "config.yml")
	require.NoError(t, os.WriteFile(configFile, []byte("key: value\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// Valid compose file with correct health path.
	composeFile := filepath.Join(tmpDir, "compose.yml")
	composeContent := `services:
  myapp:
    image: alpine:3.19
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
`
	require.NoError(t, os.WriteFile(composeFile, []byte(composeContent), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {configFile, composeFile},
	}

	err := Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestCheck_WithOtelRelatedFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Otel-related dockerfile (should be skipped).
	otelDockerfile := filepath.Join(tmpDir, "otel-collector.dockerfile")
	require.NoError(t, os.WriteFile(otelDockerfile, []byte("FROM alpine:3.19\nHEALTHCHECK ...\n"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"dockerfile": {otelDockerfile},
	}

	err := Check(logger, filesByExtension)
	require.NoError(t, err, "Otel-related files should be skipped")
}

func TestCheckHealthPathsInCompose_HealthcheckSectionExit(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "compose.yml")

	// Compose file where healthcheck section is followed by a non-indented top-level key.
	// This triggers the inHealthcheck = false branch.
	content := `services:
  myapp:
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
volumes:
  data:
`
	require.NoError(t, os.WriteFile(composeFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	violations := CheckHealthPathsInCompose(composeFile)
	require.Empty(t, violations)
}
