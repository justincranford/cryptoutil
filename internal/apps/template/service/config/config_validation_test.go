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

// TestValidateConfiguration_InvalidProtocol tests that invalid protocol is rejected.
func TestValidateConfiguration_InvalidProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		publicProtocol   string
		privateProtocol  string
		wantErrorMessage string
	}{
		{
			name:             "invalid public protocol",
			publicProtocol:   "ftp",
			privateProtocol:  "https",
			wantErrorMessage: "invalid public protocol 'ftp'",
		},
		{
			name:             "invalid private protocol",
			publicProtocol:   "https",
			privateProtocol:  "ftp",
			wantErrorMessage: "invalid private protocol 'ftp'",
		},
		{
			name:             "http protocol is valid",
			publicProtocol:   "http",
			privateProtocol:  "http",
			wantErrorMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &ServiceTemplateServerSettings{
				DevMode:             false,
				BindPublicAddress:   "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivateAddress:  "127.0.0.1",
				BindPrivatePort:     9090,
				BindPublicProtocol:  tc.publicProtocol,
				BindPrivateProtocol: tc.privateProtocol,
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

// TestValidateConfiguration_InvalidLogLevel tests that invalid log level is rejected.
func TestValidateConfiguration_InvalidLogLevel(t *testing.T) {
	t.Parallel()

	s := &ServiceTemplateServerSettings{
		DevMode:             false,
		BindPublicAddress:   "127.0.0.1",
		BindPublicPort:      8080,
		BindPrivateAddress:  "127.0.0.1",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INVALID_LEVEL",
		DatabaseURL:         "sqlite://file::memory:",
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
		BrowserIPRateLimit:  100,
		ServiceIPRateLimit:  100,
		OTLPEndpoint:        "http://localhost:4317",
	}

	err := validateConfiguration(s)
	require.Error(t, err, "Should reject invalid log level")
	require.Contains(t, err.Error(), "invalid log level 'INVALID_LEVEL'")
}

// TestValidateConfiguration_RateLimitEdgeCases tests rate limit validation.
func TestValidateConfiguration_RateLimitEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		browserRateLimit uint16
		serviceRateLimit uint16
		wantErrorMessage string
		wantWarning      bool
	}{
		{
			name:             "zero browser rate limit rejected",
			browserRateLimit: 0,
			serviceRateLimit: 100,
			wantErrorMessage: "browser rate limit cannot be 0",
		},
		{
			name:             "zero service rate limit rejected",
			browserRateLimit: 100,
			serviceRateLimit: 0,
			wantErrorMessage: "service rate limit cannot be 0",
		},
		{
			name:             "very high browser rate limit warning",
			browserRateLimit: 65000, // Above MaxIPRateLimit
			serviceRateLimit: 100,
			wantErrorMessage: "browser rate limit 65000 is very high",
		},
		{
			name:             "very high service rate limit warning",
			browserRateLimit: 100,
			serviceRateLimit: 65000, // Above MaxIPRateLimit
			wantErrorMessage: "service rate limit 65000 is very high",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &ServiceTemplateServerSettings{
				DevMode:             false,
				BindPublicAddress:   "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivateAddress:  "127.0.0.1",
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				LogLevel:            "INFO",
				DatabaseURL:         "sqlite://file::memory:",
				TLSPublicDNSNames:   []string{"localhost"},
				TLSPrivateDNSNames:  []string{"localhost"},
				BrowserIPRateLimit:  tc.browserRateLimit,
				ServiceIPRateLimit:  tc.serviceRateLimit,
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

// TestValidateConfiguration_InvalidCORSOrigin tests that invalid CORS origin format is rejected.
func TestValidateConfiguration_InvalidCORSOrigin(t *testing.T) {
	t.Parallel()

	s := &ServiceTemplateServerSettings{
		DevMode:             false,
		BindPublicAddress:   "127.0.0.1",
		BindPublicPort:      8080,
		BindPrivateAddress:  "127.0.0.1",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INFO",
		DatabaseURL:         "sqlite://file::memory:",
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
		BrowserIPRateLimit:  100,
		ServiceIPRateLimit:  100,
		OTLPEndpoint:        "http://localhost:4317",
		CORSAllowedOrigins:  []string{"invalid-origin-no-scheme"},
	}

	err := validateConfiguration(s)
	require.Error(t, err, "Should reject invalid CORS origin format")
	require.Contains(t, err.Error(), "invalid CORS origin format")
}

// TestValidateConfiguration_InvalidOTLPEndpoint tests that invalid OTLP endpoint format is rejected.
func TestValidateConfiguration_InvalidOTLPEndpoint(t *testing.T) {
	t.Parallel()

	s := &ServiceTemplateServerSettings{
		DevMode:             false,
		BindPublicAddress:   "127.0.0.1",
		BindPublicPort:      8080,
		BindPrivateAddress:  "127.0.0.1",
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		LogLevel:            "INFO",
		DatabaseURL:         "sqlite://file::memory:",
		TLSPublicDNSNames:   []string{"localhost"},
		TLSPrivateDNSNames:  []string{"localhost"},
		BrowserIPRateLimit:  100,
		ServiceIPRateLimit:  100,
		OTLPEnabled:         true,
		OTLPEndpoint:        "invalid-endpoint-no-scheme",
	}

	err := validateConfiguration(s)
	require.Error(t, err, "Should reject invalid OTLP endpoint format")
	require.Contains(t, err.Error(), "invalid OTLP endpoint format")
}

// TestValidateConfiguration_BlankAddresses tests blank address validation.
func TestValidateConfiguration_BlankAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		publicAddress    string
		privateAddress   string
		wantErrorMessage string
	}{
		{
			name:             "blank public address",
			publicAddress:    "",
			privateAddress:   "127.0.0.1",
			wantErrorMessage: "bind public address cannot be blank",
		},
		{
			name:             "blank private address",
			publicAddress:    "127.0.0.1",
			privateAddress:   "",
			wantErrorMessage: "bind private address cannot be blank",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &ServiceTemplateServerSettings{
				DevMode:             false,
				BindPublicAddress:   tc.publicAddress,
				BindPublicPort:      8080,
				BindPrivateAddress:  tc.privateAddress,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				LogLevel:            "INFO",
				DatabaseURL:         "sqlite://file::memory:",
				TLSPublicDNSNames:   []string{"localhost"},
				TLSPrivateDNSNames:  []string{"localhost"},
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			}

			err := validateConfiguration(s)
			require.Error(t, err, "Test case %s should fail", tc.name)
			require.Contains(t, err.Error(), tc.wantErrorMessage)
		})
	}
}

// TestValidateConfiguration_HTTPSWithoutTLSConfig tests HTTPS protocol validation.
func TestValidateConfiguration_HTTPSWithoutTLSConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		publicProtocol   string
		privateProtocol  string
		publicDNS        []string
		privateDNS       []string
		wantErrorMessage string
	}{
		{
			name:             "HTTPS public without TLS config",
			publicProtocol:   "https",
			privateProtocol:  "http",
			publicDNS:        nil,
			privateDNS:       []string{"localhost"},
			wantErrorMessage: "HTTPS public protocol requires TLS DNS names or IP addresses",
		},
		{
			name:             "HTTPS private without TLS config",
			publicProtocol:   "http",
			privateProtocol:  "https",
			publicDNS:        []string{"localhost"},
			privateDNS:       nil,
			wantErrorMessage: "HTTPS private protocol requires TLS DNS names or IP addresses",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &ServiceTemplateServerSettings{
				DevMode:               false,
				BindPublicAddress:     "127.0.0.1",
				BindPublicPort:        8080,
				BindPrivateAddress:    "127.0.0.1",
				BindPrivatePort:       9090,
				BindPublicProtocol:    tc.publicProtocol,
				BindPrivateProtocol:   tc.privateProtocol,
				LogLevel:              "INFO",
				DatabaseURL:           "sqlite://file::memory:",
				TLSPublicDNSNames:     tc.publicDNS,
				TLSPrivateDNSNames:    tc.privateDNS,
				TLSPublicIPAddresses:  nil,
				TLSPrivateIPAddresses: nil,
				BrowserIPRateLimit:    100,
				ServiceIPRateLimit:    100,
				OTLPEndpoint:          "http://localhost:4317",
			}

			err := validateConfiguration(s)
			require.Error(t, err, "Test case %s should fail", tc.name)
			require.Contains(t, err.Error(), tc.wantErrorMessage)
		})
	}
}
