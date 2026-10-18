// Copyright (c) 2025-2026 Justin Cranford.
package builder

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"
	"time"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestDomainConfig_Build_NilCtx verifies nil context returns an error.
func TestDomainConfig_Build_NilCtx(t *testing.T) {
	t.Parallel()

	_, err := Build(nil, getMinimalSettings(), nil) //nolint:staticcheck // SA1012: intentionally passing nil context to test error path.

	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestDomainConfig_Build_NilSettings verifies nil settings returns an error.
func TestDomainConfig_Build_NilSettings(t *testing.T) {
	t.Parallel()

	_, err := Build(context.Background(), nil, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

// TestDomainConfig_Build_NilDomain succeeds when domain is nil (template-only service).
func TestDomainConfig_Build_NilDomain(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
	defer cancel()

	resources, err := Build(ctx, getMinimalSettings(), nil)

	require.NoError(t, err)
	require.NotNil(t, resources)
	require.NotNil(t, resources.Application)

	resources.ShutdownCore()
}

// TestDomainConfig_Build_WithMigrations verifies that domain migrations are applied.
func TestDomainConfig_Build_WithMigrations(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
	defer cancel()

	domainFS := fstest.MapFS{
		testMigrations2001TestUpSQL: &fstest.MapFile{
			Data: []byte("CREATE TABLE IF NOT EXISTS test_domain_build (id TEXT PRIMARY KEY);"),
		},
		testMigrations2001TestDownSQL: &fstest.MapFile{
			Data: []byte("DROP TABLE IF EXISTS test_domain_build;"),
		},
	}

	resources, err := Build(ctx, getMinimalSettings(), &DomainConfig{
		MigrationsFS:   domainFS,
		MigrationsPath: valueMigrations,
	})

	require.NoError(t, err)
	require.NotNil(t, resources)

	resources.ShutdownCore()
}

// TestDomainConfig_Build_WithRouteRegistration verifies route registration callback is called.
func TestDomainConfig_Build_WithRouteRegistration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
	defer cancel()

	routeRegistered := false

	resources, err := Build(ctx, getMinimalSettings(), &DomainConfig{
		RouteRegistration: func(_ *cryptoutilAppsFrameworkServiceServer.PublicServerBase, _ *ServiceResources) error {
			routeRegistered = true

			return nil
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resources)
	require.True(t, routeRegistered)

	resources.ShutdownCore()
}

func TestDomainConfig_Build_Error(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
	defer cancel()

	settings := getMinimalSettings()
	settings.TLSPrivateProvisionMode = "invalid-mode"

	resources, err := Build(ctx, settings, &DomainConfig{
		RouteRegistration: func(*cryptoutilAppsFrameworkServiceServer.PublicServerBase, *ServiceResources) error {
			return errors.New("route registration must not run")
		},
	})

	require.Nil(t, resources)
	require.ErrorContains(t, err, "failed to build service")
	require.ErrorContains(t, err, "unsupported TLS admin mode")
}
