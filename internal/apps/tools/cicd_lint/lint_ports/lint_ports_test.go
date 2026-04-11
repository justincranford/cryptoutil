// Copyright (c) 2025 Justin Cranford

package lint_ports

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintPortsCommon "cryptoutil/internal/apps/tools/cicd_lint/lint_ports/common"

	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)

	if subDir := filepath.Dir(path); subDir != dir {
		require.NoError(t, os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	}

	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	return path
}

func TestLint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) map[string][]string
		wantErr string
	}{
		{
			name: "no legacy ports in go file",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "main.go", "package main\n\nconst port = 8700 // sm-im standardized port\n")}}
			},
		},
		{
			name: "detects single legacy port",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "main.go", "package main\n\nconst port = 8888 // legacy sm-im port\n")}}
			},
			wantErr: "legacy port violations",
		},
		{
			name: "detects multiple legacy ports",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "main.go", "package main\n\nconst smIMPort = 8888 // legacy\nconst josePort = 9443  // legacy\n")}}
			},
			wantErr: "2 legacy port violations",
		},
		{
			name: "skips collector files for collector ports",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "otel_config.go", "package main\n\nconst metricsPort = 8888 // OTEL internal metrics\nconst promPort = 8889    // OTEL Prometheus port\n")}}
			},
		},
		{
			name: "detects legacy port 8890 in regular file",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "config.go", "package main\n\nconst port = 8890 // legacy sm-im port\n")}}
			},
			wantErr: "legacy port",
		},
		{
			name: "detects 8888 in regular file",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()
				f := writeTestFile(t, d, "main.go", "package main\n\nconst port = 8888\n")
				require.False(t, lintPortsCommon.IsOtelRelatedFile(f), "test file path should NOT be collector-related")

				return map[string][]string{"go": {f}}
			},
			wantErr: "legacy port violations",
		},
		{
			name: "yaml file with legacy port",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"yml": {writeTestFile(t, d, "config.yml", "server:\n  port: 8443\n")}}
			},
			wantErr: "legacy port violations",
		},
		{
			name: "markdown file with legacy port",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"md": {writeTestFile(t, d, "README.md", "# Server\n\nConnect to port 9443 for JOSE.\n")}}
			},
			wantErr: "legacy port violations",
		},
		{
			name: "skips lint_ports directory",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, filepath.Join("lint_ports", "constants.go"), "package lint_ports\n\nvar LegacyPorts = []uint16{8888, 8889, 8890}\n")}}
			},
		},
		{
			name: "non-existent file silently ignored",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {filepath.Join(d, "does_not_exist.go")}}
			},
		},
		{
			name: "valid port numbers not flagged",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "main.go", "package main\n\nconst validPort = 8000\nconst anotherValid = 8700\nconst irrelevantNumber = 1234\n")}}
			},
		},
		{
			name: "empty files by extension",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()

				return map[string][]string{}
			},
		},
		{
			name: "json files ignored as unsupported extension",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"json": {writeTestFile(t, d, "config.json", "{\n  \"port\": 8888\n}\n")}}
			},
		},
		{
			name: "multiple extensions with violations",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{
					"go":   {writeTestFile(t, d, "main.go", "package main\nconst port = 8888\n")},
					"yml":  {writeTestFile(t, d, "config.yml", "port: 9443\n")},
					"yaml": {writeTestFile(t, d, "config.yaml", "port: 8443\n")},
					"md":   {writeTestFile(t, d, "README.md", "# Docs\nPort 8890 is used.\n")},
				}
			},
			wantErr: "legacy port violations",
		},
		{
			name: "non-legacy port numbers not flagged",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "main.go", "package main\n\nconst port1 = 8000\nconst port2 = 8700\nconst port3 = 9000\nconst port4 = 12345\n")}}
			},
		},
		{
			name: "skips collector ports in magic content",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()
				f := writeTestFile(t, d, "magic_network.go", "package magic\n\n// Default OpenTelemetry collector internal metrics port (Prometheus).\nconst DefaultPublicPortInternalMetrics uint16 = 8888\n// PortOtelCollectorReceivedMetrics - Default OpenTelemetry collector received metrics port.\nconst PortOtelCollectorReceivedMetrics uint16 = 8889\n")
				require.False(t, lintPortsCommon.IsOtelRelatedFile(f), "file path should NOT be collector-related")

				return map[string][]string{"go": {f}}
			},
		},
		{
			name: "all three check types pass with valid files",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{
					"go":         {writeTestFile(t, d, "main.go", "package main\n\nconst port = 8100\n")},
					"yml":        {writeTestFile(t, d, "compose.yml", "services:\n  sm-im:\n    ports:\n      - \"8100:8100\"\n    healthcheck:\n      test: [\"CMD\", \"wget\", \"-q\", \"-O\", \"/dev/null\", \"https://127.0.0.1:9090/admin/api/v1/livez\"]\n")},
					"dockerfile": {writeTestFile(t, d, "Dockerfile", "FROM alpine:latest\nHEALTHCHECK CMD wget -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez\n")},
				}
			},
		},
		{
			name: "legacy port 9443 detected",
			setup: func(t *testing.T) map[string][]string {
				t.Helper()
				d := t.TempDir()

				return map[string][]string{"go": {writeTestFile(t, d, "main.go", "package main\n\nconst port = 9443\n")}}
			},
			wantErr: "legacy port",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			filesByExtension := tc.setup(t)

			err := Lint(logger, filesByExtension)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
