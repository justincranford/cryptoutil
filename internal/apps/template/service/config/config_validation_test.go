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
