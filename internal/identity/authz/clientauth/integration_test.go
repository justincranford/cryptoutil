package clientauth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// setupTestRepository creates an in-memory SQLite database for testing.
func setupTestRepository(t *testing.T) (*cryptoutilIdentityRepository.RepositoryFactory, context.Context) {
	t.Helper()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:            "sqlite",
		DSN:             ":memory:",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		AutoMigrate:     true,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	return repoFactory, ctx
}

func TestRegistry_AllAuthMethods(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)
	defer repoFactory.Close()

	registry := NewRegistry(repoFactory)

	// Test all auth methods are registered.
	authMethods := []string{
		"client_secret_basic",
		"client_secret_post",
		"tls_client_auth",
		"self_signed_tls_client_auth",
	}

	for _, method := range authMethods {
		authenticator, ok := registry.GetAuthenticator(method)
		require.True(t, ok, "Auth method %s should be registered", method)
		require.NotNil(t, authenticator)
		require.Equal(t, method, authenticator.Method())
	}
}

func TestRegistry_UnknownMethod(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)
	defer repoFactory.Close()

	registry := NewRegistry(repoFactory)

	_, ok := registry.GetAuthenticator("unknown_method")
	require.False(t, ok)
}

func TestRegistry_RegisterCustomAuthenticator(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)
	defer repoFactory.Close()

	registry := NewRegistry(repoFactory)

	// Create a mock authenticator.
	mockAuth := NewBasicAuthenticator(repoFactory.ClientRepository())

	registry.RegisterAuthenticator(mockAuth)

	authenticator, ok := registry.GetAuthenticator(mockAuth.Method())
	require.True(t, ok)
	require.NotNil(t, authenticator)
}

func TestBasicAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)
	defer repoFactory.Close()

	authenticator := NewBasicAuthenticator(repoFactory.ClientRepository())
	require.Equal(t, "client_secret_basic", authenticator.Method())
}

func TestPostAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)
	defer repoFactory.Close()

	authenticator := NewPostAuthenticator(repoFactory.ClientRepository())
	require.Equal(t, "client_secret_post", authenticator.Method())
}
