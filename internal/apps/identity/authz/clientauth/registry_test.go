// Copyright (c) 2025 Justin Cranford
//
//

package clientauth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestRegistry_Creation validates registry creation with all authenticators.
func TestRegistry_Creation(t *testing.T) {
	t.Parallel()

	repoFactory := createRegistryTestRepoFactory(t)
	config := createRegistryTestConfig()

	registry := cryptoutilIdentityClientAuth.NewRegistry(repoFactory, config, nil)
	require.NotNil(t, registry, "Registry should not be nil")

	// Verify all expected authenticators are registered.
	methods := []string{
		"client_secret_basic",
		"client_secret_post",
		"tls_client_auth",
		"self_signed_tls_client_auth",
		"private_key_jwt",
		"client_secret_jwt",
	}

	// Secret-based methods share the same authenticator, so Method() returns "client_secret_basic" for both.
	for _, method := range methods {
		auth, ok := registry.GetAuthenticator(method)
		require.True(t, ok, "Authenticator %s should be registered", method)
		require.NotNil(t, auth, "Authenticator %s should not be nil", method)
		// Note: client_secret_post uses same authenticator as client_secret_basic,
		// so we only verify the authenticator exists, not the exact method name.
	}
}

// TestRegistry_GetAuthenticator validates authenticator retrieval.
func TestRegistry_GetAuthenticator(t *testing.T) {
	t.Parallel()

	repoFactory := createRegistryTestRepoFactory(t)
	config := createRegistryTestConfig()

	registry := cryptoutilIdentityClientAuth.NewRegistry(repoFactory, config, nil)

	auth, ok := registry.GetAuthenticator("client_secret_basic")
	require.True(t, ok, "Should find client_secret_basic authenticator")
	require.NotNil(t, auth, "Authenticator should not be nil")
	require.Equal(t, "client_secret_basic", auth.Method(), "Method should match")
}

// TestRegistry_GetAuthenticator_NotFound validates missing authenticator handling.
func TestRegistry_GetAuthenticator_NotFound(t *testing.T) {
	t.Parallel()

	repoFactory := createRegistryTestRepoFactory(t)
	config := createRegistryTestConfig()

	registry := cryptoutilIdentityClientAuth.NewRegistry(repoFactory, config, nil)

	auth, ok := registry.GetAuthenticator("nonexistent_method")
	require.False(t, ok, "Should not find nonexistent authenticator")
	require.Nil(t, auth, "Authenticator should be nil for nonexistent method")
}

// TestRegistry_GetHasher validates secret hasher retrieval.
func TestRegistry_GetHasher(t *testing.T) {
	t.Parallel()

	repoFactory := createRegistryTestRepoFactory(t)
	config := createRegistryTestConfig()

	registry := cryptoutilIdentityClientAuth.NewRegistry(repoFactory, config, nil)

	hasher := registry.GetHasher()
	require.NotNil(t, hasher, "Hasher should not be nil")
}

// createRegistryTestRepoFactory creates repository factory for registry testing.
func createRegistryTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createRegistryTestConfig()
	ctx := context.Background()

	factory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Repository factory creation should succeed")

	err = factory.AutoMigrate(ctx)
	require.NoError(t, err, "Auto migration should succeed")

	return factory
}

// createRegistryTestConfig creates config for registry testing.
func createRegistryTestConfig() *cryptoutilIdentityConfig.Config {
	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  "file::memory:?cache=private",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}
