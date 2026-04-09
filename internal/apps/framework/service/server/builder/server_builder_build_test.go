// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"context"
	"fmt"
	"testing"
	"testing/fstest"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"

	"github.com/stretchr/testify/require"
)

// TestBuild_SimpleErrors tests Build error paths that do not require full service initialization.
func TestBuild_SimpleErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFn       func() *ServerBuilder
		wantErrSubstr string
	}{
		{
			name: "early error from nil config",
			setupFn: func() *ServerBuilder {
				return NewServerBuilder(context.Background(), nil)
			},
			wantErrSubstr: "config cannot be nil",
		},
		{
			name: "admin TLS invalid mode",
			setupFn: func() *ServerBuilder {
				settings := getMinimalSettings()
				settings.TLSPrivateMode = cryptoutilAppsFrameworkServiceConfig.TLSMode("invalid-mode")

				return NewServerBuilder(context.Background(), settings)
			},
			wantErrSubstr: "unsupported TLS admin mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := tt.setupFn()
			resources, err := builder.Build()

			require.Nil(t, resources)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErrSubstr)
		})
	}
}

// TestBuild_PublicRouteRegistrationError tests Build failure when route registration fails.
func TestBuild_PublicRouteRegistrationError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
	defer cancel()

	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)
	builder.WithPublicRouteRegistration(func(
		_ *cryptoutilAppsFrameworkServiceServer.PublicServerBase,
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
func TestBuild_PublicTLSError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
	defer cancel()

	settings := getMinimalSettings()
	// Valid admin TLS mode (auto).
	settings.TLSPrivateMode = cryptoutilAppsFrameworkServiceConfig.TLSModeAuto
	// Invalid public TLS mode causes error AFTER services are initialized.
	settings.TLSPublicMode = cryptoutilAppsFrameworkServiceConfig.TLSMode("invalid_mode")

	builder := NewServerBuilder(ctx, settings)

	resources, err := builder.Build()

	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported TLS public mode: invalid_mode")
}

// TestBuild_MigrationError tests Build failure when migration fails.
func TestBuild_MigrationError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
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

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
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
		base *cryptoutilAppsFrameworkServiceServer.PublicServerBase,
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

// TestBuild_AutoConfig tests that Build auto-populates configs when With* methods are not called.
func TestBuild_AutoConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		checkFn func(t *testing.T, resources *ServiceResources)
	}{
		{
			name: "auto-configures JWTAuthConfig",
			checkFn: func(t *testing.T, resources *ServiceResources) {
				t.Helper()
				require.NotNil(t, resources.JWTAuthConfig, "JWTAuthConfig must be auto-configured when WithJWTAuth is not called")
				require.Equal(t, JWTAuthModeSession, resources.JWTAuthConfig.Mode)
			},
		},
		{
			name: "auto-configures StrictServerConfig",
			checkFn: func(t *testing.T, resources *ServiceResources) {
				t.Helper()
				require.NotNil(t, resources.StrictServerConfig, "StrictServerConfig must be auto-configured when WithStrictServer is not called")
				require.Equal(t, cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath, resources.StrictServerConfig.BrowserAPIBasePath)
				require.Equal(t, cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath, resources.StrictServerConfig.ServiceAPIBasePath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
			defer cancel()

			settings := getMinimalSettings()

			builder := NewServerBuilder(ctx, settings)
			resources, err := builder.Build()

			require.NoError(t, err)
			require.NotNil(t, resources)

			tt.checkFn(t, resources)

			// Cleanup.
			if resources.ShutdownCore != nil {
				resources.ShutdownCore()
			}

			if resources.ShutdownContainer != nil {
				resources.ShutdownContainer()
			}
		})
	}
}

// TestBuild_ExplicitConfigPreserved tests that Build preserves explicitly-set configs (not overwritten by auto-config).
func TestBuild_ExplicitConfigPreserved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(builder *ServerBuilder)
		checkFn func(t *testing.T, resources *ServiceResources)
	}{
		{
			name: "explicit JWTAuth preserved",
			setupFn: func(builder *ServerBuilder) {
				explicitConfig := NewKMSJWTAuthConfig("https://example.com/.well-known/jwks.json", "https://auth.example.com", "kms-service")
				builder.WithJWTAuth(explicitConfig)
			},
			checkFn: func(t *testing.T, resources *ServiceResources) {
				t.Helper()
				require.NotNil(t, resources.JWTAuthConfig)
				require.Equal(t, JWTAuthModeRequired, resources.JWTAuthConfig.Mode, "explicit JWTAuthMode must be preserved")
				require.Equal(t, "https://example.com/.well-known/jwks.json", resources.JWTAuthConfig.JWKSURL)
			},
		},
		{
			name: "explicit StrictServer preserved",
			setupFn: func(builder *ServerBuilder) {
				const (
					customBrowserPath = "/custom/browser/v99"
					customServicePath = "/custom/service/v99"
				)

				explicitStrictConfig := NewDefaultStrictServerConfig().
					WithBrowserBasePath(customBrowserPath).
					WithServiceBasePath(customServicePath)
				builder.WithStrictServer(explicitStrictConfig)
			},
			checkFn: func(t *testing.T, resources *ServiceResources) {
				t.Helper()
				require.NotNil(t, resources.StrictServerConfig)
				require.Equal(t, "/custom/browser/v99", resources.StrictServerConfig.BrowserAPIBasePath, "explicit BrowserAPIBasePath must be preserved")
				require.Equal(t, "/custom/service/v99", resources.StrictServerConfig.ServiceAPIBasePath, "explicit ServiceAPIBasePath must be preserved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*time.Second)
			defer cancel()

			settings := getMinimalSettings()

			builder := NewServerBuilder(ctx, settings)
			tt.setupFn(builder)

			resources, err := builder.Build()

			require.NoError(t, err)
			require.NotNil(t, resources)

			tt.checkFn(t, resources)

			// Cleanup.
			if resources.ShutdownCore != nil {
				resources.ShutdownCore()
			}

			if resources.ShutdownContainer != nil {
				resources.ShutdownContainer()
			}
		})
	}
}
