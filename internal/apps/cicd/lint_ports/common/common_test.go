// Copyright (c) 2025 Justin Cranford

package common

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	
	"github.com/stretchr/testify/require"
)

func TestAllLegacyPorts(t *testing.T) {
	t.Parallel()

	ports := AllLegacyPorts()

	// Verify known legacy ports are included.
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.DefaultPublicPortInternalMetrics)) // sm-im legacy
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.PortOtelCollectorReceivedMetrics)) // sm-im legacy
	require.Contains(t, ports, uint16(8890)) // sm-im legacy
	require.Contains(t, ports, uint16(9443)) // jose-ja legacy
	require.Contains(t, ports, uint16(8092)) // jose-ja legacy
	require.Contains(t, ports, uint16(8443)) // pki-ca legacy
}

func TestAllValidPublicPorts(t *testing.T) {
	t.Parallel()

	ports := AllValidPublicPorts()

	// Verify standardized ports are included.
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.IMServicePort)) // sm-im
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.IME2EPostgreSQL1PublicPort)) // sm-im
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.IME2EPostgreSQL2PublicPort)) // sm-im
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.JoseJAServicePort)) // jose-ja
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.PKICAServicePort)) // pki-ca
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.KMSServicePort)) // sm-kms
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.KMSE2EPostgreSQL1PublicPort)) // sm-kms
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.KMSE2EPostgreSQL2PublicPort)) // sm-kms
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.IdentityDefaultAuthZPort)) // identity-authz
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.IdentityDefaultIDPPort)) // identity-idp
	require.Contains(t, ports, uint16(8301)) // identity-idp E2E (avoids conflict with authz)
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.IdentityDefaultRSPort)) // identity-rs
	require.Contains(t, ports, uint16(cryptoutilSharedMagic.DefaultSPARPPort)) // identity-rp
	require.Contains(t, ports, uint16(8600)) // identity-spa
}

func TestIsOtelCollectorPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		port uint16
		want bool
	}{
		{name: "OTEL internal metrics", port: cryptoutilSharedMagic.DefaultPublicPortInternalMetrics, want: true},
		{name: "OTEL Prometheus", port: cryptoutilSharedMagic.PortOtelCollectorReceivedMetrics, want: true},
		{name: "sm-im standardized", port: cryptoutilSharedMagic.IMServicePort, want: false}, // New standardized port
		{name: "jose-ja standardized", port: cryptoutilSharedMagic.JoseJAServicePort, want: false},   // New standardized port
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

			got := IsOtelRelatedFile(tt.filePath)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServicePorts_AllServicesPresent(t *testing.T) {
	t.Parallel()

	expectedServices := []string{
		cryptoutilSharedMagic.OTLPServiceSMIM,
		cryptoutilSharedMagic.OTLPServiceJoseJA,
		cryptoutilSharedMagic.OTLPServicePKICA,
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		cryptoutilSharedMagic.OTLPServiceIdentityRS,
		cryptoutilSharedMagic.OTLPServiceIdentityRP,
		cryptoutilSharedMagic.OTLPServiceIdentitySPA,
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
			{name: "sm-im port", content: "const smIMPort = 8888", want: false},
		{name: "empty line", content: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsOtelRelatedContent(tt.content)
			require.Equal(t, tt.want, got)
		})
	}
}

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

			got := IsComposeFile(tt.filePath)
			require.Equal(t, tt.want, got)
		})
	}
}
