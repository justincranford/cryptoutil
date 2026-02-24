// Copyright (c) 2025 Justin Cranford

package host_port_ranges

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintPortsCommon "cryptoutil/internal/apps/cicd/lint_ports/common"

	"github.com/stretchr/testify/require"
)

func TestCheckHostPortRangesInFile_FileNotExists(t *testing.T) {
	t.Parallel()

	violations := CheckHostPortRangesInFile("/nonexistent/path/compose.yml")
	require.Empty(t, violations)
}

func TestCheckHostPortRangesInFile_InvalidPorts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  sm-im:
    ports:
      - "8070:8700"
`), 0o600)
	require.NoError(t, err)

	violations := CheckHostPortRangesInFile(composeFile)
	require.Len(t, violations, 1)
	require.Equal(t, uint16(8070), violations[0].Port)
	require.Contains(t, violations[0].Reason, "outside valid range")
}

func TestCheckHostPortRangesInFile_TopLevelReset(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  sm-im:
    ports:
      - "8700:8700"
networks:
  default:
    driver: bridge
`), 0o600)
	require.NoError(t, err)

	violations := CheckHostPortRangesInFile(composeFile)
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
	violations := CheckHostPortRangesInFile(composeFile)
	require.Empty(t, violations)
}

func TestCheckHostPortRangesInFile_ValidPorts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  sm-im:
    ports:
      - "8700:8700"
      - "9090:9090"
`), 0o600)
	require.NoError(t, err)

	violations := CheckHostPortRangesInFile(composeFile)
	require.Empty(t, violations)
}

func TestGetServiceConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		wantNil     bool
		wantName    string
	}{
		{name: "exact match sm-im", serviceName: "sm-im", wantNil: false, wantName: "sm-im"},
		{name: "exact match jose-ja", serviceName: "jose-ja", wantNil: false, wantName: "jose-ja"},
		{name: "exact match sm-kms", serviceName: "sm-kms", wantNil: false, wantName: "sm-kms"},
		{name: "prefix match sm-im-postgres", serviceName: "sm-im-postgres", wantNil: false, wantName: "sm-im"},
		{name: "prefix match jose-ja-sqlite", serviceName: "jose-ja-sqlite", wantNil: false, wantName: "jose-ja"},
		{name: "unknown service", serviceName: "unknown-service", wantNil: true, wantName: ""},
		{name: "empty string", serviceName: "", wantNil: true, wantName: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetServiceConfig(tt.serviceName)
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

	cipherConfig := &lintPortsCommon.ServicePortConfig{
		Name:        "sm-im",
		PublicPorts: []uint16{8700, 8701, 8702},
		AdminPort:   9090,
	}

	tests := []struct {
		name string
		port uint16
		cfg  *lintPortsCommon.ServicePortConfig
		want bool
	}{
		{name: "public port 8700", port: 8700, cfg: cipherConfig, want: true},
		{name: "public port 8701", port: 8701, cfg: cipherConfig, want: true},
		{name: "public port 8702", port: 8702, cfg: cipherConfig, want: true},
		{name: "admin port 9090", port: 9090, cfg: cipherConfig, want: true},
		{name: "range port 8703", port: 8703, cfg: cipherConfig, want: true},    // In range 8700-8799
		{name: "range port 8799", port: 8799, cfg: cipherConfig, want: true},    // Last in range
		{name: "out of range 8800", port: 8800, cfg: cipherConfig, want: false}, // Out of range (jose-ja territory)
		{name: "out of range 8060", port: 8060, cfg: cipherConfig, want: false}, // Legacy jose-ja port
		{name: "legacy port 8888", port: 8888, cfg: cipherConfig, want: false},  // Legacy
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsPortInValidRange(tt.port, tt.cfg)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLintHostPortRanges_NoViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  jose-ja:
    ports:
      - "8800:8800"
      - "9090:9090"
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestLintHostPortRanges_WithViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  jose-ja:
    ports:
      - "9443:8800"
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = Check(logger, filesByExtension)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host port range violations")
}

func TestLintHostPortRanges_NonComposeYAMLSkipped(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a YAML file that is NOT a compose file (should be skipped by IsComposeFile).
	nonComposeFile := filepath.Join(tempDir, "config.yml")
	err := os.WriteFile(nonComposeFile, []byte(`key: value
`), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {nonComposeFile},
	}

	err = Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestCheckHostPortRangesInFile_PortParseUintError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Port 99999 exceeds uint16 max (65535), so ParseUint with bitSize=16 will fail.
	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
  sm-im:
    ports:
      - "99999:8700"
`), 0o600)
	require.NoError(t, err)

	violations := CheckHostPortRangesInFile(composeFile)
	require.Empty(t, violations)
}
