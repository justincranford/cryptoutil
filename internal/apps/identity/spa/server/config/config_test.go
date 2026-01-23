// Copyright (c) 2025 Justin Cranford

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

func TestNewTestConfig(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)

	require.NotNil(t, cfg)
	require.Equal(t, cryptoutilMagic.OTLPServiceIdentitySPA, cfg.OTLPService)
	require.Equal(t, defaultStaticFilesPath, cfg.StaticFilesPath)
	require.Equal(t, defaultIndexFile, cfg.IndexFile)
	require.Equal(t, defaultCacheMaxAgeDev, cfg.CacheControlMaxAge)
	require.False(t, cfg.EnableGzip, "Gzip should be disabled for tests")
	require.False(t, cfg.EnableBrotli, "Brotli should be disabled for tests")
}

func TestDefaultTestConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	require.NotNil(t, cfg)
	require.True(t, cfg.DevMode)
	require.Equal(t, defaultStaticFilesPath, cfg.StaticFilesPath)
	require.Equal(t, defaultIndexFile, cfg.IndexFile)
}

func TestNewTestConfig_ProductionMode(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, false)

	require.NotNil(t, cfg)
	require.False(t, cfg.DevMode)
}

func TestIdentitySPAServerSettings_FullConfig(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)

	// Verify all fields have expected values.
	require.NotEmpty(t, cfg.StaticFilesPath)
	require.NotEmpty(t, cfg.IndexFile)
	require.NotEmpty(t, cfg.CSPDirectives)
	require.GreaterOrEqual(t, cfg.CacheControlMaxAge, 0)
}

func TestValidateIdentitySPASettings_RPOriginFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rpOrigin  string
		wantError bool
	}{
		{"valid_https", "https://localhost:18300", false},
		{"valid_http", "http://localhost:18300", false},
		{"valid_with_path", "https://example.com:8080", false},
		{"invalid_no_scheme", "localhost:18300", true},
		{"invalid_ftp_scheme", "ftp://localhost:18300", true},
		{"empty_allowed", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)
			cfg.RPOrigin = tt.rpOrigin

			err := validateIdentitySPASettings(cfg)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateIdentitySPASettings_RequiredFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		modifyFunc func(*IdentitySPAServerSettings)
		wantError  bool
	}{
		{
			name:       "valid_default",
			modifyFunc: func(_ *IdentitySPAServerSettings) {},
			wantError:  false,
		},
		{
			name:       "empty_static_path",
			modifyFunc: func(cfg *IdentitySPAServerSettings) { cfg.StaticFilesPath = "" },
			wantError:  true,
		},
		{
			name:       "empty_index_file",
			modifyFunc: func(cfg *IdentitySPAServerSettings) { cfg.IndexFile = "" },
			wantError:  true,
		},
		{
			name:       "negative_cache_max_age",
			modifyFunc: func(cfg *IdentitySPAServerSettings) { cfg.CacheControlMaxAge = -1 },
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)
			tt.modifyFunc(cfg)

			err := validateIdentitySPASettings(cfg)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
