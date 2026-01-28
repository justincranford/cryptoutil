// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidateConfiguration_AddressValidation tests address validation (0.0.0.0, blank) in different modes.
func TestValidateConfiguration_AddressValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		publicAddress    string
		privateAddress   string
		devMode          bool
		wantErrorMessage string
	}{
		{
			name:             "reject 0.0.0.0 public address in dev mode",
			publicAddress:    "0.0.0.0",
			privateAddress:   "127.0.0.1",
			devMode:          true,
			wantErrorMessage: "CRITICAL: bind public address cannot be 0.0.0.0 in test/dev mode",
		},
		{
			name:             "reject 0.0.0.0 private address in dev mode",
			publicAddress:    "127.0.0.1",
			privateAddress:   "0.0.0.0",
			devMode:          true,
			wantErrorMessage: "CRITICAL: bind private address cannot be 0.0.0.0 in test/dev mode",
		},
		{
			name:             "allow 0.0.0.0 public address in prod mode (containers)",
			publicAddress:    "0.0.0.0",
			privateAddress:   "127.0.0.1",
			devMode:          false,
			wantErrorMessage: "",
		},
		{
			name:             "blank public address",
			publicAddress:    "",
			privateAddress:   "127.0.0.1",
			devMode:          false,
			wantErrorMessage: "bind public address cannot be blank",
		},
		{
			name:             "blank private address",
			publicAddress:    "127.0.0.1",
			privateAddress:   "",
			devMode:          false,
			wantErrorMessage: "bind private address cannot be blank",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &ServiceTemplateServerSettings{
				DevMode:             tc.devMode,
				BindPublicAddress:   tc.publicAddress,
				BindPublicPort:      8080,
				BindPrivateAddress:  tc.privateAddress,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				LogLevel:            "INFO",
				DatabaseURL:         "postgres://user:pass@localhost:5432/db",
				TLSPublicDNSNames:   []string{"localhost"},
				TLSPrivateDNSNames:  []string{"localhost"},
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
				OTLPEndpoint:        "http://localhost:4317",
			}

			err := validateConfiguration(s)

			if tc.wantErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrorMessage)
			} else {
				require.NoError(t, err, "Should allow 0.0.0.0 public address in production mode (for containers)")
			}
		})
	}
}

// TestValidateConfiguration_Reject127InTestHelper tests that NewTestConfig accepts 127.0.0.1.
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

// TestValidateConfiguration_ConfigFormatValidation tests various config format validations.
func TestValidateConfiguration_ConfigFormatValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		databaseURL      string
		logLevel         string
		corsOrigins      []string
		otlpEndpoint     string
		otlpEnabled      bool
		wantErrorMessage string
	}{
		{
			name:             "invalid database URL format (missing ://)",
			databaseURL:      "invalid-format-no-scheme",
			logLevel:         "INFO",
			corsOrigins:      []string{},
			otlpEndpoint:     "http://localhost:4317",
			otlpEnabled:      false,
			wantErrorMessage: "invalid database URL format",
		},
		{
			name:             "invalid log level",
			databaseURL:      "sqlite://file::memory:",
			logLevel:         "INVALID_LEVEL",
			corsOrigins:      []string{},
			otlpEndpoint:     "http://localhost:4317",
			otlpEnabled:      false,
			wantErrorMessage: "invalid log level 'INVALID_LEVEL'",
		},
		{
			name:             "invalid CORS origin format (missing scheme)",
			databaseURL:      "sqlite://file::memory:",
			logLevel:         "INFO",
			corsOrigins:      []string{"invalid-origin-no-scheme"},
			otlpEndpoint:     "http://localhost:4317",
			otlpEnabled:      false,
			wantErrorMessage: "invalid CORS origin format",
		},
		{
			name:             "invalid OTLP endpoint format (missing scheme)",
			databaseURL:      "sqlite://file::memory:",
			logLevel:         "INFO",
			corsOrigins:      []string{},
			otlpEndpoint:     "invalid-endpoint-no-scheme",
			otlpEnabled:      true,
			wantErrorMessage: "invalid OTLP endpoint format",
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
				LogLevel:            tc.logLevel,
				DatabaseURL:         tc.databaseURL,
				TLSPublicDNSNames:   []string{"localhost"},
				TLSPrivateDNSNames:  []string{"localhost"},
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
				OTLPEndpoint:        tc.otlpEndpoint,
				OTLPEnabled:         tc.otlpEnabled,
				CORSAllowedOrigins:  tc.corsOrigins,
			}

			err := validateConfiguration(s)
			require.Error(t, err, "Test case %s should fail", tc.name)
			require.Contains(t, err.Error(), tc.wantErrorMessage)
		})
	}
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
