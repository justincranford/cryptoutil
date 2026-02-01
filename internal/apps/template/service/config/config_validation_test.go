// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// validBaseSettings returns a valid base configuration for testing.
// Test cases override specific fields to test validation.
func validBaseSettings() *ServiceTemplateServerSettings {
	return &ServiceTemplateServerSettings{
		DevMode:             false,
		BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
		BindPublicPort:      8080,
		BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
		BindPrivatePort:     9090,
		BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
		BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
		LogLevel:            "INFO",
		DatabaseURL:         "sqlite://file::memory:",
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
		BrowserIPRateLimit:  100,
		ServiceIPRateLimit:  100,
		OTLPEndpoint:        "http://localhost:4317",
	}
}

// TestValidateConfiguration tests all configuration validation scenarios.
func TestValidateConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		modify           func(*ServiceTemplateServerSettings)
		wantErrorMessage string
	}{
		// Address validation (dev mode).
		{
			name: "reject 0.0.0.0 public address in dev mode",
			modify: func(s *ServiceTemplateServerSettings) {
				s.DevMode = true
				s.BindPublicAddress = cryptoutilSharedMagic.IPv4AnyAddress
			},
			wantErrorMessage: "CRITICAL: bind public address cannot be 0.0.0.0 in test/dev mode",
		},
		{
			name: "reject 0.0.0.0 private address in dev mode",
			modify: func(s *ServiceTemplateServerSettings) {
				s.DevMode = true
				s.BindPrivateAddress = cryptoutilSharedMagic.IPv4AnyAddress
			},
			wantErrorMessage: "CRITICAL: bind private address cannot be 0.0.0.0 in test/dev mode",
		},
		{
			name: "allow 0.0.0.0 public address in prod mode (containers)",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicAddress = cryptoutilSharedMagic.IPv4AnyAddress
			},
			wantErrorMessage: "",
		},
		{
			name: "blank public address",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicAddress = ""
			},
			wantErrorMessage: "bind public address cannot be blank",
		},
		{
			name: "blank private address",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPrivateAddress = ""
			},
			wantErrorMessage: "bind private address cannot be blank",
		},
		// Valid production configurations.
		{
			name: "valid production PostgreSQL config",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicAddress = "192.168.1.100"
				s.BindPublicPort = 8443
				s.DatabaseURL = "postgres://user:pass@db.example.com:5432/production"
				s.TLSPublicDNSNames = []string{"api.example.com"}
				s.OTLPEndpoint = "http://otel-collector:4317"
			},
			wantErrorMessage: "",
		},
		// Config format validation.
		{
			name: "invalid database URL format (missing ://)",
			modify: func(s *ServiceTemplateServerSettings) {
				s.DatabaseURL = "invalid-format-no-scheme"
			},
			wantErrorMessage: "invalid database URL format",
		},
		{
			name: "invalid log level",
			modify: func(s *ServiceTemplateServerSettings) {
				s.LogLevel = "INVALID_LEVEL"
			},
			wantErrorMessage: "invalid log level 'INVALID_LEVEL'",
		},
		{
			name: "invalid CORS origin format (missing scheme)",
			modify: func(s *ServiceTemplateServerSettings) {
				s.CORSAllowedOrigins = []string{"invalid-origin-no-scheme"}
			},
			wantErrorMessage: "invalid CORS origin format",
		},
		{
			name: "invalid OTLP endpoint format (missing scheme)",
			modify: func(s *ServiceTemplateServerSettings) {
				s.OTLPEnabled = true
				s.OTLPEndpoint = "invalid-endpoint-no-scheme"
			},
			wantErrorMessage: "invalid OTLP endpoint format",
		},
		// Port edge cases.
		{
			name: "both ports 0 is valid (dynamic allocation)",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicPort = 0
				s.BindPrivatePort = 0
			},
			wantErrorMessage: "",
		},
		{
			name: "same non-zero ports rejected",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicPort = 8080
				s.BindPrivatePort = 8080
			},
			wantErrorMessage: "public port (8080) and private port (8080) cannot be the same",
		},
		{
			name: "public port 0 with non-zero private port is valid",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicPort = 0
				s.BindPrivatePort = 9090
			},
			wantErrorMessage: "",
		},
		// Protocol validation.
		{
			name: "invalid public protocol",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicProtocol = "ftp"
			},
			wantErrorMessage: "invalid public protocol 'ftp'",
		},
		{
			name: "invalid private protocol",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPrivateProtocol = "gopher"
			},
			wantErrorMessage: "invalid private protocol 'gopher'",
		},
		{
			name: "http protocol is valid",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicProtocol = cryptoutilSharedMagic.ProtocolHTTP
				s.BindPrivateProtocol = cryptoutilSharedMagic.ProtocolHTTP
			},
			wantErrorMessage: "",
		},
		// Rate limit validation.
		{
			name: "zero browser rate limit rejected",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BrowserIPRateLimit = 0
			},
			wantErrorMessage: "browser rate limit cannot be 0",
		},
		{
			name: "zero service rate limit rejected",
			modify: func(s *ServiceTemplateServerSettings) {
				s.ServiceIPRateLimit = 0
			},
			wantErrorMessage: "service rate limit cannot be 0",
		},
		{
			name: "very high browser rate limit warning",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BrowserIPRateLimit = 65000
			},
			wantErrorMessage: "browser rate limit 65000 is very high",
		},
		{
			name: "very high service rate limit warning",
			modify: func(s *ServiceTemplateServerSettings) {
				s.ServiceIPRateLimit = 65000
			},
			wantErrorMessage: "service rate limit 65000 is very high",
		},
		// HTTPS without TLS config.
		{
			name: "HTTPS public without TLS config",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicProtocol = cryptoutilSharedMagic.ProtocolHTTPS
				s.BindPrivateProtocol = cryptoutilSharedMagic.ProtocolHTTP
				s.TLSPublicDNSNames = nil
				s.TLSPublicIPAddresses = nil
			},
			wantErrorMessage: "HTTPS public protocol requires TLS DNS names or IP addresses",
		},
		{
			name: "HTTPS private without TLS config",
			modify: func(s *ServiceTemplateServerSettings) {
				s.BindPublicProtocol = cryptoutilSharedMagic.ProtocolHTTP
				s.BindPrivateProtocol = cryptoutilSharedMagic.ProtocolHTTPS
				s.TLSPrivateDNSNames = nil
				s.TLSPrivateIPAddresses = nil
			},
			wantErrorMessage: "HTTPS private protocol requires TLS DNS names or IP addresses",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := validBaseSettings()
			tc.modify(s)
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

// TestNewTestConfig tests that NewTestConfig helper works correctly.
func TestNewTestConfig(t *testing.T) {
	t.Parallel()

	s := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	require.NotNil(t, s)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, s.BindPublicAddress)
}
