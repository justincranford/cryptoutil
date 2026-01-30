// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"context"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	defer func() { _ = domainFile.Close() }()

	// Test opening template file (fallback).
	templateFile, err := merged.Open("1001_template.up.sql")
	require.NoError(t, err)
	require.NotNil(t, templateFile)

	defer func() { _ = templateFile.Close() }()

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

// TestGenerateTLSConfig_StaticMode tests TLS config generation in static mode.
func TestGenerateTLSConfig_StaticMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()
	settings.TLSPrivateMode = cryptoutilAppsTemplateServiceConfig.TLSModeStatic
	settings.TLSStaticCertPEM = []byte("test-cert-pem")
	settings.TLSStaticKeyPEM = []byte("test-key-pem")

	builder := NewServerBuilder(ctx, settings)

	cfg, err := builder.generateTLSConfig(
		cryptoutilAppsTemplateServiceConfig.TLSModeStatic,
		[]byte("test-cert-pem"),
		[]byte("test-key-pem"),
		nil,
		nil,
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		"admin",
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, []byte("test-cert-pem"), cfg.StaticCertPEM)
	require.Equal(t, []byte("test-key-pem"), cfg.StaticKeyPEM)
}

// TestGenerateTLSConfig_MixedMode tests TLS config generation in mixed mode.
func TestGenerateTLSConfig_MixedMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)

	// Generate a valid CA certificate and key for mixed mode testing.
	caCertPEM, caKeyPEM := generateTestCA(t)

	cfg, err := builder.generateTLSConfig(
		cryptoutilAppsTemplateServiceConfig.TLSModeMixed,
		nil,
		nil,
		caCertPEM,
		caKeyPEM,
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		"public",
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
}

// TestGenerateTLSConfig_MixedModeError tests TLS config error handling in mixed mode.
func TestGenerateTLSConfig_MixedModeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)

	// Invalid CA certificate should cause error.
	cfg, err := builder.generateTLSConfig(
		cryptoutilAppsTemplateServiceConfig.TLSModeMixed,
		nil,
		nil,
		[]byte("invalid-ca-cert"),
		[]byte("invalid-ca-key"),
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		"admin",
	)

	require.Error(t, err)
	require.Nil(t, cfg)
	require.Contains(t, err.Error(), "failed to generate admin TLS config (mixed mode)")
}

// TestGenerateTLSConfig_AutoModeError tests TLS config error handling in auto mode.
// The auto mode fails when GenerateAutoTLSGeneratedSettings receives invalid IP addresses.
func TestGenerateTLSConfig_AutoModeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)

	// Invalid IP address triggers GenerateAutoTLSGeneratedSettings error.
	cfg, err := builder.generateTLSConfig(
		cryptoutilAppsTemplateServiceConfig.TLSModeAuto,
		nil,
		nil,
		nil,
		nil,
		[]string{"localhost"},
		[]string{"not-a-valid-ip"}, // Invalid IP address causes error
		"public",
	)

	require.Error(t, err)
	require.Nil(t, cfg)
	require.Contains(t, err.Error(), "failed to generate public TLS config (auto mode)")
}

// TestGenerateTLSConfig_UnsupportedMode tests error handling for unsupported TLS mode.
func TestGenerateTLSConfig_UnsupportedMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)

	cfg, err := builder.generateTLSConfig(
		cryptoutilAppsTemplateServiceConfig.TLSMode("unsupported"),
		nil,
		nil,
		nil,
		nil,
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		"admin",
	)

	require.Error(t, err)
	require.Nil(t, cfg)
	require.Contains(t, err.Error(), "unsupported TLS admin mode: unsupported")
}

// TestGenerateTLSConfig_DefaultMode tests that empty mode defaults to auto.
func TestGenerateTLSConfig_DefaultMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)

	// Empty string mode should default to auto.
	cfg, err := builder.generateTLSConfig(
		"", // Empty mode
		nil,
		nil,
		nil,
		nil,
		[]string{"localhost"},
		[]string{"127.0.0.1"},
		"public",
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
}

// TestMergedMigrations_ReadDir_SubPath tests ReadDir with a sub-path.
func TestMergedMigrations_ReadDir_SubPath(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/subdir/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template_sub (id TEXT);"),
		},
	}

	domainFS := fstest.MapFS{
		"migrations/subdir/2001_domain.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE domain_sub (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     domainFS,
		domainPath:   "migrations",
	}

	// Read merged subdirectory.
	entries, err := merged.ReadDir("subdir")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 1)
}

// TestMergedMigrations_Open_RootDir tests Open with current directory.
func TestMergedMigrations_Open_RootDir(t *testing.T) {
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

	// Open root directory (should work for both "." and "").
	rootFile, err := merged.Open(".")
	require.NoError(t, err)
	require.NotNil(t, rootFile)

	defer func() { _ = rootFile.Close() }()

	emptyPathFile, err := merged.Open("")
	require.NoError(t, err)
	require.NotNil(t, emptyPathFile)

	defer func() { _ = emptyPathFile.Close() }()
}

// TestMergedMigrations_ReadDir_NilDomainFS tests ReadDir when domainFS is nil.
func TestMergedMigrations_ReadDir_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Read merged directory with nil domainFS.
	entries, err := merged.ReadDir(".")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 1)
}

// TestMergedMigrations_Open_NilDomainFS tests Open when domainFS is nil.
func TestMergedMigrations_Open_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Open template file when domain FS is nil.
	templateFile, err := merged.Open("1001_template.up.sql")
	require.NoError(t, err)
	require.NotNil(t, templateFile)

	defer func() { _ = templateFile.Close() }()
}

// TestMergedMigrations_ReadFile_NilDomainFS tests ReadFile when domainFS is nil.
func TestMergedMigrations_ReadFile_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template_only (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Read template file when domain FS is nil.
	templateData, err := merged.ReadFile("1001_template.up.sql")
	require.NoError(t, err)
	require.Contains(t, string(templateData), "CREATE TABLE template_only")
}

// TestMergedMigrations_Stat_NilDomainFS tests Stat when domainFS is nil.
func TestMergedMigrations_Stat_NilDomainFS(t *testing.T) {
	t.Parallel()

	templateFS := fstest.MapFS{
		"migrations/1001_template.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE template_stat (id TEXT);"),
		},
	}

	merged := &mergedMigrations{
		templateFS:   templateFS,
		templatePath: "migrations",
		domainFS:     nil, // No domain FS
		domainPath:   "",
	}

	// Stat template file when domain FS is nil.
	templateInfo, err := merged.Stat("1001_template.up.sql")
	require.NoError(t, err)
	require.NotNil(t, templateInfo)
}

// generateTestCA generates a valid CA certificate and key for testing.
func generateTestCA(t *testing.T) (caCertPEM, caKeyPEM []byte) {
	t.Helper()

	// Generate CA key pair.
	caKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	require.NoError(t, err)

	// Generate CA certificate.
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * cryptoutilSharedMagic.HoursPerDay * time.Hour //nolint:mnd // Duration calculation.
	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "Test CA", duration)
	require.NoError(t, err)
	require.Len(t, caSubjects, 1)

	caCert := caSubjects[0].KeyMaterial.CertificateChain[0]

	// Serialize CA certificate to PEM.
	caCertPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	// Serialize CA private key to PEM.
	caKeyBytes, err := x509.MarshalPKCS8PrivateKey(caKeyPair.Private)
	require.NoError(t, err)

	caKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caKeyBytes,
	})

	return caCertPEM, caKeyPEM
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
