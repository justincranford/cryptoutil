// Copyright (c) 2025 Justin Cranford

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPrivateBaseURL tests the PrivateBaseURL method.
func TestPrivateBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings ServiceTemplateServerSettings
		expected string
	}{
		{
			name: "https localhost 9090",
			settings: ServiceTemplateServerSettings{
				BindPrivateProtocol: "https",
				BindPrivateAddress:  "localhost",
				BindPrivatePort:     9090,
			},
			expected: "https://localhost:9090",
		},
		{
			name: "http 127.0.0.1 8080",
			settings: ServiceTemplateServerSettings{
				BindPrivateProtocol: "http",
				BindPrivateAddress:  "127.0.0.1",
				BindPrivatePort:     8080,
			},
			expected: "http://127.0.0.1:8080",
		},
		{
			name: "https IPv6 9999",
			settings: ServiceTemplateServerSettings{
				BindPrivateProtocol: "https",
				BindPrivateAddress:  "::1",
				BindPrivatePort:     9999,
			},
			expected: "https://::1:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.settings.PrivateBaseURL()
			require.Equal(t, tt.expected, result, "PrivateBaseURL should match")
		})
	}
}

// TestPublicBaseURL tests the PublicBaseURL method.
func TestPublicBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings ServiceTemplateServerSettings
		expected string
	}{
		{
			name: "https localhost 8080",
			settings: ServiceTemplateServerSettings{
				BindPublicProtocol: "https",
				BindPublicAddress:  "localhost",
				BindPublicPort:     8080,
			},
			expected: "https://localhost:8080",
		},
		{
			name: "http 0.0.0.0 3000",
			settings: ServiceTemplateServerSettings{
				BindPublicProtocol: "http",
				BindPublicAddress:  "0.0.0.0",
				BindPublicPort:     3000,
			},
			expected: "http://0.0.0.0:3000",
		},
		{
			name: "https IPv6 443",
			settings: ServiceTemplateServerSettings{
				BindPublicProtocol: "https",
				BindPublicAddress:  "[::]",
				BindPublicPort:     443,
			},
			expected: "https://[::]:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.settings.PublicBaseURL()
			require.Equal(t, tt.expected, result, "PublicBaseURL should match")
		})
	}
}

// TestRequireNewForTest tests the test utility function.
func TestRequireNewForTest(t *testing.T) {
	t.Parallel()

	// Call the test utility function.
	settings := RequireNewForTest("test-app")

	// Verify it returns a valid settings object.
	require.NotNil(t, settings, "Settings should not be nil")

	// Verify some expected fields are set (may be defaults or zeros).
	require.NotEmpty(t, settings.LogLevel, "LogLevel should be set")
	// Note: Ports may be zero in test settings, so just verify structure exists.
}
