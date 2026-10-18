// Copyright (c) 2025-2026 Justin Cranford.
//

package server

import (
	"context"
	"path/filepath"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps-framework/service/server/builder"
	cryptoutilAppsSmKmsPKIServer "cryptoutil/internal/apps/sm-kms/server/pki/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestNewKMSServerFromConfig_PKIConfigReadError(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(cryptoutilSharedMagic.OTLPServiceSMKMS)
	settings.ConfigFile = []string{filepath.Join(t.TempDir(), "missing-config.yml")}

	srv, err := NewKMSServerFromConfig(context.Background(), settings)
	require.ErrorContains(t, err, "failed to load PKI authority configuration")
	require.Nil(t, srv)
}

func TestRegisterKMSRoutes_ValidationErrors(t *testing.T) {
	t.Parallel()

	var nilContext context.Context

	tests := []struct {
		name string
		call func() error
		want string
	}{
		{
			name: "missing context",
			call: func() error {
				return registerKMSRoutes(nilContext, nil, nil, nil, nil, cryptoutilAppsSmKmsPKIServer.PKIAuthorityConfig{})
			},
			want: "context is required",
		},
		{
			name: "missing resources",
			call: func() error {
				return registerKMSRoutes(context.Background(), nil, nil, nil, nil, cryptoutilAppsSmKmsPKIServer.PKIAuthorityConfig{})
			},
			want: "database and barrier service resources are required",
		},
		{
			name: "missing PKI handler",
			call: func() error {
				return registerKMSRoutesWithPKIComponents(
					fiber.New(),
					nil,
					nil,
					&cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources{},
					nil,
				)
			},
			want: "failed to register public PKI revocation endpoints",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.call()
			require.ErrorContains(t, err, tc.want)
		})
	}
}
