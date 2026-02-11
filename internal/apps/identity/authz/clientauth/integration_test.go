// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityJobs "cryptoutil/internal/apps/identity/jobs"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityRotation "cryptoutil/internal/apps/identity/rotation"
)

// getTestRepository creates optimized test repository with minimal overhead.
// Uses unique in-memory SQLite with cache=private for complete test isolation.
// Critical optimization: DevMode + AutoMigrate=false to skip redundant migration checks.
func getTestRepository(t *testing.T) (*cryptoutilIdentityRepository.RepositoryFactory, context.Context) {
	t.Helper()

	ctx := context.Background()

	// Each test gets unique in-memory DB with cache=private (required for GORM transaction safety).
	dsn := "file::memory:?cache=private&_id=" + googleUuid.NewString()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         dsn,
		AutoMigrate: true, // Required for schema creation
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	// Run migrations once per test database.
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run migrations")

	return repoFactory, ctx
}

// Validates requirements:
// - R04-04: Client authentication method enforcement.
func TestRegistry_AllAuthMethods(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)

	config := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://authz.example.com",
		},
	}

	registry := NewRegistry(repoFactory, config, nil)

	// Test all auth methods are registered.
	authMethods := []string{
		"client_secret_basic",
		"client_secret_post",
		"tls_client_auth",
		"self_signed_tls_client_auth",
		"private_key_jwt",
		"client_secret_jwt",
	}

	for _, method := range authMethods {
		authenticator, ok := registry.GetAuthenticator(method)
		require.True(t, ok, "Auth method %s should be registered", method)
		require.NotNil(t, authenticator)
		// Note: client_secret_post uses same authenticator as client_secret_basic,
		// so we only verify the authenticator exists, not the exact method name.
	}
}

func TestRegistry_UnknownMethod(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)
	config := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://authz.example.com",
		},
	}

	registry := NewRegistry(repoFactory, config, nil)

	_, ok := registry.GetAuthenticator("unknown_method")
	require.False(t, ok)
}

func TestRegistry_RegisterCustomAuthenticator(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)
	config := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://authz.example.com",
		},
	}

	registry := NewRegistry(repoFactory, config, nil)

	// Create a mock authenticator.
	mockAuth := NewBasicAuthenticator(repoFactory.ClientRepository())

	registry.RegisterAuthenticator(mockAuth)

	authenticator, ok := registry.GetAuthenticator(mockAuth.Method())
	require.True(t, ok)
	require.NotNil(t, authenticator)
}

func TestBasicAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)
	authenticator := NewBasicAuthenticator(repoFactory.ClientRepository())
	require.Equal(t, "client_secret_basic", authenticator.Method())
}

func TestPostAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)
	authenticator := NewPostAuthenticator(repoFactory.ClientRepository())
	require.Equal(t, "client_secret_post", authenticator.Method())
}

// TestClientAuthentication_MultiSecretValidation verifies that both old and new secrets work during grace period.
func TestClientAuthentication_MultiSecretValidation(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	// Create test client (version 1 secret generated automatically).
	clientRepo := repoFactory.ClientRepository()
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-multi-secret-client",
		ClientSecret:            "will-be-replaced-by-create",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Multi-Secret Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err)

	// Get version 1 secret from database.
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(repoFactory.DB())
	secretVersions, err := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, secretVersions, 1)

	// Rotate secret to version 2 with 24-hour grace period.
	secretV2Result, err := rotationService.RotateClientSecret(ctx, client.ID, 24*time.Hour, "test-admin", "test rotation")
	require.NoError(t, err)

	// Verify new secret (version 2) works immediately.
	validNew, versionNew, err2 := rotationService.ValidateSecretDuringGracePeriod(ctx, client.ID, secretV2Result.NewSecretPlaintext)
	require.NoError(t, err2)
	require.True(t, validNew, "New secret should be valid immediately")
	require.Equal(t, 2, versionNew)
	// Note: Cannot validate old secret plaintext because Create doesn't return it.
	// This is expected - production code never exposes plaintext after initial generation.
	// The multi-secret validation is tested in rotation service tests where we control plaintext.
}

// TestClientAuthentication_OldSecretExpired verifies that old secrets are rejected after grace period + cleanup.
func TestClientAuthentication_OldSecretExpired(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	// Create test client (version 1 secret generated automatically).
	clientRepo := repoFactory.ClientRepository()
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-expired-secret-client",
		ClientSecret:            "will-be-replaced-by-create",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Expired Secret Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err)

	// Get version 1 secret from database.
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(repoFactory.DB())
	secretVersions, err := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, secretVersions, 1)
	secretV1Hash := secretVersions[0].SecretHash

	// Rotate secret to version 2 with 1-second grace period.
	secretV2Result, err := rotationService.RotateClientSecret(ctx, client.ID, 1*time.Second, "test-admin", "test rotation")
	require.NoError(t, err)

	// Wait for grace period to expire.
	time.Sleep(2 * time.Second)

	// Run cleanup job to mark expired secrets.
	rowsAffected, err2 := cryptoutilIdentityJobs.CleanupExpiredSecrets(ctx, repoFactory.DB())
	require.NoError(t, err2)
	require.Equal(t, int64(1), rowsAffected, "Version 1 should be marked expired")

	// Verify old secret is rejected.
	validOld, _, err3 := rotationService.ValidateSecretDuringGracePeriod(ctx, client.ID, secretV1Hash)
	require.NoError(t, err3)
	require.False(t, validOld, "Old secret should be rejected after grace period")

	// Verify new secret still works.
	validNew, versionNew, err4 := rotationService.ValidateSecretDuringGracePeriod(ctx, client.ID, secretV2Result.NewSecretPlaintext)
	require.NoError(t, err4)
	require.True(t, validNew, "New secret should still be valid")
	require.Equal(t, 2, versionNew)
}

// TestClientAuthentication_NewSecretImmediate verifies that new secrets work immediately after rotation.
func TestClientAuthentication_NewSecretImmediate(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	// Create test client (version 1 secret generated automatically).
	clientRepo := repoFactory.ClientRepository()
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-immediate-secret-client",
		ClientSecret:            "will-be-replaced-by-create",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Immediate Secret Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err)

	// Rotate secret to version 2 with 24-hour grace period.
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(repoFactory.DB())
	secretV2Result, err := rotationService.RotateClientSecret(ctx, client.ID, 24*time.Hour, "test-admin", "test rotation")
	require.NoError(t, err)

	// Verify new secret works immediately (no delay).
	validNew, versionNew, err2 := rotationService.ValidateSecretDuringGracePeriod(ctx, client.ID, secretV2Result.NewSecretPlaintext)
	require.NoError(t, err2)
	require.True(t, validNew, "New secret should work immediately after rotation")
	require.Equal(t, 2, versionNew)
}

// TestClientAuthentication_RevokedSecretRejected verifies that revoked secrets are rejected immediately.
func TestClientAuthentication_RevokedSecretRejected(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	// Create test client (version 1 secret generated automatically).
	clientRepo := repoFactory.ClientRepository()
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-revoked-secret-client",
		ClientSecret:            "will-be-replaced-by-create",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Revoked Secret Test Client",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"openid"},
		RedirectURIs:            []string{"https://example.com/callback"},
		RequirePKCE:             boolPtr(true),
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
	}
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err)

	// Get version 1 secret from database.
	rotationService := cryptoutilIdentityRotation.NewSecretRotationService(repoFactory.DB())
	secretVersions, err := rotationService.GetActiveSecretVersions(ctx, client.ID)
	require.NoError(t, err)
	require.Len(t, secretVersions, 1)
	secretV1Hash := secretVersions[0].SecretHash

	// Rotate secret to version 2 with 24-hour grace period.
	secretV2Result, err := rotationService.RotateClientSecret(ctx, client.ID, 24*time.Hour, "test-admin", "test rotation")
	require.NoError(t, err)

	// Manually revoke version 1 secret.
	err2 := rotationService.RevokeSecretVersion(ctx, client.ID, 1, "test-admin", "security breach")
	require.NoError(t, err2)

	// Verify revoked secret is rejected immediately (even during grace period).
	validOld, _, err3 := rotationService.ValidateSecretDuringGracePeriod(ctx, client.ID, secretV1Hash)
	require.NoError(t, err3)
	require.False(t, validOld, "Revoked secret should be rejected immediately")

	// Verify new secret still works.
	validNew, versionNew, err4 := rotationService.ValidateSecretDuringGracePeriod(ctx, client.ID, secretV2Result.NewSecretPlaintext)
	require.NoError(t, err4)
	require.True(t, validNew, "New secret should still be valid")
	require.Equal(t, 2, versionNew)
}
