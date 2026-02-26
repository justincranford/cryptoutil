// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"testing/fstest"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"

	"github.com/stretchr/testify/require"
)

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
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
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
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.SubjectTypePublic,
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
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
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
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{"not-a-valid-ip"}, // Invalid IP address causes error
		cryptoutilSharedMagic.SubjectTypePublic,
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
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
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
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.SubjectTypePublic,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
}

// TestMergedMigrations_ReadDir_SubPath tests ReadDir with a sub-path.
