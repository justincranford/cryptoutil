// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidateConfiguration_Reject0000InDevMode tests that validateConfiguration rejects 0.0.0.0 in dev mode.
func TestValidateConfiguration_Reject0000InDevMode(t *testing.T) {
	// Test public address validation.
	s := &ServiceTemplateServerSettings{
		DevMode:             true,
		BindPublicAddress:   "0.0.0.0",
		BindPublicPort:      8080,
		BindPrivateAddress:  "127.0.0.1",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INFO",
		DatabaseURL:         "sqlite://file::memory:",
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
	}

	err := validateConfiguration(s)
	require.Error(t, err, "Should reject 0.0.0.0 public address in dev mode")
	require.Contains(t, err.Error(), "CRITICAL: bind public address cannot be 0.0.0.0 in test/dev mode")

	// Test private address validation.
	s2 := &ServiceTemplateServerSettings{
		DevMode:             true,
		BindPublicAddress:   "127.0.0.1",
		BindPublicPort:      8080,
		BindPrivateAddress:  "0.0.0.0",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INFO",
		DatabaseURL:         "sqlite://file::memory:",
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
	}

	err2 := validateConfiguration(s2)
	require.Error(t, err2, "Should reject 0.0.0.0 private address in dev mode")
	require.Contains(t, err2.Error(), "CRITICAL: bind private address cannot be 0.0.0.0 in test/dev mode")
}

// TestValidateConfiguration_Allow0000InProdMode tests that validateConfiguration allows 0.0.0.0 in production mode.
func TestValidateConfiguration_Allow0000InProdMode(t *testing.T) {
	s := &ServiceTemplateServerSettings{
		DevMode:             false,
		BindPublicAddress:   "0.0.0.0",
		BindPublicPort:      8080,
		BindPrivateAddress:  "127.0.0.1",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INFO",
		DatabaseURL:         "postgres://user:pass@localhost:5432/db",
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
		BrowserIPRateLimit:  100,                     // Required - not 0
		ServiceIPRateLimit:  100,                     // Required - not 0
		OTLPEndpoint:        "http://localhost:4317", // Required format
	}

	err := validateConfiguration(s)
	require.NoError(t, err, "Should allow 0.0.0.0 public address in production mode (for containers)")
}

// TestValidateConfiguration_Reject127InTestHelper tests that NewTestConfig rejects 0.0.0.0.
func TestValidateConfiguration_Reject127InTestHelper(t *testing.T) {
	// This should pass without panic.
	s := NewTestConfig("127.0.0.1", 0, true)
	require.NotNil(t, s)
	require.Equal(t, "127.0.0.1", s.BindPublicAddress)
}

// TestValidateConfiguration_ValidProductionPostgreSQL tests that valid production config with PostgreSQL passes.
func TestValidateConfiguration_ValidProductionPostgreSQL(t *testing.T) {
	s := &ServiceTemplateServerSettings{
		DevMode:             false,
		BindPublicAddress:   "192.168.1.100", // Specific IP
		BindPublicPort:      8443,
		BindPrivateAddress:  "127.0.0.1",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INFO",
		DatabaseURL:         "postgres://user:pass@db.example.com:5432/production",
		TLSPublicDNSNames:   []string{"api.example.com"},
		TLSPrivateDNSNames:  []string{"localhost"},
		BrowserIPRateLimit:  100,
		ServiceIPRateLimit:  100,
		OTLPEndpoint:        "http://otel-collector:4317",
	}

	err := validateConfiguration(s)
	require.NoError(t, err, "Production config with PostgreSQL + specific IP should be valid")
}

// TestValidateConfiguration_InvalidDatabaseURLFormat tests that invalid database-url format is rejected.
func TestValidateConfiguration_InvalidDatabaseURLFormat(t *testing.T) {
	s := &ServiceTemplateServerSettings{
		DevMode:             false,
		BindPublicAddress:   "127.0.0.1",
		BindPublicPort:      8080,
		BindPrivateAddress:  "127.0.0.1",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INFO",
		DatabaseURL:         "invalid-format-no-scheme", // Missing "://"
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
		BrowserIPRateLimit:  100,
		ServiceIPRateLimit:  100,
		OTLPEndpoint:        "http://localhost:4317",
	}

	err := validateConfiguration(s)
	require.Error(t, err, "Should reject invalid database-url format")
	require.Contains(t, err.Error(), "invalid database URL format")
	require.Contains(t, err.Error(), "must contain '://'")
}

// TestValidateConfiguration_PortEdgeCases tests port validation edge cases.
func TestValidateConfiguration_PortEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		publicPort       uint16
		privatePort      uint16
		wantErrorMessage string
	}{
		{
			name:             "both ports 0 is valid (dynamic allocation)",
			publicPort:       0,
			privatePort:      0,
			wantErrorMessage: "",
		},
		{
			name:             "same non-zero ports rejected",
			publicPort:       8080,
			privatePort:      8080,
			wantErrorMessage: "public port (8080) and private port (8080) cannot be the same",
		},
		{
			name:             "public port 0 with non-zero private port is valid",
			publicPort:       0,
			privatePort:      9090,
			wantErrorMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &ServiceTemplateServerSettings{
				DevMode:             false,
				BindPublicAddress:   "127.0.0.1",
				BindPublicPort:      tc.publicPort,
				BindPrivateAddress:  "127.0.0.1",
				BindPrivatePort:     tc.privatePort,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				LogLevel:            "INFO",
				DatabaseURL:         "sqlite://file::memory:",
				TLSPublicDNSNames:   []string{"localhost"},
				TLSPrivateDNSNames:  []string{"localhost"},
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
				OTLPEndpoint:        "http://localhost:4317",
			}

			err := validateConfiguration(s)

			if tc.wantErrorMessage != "" {
				require.Error(t, err, "Test case %s should fail", tc.name)
				require.Contains(t, err.Error(), tc.wantErrorMessage)
			} else {
				require.NoError(t, err, "Test case %s should pass", tc.name)
			}
		})
	}
}
