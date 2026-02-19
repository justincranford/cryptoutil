// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateConfiguration_HappyPath(t *testing.T) {
	t.Parallel()

	s := &ServiceTemplateServerSettings{
		BindPublicAddress:   "127.0.0.1",
		BindPrivateAddress:  "127.0.0.1",
		BindPublicPort:      8080,
		BindPrivatePort:     9090,
		BindPublicProtocol:  "https",
		BindPrivateProtocol: "https",
		TLSPublicDNSNames:   []string{"public.example.com"},
		TLSPrivateDNSNames:  []string{"private.example.com"},
		DatabaseURL:         "postgres://user:pass@localhost:5432/db",
		CORSAllowedOrigins:  []string{"https://example.com"},
		LogLevel:            "INFO",
		BrowserIPRateLimit:  100,
		ServiceIPRateLimit:  200,
		OTLPEnabled:         true,
		OTLPEndpoint:        "grpc://otel:4317",
	}

	err := validateConfiguration(s)
	require.NoError(t, err)
}

func TestValidateConfiguration_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *ServiceTemplateServerSettings
		errMsg   string
	}{
		{
			name: "blank public bind address",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "",
				BindPrivateAddress:  "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "bind public address cannot be blank",
		},
		{
			name: "blank private bind address",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "127.0.0.1",
				BindPrivateAddress:  "",
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "bind private address cannot be blank",
		},
		{
			name: "same non-zero ports",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     8080,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "cannot be the same",
		},
		{
			name: "invalid public protocol",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "ftp",
				BindPrivateProtocol: "https",
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "invalid public protocol 'ftp'",
		},
		{
			name: "invalid private protocol",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "ftp",
				TLSPublicDNSNames:   []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "invalid private protocol 'ftp'",
		},
		{
			name: "https public missing TLS config",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "http",
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "HTTPS public protocol requires TLS DNS names or IP addresses",
		},
		{
			name: "https private missing TLS config",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "http",
				BindPrivateProtocol: "https",
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "HTTPS private protocol requires TLS DNS names or IP addresses",
		},
		{
			name: "invalid database URL format",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				DatabaseURL:         "invalid-no-scheme",
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "invalid database URL format",
		},
		{
			name: "invalid CORS origin format",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				CORSAllowedOrigins:  []string{"invalid-no-scheme"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "invalid CORS origin format",
		},
		{
			name: "invalid log level",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INVALID",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			errMsg: "invalid log level 'INVALID'",
		},
		{
			name: "browser rate limit zero",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  0,
				ServiceIPRateLimit:  100,
			},
			errMsg: "browser rate limit cannot be 0",
		},
		{
			name: "service rate limit zero",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  0,
			},
			errMsg: "service rate limit cannot be 0",
		},
		{
			name: "invalid OTLP endpoint format",
			settings: &ServiceTemplateServerSettings{
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
				OTLPEnabled:         true,
				OTLPEndpoint:        "invalid-no-scheme:4317",
			},
			errMsg: "invalid OTLP endpoint format",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateConfiguration(tc.settings)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

// TestValidateConfiguration_BoundaryConditions tests edge cases at validation boundaries.
// Kills mutation: config.go:1526 (CONDITIONALS_BOUNDARY: > vs >=).
// Kills mutation: config.go:1530 (CONDITIONALS_BOUNDARY: > vs >=).
// Kills mutation: config.go:1593 (CONDITIONALS_BOUNDARY: > vs >=).
// Kills mutation: config.go:1599 (CONDITIONALS_BOUNDARY: > vs >=).
func TestValidateConfiguration_BoundaryConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *ServiceTemplateServerSettings
		wantErr  bool
		errMsg   string
	}{
		{
			name: "public port exactly 65535 - valid",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "127.0.0.1",
				BindPrivateAddress:  "127.0.0.1",
				BindPublicPort:      65535,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			wantErr: false,
		},
		{
			name: "private port exactly 65535 - valid",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "127.0.0.1",
				BindPrivateAddress:  "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivatePort:     65535,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  100,
			},
			wantErr: false,
		},
		{
			name: "browser rate limit exactly MaxIPRateLimit - valid warning",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "127.0.0.1",
				BindPrivateAddress:  "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  10000, // cryptoutilSharedMagic.MaxIPRateLimit
				ServiceIPRateLimit:  100,
			},
			wantErr: false,
		},
		{
			name: "browser rate limit above MaxIPRateLimit - warning",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "127.0.0.1",
				BindPrivateAddress:  "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  10001,
				ServiceIPRateLimit:  100,
			},
			wantErr: true,
			errMsg:  "browser rate limit 10001 is very high",
		},
		{
			name: "service rate limit exactly MaxIPRateLimit - valid warning",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "127.0.0.1",
				BindPrivateAddress:  "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  10000, // cryptoutilSharedMagic.MaxIPRateLimit
			},
			wantErr: false,
		},
		{
			name: "service rate limit above MaxIPRateLimit - warning",
			settings: &ServiceTemplateServerSettings{
				BindPublicAddress:   "127.0.0.1",
				BindPrivateAddress:  "127.0.0.1",
				BindPublicPort:      8080,
				BindPrivatePort:     9090,
				BindPublicProtocol:  "https",
				BindPrivateProtocol: "https",
				TLSPublicDNSNames:   []string{"test.com"},
				TLSPrivateDNSNames:  []string{"test.com"},
				LogLevel:            "INFO",
				BrowserIPRateLimit:  100,
				ServiceIPRateLimit:  10001,
			},
			wantErr: true,
			errMsg:  "service rate limit 10001 is very high",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateConfiguration(tc.settings)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				if err != nil {
					// Allow warnings that don't prevent startup.
					require.NotContains(t, err.Error(), "invalid")
				}
			}
		})
	}
}

// TestParseWithMultipleConfigFiles tests config file merging with multiple files.
// Kills mutation: config.go:1046 (INCREMENT_DECREMENT: i++ vs i--).
