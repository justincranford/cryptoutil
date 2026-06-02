// Copyright (c) 2025-2026 Justin Cranford.
package host_port_ranges

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintPortsCommon "cryptoutil/internal/apps-tools/cicd_lint/lint_ports/common"

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
  sm-kms:
    ports:
      - "8443:8000"
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHostPortRangesInFile(composeFile)
	require.Len(t, violations, 1)
	require.Equal(t, uint16(8443), violations[0].Port)
	require.Contains(t, violations[0].Reason, "outside valid range")
}

func TestCheckHostPortRangesInFile_TopLevelReset(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	composeFile := filepath.Join(tempDir, "compose.yml")
	err := os.WriteFile(composeFile, []byte(`services:
	sm-kms:
    ports:
			- "8000:8000"
networks:
  default:
    driver: bridge
`), cryptoutilSharedMagic.CacheFilePermissions)
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
`), cryptoutilSharedMagic.CacheFilePermissions)
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
  sm-kms:
    ports:
      - "8000:8000"
      - "9090:9090"
`), cryptoutilSharedMagic.CacheFilePermissions)
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
		{name: "exact match sm-kms", serviceName: cryptoutilSharedMagic.OTLPServiceSMKMS, wantNil: false, wantName: cryptoutilSharedMagic.OTLPServiceSMKMS},
		{name: "prefix match sm-kms-postgres", serviceName: "sm-kms-postgres", wantNil: false, wantName: cryptoutilSharedMagic.OTLPServiceSMKMS},
		{name: "prefix match pki-ca-sqlite", serviceName: "pki-ca-sqlite", wantNil: false, wantName: cryptoutilSharedMagic.OTLPServicePKICA},
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

	smIMConfig := &lintPortsCommon.ServicePortConfig{
		Name:        cryptoutilSharedMagic.OTLPServiceSMKMS,
		PublicPorts: []uint16{cryptoutilSharedMagic.KMSServicePort, cryptoutilSharedMagic.KMSE2EPostgreSQL1PublicPort, cryptoutilSharedMagic.KMSE2EPostgreSQL2PublicPort},
		AdminPort:   cryptoutilSharedMagic.DefaultPrivatePortCryptoutil,
	}

	tests := []struct {
		name string
		port uint16
		cfg  *lintPortsCommon.ServicePortConfig
		want bool
	}{
		{name: "public port cryptoutilSharedMagic.KMSServicePort", port: cryptoutilSharedMagic.KMSServicePort, cfg: smIMConfig, want: true},
		{name: "public port cryptoutilSharedMagic.KMSE2EPostgreSQL1PublicPort", port: cryptoutilSharedMagic.KMSE2EPostgreSQL1PublicPort, cfg: smIMConfig, want: true},
		{name: "public port cryptoutilSharedMagic.KMSE2EPostgreSQL2PublicPort", port: cryptoutilSharedMagic.KMSE2EPostgreSQL2PublicPort, cfg: smIMConfig, want: true},
		{name: "admin port cryptoutilSharedMagic.TestAdminPort", port: cryptoutilSharedMagic.IdentityDefaultAuthZAdminPort, cfg: smIMConfig, want: true},
		{name: "range port 8050", port: 8050, cfg: smIMConfig, want: true},                                                                                                       // In range 8000-8099
		{name: "range port 8099", port: 8099, cfg: smIMConfig, want: true},                                                                                                       // Last in range
		{name: "out of range cryptoutilSharedMagic.PKICAServicePort", port: cryptoutilSharedMagic.PKICAServicePort, cfg: smIMConfig, want: false},                                // Out of range (pki-ca territory)
		{name: "range port 8060", port: 8060, cfg: smIMConfig, want: true},                                                                                                       // In range 8000-8099
		{name: "legacy port cryptoutilSharedMagic.DefaultPublicPortInternalMetrics", port: cryptoutilSharedMagic.DefaultPublicPortInternalMetrics, cfg: smIMConfig, want: false}, // Legacy
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
  sm-kms:
    ports:
      - "8000:8000"
      - "9090:9090"
`), cryptoutilSharedMagic.CacheFilePermissions)
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
  sm-kms:
    ports:
      - "9443:8800"
`), cryptoutilSharedMagic.CacheFilePermissions)
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
`), cryptoutilSharedMagic.CacheFilePermissions)
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
  sm-kms:
    ports:
			- "99999:8000"
`), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	violations := CheckHostPortRangesInFile(composeFile)
	require.Empty(t, violations)
}
