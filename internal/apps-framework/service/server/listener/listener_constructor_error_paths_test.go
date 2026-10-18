// Copyright (c) 2025-2026 Justin Cranford.
package listener

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testListenerAdmin          = "admin"
	testInvalidTLSClientPolicy = "invalid-policy"
)

func TestHTTPServerConstructors_InvalidMTLSPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		newServer func(*testing.T) error
		wantError string
	}{
		{
			name: cryptoutilSharedMagic.PublicLabel,
			newServer: func(t *testing.T) error {
				t.Helper()

				settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
				settings.PublicTLSClientPolicy = testInvalidTLSClientPolicy
				_, err := NewPublicHTTPServer(context.Background(), settings, validAutoTLSSettings(t))

				return err
			},
			wantError: "failed to apply public mTLS configuration",
		},
		{
			name: testListenerAdmin,
			newServer: func(t *testing.T) error {
				t.Helper()

				settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
				settings.AdminTLSClientPolicy = testInvalidTLSClientPolicy
				_, err := NewAdminHTTPServer(context.Background(), settings, validAutoTLSSettings(t))

				return err
			},
			wantError: "failed to apply admin mTLS configuration",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorContains(t, tc.newServer(t), tc.wantError)
		})
	}
}
