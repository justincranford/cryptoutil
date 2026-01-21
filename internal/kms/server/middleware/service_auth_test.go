// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServiceAuthMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  ServiceAuthConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with JWT",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodJWT},
				JWTConfig: &JWTValidatorConfig{
					JWKSURL: "https://example.com/.well-known/jwks.json",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with mTLS",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodMTLS},
				MTLSConfig: &MTLSConfig{
					RequireClientCert: true,
					AllowedCNs:        []string{"service-a.example.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with API key",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodAPIKey},
				APIKeyConfig: &APIKeyConfig{
					HeaderName: "X-API-Key",
					ValidKeys: map[string]string{
						"secret-key-123": "service-a",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with multiple methods",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodJWT, AuthMethodMTLS, AuthMethodAPIKey},
				JWTConfig: &JWTValidatorConfig{
					JWKSURL: "https://example.com/.well-known/jwks.json",
				},
				MTLSConfig: &MTLSConfig{
					RequireClientCert: false,
				},
				APIKeyConfig: &APIKeyConfig{
					ValidKeys: map[string]string{"key": "service"},
				},
			},
			wantErr: false,
		},
		{
			name:    "no allowed methods",
			config:  ServiceAuthConfig{},
			wantErr: true,
			errMsg:  "at least one auth method must be allowed",
		},
		{
			name: "invalid JWT config",
			config: ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodJWT},
				JWTConfig:      &JWTValidatorConfig{}, // Missing JWKS URL.
			},
			wantErr: true,
			errMsg:  "JWKS URL is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			middleware, err := NewServiceAuthMiddleware(tc.config)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, middleware)
			} else {
				require.NoError(t, err)
				require.NotNil(t, middleware)
			}
		})
	}
}

func TestAuthMethodConstants(t *testing.T) {
	t.Parallel()

	// Verify auth method constants have expected string values.
	tests := []struct {
		method   AuthMethod
		expected string
	}{
		{AuthMethodJWT, "jwt"},
		{AuthMethodMTLS, "mtls"},
		{AuthMethodAPIKey, "api-key"},
		{AuthMethodClientCredentials, "client-credentials"},
	}

	for _, tc := range tests {
		t.Run(string(tc.method), func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, string(tc.method))
		})
	}
}

func TestGetServiceAuthInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() context.Context
		expected *ServiceAuthInfo
	}{
		{
			name: "auth info present",
			setup: func() context.Context {
				info := &ServiceAuthInfo{
					Method:      AuthMethodJWT,
					ServiceName: "test-service",
					Subject:     "user@example.com",
					Scopes:      []string{"read", "write"},
				}

				return context.WithValue(context.Background(), ServiceAuthContextKey{}, info)
			},
			expected: &ServiceAuthInfo{
				Method:      AuthMethodJWT,
				ServiceName: "test-service",
				Subject:     "user@example.com",
				Scopes:      []string{"read", "write"},
			},
		},
		{
			name: "auth info absent",
			setup: func() context.Context {
				return context.Background()
			},
			expected: nil,
		},
		{
			name: "wrong type in context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ServiceAuthContextKey{}, "wrong-type")
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setup()
			info := GetServiceAuthInfo(ctx)

			if tc.expected == nil {
				require.Nil(t, info)
			} else {
				require.NotNil(t, info)
				require.Equal(t, tc.expected.Method, info.Method)
				require.Equal(t, tc.expected.ServiceName, info.ServiceName)
				require.Equal(t, tc.expected.Subject, info.Subject)
				require.Equal(t, tc.expected.Scopes, info.Scopes)
			}
		})
	}
}

func TestServiceAuthMiddleware_IsAllowedValue(t *testing.T) {
	t.Parallel()

	middleware := &ServiceAuthMiddleware{
		config: ServiceAuthConfig{
			AllowedMethods: []AuthMethod{AuthMethodAPIKey},
		},
	}

	tests := []struct {
		name     string
		value    string
		allowed  []string
		expected bool
	}{
		{
			name:     "value in list",
			value:    "service-a",
			allowed:  []string{"service-a", "service-b"},
			expected: true,
		},
		{
			name:     "value not in list",
			value:    "service-c",
			allowed:  []string{"service-a", "service-b"},
			expected: false,
		},
		{
			name:     "empty allowed list",
			value:    "service-a",
			allowed:  []string{},
			expected: false,
		},
		{
			name:     "empty value",
			value:    "",
			allowed:  []string{"service-a"},
			expected: false,
		},
		{
			name:     "exact match required",
			value:    "service",
			allowed:  []string{"service-a", "service-b"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := middleware.isAllowedValue(tc.value, tc.allowed)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestMTLSConfig_Validation(t *testing.T) {
	t.Parallel()

	// Test that MTLSConfig fields work correctly.
	tests := []struct {
		name   string
		config MTLSConfig
	}{
		{
			name: "require client cert",
			config: MTLSConfig{
				RequireClientCert: true,
			},
		},
		{
			name: "optional client cert",
			config: MTLSConfig{
				RequireClientCert: false,
			},
		},
		{
			name: "with allowed CNs",
			config: MTLSConfig{
				AllowedCNs: []string{"cn1.example.com", "cn2.example.com"},
			},
		},
		{
			name: "with allowed OUs",
			config: MTLSConfig{
				AllowedOUs: []string{"Engineering", "Operations"},
			},
		},
		{
			name: "with DNS SANs",
			config: MTLSConfig{
				AllowedDNSSANs: []string{"service.example.com"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			middleware, err := NewServiceAuthMiddleware(ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodMTLS},
				MTLSConfig:     &tc.config,
			})
			require.NoError(t, err)
			require.NotNil(t, middleware)
		})
	}
}

func TestAPIKeyConfig_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config APIKeyConfig
	}{
		{
			name: "with static keys",
			config: APIKeyConfig{
				ValidKeys: map[string]string{
					"key1": "service-a",
					"key2": "service-b",
				},
			},
		},
		{
			name: "with custom header",
			config: APIKeyConfig{
				HeaderName: "X-Custom-Key",
				ValidKeys: map[string]string{
					"key": "service",
				},
			},
		},
		{
			name: "with validator function",
			config: APIKeyConfig{
				KeyValidator: func(_ context.Context, _ string) (string, bool, error) {
					return "validated-service", true, nil
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			middleware, err := NewServiceAuthMiddleware(ServiceAuthConfig{
				AllowedMethods: []AuthMethod{AuthMethodAPIKey},
				APIKeyConfig:   &tc.config,
			})
			require.NoError(t, err)
			require.NotNil(t, middleware)
		})
	}
}

func TestClientCredentialsConfig_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config ClientCredentialsConfig
	}{
		{
			name: "with endpoints",
			config: ClientCredentialsConfig{
				TokenEndpoint:         "https://example.com/oauth2/token",
				IntrospectionEndpoint: "https://example.com/oauth2/introspect",
			},
		},
		{
			name: "with client ID validator",
			config: ClientCredentialsConfig{
				IntrospectionEndpoint: "https://example.com/oauth2/introspect",
				ValidateClientID: func(clientID string) bool {
					return clientID == "allowed-client"
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			middleware, err := NewServiceAuthMiddleware(ServiceAuthConfig{
				AllowedMethods:          []AuthMethod{AuthMethodClientCredentials},
				ClientCredentialsConfig: &tc.config,
			})
			require.NoError(t, err)
			require.NotNil(t, middleware)
		})
	}
}

func TestServiceAuthInfo_Fields(t *testing.T) {
	t.Parallel()

	info := &ServiceAuthInfo{
		Method:        AuthMethodMTLS,
		ServiceName:   "backend-service",
		ClientID:      "client-123",
		Subject:       "service@example.com",
		CertificateCN: "backend-service.internal",
		Scopes:        []string{"kms:read", "kms:write"},
		Metadata: map[string]any{
			"tenant_id": "tenant-456",
			"role":      "admin",
		},
	}

	require.Equal(t, AuthMethodMTLS, info.Method)
	require.Equal(t, "backend-service", info.ServiceName)
	require.Equal(t, "client-123", info.ClientID)
	require.Equal(t, "service@example.com", info.Subject)
	require.Equal(t, "backend-service.internal", info.CertificateCN)
	require.Len(t, info.Scopes, 2)
	require.Contains(t, info.Scopes, "kms:read")
	require.Contains(t, info.Scopes, "kms:write")
	require.Equal(t, "tenant-456", info.Metadata["tenant_id"])
	require.Equal(t, "admin", info.Metadata["role"])
}

func TestConfigureTLSForMTLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		requireClientCert bool
	}{
		{
			name:              "require client cert",
			requireClientCert: true,
		},
		{
			name:              "optional client cert",
			requireClientCert: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tlsConfig := ConfigureTLSForMTLS(tc.requireClientCert)
			require.NotNil(t, tlsConfig)

			// Verify TLS 1.3 minimum.
			const tlsVersion13 uint16 = 0x0304

			require.Equal(t, tlsVersion13, tlsConfig.MinVersion)
		})
	}
}

func TestServiceAuthMiddleware_ErrorDetailLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		errorLevel    string
		expectedLevel string
		expectDefault bool
	}{
		{
			name:          "explicit minimal level",
			errorLevel:    "minimal",
			expectedLevel: "minimal",
		},
		{
			name:          "explicit detailed level",
			errorLevel:    "detailed",
			expectedLevel: "detailed",
		},
		{
			name:          "explicit debug level",
			errorLevel:    "debug",
			expectedLevel: "debug",
		},
		{
			name:          "empty defaults to minimal",
			errorLevel:    "",
			expectedLevel: "minimal",
			expectDefault: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			middleware, err := NewServiceAuthMiddleware(ServiceAuthConfig{
				AllowedMethods:   []AuthMethod{AuthMethodAPIKey},
				APIKeyConfig:     &APIKeyConfig{ValidKeys: map[string]string{"k": "v"}},
				ErrorDetailLevel: tc.errorLevel,
			})
			require.NoError(t, err)
			require.NotNil(t, middleware)
			require.Equal(t, tc.expectedLevel, middleware.config.ErrorDetailLevel)
		})
	}
}
