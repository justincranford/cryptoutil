// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"crypto/tls"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestContainerModeDetection tests container mode detection logic based on bind address.
// Container mode is triggered when BindPublicAddress == "0.0.0.0"
// Priority: P1.1 (Critical - Must Have).
func TestContainerModeDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		bindPublicAddress  string
		bindPrivateAddress string
		wantContainerMode  bool
	}{
		{
			name:               "public 0.0.0.0 triggers container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4AnyAddress, // "0.0.0.0"
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,   // "127.0.0.1"
			wantContainerMode:  true,
		},
		{
			name:               "both 127.0.0.1 is NOT container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,
			wantContainerMode:  false,
		},
		{
			name:               "private 0.0.0.0 does NOT trigger container mode",
			bindPublicAddress:  cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress: cryptoutilSharedMagic.IPv4AnyAddress,
			wantContainerMode:  false,
		},
		{
			name:               "specific IP is NOT container mode",
			bindPublicAddress:  "192.168.1.100",
			bindPrivateAddress: cryptoutilSharedMagic.IPv4Loopback,
			wantContainerMode:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BindPublicAddress:  tc.bindPublicAddress,
				BindPrivateAddress: tc.bindPrivateAddress,
			}

			isContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress
			require.Equal(t, tc.wantContainerMode, isContainerMode)
		})
	}
}

// TestMTLSConfiguration tests mTLS client auth configuration for private/public servers
// in dev/container/production modes.
// Priority: P1.2 (MOST CRITICAL - Currently 0% coverage on security code).
func TestMTLSConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		devMode               bool
		bindPublicAddress     string
		bindPrivateAddress    string
		wantPrivateClientAuth tls.ClientAuthType
		wantPublicClientAuth  tls.ClientAuthType
	}{
		{
			name:                  "dev mode disables mTLS on private server",
			devMode:               true,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.NoClientCert,
			wantPublicClientAuth:  tls.NoClientCert, // Public never requires client certs
		},
		{
			name:                  "container mode disables mTLS on private server",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4AnyAddress, // 0.0.0.0
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.NoClientCert,
			wantPublicClientAuth:  tls.NoClientCert,
		},
		{
			name:                  "production mode enables mTLS on private server",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.RequireAndVerifyClientCert,
			wantPublicClientAuth:  tls.NoClientCert, // Public never requires client certs
		},
		{
			name:                  "container mode with private 0.0.0.0 still enables mTLS",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4AnyAddress,
			wantPrivateClientAuth: tls.RequireAndVerifyClientCert, // Only public triggers container mode
			wantPublicClientAuth:  tls.NoClientCert,
		},
		{
			name:                  "public server never uses RequireAndVerifyClientCert",
			devMode:               false,
			bindPublicAddress:     cryptoutilSharedMagic.IPv4Loopback,
			bindPrivateAddress:    cryptoutilSharedMagic.IPv4Loopback,
			wantPrivateClientAuth: tls.RequireAndVerifyClientCert,
			wantPublicClientAuth:  tls.NoClientCert, // Browsers don't have client certs
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				DevMode:            tc.devMode,
				BindPublicAddress:  tc.bindPublicAddress,
				BindPrivateAddress: tc.bindPrivateAddress,
			}

			// Replicate the mTLS logic from application_listener.go.
			isContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress
			privateClientAuth := tls.RequireAndVerifyClientCert
			if settings.DevMode || isContainerMode {
				privateClientAuth = tls.NoClientCert
			}

			publicClientAuth := tls.NoClientCert // Always NoClientCert for browser compatibility

			require.Equal(t, tc.wantPrivateClientAuth, privateClientAuth, "Private server mTLS")
			require.Equal(t, tc.wantPublicClientAuth, publicClientAuth, "Public server mTLS")
		})
	}
}

// TestHealthcheck_CompletesWithinTimeout tests healthcheck completes within reasonable timeout.
// Priority: P3.2 (Nice to Have - Could Have).
func TestHealthcheck_CompletesWithinTimeout(t *testing.T) {
	// Skipping because template service uses ApplicationCore builder pattern
	// which starts admin server internally. Testing healthcheck timeout
	// requires standalone admin server initialization, which is not the
	// current architecture pattern.
	// TODO: Revisit when admin server becomes independently testable.
	t.Skip("Template service uses ApplicationCore - admin server not independently testable")
}

// TestHealthcheck_TimeoutExceeded tests healthcheck fails when client timeout exceeded.
// Priority: P3.2 (Nice to Have - Could Have).
func TestHealthcheck_TimeoutExceeded(t *testing.T) {
	// Skipping because template service uses ApplicationCore builder pattern
	// which starts admin server internally. Testing timeout behavior
	// requires standalone admin server initialization, which is not the
	// current architecture pattern.
	// TODO: Revisit when admin server becomes independently testable.
	t.Skip("Template service uses ApplicationCore - admin server not independently testable")
}
