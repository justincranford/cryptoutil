// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"

	"github.com/stretchr/testify/require"
)

// TestNewServerBuilder_Success tests successful server builder creation.
func TestNewServerBuilder_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)

	require.NotNil(t, builder)
	require.Nil(t, builder.err)
	require.Equal(t, ctx, builder.ctx)
	require.Equal(t, settings, builder.config)
}

// TestNewServerBuilder_ValidationErrors tests error handling for invalid inputs using table-driven pattern.
func TestNewServerBuilder_ValidationErrors(t *testing.T) {
	t.Parallel()

	settings := getMinimalSettings()

	tests := []struct {
		name          string
		ctx           context.Context
		config        *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
		wantErrSubstr string
	}{
		{
			name:          "nil context",
			ctx:           nil,
			config:        settings,
			wantErrSubstr: "context cannot be nil",
		},
		{
			name:          "nil config",
			ctx:           context.Background(),
			config:        nil,
			wantErrSubstr: "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := NewServerBuilder(tt.ctx, tt.config)

			require.NotNil(t, builder)
			require.Error(t, builder.err)
			require.Contains(t, builder.err.Error(), tt.wantErrSubstr)
		})
	}
}

// TestWithDomainMigrations_Success tests successful domain migration registration.
func TestWithDomainMigrations_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	// Create test migration filesystem.
	migrationFS := fstest.MapFS{
		"migrations/2001_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test (id TEXT);"),
		},
		"migrations/2001_test.down.sql": &fstest.MapFile{
			Data: []byte("DROP TABLE test;"),
		},
	}

	builder := NewServerBuilder(ctx, settings).
		WithDomainMigrations(migrationFS, "migrations")

	require.NotNil(t, builder)
	require.Nil(t, builder.err)
	require.NotNil(t, builder.migrationFS)
	require.Equal(t, "migrations", builder.migrationsPath)
}

// TestWithDomainMigrations_ValidationErrors tests error handling using table-driven pattern.
func TestWithDomainMigrations_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	migrationFS := fstest.MapFS{
		"migrations/2001_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test (id TEXT);"),
		},
	}

	tests := []struct {
		name          string
		fs            fs.FS
		path          string
		wantErrSubstr string
	}{
		{
			name:          "nil filesystem",
			fs:            nil,
			path:          "migrations",
			wantErrSubstr: "migration FS cannot be nil",
		},
		{
			name:          "empty path",
			fs:            migrationFS,
			path:          "",
			wantErrSubstr: "migrations path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := NewServerBuilder(ctx, settings).
				WithDomainMigrations(tt.fs, tt.path)

			require.NotNil(t, builder)
			require.Error(t, builder.err)
			require.Contains(t, builder.err.Error(), tt.wantErrSubstr)
		})
	}
}

// TestWithDomainMigrations_ErrorAccumulation tests error accumulation in fluent chain.
func TestWithDomainMigrations_ErrorAccumulation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start with error from NewServerBuilder.
	builder := NewServerBuilder(ctx, nil).
		WithDomainMigrations(fstest.MapFS{}, "migrations")

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "config cannot be nil")
	// Should NOT have migration error because first error short-circuits.
	require.NotContains(t, builder.err.Error(), "migration")
}

// TestWithPublicRouteRegistration_Success tests successful route registration function.
func TestWithPublicRouteRegistration_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	registerFunc := func(_ *cryptoutilAppsTemplateServiceServer.PublicServerBase, _ *ServiceResources) error {
		return nil
	}

	builder := NewServerBuilder(ctx, settings).
		WithPublicRouteRegistration(registerFunc)

	require.NotNil(t, builder)
	require.Nil(t, builder.err)
	require.NotNil(t, builder.publicRouteRegister)
}

// TestWithPublicRouteRegistration_NilFunc tests error handling for nil registration function.
func TestWithPublicRouteRegistration_NilFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings).
		WithPublicRouteRegistration(nil)

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "route registration function cannot be nil")
}

// TestWithPublicRouteRegistration_ErrorAccumulation tests error accumulation.
func TestWithPublicRouteRegistration_ErrorAccumulation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start with error from NewServerBuilder.
	registerFunc := func(_ *cryptoutilAppsTemplateServiceServer.PublicServerBase, _ *ServiceResources) error {
		return nil
	}

	builder := NewServerBuilder(ctx, nil).
		WithPublicRouteRegistration(registerFunc)

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "config cannot be nil")
	// Should NOT have route registration error because first error short-circuits.
	require.NotContains(t, builder.err.Error(), "route registration")
}

// TestBuild_EarlyError tests that Build returns accumulated error from fluent chain.
func TestBuild_EarlyError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start with error from NewServerBuilder.
	builder := NewServerBuilder(ctx, nil)

	resources, err := builder.Build()

	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestBuild_AdminTLSError tests Build failure when admin TLS configuration fails.
func TestBuild_AdminTLSError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	// Set invalid TLS mode to trigger error in generateTLSConfig.
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSMode("invalid-mode")

	builder := NewServerBuilder(ctx, settings)
	resources, err := builder.Build()

	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported TLS admin mode")
}

// TestBuild_PublicRouteRegistrationError tests Build failure when route registration fails.
// Note: This test cannot run in parallel with other Build tests due to shared in-memory SQLite cache.
func TestBuild_PublicRouteRegistrationError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)
	builder.WithPublicRouteRegistration(func(
		_ *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		_ *ServiceResources,
	) error {
		return fmt.Errorf("intentional route registration failure")
	})

	resources, err := builder.Build()

	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to register public routes")
	require.Contains(t, err.Error(), "intentional route registration failure")
}

// TestBuild_PublicTLSError tests Build failure when public TLS config fails.
// This covers the error path after services are initialized but before public server creation.
// Note: This test cannot run in parallel due to shared in-memory SQLite cache.
func TestBuild_PublicTLSError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	settings := getMinimalSettings()
	// Valid admin TLS mode (auto).
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeAuto
	// Invalid public TLS mode causes error AFTER services are initialized.
	settings.TLSPublicMode = cryptoutilAppsTemplateServiceConfig.TLSMode("invalid_mode")

	builder := NewServerBuilder(ctx, settings)

	resources, err := builder.Build()

	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported TLS public mode: invalid_mode")
}

// TestBuild_MigrationError tests Build failure when migration fails.
func TestBuild_MigrationError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	settings := getMinimalSettings()

	// Create invalid migration that will fail.
	invalidMigrations := fstest.MapFS{
		"migrations/9999_invalid.up.sql": {
			Data: []byte("INVALID SQL SYNTAX THAT WILL FAIL;"),
		},
	}

	builder := NewServerBuilder(ctx, settings)
	builder.WithDomainMigrations(invalidMigrations, "migrations")

	resources, err := builder.Build()

	require.Nil(t, resources)
	require.Error(t, err)
	// Migration error should propagate.
}

// TestBuild_Success tests full Build pipeline with domain migrations and route registration.
func TestBuild_Success(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	settings := getMinimalSettings()

	// Create test migrations (template already exists, add domain).
	domainMigrations := fstest.MapFS{
		"migrations/2001_test_domain.up.sql": {
			Data: []byte("CREATE TABLE IF NOT EXISTS test_domain (id TEXT PRIMARY KEY);"),
		},
		"migrations/2001_test_domain.down.sql": {
			Data: []byte("DROP TABLE IF EXISTS test_domain;"),
		},
	}

	// Build with domain migrations and route registration.
	builder := NewServerBuilder(ctx, settings)
	builder.WithDomainMigrations(domainMigrations, "migrations")

	routeRegistered := false

	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		res *ServiceResources,
	) error {
		routeRegistered = true

		require.NotNil(t, base)
		require.NotNil(t, res)
		require.NotNil(t, res.DB)
		require.NotNil(t, res.TelemetryService)
		require.NotNil(t, res.JWKGenService)
		require.NotNil(t, res.BarrierService)
		require.NotNil(t, res.UnsealKeysService)
		require.NotNil(t, res.SessionManager)
		require.NotNil(t, res.RealmService)
		require.NotNil(t, res.RealmRepository)
		require.NotNil(t, res.ShutdownCore)
		require.NotNil(t, res.ShutdownContainer)
		// Note: res.Application is nil during callback - it's created after route registration.
		return nil
	})

	resources, err := builder.Build()

	// Verify Build succeeded.
	require.NoError(t, err)
	require.NotNil(t, resources)

	// Verify route registration callback was invoked.
	require.True(t, routeRegistered)

	// Verify all resources populated.
	require.NotNil(t, resources.DB)
	require.NotNil(t, resources.TelemetryService)
	require.NotNil(t, resources.JWKGenService)
	require.NotNil(t, resources.BarrierService)
	require.NotNil(t, resources.UnsealKeysService)
	require.NotNil(t, resources.SessionManager)
	require.NotNil(t, resources.RegistrationService)
	require.NotNil(t, resources.RealmService)
	require.NotNil(t, resources.RealmRepository)
	require.NotNil(t, resources.Application)
	require.NotNil(t, resources.ShutdownCore)
	require.NotNil(t, resources.ShutdownContainer)

	// Verify domain migration was applied.
	var domainTableName string

	err = resources.DB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='test_domain'").Scan(&domainTableName).Error
	require.NoError(t, err)
	require.Equal(t, "test_domain", domainTableName)

	// Verify template migrations were applied (check for any template table - barrier_root_keys is created by template).
	var barrierTableName string

	err = resources.DB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='barrier_root_keys'").Scan(&barrierTableName).Error
	require.NoError(t, err)
	require.Equal(t, "barrier_root_keys", barrierTableName)

	// Cleanup.
	if resources.ShutdownCore != nil {
		resources.ShutdownCore()
	}

	if resources.ShutdownContainer != nil {
		resources.ShutdownContainer()
	}
}

// TestMergedMigrations_Open tests mergedMigrations.Open method.
