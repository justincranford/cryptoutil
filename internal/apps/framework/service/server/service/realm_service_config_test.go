// Copyright 2025 Cisco Systems, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

func TestRealmService_GetRealmConfig(t *testing.T) {
	t.Parallel()

	svc, db := setupRealmService(t)
	ctx := context.Background()

	tenant := createRealmTestTenant(t, db, "realm-config-"+googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	originalConfig := &UsernamePasswordConfig{
		MinPasswordLength: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireDigit:      true,
		RequireSpecial:    true,
	}

	created, err := svc.CreateRealm(ctx, tenant.ID, string(RealmTypeUsernamePassword), originalConfig)
	require.NoError(t, err)

	// Get and parse config.
	parsedConfig, err := svc.GetRealmConfig(ctx, tenant.ID, created.RealmID)
	require.NoError(t, err)

	pwConfig, ok := parsedConfig.(*UsernamePasswordConfig)
	require.True(t, ok)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, pwConfig.MinPasswordLength)
	require.True(t, pwConfig.RequireUppercase)
	require.True(t, pwConfig.RequireLowercase)
	require.True(t, pwConfig.RequireDigit)
	require.True(t, pwConfig.RequireSpecial)
}

// TestUsernamePasswordConfig_Validate tests UsernamePasswordConfig validation.
func TestUsernamePasswordConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *UsernamePasswordConfig
		wantErr bool
	}{
		{
			name:    "valid",
			config:  &UsernamePasswordConfig{MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength},
			wantErr: false,
		},
		{
			name:    "invalid_zero_length",
			config:  &UsernamePasswordConfig{MinPasswordLength: 0},
			wantErr: true,
		},
		{
			name:    "invalid_negative_length",
			config:  &UsernamePasswordConfig{MinPasswordLength: -1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestLDAPConfig_Validate tests LDAPConfig validation.
func TestLDAPConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *LDAPConfig
		wantErr bool
	}{
		{
			name: "valid",
			config: &LDAPConfig{
				URL:    "ldap://ldap.example.com",
				BaseDN: "dc=example,dc=com",
			},
			wantErr: false,
		},
		{
			name:    "missing_url",
			config:  &LDAPConfig{BaseDN: "dc=example,dc=com"},
			wantErr: true,
		},
		{
			name:    "missing_basedn",
			config:  &LDAPConfig{URL: "ldap://ldap.example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestOAuth2Config_Validate tests OAuth2Config validation.
func TestOAuth2Config_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *OAuth2Config
		wantErr bool
	}{
		{
			name: "valid_with_discovery",
			config: &OAuth2Config{
				ClientID:     "my-client",
				ProviderURL:  "https://auth.example.com",
				UseDiscovery: true,
			},
			wantErr: false,
		},
		{
			name: "valid_without_discovery",
			config: &OAuth2Config{
				ClientID:     "my-client",
				AuthorizeURL: "https://auth.example.com/authorize",
				TokenURL:     "https://auth.example.com/token",
				UseDiscovery: false,
			},
			wantErr: false,
		},
		{
			name:    "missing_client_id",
			config:  &OAuth2Config{ProviderURL: "https://auth.example.com", UseDiscovery: true},
			wantErr: true,
		},
		{
			name: "discovery_without_provider_url",
			config: &OAuth2Config{
				ClientID:     "my-client",
				UseDiscovery: true,
			},
			wantErr: true,
		},
		{
			name: "no_discovery_missing_urls",
			config: &OAuth2Config{
				ClientID:     "my-client",
				UseDiscovery: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSAMLConfig_Validate tests SAMLConfig validation.
func TestSAMLConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *SAMLConfig
		wantErr bool
	}{
		{
			name: "valid_with_metadata_url",
			config: &SAMLConfig{
				MetadataURL: "https://idp.example.com/metadata",
				EntityID:    "https://myapp.example.com",
			},
			wantErr: false,
		},
		{
			name: "valid_with_metadata_xml",
			config: &SAMLConfig{
				MetadataXML: "<xml>...</xml>",
				EntityID:    "https://myapp.example.com",
			},
			wantErr: false,
		},
		{
			name:    "missing_metadata",
			config:  &SAMLConfig{EntityID: "https://myapp.example.com"},
			wantErr: true,
		},
		{
			name: "missing_entity_id",
			config: &SAMLConfig{
				MetadataURL: "https://idp.example.com/metadata",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
