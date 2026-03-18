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

// TestRealmConfig_GetType tests all 14 config types implement GetType() correctly.
func TestRealmConfig_GetType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   RealmConfig
		expected RealmType
	}{
		// Federated realm types (4).
		{
			name:     "UsernamePasswordConfig",
			config:   &UsernamePasswordConfig{},
			expected: RealmTypeUsernamePassword,
		},
		{
			name:     "LDAPConfig",
			config:   &LDAPConfig{},
			expected: RealmTypeLDAP,
		},
		{
			name:     "OAuth2Config",
			config:   &OAuth2Config{},
			expected: RealmTypeOAuth2,
		},
		{
			name:     "SAMLConfig",
			config:   &SAMLConfig{},
			expected: RealmTypeSAML,
		},
		// Browser realm types (6).
		{
			name:     "JWESessionCookieConfig",
			config:   &JWESessionCookieConfig{},
			expected: RealmTypeJWESessionCookie,
		},
		{
			name:     "JWSSessionCookieConfig",
			config:   &JWSSessionCookieConfig{},
			expected: RealmTypeJWSSessionCookie,
		},
		{
			name:     "OpaqueSessionCookieConfig",
			config:   &OpaqueSessionCookieConfig{},
			expected: RealmTypeOpaqueSessionCookie,
		},
		{
			name:     "BasicUsernamePasswordConfig",
			config:   &BasicUsernamePasswordConfig{},
			expected: RealmTypeBasicUsernamePassword,
		},
		{
			name:     "BearerAPITokenConfig",
			config:   &BearerAPITokenConfig{},
			expected: RealmTypeBearerAPIToken,
		},
		{
			name:     "HTTPSClientCertConfig",
			config:   &HTTPSClientCertConfig{},
			expected: RealmTypeHTTPSClientCert,
		},
		// Service realm types (4).
		{
			name:     "JWESessionTokenConfig",
			config:   &JWESessionTokenConfig{},
			expected: RealmTypeJWESessionToken,
		},
		{
			name:     "JWSSessionTokenConfig",
			config:   &JWSSessionTokenConfig{},
			expected: RealmTypeJWSSessionToken,
		},
		{
			name:     "OpaqueSessionTokenConfig",
			config:   &OpaqueSessionTokenConfig{},
			expected: RealmTypeOpaqueSessionToken,
		},
		{
			name:     "BasicClientIDSecretConfig",
			config:   &BasicClientIDSecretConfig{},
			expected: RealmTypeBasicClientIDSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.config.GetType()
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestRealmConfig_Validate_Valid tests valid configurations pass validation.
func TestRealmConfig_Validate_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config RealmConfig
	}{
		// Federated realm types (4).
		{
			name: "UsernamePasswordConfig_Valid",
			config: &UsernamePasswordConfig{
				MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength,
				RequireUppercase:  true,
				RequireLowercase:  true,
				RequireDigit:      true,
				RequireSpecial:    false,
			},
		},
		{
			name: "LDAPConfig_Valid",
			config: &LDAPConfig{
				URL:    "ldap://ldap.example.com:389",
				BaseDN: "dc=example,dc=com",
			},
		},
		{
			name: "OAuth2Config_Valid_WithDiscovery",
			config: &OAuth2Config{
				ClientID:     cryptoutilSharedMagic.TestClientID,
				ProviderURL:  "https://auth.example.com",
				UseDiscovery: true,
			},
		},
		{
			name: "OAuth2Config_Valid_WithoutDiscovery",
			config: &OAuth2Config{
				ClientID:     cryptoutilSharedMagic.TestClientID,
				UseDiscovery: false,
				AuthorizeURL: "https://auth.example.com/authorize",
				TokenURL:     "https://auth.example.com/token",
			},
		},
		{
			name: "SAMLConfig_Valid_WithMetadataURL",
			config: &SAMLConfig{
				MetadataURL: "https://idp.example.com/metadata",
				EntityID:    "test-entity-id",
			},
		},
		{
			name: "SAMLConfig_Valid_WithMetadataXML",
			config: &SAMLConfig{
				MetadataXML: "<EntityDescriptor>...</EntityDescriptor>",
				EntityID:    "test-entity-id",
			},
		},
		// Browser realm types (6).
		{
			name: "JWESessionCookieConfig_Valid",
			config: &JWESessionCookieConfig{
				EncryptionAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWEAlgorithm,
				SessionExpiryMinutes: 15,
			},
		},
		{
			name: "JWSSessionCookieConfig_Valid",
			config: &JWSSessionCookieConfig{
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
				SessionExpiryMinutes: 15,
			},
		},
		{
			name: "OpaqueSessionCookieConfig_Valid",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
				SessionExpiryMinutes: 15,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
		},
		{
			name: "OpaqueSessionCookieConfig_Valid_Redis",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
				SessionExpiryMinutes: 15,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeRedis,
			},
		},
		{
			name: "BasicUsernamePasswordConfig_Valid",
			config: &BasicUsernamePasswordConfig{
				MinPasswordLength: cryptoutilSharedMagic.IMMinPasswordLength,
				RequireUppercase:  true,
				RequireLowercase:  true,
				RequireDigit:      true,
				RequireSpecial:    false,
			},
		},
		{
			name: "BearerAPITokenConfig_Valid",
			config: &BearerAPITokenConfig{
				TokenExpiryDays:   cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days,
				TokenLengthBytes:  cryptoutilSharedMagic.MinSerialNumberBits,
				AllowRefreshToken: true,
			},
		},
		{
			name: "HTTPSClientCertConfig_Valid_WithCAs",
			config: &HTTPSClientCertConfig{
				RequireClientCert: true,
				TrustedCAs:        []string{"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"},
				ValidateOCSP:      true,
				ValidateCRL:       true,
			},
		},
		{
			name: "HTTPSClientCertConfig_Valid_WithoutRequire",
			config: &HTTPSClientCertConfig{
				RequireClientCert: false,
				TrustedCAs:        nil, // Empty is valid when RequireClientCert is false
			},
		},
		// Service realm types (4).
		{
			name: "JWESessionTokenConfig_Valid",
			config: &JWESessionTokenConfig{
				EncryptionAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWEAlgorithm,
				TokenExpiryMinutes:  cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds,
			},
		},
		{
			name: "JWSSessionTokenConfig_Valid",
			config: &JWSSessionTokenConfig{
				SigningAlgorithm:   cryptoutilSharedMagic.JoseAlgES256,
				TokenExpiryMinutes: cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds,
			},
		},
		{
			name: "OpaqueSessionTokenConfig_Valid",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
				TokenExpiryMinutes: cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds,
				StorageType:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
		},
		{
			name: "BasicClientIDSecretConfig_Valid",
			config: &BasicClientIDSecretConfig{
				MinSecretLength:  cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireDigit:     true,
				RequireSpecial:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			require.NoError(t, err)
		})
	}
}

// TestRealmConfig_Validate_Errors tests that invalid configurations return validation errors.
