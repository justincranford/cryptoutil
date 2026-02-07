// Copyright (c) 2025 Justin Cranford

package lint_ports

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllLegacyPorts(t *testing.T) {
	t.Parallel()

	ports := AllLegacyPorts()

	// Verify known legacy ports are included.
	require.Contains(t, ports, uint16(8888)) // cipher-im legacy
	require.Contains(t, ports, uint16(8889)) // cipher-im legacy
	require.Contains(t, ports, uint16(8890)) // cipher-im legacy
	require.Contains(t, ports, uint16(9443)) // jose-ja legacy
	require.Contains(t, ports, uint16(8092)) // jose-ja legacy
	require.Contains(t, ports, uint16(8443)) // pki-ca legacy
}

func TestAllValidPublicPorts(t *testing.T) {
	t.Parallel()

	ports := AllValidPublicPorts()

	// Verify standardized ports are included.
	require.Contains(t, ports, uint16(8070)) // cipher-im
	require.Contains(t, ports, uint16(8071)) // cipher-im
	require.Contains(t, ports, uint16(8072)) // cipher-im
	require.Contains(t, ports, uint16(8060)) // jose-ja
	require.Contains(t, ports, uint16(8050)) // pki-ca
	require.Contains(t, ports, uint16(8080)) // sm-kms
	require.Contains(t, ports, uint16(8081)) // sm-kms
	require.Contains(t, ports, uint16(8082)) // sm-kms
	require.Contains(t, ports, uint16(8100)) // identity-authz
	require.Contains(t, ports, uint16(8100)) // identity-idp (same port as authz)
	require.Contains(t, ports, uint16(8110)) // identity-rs
	require.Contains(t, ports, uint16(8120)) // identity-rp
	require.Contains(t, ports, uint16(8130)) // identity-spa
}

func TestIsOtelCollectorPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		port uint16
		want bool
	}{
		{name: "OTEL internal metrics", port: 8888, want: true},
		{name: "OTEL Prometheus", port: 8889, want: true},
		{name: "cipher-im standardized", port: 8070, want: false},
		{name: "jose-ja standardized", port: 8060, want: false},
		{name: "random port", port: 12345, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsOtelCollectorPort(tt.port)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsOtelRelatedFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{name: "otel in path", filePath: "/path/to/otel_config.go", want: true},
		{name: "opentelemetry in path", filePath: "/path/opentelemetry/main.go", want: true},
		{name: "telemetry in path", filePath: "/internal/telemetry/setup.go", want: true},
		{name: "regular go file", filePath: "/internal/server/main.go", want: false},
		{name: "config yaml", filePath: "/configs/app.yml", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isOtelRelatedFile(tt.filePath)
			require.Equal(t, tt.want, got)
		})
	}
}

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

			got := getServiceForLegacyPort(tt.port)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServicePorts_AllServicesPresent(t *testing.T) {
	t.Parallel()

	expectedServices := []string{
		"cipher-im",
		"jose-ja",
		"pki-ca",
		"sm-kms",
		"identity-authz",
		"identity-idp",
		"identity-rs",
		"identity-rp",
		"identity-spa",
	}

	for _, svc := range expectedServices {
		t.Run(svc, func(t *testing.T) {
			t.Parallel()

			cfg, ok := ServicePorts[svc]
			require.True(t, ok, "Service %s should be in ServicePorts", svc)
			require.Equal(t, svc, cfg.Name)
			require.Equal(t, StandardAdminPort, cfg.AdminPort)
			require.NotEmpty(t, cfg.PublicPorts)
		})
	}
}

// TestLint_FileOpenError tests that checkFile handles file open errors gracefully.

func TestIsOtelRelatedContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "otel in constant name", content: "PortOtelCollectorReceivedMetrics uint16 = 8889", want: true},
		{name: "telemetry in comment", content: "// OpenTelemetry metrics port", want: true},
		{name: "opentelemetry in text", content: "// Use OpenTelemetry for observability", want: true},
		{name: "OTEL uppercase", content: "const OTEL_PORT = 8888", want: true},
		{name: "no otel terms", content: "const port = 8080", want: false},
		{name: "cipher-im port", content: "const cipherPort = 8888", want: false},
		{name: "empty line", content: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isOtelRelatedContent(tt.content)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestLint_SkipsCollectorPortsInMagicFile tests that collector ports are skipped
// when the line content contains related terms (even if file path doesn't).
// NOTE: Function name avoids "otel/telemetry" to prevent t.TempDir() path matching.
