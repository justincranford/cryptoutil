// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"context"
	"testing"
	"testing/fstest"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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

// TestNewServerBuilder_NilContext tests error handling for nil context.
func TestNewServerBuilder_NilContext(t *testing.T) {
	t.Parallel()

	settings := getMinimalSettings()

	builder := NewServerBuilder(nil, settings)

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "context cannot be nil")
}

// TestNewServerBuilder_NilConfig tests error handling for nil configuration.
func TestNewServerBuilder_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	builder := NewServerBuilder(ctx, nil)

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "config cannot be nil")
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

// TestWithDomainMigrations_NilFS tests error handling for nil migration filesystem.
func TestWithDomainMigrations_NilFS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings).
		WithDomainMigrations(nil, "migrations")

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "migration FS cannot be nil")
}

// TestWithDomainMigrations_EmptyPath tests error handling for empty migration path.
func TestWithDomainMigrations_EmptyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	migrationFS := fstest.MapFS{
		"migrations/2001_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test (id TEXT);"),
		},
	}

	builder := NewServerBuilder(ctx, settings).
		WithDomainMigrations(migrationFS, "")

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "migrations path cannot be empty")
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

// TestBuild_Success tests full Build pipeline with domain migrations and route registration.
func TestBuild_Success(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
func TestMergedMigrations_Open(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	domainFS := fstest.MapFS{
		"migrations/2001_domain.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE domain (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     domainFS,
		domainPath:   "migrations",
	}

	// Test opening domain file (higher priority).
	domainFile, err := merged.Open("2001_domain.up.sql")
	require.NoError(t, err)
	require.NotNil(t, domainFile)
	defer domainFile.Close()

	// Test opening template file (fallback).
	templateFile, err := merged.Open("1001_template.up.sql")
	require.NoError(t, err)
	require.NotNil(t, templateFile)
	defer templateFile.Close()

	// Test opening non-existent file.
	_, err = merged.Open("9999_missing.up.sql")
	require.Error(t, err)
}

// TestMergedMigrations_ReadDir tests mergedMigrations.ReadDir method.
func TestMergedMigrations_ReadDir(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	domainFS := fstest.MapFS{
		"migrations/2001_domain.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE domain (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     domainFS,
		domainPath:   "migrations",
	}

	// Read merged directory (should contain both template and domain files).
	entries, err := merged.ReadDir(".")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 1) // At least domain entry.
}

// TestMergedMigrations_ReadFile tests mergedMigrations.ReadFile method.
func TestMergedMigrations_ReadFile(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	domainFS := fstest.MapFS{
		"migrations/2001_domain.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE domain (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     domainFS,
		domainPath:   "migrations",
	}

	// Read domain file (higher priority).
	domainData, err := merged.ReadFile("2001_domain.up.sql")
	require.NoError(t, err)
	require.Contains(t, string(domainData), "CREATE TABLE domain")

	// Read template file (fallback).
	templateData, err := merged.ReadFile("1001_template.up.sql")
	require.NoError(t, err)
	require.Contains(t, string(templateData), "CREATE TABLE template")

	// Read non-existent file.
	_, err = merged.ReadFile("9999_missing.up.sql")
	require.Error(t, err)
}

// TestMergedMigrations_Stat tests mergedMigrations.Stat method.
func TestMergedMigrations_Stat(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	domainFS := fstest.MapFS{
		"migrations/2001_domain.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE domain (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     domainFS,
		domainPath:   "migrations",
	}

	// Stat domain file (higher priority).
	domainInfo, err := merged.Stat("2001_domain.up.sql")
	require.NoError(t, err)
	require.NotNil(t, domainInfo)

	// Stat template file (fallback).
	templateInfo, err := merged.Stat("1001_template.up.sql")
	require.NoError(t, err)
	require.NotNil(t, templateInfo)

	// Stat non-existent file.
	_, err = merged.Stat("9999_missing.up.sql")
	require.Error(t, err)
}

// getMinimalSettings returns minimal valid settings for testing.
// Uses same pattern as application_listener_test.go.
func getMinimalSettings() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:                    true,
		VerboseMode:                false,
		DatabaseURL:                cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:                "template-service-test",
		OTLPEnabled:                false,
		OTLPEndpoint:               "grpc://127.0.0.1:4317",
		LogLevel:                   "INFO",
		BrowserSessionAlgorithm:    "JWS",
		BrowserSessionJWSAlgorithm: "RS256",
		BrowserSessionJWEAlgorithm: "RSA-OAEP",
		BrowserSessionExpiration:   15 * time.Minute,
		ServiceSessionAlgorithm:    "JWS",
		ServiceSessionJWSAlgorithm: "RS256",
		ServiceSessionJWEAlgorithm: "RSA-OAEP",
		ServiceSessionExpiration:   1 * time.Hour,
		SessionIdleTimeout:         30 * time.Minute,
		SessionCleanupInterval:     1 * time.Hour,
		BindPublicProtocol:         "https",
		BindPublicAddress:          cryptoutilSharedMagic.IPv4Loopback,
		BindPublicPort:             0,
		BindPrivateProtocol:        "https",
		BindPrivateAddress:         cryptoutilSharedMagic.IPv4Loopback,
		BindPrivatePort:            0,
		TLSPublicMode:              cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
		TLSPrivateMode:             cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
		TLSPublicDNSNames:          []string{"localhost"},
		TLSPublicIPAddresses:       []string{"127.0.0.1"},
		TLSPrivateDNSNames:         []string{"localhost"},
		TLSPrivateIPAddresses:      []string{"127.0.0.1"},
	}
}
