// Copyright (c) 2025 Justin Cranford

package lint_ports

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestIsComposeFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{name: "compose.yml", filePath: "deployments/compose.yml", want: true},
		{name: "compose.yaml", filePath: "deployments/compose.yaml", want: true},
		{name: "docker-compose.yml", filePath: "docker-compose.yml", want: true},
		{name: "docker-compose.yaml", filePath: "docker-compose.yaml", want: true},
		{name: "compose.e2e.yml", filePath: "deployments/identity/compose.e2e.yml", want: true},
		{name: "compose.advanced.yml", filePath: "compose.advanced.yml", want: true},
		{name: "regular yaml", filePath: "config/settings.yml", want: false},
		{name: "regular yaml 2", filePath: "configs/app.yaml", want: false},
		{name: "go file", filePath: "main.go", want: false},
		{name: "dockerfile", filePath: "Dockerfile", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isComposeFile(tt.filePath)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetServiceConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		wantNil     bool
		wantName    string
	}{
		{name: "exact match cipher-im", serviceName: "cipher-im", wantNil: false, wantName: "cipher-im"},
		{name: "exact match jose-ja", serviceName: "jose-ja", wantNil: false, wantName: "jose-ja"},
		{name: "exact match sm-kms", serviceName: "sm-kms", wantNil: false, wantName: "sm-kms"},
		{name: "prefix match cipher-im-postgres", serviceName: "cipher-im-postgres", wantNil: false, wantName: "cipher-im"},
		{name: "prefix match jose-ja-sqlite", serviceName: "jose-ja-sqlite", wantNil: false, wantName: "jose-ja"},
		{name: "unknown service", serviceName: "unknown-service", wantNil: true, wantName: ""},
		{name: "empty string", serviceName: "", wantNil: true, wantName: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getServiceConfig(tt.serviceName)
			if tt.wantNil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, tt.wantName, got.Name)
			}
		})
	}
}

func TestIsPortInValidRange(t *testing.T) {
	t.Parallel()

	cipherConfig := &ServicePortConfig{
		Name:        "cipher-im",
		PublicPorts: []uint16{8070, 8071, 8072},
		AdminPort:   9090,
	}

	tests := []struct {
		name string
		port uint16
		cfg  *ServicePortConfig
		want bool
	}{
		{name: "public port 8070", port: 8070, cfg: cipherConfig, want: true},
		{name: "public port 8071", port: 8071, cfg: cipherConfig, want: true},
		{name: "public port 8072", port: 8072, cfg: cipherConfig, want: true},
		{name: "admin port 9090", port: 9090, cfg: cipherConfig, want: true},
		{name: "range port 8073", port: 8073, cfg: cipherConfig, want: true},    // In range 8070-8079
		{name: "range port 8079", port: 8079, cfg: cipherConfig, want: true},    // Last in range
		{name: "out of range 8080", port: 8080, cfg: cipherConfig, want: false}, // Out of range
		{name: "out of range 8060", port: 8060, cfg: cipherConfig, want: false}, // Different service
		{name: "legacy port 8888", port: 8888, cfg: cipherConfig, want: false},  // Legacy
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isPortInValidRange(tt.port, tt.cfg)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCheckHostPortRangesInFile_ValidPorts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  cipher-im:
    ports:
      - "8070:8070"
      - "9090:9090"
`), 0o600)
	require.NoError(t, err)

	violations := checkHostPortRangesInFile(composeFile)
	require.Empty(t, violations)
}

func TestCheckHostPortRangesInFile_InvalidPorts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  cipher-im:
    ports:
      - "8888:8070"
`), 0o600)
	require.NoError(t, err)

	violations := checkHostPortRangesInFile(composeFile)
	require.Len(t, violations, 1)
	require.Equal(t, uint16(8888), violations[0].Port)
	require.Contains(t, violations[0].Reason, "outside valid range")
}

func TestLintHostPortRanges_NoViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  jose-ja:
    ports:
      - "8060:8060"
      - "9090:9090"
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = lintHostPortRanges(logger, filesByExtension)
	require.NoError(t, err)
}

func TestLintHostPortRanges_WithViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  jose-ja:
    ports:
      - "9443:8060"
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = lintHostPortRanges(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host port range violations")
}

// =============================================================================
// Health Path Validation Tests
// =============================================================================

func TestCheckHostPortRangesInFile_FileNotExists(t *testing.T) {
	t.Parallel()

	violations := checkHostPortRangesInFile("/nonexistent/path/compose.yml")
	require.Empty(t, violations)
}

func TestCheckHostPortRangesInFile_UnknownService(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  unknown-service:
    ports:
      - "9999:8080"
`), 0o600)
	require.NoError(t, err)

	// Unknown services should not cause violations (no config to validate against).
	violations := checkHostPortRangesInFile(composeFile)
	require.Empty(t, violations)
}

func TestCheckHostPortRangesInFile_TopLevelReset(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  cipher-im:
    ports:
      - "8070:8070"
networks:
  default:
    driver: bridge
`), 0o600)
	require.NoError(t, err)

	violations := checkHostPortRangesInFile(composeFile)
	require.Empty(t, violations)
}
