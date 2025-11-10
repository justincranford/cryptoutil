package security

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func TestDefaultClientAuthPolicy(t *testing.T) {
	t.Parallel()

	policy := DefaultClientAuthPolicy()

	require.NotNil(t, policy)
	require.Len(t, policy.AllowedMethods, 5)
	require.False(t, policy.RequireMTLS)
	require.True(t, policy.RequireJWTSignature)
	require.True(t, policy.RequireCertificateValidation)
	require.False(t, policy.AllowSelfSignedCertificates)
}

func TestStrictClientAuthPolicy(t *testing.T) {
	t.Parallel()

	policy := StrictClientAuthPolicy()

	require.NotNil(t, policy)
	require.Len(t, policy.AllowedMethods, 2)
	require.True(t, policy.RequireMTLS)
	require.True(t, policy.RequireJWTSignature)
	require.Len(t, policy.AllowedJWTAlgorithms, 2)
	require.Contains(t, policy.AllowedJWTAlgorithms, "RS256")
	require.Contains(t, policy.AllowedJWTAlgorithms, "ES256")
}

func TestPublicClientAuthPolicy(t *testing.T) {
	t.Parallel()

	policy := PublicClientAuthPolicy()

	require.NotNil(t, policy)
	require.Len(t, policy.AllowedMethods, 1)
	require.Contains(t, policy.AllowedMethods, cryptoutilIdentityDomain.ClientAuthMethodNone)
	require.False(t, policy.RequireMTLS)
	require.False(t, policy.RequireJWTSignature)
}

func TestDevelopmentClientAuthPolicy(t *testing.T) {
	t.Parallel()

	policy := DevelopmentClientAuthPolicy()

	require.NotNil(t, policy)
	require.Len(t, policy.AllowedMethods, 7)
	require.False(t, policy.RequireMTLS)
	require.False(t, policy.RequireJWTSignature)
	require.True(t, policy.AllowSelfSignedCertificates)
}

func TestClientAuthPolicyManager_GetPolicy(t *testing.T) {
	t.Parallel()

	manager := NewClientAuthPolicyManager()

	// Test default policy.
	policy, err := manager.GetPolicy("default")
	require.NoError(t, err)
	require.NotNil(t, policy)

	// Test strict policy.
	policy, err = manager.GetPolicy("strict")
	require.NoError(t, err)
	require.NotNil(t, policy)

	// Test public policy.
	policy, err = manager.GetPolicy("public")
	require.NoError(t, err)
	require.NotNil(t, policy)

	// Test development policy.
	policy, err = manager.GetPolicy("development")
	require.NoError(t, err)
	require.NotNil(t, policy)

	// Test non-existent policy.
	_, err = manager.GetPolicy("nonexistent")
	require.Error(t, err)
}

func TestClientAuthPolicyManager_RegisterPolicy(t *testing.T) {
	t.Parallel()

	manager := NewClientAuthPolicyManager()

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
}

func TestClientAuthPolicy_ValidateClientAuthMethod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	policy := DefaultClientAuthPolicy()

	// Test allowed method.
	err := policy.ValidateClientAuthMethod(ctx, cryptoutilIdentityDomain.ClientAuthMethodSecretBasic)
	require.NoError(t, err)

	// Test disallowed method.
	err = policy.ValidateClientAuthMethod(ctx, cryptoutilIdentityDomain.ClientAuthMethodNone)
	require.Error(t, err)
}

func TestClientAuthPolicy_ValidateJWTAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	policy := StrictClientAuthPolicy()

	// Test allowed algorithm.
	err := policy.ValidateJWTAlgorithm(ctx, "RS256")
	require.NoError(t, err)

	// Test disallowed algorithm.
	err = policy.ValidateJWTAlgorithm(ctx, "HS256")
	require.Error(t, err)
}

func TestClientAuthPolicy_ValidateJWTAlgorithm_NoRequirement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	policy := DevelopmentClientAuthPolicy()
	policy.RequireJWTSignature = false

	// When JWT signature not required, any algorithm passes.
	err := policy.ValidateJWTAlgorithm(ctx, "HS256")
	require.NoError(t, err)
}
