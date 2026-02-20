// Copyright (c) 2025 Justin Cranford

package legacy_ports

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintPortsCommon "cryptoutil/internal/apps/cicd/lint_ports/common"
)

func TestGetServiceForLegacyPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		port uint16
		want string
	}{
		{name: "cipher-im 8888", port: 8888, want: "cipher-im"},
		{name: "cipher-im 8889", port: 8889, want: "cipher-im"},
		{name: "cipher-im 8890", port: 8890, want: "cipher-im"},
		{name: "jose-ja 9443", port: 9443, want: "jose-ja"},
		{name: "jose-ja 8092", port: 8092, want: "jose-ja"},
		{name: "pki-ca 8443", port: 8443, want: "pki-ca"},
		{name: "unknown port", port: 12345, want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetServiceForLegacyPort(tt.port)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCheck_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestCheck_WithGoFileContainingLegacyPort(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "server.go")

	// Write a Go file that contains a legacy port number (8888 used by cipher-im).
	content := "package main\nconst port = 8888\n"
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"go": {goFile},
	}

	err := Check(logger, filesByExtension)
	require.Error(t, err, "Should detect legacy port violations")
	require.Contains(t, err.Error(), "legacy port violations")
}

func TestCheck_WithMarkdownAndYamlFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "README.md")
	yamlFile := filepath.Join(tmpDir, "config.yml")

	require.NoError(t, os.WriteFile(mdFile, []byte("# Config\nport: 9000\n"), 0o600))
	require.NoError(t, os.WriteFile(yamlFile, []byte("port: 9000\n"), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"md":  {mdFile},
		"yml": {yamlFile},
	}

	err := Check(logger, filesByExtension)
	require.NoError(t, err)
}

func TestCheckFile_WithLegacyPort(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "handler.go")

	// Port 9443 belongs to jose-ja range.
	content := "package main\nconst tlsPort = 9443\n"
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	violations := CheckFile(goFile, lintPortsCommon.AllLegacyPorts())
	require.NotEmpty(t, violations, "Should find legacy port 9443")
}

func TestCheckFile_NonExistentFile(t *testing.T) {
	t.Parallel()

	violations := CheckFile("/nonexistent/path/file.go", lintPortsCommon.AllLegacyPorts())
	require.Empty(t, violations, "Non-existent file returns empty violations")
}

func TestCheck_WithYamlExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(yamlFile, []byte("port: 9000\n"), 0o600))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yaml": {yamlFile},
	}

	err := Check(logger, filesByExtension)
	require.NoError(t, err, "Non-legacy yaml file should pass")
}

func TestCheckFile_ParseUintOverflow(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "overflow.go")

	// 99999 is 5 digits but > 65535 (uint16 max), triggering ParseUint error.
	content := "package main\nconst bigPort = 99999\n"
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0o600))

	violations := CheckFile(goFile, lintPortsCommon.AllLegacyPorts())
	require.Empty(t, violations, "99999 cannot be uint16, should be skipped")
}

func TestCheckFile_OtelPortSkipped(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// File with "otel" in path, so OTEL port 4317 should be skipped.
	otelDir := filepath.Join(tmpDir, "otel-collector")
	require.NoError(t, os.MkdirAll(otelDir, 0o700))
	otelFile := filepath.Join(otelDir, "config.yml")

	// 4317 is an OTEL gRPC port that should be allowed in OTEL context.
	content := "grpc_port: 4317\n"
	require.NoError(t, os.WriteFile(otelFile, []byte(content), 0o600))

	violations := CheckFile(otelFile, lintPortsCommon.AllLegacyPorts())
	// 4317 is not a legacy port anyway, but this exercises the OTEL skip path.
	require.Empty(t, violations)
}

func TestCheckFile_LintPortsDirectorySkipped(t *testing.T) {
	t.Parallel()

	// Files inside "lint_ports/" directory are skipped (port definitions are
	// legitimate there). This covers the early return nil in CheckFile.
	skipPath := "/some/project/lint_ports/common/ports.go"
	violations := CheckFile(skipPath, lintPortsCommon.AllLegacyPorts())
	require.Nil(t, violations, "Files in lint_ports/ should be skipped entirely")
}

func TestCheckFile_OtelPortInOtelContext(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// OtelCollectorPorts are {8888, 8889}. Use 8888 in an OTEL-related path
	// to trigger the OTEL continue (prevents false positive for 8888).
	otelDir := filepath.Join(tmpDir, "otel-collector-config")
	require.NoError(t, os.MkdirAll(otelDir, 0o700))
	otelFile := filepath.Join(otelDir, "config.yml")

	content := "metrics_port: 8888\n"
	require.NoError(t, os.WriteFile(otelFile, []byte(content), 0o600))

	violations := CheckFile(otelFile, lintPortsCommon.AllLegacyPorts())
	// 8888 is both an OTEL collector port AND a legacy cipher-im port.
	// In an OTEL-related file, the OTEL continue fires and no violation is flagged.
	require.Empty(t, violations, "8888 in OTEL context should be skipped")
}
