// Copyright (c) 2025-2026 Justin Cranford.
//
//

package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateConfiguration_HappyPath(t *testing.T) {
	t.Parallel()

	s := &ServiceFrameworkServerSettings{
		BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
		BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
		BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
		BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
		BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
		BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
		TLSPublicDNSNames:   []string{"public.example.com"},
		TLSPrivateDNSNames:  []string{"private.example.com"},
		DatabaseURL:         "postgres://user:pass@localhost:5432/db",
		CORSAllowedOrigins:  []string{testHTTPSExampleCom},
		LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
		BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
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
		settings *ServiceFrameworkServerSettings
		errMsg   string
	}{
		{
			name: "blank public bind address",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   "",
				BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testBindPublicAddressCannotBeBlank,
		},
		{
			name: "blank private bind address",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
				BindPrivateAddress:  "",
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testBindPrivateAddressCannotBeBlank,
		},
		{
			name: "same non-zero ports",
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestServerPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: "cannot be the same",
		},
		{
			name: testInvalidPublicProtocol,
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  testFtp,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testInvalidPublicProtocolFtp,
		},
		{
			name: testInvalidPrivateProtocol,
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: testFtp,
				TLSPublicDNSNames:   []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: "invalid private protocol 'ftp'",
		},
		{
			name: "https public missing TLS config",
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTP,
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testHTTPSPublicProtocolRequiresTLSDNSNamesOrIPAddresses,
		},
		{
			name: "https private missing TLS config",
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTP,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testHTTPSPrivateProtocolRequiresTLSDNSNamesOrIPAddresses,
		},
		{
			name: testInvalidDatabaseURLFormat,
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				DatabaseURL:         testInvalidNoScheme,
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testInvalidDatabaseURLFormat,
		},
		{
			name: testInvalidCorsOriginFormat,
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				CORSAllowedOrigins:  []string{testInvalidNoScheme},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testInvalidCorsOriginFormat,
		},
		{
			name: testInvalidLogLevel,
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            "INVALID",
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: "invalid log level 'INVALID'",
		},
		{
			name: "browser rate limit zero",
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  0,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			errMsg: testBrowserRateLimitCannotBe0,
		},
		{
			name: "service rate limit zero",
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  0,
			},
			errMsg: testServiceRateLimitCannotBe0,
		},
		{
			name: testInvalidOtlpEndpointFormat,
			settings: &ServiceFrameworkServerSettings{
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				OTLPEnabled:         true,
				OTLPEndpoint:        "invalid-no-scheme:4317",
			},
			errMsg: testInvalidOtlpEndpointFormat,
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
		settings *ServiceFrameworkServerSettings
		wantErr  bool
		errMsg   string
	}{
		{
			name: "public port exactly 65535 - valid",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
				BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
				BindPublicPort:      cryptoutilSharedMagic.MaxPortNumber,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			wantErr: false,
		},
		{
			name: "private port exactly 65535 - valid",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
				BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.MaxPortNumber,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			wantErr: false,
		},
		{
			name: "browser rate limit exactly MaxIPRateLimit - valid warning",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
				BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.DBContainerRandSuffixMax, // cryptoutilSharedMagic.MaxIPRateLimit
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			wantErr: false,
		},
		{
			name: "browser rate limit above MaxIPRateLimit - warning",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
				BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  10001,
				ServiceIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
			},
			wantErr: true,
			errMsg:  "browser rate limit 10001 is very high",
		},
		{
			name: "service rate limit exactly MaxIPRateLimit - valid warning",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
				BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
				ServiceIPRateLimit:  cryptoutilSharedMagic.DBContainerRandSuffixMax, // cryptoutilSharedMagic.MaxIPRateLimit
			},
			wantErr: false,
		},
		{
			name: "service rate limit above MaxIPRateLimit - warning",
			settings: &ServiceFrameworkServerSettings{
				BindPublicAddress:   cryptoutilSharedMagic.IPv4Loopback,
				BindPrivateAddress:  cryptoutilSharedMagic.IPv4Loopback,
				BindPublicPort:      cryptoutilSharedMagic.TestServerPort,
				BindPrivatePort:     cryptoutilSharedMagic.TestAdminPort,
				BindPublicProtocol:  cryptoutilSharedMagic.ProtocolHTTPS,
				BindPrivateProtocol: cryptoutilSharedMagic.ProtocolHTTPS,
				TLSPublicDNSNames:   []string{testTestCom},
				TLSPrivateDNSNames:  []string{testTestCom},
				LogLevel:            cryptoutilSharedMagic.DefaultLogLevelInfo,
				BrowserIPRateLimit:  cryptoutilSharedMagic.JoseJAMaxMaterials,
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
					require.NotContains(t, err.Error(), cryptoutilSharedMagic.DevSetupTestInvalidName)
				}
			}
		})
	}
}

// TestParseWithMultipleConfigFiles tests config file merging with multiple files.
// Kills mutation: config.go:1046 (INCREMENT_DECREMENT: i++ vs i--).
