// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"

	"github.com/stretchr/testify/require"
)

func TestKMSBuilderAdapterSettings_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *KMSBuilderAdapterSettings
		wantErr  bool
	}{
		{
			name:     "empty settings valid",
			settings: &KMSBuilderAdapterSettings{},
			wantErr:  false,
		},
		{
			name: "with JWKS URL valid",
			settings: &KMSBuilderAdapterSettings{
				JWKSURL:     "https://example.com/.well-known/jwks.json",
				JWTIssuer:   "https://example.com",
				JWTAudience: "my-audience",
			},
			wantErr: false,
		},
		{
			name: "partial settings valid",
			settings: &KMSBuilderAdapterSettings{
				JWKSURL: "https://example.com/.well-known/jwks.json",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.settings.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewKMSBuilderAdapter(t *testing.T) {
	t.Parallel()

	// Create a minimal valid settings for testing.
	validSettings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BindPublicProtocol:          "https",
		BindPublicAddress:           "127.0.0.1",
		BindPublicPort:              0,
		BindPrivateProtocol:         "https",
		BindPrivateAddress:          "127.0.0.1",
		BindPrivatePort:             0,
		PublicBrowserAPIContextPath: "/browser/api/v1",
		PublicServiceAPIContextPath: "/service/api/v1",
	}

	tests := []struct {
		name        string
		ctx         context.Context
		settings    *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
		kmsSettings *KMSBuilderAdapterSettings
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid with all settings",
			ctx:         context.Background(),
			settings:    validSettings,
			kmsSettings: &KMSBuilderAdapterSettings{JWKSURL: "https://example.com/.well-known/jwks.json"},
			wantErr:     false,
		},
		{
			name:        "valid with nil KMS settings",
			ctx:         context.Background(),
			settings:    validSettings,
			kmsSettings: nil,
			wantErr:     false,
		},
		{
			name:        "nil context",
			ctx:         nil,
			settings:    validSettings,
			kmsSettings: nil,
			wantErr:     true,
			errContains: "context cannot be nil",
		},
		{
			name:        "nil settings",
			ctx:         context.Background(),
			settings:    nil,
			kmsSettings: nil,
			wantErr:     true,
			errContains: "settings cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter, err := NewKMSBuilderAdapter(tt.ctx, tt.settings, tt.kmsSettings)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, adapter)

				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, adapter)
			}
		})
	}
}

func TestKMSBuilderAdapter_ConfigureBuilder(t *testing.T) {
	t.Parallel()

	validSettings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BindPublicProtocol:          "https",
		BindPublicAddress:           "127.0.0.1",
		BindPublicPort:              0,
		BindPrivateProtocol:         "https",
		BindPrivateAddress:          "127.0.0.1",
		BindPrivatePort:             0,
		PublicBrowserAPIContextPath: "/browser/api/v1",
		PublicServiceAPIContextPath: "/service/api/v1",
	}

	tests := []struct {
		name        string
		kmsSettings *KMSBuilderAdapterSettings
	}{
		{
			name:        "without JWT",
			kmsSettings: nil,
		},
		{
			name: "with JWT",
			kmsSettings: &KMSBuilderAdapterSettings{
				JWKSURL:     "https://example.com/.well-known/jwks.json",
				JWTIssuer:   "https://example.com",
				JWTAudience: "my-audience",
			},
		},
		{
			name: "with JWT URL only",
			kmsSettings: &KMSBuilderAdapterSettings{
				JWKSURL: "https://example.com/.well-known/jwks.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter, err := NewKMSBuilderAdapter(context.Background(), validSettings, tt.kmsSettings)
			require.NoError(t, err)
			require.NotNil(t, adapter)

			builder := adapter.ConfigureBuilder()
			require.NotNil(t, builder)
		})
	}
}

func TestKMSBuilderAdapter_Accessors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BindPublicProtocol:          "https",
		BindPublicAddress:           "127.0.0.1",
		BindPublicPort:              8080,
		BindPrivateProtocol:         "https",
		BindPrivateAddress:          "127.0.0.1",
		BindPrivatePort:             9090,
		PublicBrowserAPIContextPath: "/browser/api/v1",
		PublicServiceAPIContextPath: "/service/api/v1",
	}
	kmsSettings := &KMSBuilderAdapterSettings{
		JWKSURL:     "https://example.com/.well-known/jwks.json",
		JWTIssuer:   "https://example.com",
		JWTAudience: "my-audience",
	}

	adapter, err := NewKMSBuilderAdapter(ctx, settings, kmsSettings)
	require.NoError(t, err)

	// Test Context accessor.
	require.Equal(t, ctx, adapter.Context())

	// Test Settings accessor.
	require.Equal(t, settings, adapter.Settings())
	require.Equal(t, uint16(8080), adapter.Settings().BindPublicPort)
	require.Equal(t, uint16(9090), adapter.Settings().BindPrivatePort)

	// Test KMSSettings accessor.
	require.Equal(t, kmsSettings, adapter.KMSSettings())
	require.Equal(t, "https://example.com/.well-known/jwks.json", adapter.KMSSettings().JWKSURL)
	require.Equal(t, "https://example.com", adapter.KMSSettings().JWTIssuer)
	require.Equal(t, "my-audience", adapter.KMSSettings().JWTAudience)
}

func TestKMSBuilderAdapter_NilKMSSettingsDefaults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BindPublicProtocol:          "https",
		BindPublicAddress:           "127.0.0.1",
		BindPublicPort:              0,
		BindPrivateProtocol:         "https",
		BindPrivateAddress:          "127.0.0.1",
		BindPrivatePort:             0,
		PublicBrowserAPIContextPath: "/browser/api/v1",
		PublicServiceAPIContextPath: "/service/api/v1",
	}

	adapter, err := NewKMSBuilderAdapter(ctx, settings, nil)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// KMSSettings should be populated with empty defaults.
	require.NotNil(t, adapter.KMSSettings())
	require.Empty(t, adapter.KMSSettings().JWKSURL)
	require.Empty(t, adapter.KMSSettings().JWTIssuer)
	require.Empty(t, adapter.KMSSettings().JWTAudience)
}
