// Copyright (c) 2025 Justin Cranford
//
//

package security

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

//nolint:thelper // testFn inline functions are NOT test helpers - they're test implementations
func TestPolicyConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                          string
		constructor                   func() *ClientAuthPolicy
		wantAllowedMethodsCount       int
		wantRequireMTLS               bool
		wantRequireJWTSignature       bool
		wantRequireCertValidation     bool
		wantAllowSelfSigned           bool
		wantAllowedJWTAlgorithmsCount int
		wantSpecificMethods           []cryptoutilIdentityDomain.ClientAuthMethod
		wantSpecificAlgorithms        []string
	}{
		{
			name:                      "default_policy",
			constructor:               DefaultClientAuthPolicy,
			wantAllowedMethodsCount:   cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
			wantRequireMTLS:           false,
			wantRequireJWTSignature:   true,
			wantRequireCertValidation: true,
			wantAllowSelfSigned:       false,
		},
		{
			name:                          "strict_policy",
			constructor:                   StrictClientAuthPolicy,
			wantAllowedMethodsCount:       2,
			wantRequireMTLS:               true,
			wantRequireJWTSignature:       true,
			wantAllowedJWTAlgorithmsCount: 2,
			wantSpecificAlgorithms:        []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgES256},
		},
		{
			name:                    "public_client_policy",
			constructor:             PublicClientAuthPolicy,
			wantAllowedMethodsCount: 1,
			wantRequireMTLS:         false,
			wantRequireJWTSignature: false,
			wantSpecificMethods:     []cryptoutilIdentityDomain.ClientAuthMethod{cryptoutilIdentityDomain.ClientAuthMethodNone},
		},
		{
			name:                    "development_policy",
			constructor:             DevelopmentClientAuthPolicy,
			wantAllowedMethodsCount: cryptoutilSharedMagic.GitRecentActivityDays,
			wantRequireMTLS:         false,
			wantRequireJWTSignature: false,
			wantAllowSelfSigned:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			policy := tc.constructor()
			require.NotNil(t, policy)
			require.Len(t, policy.AllowedMethods, tc.wantAllowedMethodsCount)

			if tc.wantRequireMTLS {
				require.True(t, policy.RequireMTLS)
			}

			if tc.wantRequireJWTSignature {
				require.True(t, policy.RequireJWTSignature)
			}

			if tc.wantRequireCertValidation {
				require.True(t, policy.RequireCertificateValidation)
			}

			if tc.wantAllowSelfSigned {
				require.True(t, policy.AllowSelfSignedCertificates)
			}

			if tc.wantAllowedJWTAlgorithmsCount > 0 {
				require.Len(t, policy.AllowedJWTAlgorithms, tc.wantAllowedJWTAlgorithmsCount)
			}

			for _, method := range tc.wantSpecificMethods {
				require.Contains(t, policy.AllowedMethods, method)
			}

			for _, algo := range tc.wantSpecificAlgorithms {
				require.Contains(t, policy.AllowedJWTAlgorithms, algo)
			}
		})
	}
}

func TestClientAuthPolicyManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		testFn func(t *testing.T, manager *ClientAuthPolicyManager)
	}{
		{
			name: "get_default_policy",
			testFn: func(t *testing.T, manager *ClientAuthPolicyManager) {
				t.Helper()

				policy, err := manager.GetPolicy("default")
				require.NoError(t, err)
				require.NotNil(t, policy)
			},
		},
		{
			name: "get_strict_policy",
			testFn: func(t *testing.T, manager *ClientAuthPolicyManager) {
				t.Helper()

				policy, err := manager.GetPolicy("strict")
				require.NoError(t, err)
				require.NotNil(t, policy)
			},
		},
		{
			name: "get_public_policy",
			testFn: func(t *testing.T, manager *ClientAuthPolicyManager) {
				t.Helper()

				policy, err := manager.GetPolicy(cryptoutilSharedMagic.SubjectTypePublic)
				require.NoError(t, err)
				require.NotNil(t, policy)
			},
		},
		{
			name: "get_development_policy",
			testFn: func(t *testing.T, manager *ClientAuthPolicyManager) {
				t.Helper()

				policy, err := manager.GetPolicy("development")
				require.NoError(t, err)
				require.NotNil(t, policy)
			},
		},
		{
			name: "get_nonexistent_policy",
			testFn: func(t *testing.T, manager *ClientAuthPolicyManager) {
				t.Helper()

				_, err := manager.GetPolicy("nonexistent")
				require.Error(t, err)
			},
		},
		{
			name: "register_custom_policy",
			testFn: func(t *testing.T, manager *ClientAuthPolicyManager) {
				t.Helper()

				customPolicy := &ClientAuthPolicy{
					AllowedMethods: []cryptoutilIdentityDomain.ClientAuthMethod{
						cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
					},
					RequireMTLS: true,
				}

				manager.RegisterPolicy("custom", customPolicy)

				retrievedPolicy, err := manager.GetPolicy("custom")
				require.NoError(t, err)
				require.Equal(t, customPolicy, retrievedPolicy)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			manager := NewClientAuthPolicyManager()
			tc.testFn(t, manager)
		})
	}
}

func TestClientAuthPolicy_ValidateClientAuthMethod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		policy  *ClientAuthPolicy
		method  cryptoutilIdentityDomain.ClientAuthMethod
		wantErr bool
	}{
		{
			name:    "allowed_method",
			policy:  DefaultClientAuthPolicy(),
			method:  cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
			wantErr: false,
		},
		{
			name:    "disallowed_method",
			policy:  DefaultClientAuthPolicy(),
			method:  cryptoutilIdentityDomain.ClientAuthMethodNone,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.policy.ValidateClientAuthMethod(ctx, tc.method)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClientAuthPolicy_ValidateJWTAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		policy    *ClientAuthPolicy
		algorithm string
		wantErr   bool
	}{
		{
			name:      "allowed_algorithm",
			policy:    StrictClientAuthPolicy(),
			algorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			wantErr:   false,
		},
		{
			name:      "disallowed_algorithm",
			policy:    StrictClientAuthPolicy(),
			algorithm: cryptoutilSharedMagic.JoseAlgHS256,
			wantErr:   true,
		},
		{
			name: "no_requirement_any_algorithm_passes",
			policy: func() *ClientAuthPolicy {
				policy := DevelopmentClientAuthPolicy()
				policy.RequireJWTSignature = false

				return policy
			}(),
			algorithm: cryptoutilSharedMagic.JoseAlgHS256,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.policy.ValidateJWTAlgorithm(ctx, tc.algorithm)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
