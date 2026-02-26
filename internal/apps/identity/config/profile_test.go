// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		profileName string
		wantErr     bool
	}{
		{
			name:        "demo profile exists",
			profileName: "demo",
			wantErr:     false,
		},
		{
			name:        "authz-only profile exists",
			profileName: "authz-only",
			wantErr:     false,
		},
		{
			name:        "nonexistent profile",
			profileName: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := LoadProfile(tc.profileName)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
			}
		})
	}
}

func TestLoadProfileFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid profile file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		profileFile := filepath.Join(tempDir, "test.yml")

		profileContent := `services:
  authz:
    enabled: true
    bind_address: "127.0.0.1:8080"
    database_url: ":memory:"
    log_level: "debug"
  idp:
    enabled: false
  rs:
    enabled: false
`
		err := os.WriteFile(profileFile, []byte(profileContent), cryptoutilSharedMagic.CacheFilePermissions)
		require.NoError(t, err)

		cfg, err := LoadProfileFromFile(profileFile)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.True(t, cfg.Services.AuthZ.Enabled)
		require.False(t, cfg.Services.IDP.Enabled)
		require.Equal(t, "127.0.0.1:8080", cfg.Services.AuthZ.BindAddress)
	})

	t.Run("invalid YAML", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		profileFile := filepath.Join(tempDir, "invalid.yml")

		invalidContent := `this is not: valid: yaml: content`
		err := os.WriteFile(profileFile, []byte(invalidContent), cryptoutilSharedMagic.CacheFilePermissions)
		require.NoError(t, err)

		cfg, err := LoadProfileFromFile(profileFile)
		require.Error(t, err)
		require.Nil(t, cfg)
		require.Contains(t, err.Error(), "failed to parse profile YAML")
	})
}

func TestProfileConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     ProfileConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid authz service",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					AuthZ: ServiceConfig{
						Enabled:     true,
						BindAddress: "127.0.0.1:8080",
						DatabaseURL: cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
						LogLevel:    "debug",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no services enabled",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					AuthZ: ServiceConfig{Enabled: false},
					IDP:   ServiceConfig{Enabled: false},
					RS:    ServiceConfig{Enabled: false},
				},
			},
			wantErr: true,
			errMsg:  "at least one service must be enabled",
		},
		{
			name: "missing bind address",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					AuthZ: ServiceConfig{
						Enabled:     true,
						DatabaseURL: cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
						LogLevel:    "info",
					},
				},
			},
			wantErr: true,
			errMsg:  "bind_address is required",
		},
		{
			name: "invalid log level",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					AuthZ: ServiceConfig{
						Enabled:     true,
						BindAddress: "127.0.0.1:8080",
						DatabaseURL: cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
						LogLevel:    "invalid",
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid log_level",
		},
		{
			name: "valid idp service",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					IDP: ServiceConfig{
						Enabled:     true,
						BindAddress: "127.0.0.1:8081",
						DatabaseURL: cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
						LogLevel:    "info",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "idp missing database_url",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					IDP: ServiceConfig{
						Enabled:     true,
						BindAddress: "127.0.0.1:8081",
						LogLevel:    "info",
					},
				},
			},
			wantErr: true,
			errMsg:  "database_url is required",
		},
		{
			name: "valid rs service",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					RS: ServiceConfig{
						Enabled:     true,
						BindAddress: "127.0.0.1:8082",
						LogLevel:    "warn",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "rs missing bind_address",
			cfg: ProfileConfig{
				Services: ServiceConfigs{
					RS: ServiceConfig{
						Enabled:  true,
						LogLevel: cryptoutilSharedMagic.StringError,
					},
				},
			},
			wantErr: true,
			errMsg:  "bind_address is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
