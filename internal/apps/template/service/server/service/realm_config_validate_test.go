// Copyright (c) 2025 Justin Cranford
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
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestRealmConfig_Validate_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         RealmConfig
		expectedSubstr string
	}{
		// UsernamePasswordConfig errors.
		{
			name: "UsernamePasswordConfig_MinPasswordLengthZero",
			config: &UsernamePasswordConfig{
				MinPasswordLength: 0,
			},
			expectedSubstr: "min_password_length must be at least 1",
		},
		// LDAPConfig errors.
		{
			name: "LDAPConfig_MissingURL",
			config: &LDAPConfig{
				BaseDN: "dc=example,dc=com",
			},
			expectedSubstr: "url is required",
		},
		{
			name: "LDAPConfig_MissingBaseDN",
			config: &LDAPConfig{
				URL: "ldap://ldap.example.com:389",
			},
			expectedSubstr: "base_dn is required",
		},
		// OAuth2Config errors.
		{
			name: "OAuth2Config_MissingClientID",
			config: &OAuth2Config{
				UseDiscovery: true,
				ProviderURL:  "https://auth.example.com",
			},
			expectedSubstr: "client_id is required",
		},
		{
			name: "OAuth2Config_MissingProviderURL_WithDiscovery",
			config: &OAuth2Config{
				ClientID:     "test-client-id",
				UseDiscovery: true,
			},
			expectedSubstr: "provider_url is required when use_discovery is true",
		},
		{
			name: "OAuth2Config_MissingURLs_WithoutDiscovery",
			config: &OAuth2Config{
				ClientID:     "test-client-id",
				UseDiscovery: false,
			},
			expectedSubstr: "authorize_url and token_url are required when use_discovery is false",
		},
		// SAMLConfig errors.
		{
			name: "SAMLConfig_MissingMetadata",
			config: &SAMLConfig{
				EntityID: "test-entity-id",
			},
			expectedSubstr: "either metadata_url or metadata_xml is required",
		},
		{
			name: "SAMLConfig_MissingEntityID",
			config: &SAMLConfig{
				MetadataURL: "https://idp.example.com/metadata",
			},
			expectedSubstr: "entity_id is required",
		},
		// JWESessionCookieConfig errors.
		{
			name: "JWESessionCookieConfig_SessionExpiryZero",
			config: &JWESessionCookieConfig{
				SessionExpiryMinutes: 0,
			},
			expectedSubstr: "session_expiry_minutes must be at least 1",
		},
		// JWSSessionCookieConfig errors.
		{
			name: "JWSSessionCookieConfig_SessionExpiryZero",
			config: &JWSSessionCookieConfig{
				SessionExpiryMinutes: 0,
			},
			expectedSubstr: "session_expiry_minutes must be at least 1",
		},
		// OpaqueSessionCookieConfig errors.
		{
			name: "OpaqueSessionCookieConfig_TokenLengthTooSmall",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     8, // Below minimum
				SessionExpiryMinutes: 15,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "token_length_bytes must be at least",
		},
		{
			name: "OpaqueSessionCookieConfig_SessionExpiryZero",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     32,
				SessionExpiryMinutes: 0,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "session_expiry_minutes must be at least 1",
		},
		{
			name: "OpaqueSessionCookieConfig_InvalidStorageType",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     32,
				SessionExpiryMinutes: 15,
				StorageType:          "invalid",
			},
			expectedSubstr: "storage_type must be",
		},
		// BasicUsernamePasswordConfig errors.
		{
			name: "BasicUsernamePasswordConfig_MinPasswordLengthZero",
			config: &BasicUsernamePasswordConfig{
				MinPasswordLength: 0,
			},
			expectedSubstr: "min_password_length must be at least 1",
		},
		// BearerAPITokenConfig errors.
		{
			name: "BearerAPITokenConfig_TokenExpiryZero",
			config: &BearerAPITokenConfig{
				TokenExpiryDays:  0,
				TokenLengthBytes: 64,
			},
			expectedSubstr: "token_expiry_days must be at least 1",
		},
		{
			name: "BearerAPITokenConfig_TokenLengthTooSmall",
			config: &BearerAPITokenConfig{
				TokenExpiryDays:  30,
				TokenLengthBytes: 16, // Below minimum
			},
			expectedSubstr: "token_length_bytes must be at least",
		},
		// HTTPSClientCertConfig errors.
		{
			name: "HTTPSClientCertConfig_RequireClientCertWithoutCAs",
			config: &HTTPSClientCertConfig{
				RequireClientCert: true,
				TrustedCAs:        nil,
			},
			expectedSubstr: "trusted_cas is required when require_client_cert is true",
		},
		// JWESessionTokenConfig errors.
		{
			name: "JWESessionTokenConfig_TokenExpiryZero",
			config: &JWESessionTokenConfig{
				TokenExpiryMinutes: 0,
			},
			expectedSubstr: "token_expiry_minutes must be at least 1",
		},
		// JWSSessionTokenConfig errors.
		{
			name: "JWSSessionTokenConfig_TokenExpiryZero",
			config: &JWSSessionTokenConfig{
				TokenExpiryMinutes: 0,
			},
			expectedSubstr: "token_expiry_minutes must be at least 1",
		},
		// OpaqueSessionTokenConfig errors.
		{
			name: "OpaqueSessionTokenConfig_TokenLengthTooSmall",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   8, // Below minimum
				TokenExpiryMinutes: 60,
				StorageType:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "token_length_bytes must be at least",
		},
		{
			name: "OpaqueSessionTokenConfig_TokenExpiryZero",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   32,
				TokenExpiryMinutes: 0,
				StorageType:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
			expectedSubstr: "token_expiry_minutes must be at least 1",
		},
		{
			name: "OpaqueSessionTokenConfig_InvalidStorageType",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   32,
				TokenExpiryMinutes: 60,
				StorageType:        "invalid",
			},
			expectedSubstr: "storage_type must be",
		},
		// BasicClientIDSecretConfig errors.
		{
			name: "BasicClientIDSecretConfig_MinSecretLengthZero",
			config: &BasicClientIDSecretConfig{
				MinSecretLength: 0,
			},
			expectedSubstr: "min_secret_length must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedSubstr)
		})
	}
}

// TestParseRealmConfig_AllTypes tests parsing JSON config for all 14 realm types.
