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
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
			name:    testValid,
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
			name: testValid,
			config: &LDAPConfig{
				URL:    testLdapLdapExampleCom,
				BaseDN: testDcExampleDcCom,
			},
			wantErr: false,
		},
		{
			name:    "missing_url",
			config:  &LDAPConfig{BaseDN: testDcExampleDcCom},
			wantErr: true,
		},
		{
			name:    "missing_basedn",
			config:  &LDAPConfig{URL: testLdapLdapExampleCom},
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
				ClientID:     testMyClient,
				ProviderURL:  testHTTPSAuthExampleCom,
				UseDiscovery: true,
			},
			wantErr: false,
		},
		{
			name: "valid_without_discovery",
			config: &OAuth2Config{
				ClientID:     testMyClient,
				AuthorizeURL: testHTTPSAuthExampleComAuthorize,
				TokenURL:     testHTTPSAuthExampleComToken,
				UseDiscovery: false,
			},
			wantErr: false,
		},
		{
			name:    "missing_client_id",
			config:  &OAuth2Config{ProviderURL: testHTTPSAuthExampleCom, UseDiscovery: true},
			wantErr: true,
		},
		{
			name: "discovery_without_provider_url",
			config: &OAuth2Config{
				ClientID:     testMyClient,
				UseDiscovery: true,
			},
			wantErr: true,
		},
		{
			name: "no_discovery_missing_urls",
			config: &OAuth2Config{
				ClientID:     testMyClient,
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
				MetadataURL: testHTTPSIdpExampleComMetadata,
				EntityID:    testHTTPSMyappExampleCom,
			},
			wantErr: false,
		},
		{
			name: "valid_with_metadata_xml",
			config: &SAMLConfig{
				MetadataXML: "<xml>...</xml>",
				EntityID:    testHTTPSMyappExampleCom,
			},
			wantErr: false,
		},
		{
			name:    "missing_metadata",
			config:  &SAMLConfig{EntityID: testHTTPSMyappExampleCom},
			wantErr: true,
		},
		{
			name: "missing_entity_id",
			config: &SAMLConfig{
				MetadataURL: testHTTPSIdpExampleComMetadata,
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
