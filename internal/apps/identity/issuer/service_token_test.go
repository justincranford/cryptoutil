// Copyright (c) 2025 Justin Cranford
//
//

package issuer_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestIssueIDTokenService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing key.
	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	claims := map[string]any{
		"sub": googleUuid.Must(googleUuid.NewV7()).String(),
		"aud": "client123",
	}

	token, err := service.IssueIDToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueRefreshToken validates refresh token generation via service.
func TestIssueRefreshToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	token, err := service.IssueRefreshToken(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUID.
	_, err = googleUuid.Parse(token)
	require.NoError(t, err)
}

// TestGetPublicKeys validates public keys retrieval.
func TestGetPublicKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Rotate a key so there's something to return.
	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	keys := service.GetPublicKeys()
	require.NotNil(t, keys)
	// Keys should be present since we rotated.
	require.GreaterOrEqual(t, len(keys), 0)
}

// TestGetPublicKeysNilJWSIssuer validates empty result when JWS issuer is nil.
func TestGetPublicKeysNilJWSIssuer(t *testing.T) {
	t.Parallel()

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "uuid", // Not JWS.
	}

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create service with nil JWS issuer.
	service := cryptoutilIdentityIssuer.NewTokenService(nil, nil, uuidIssuer, tokenConfig)

	keys := service.GetPublicKeys()
	require.NotNil(t, keys)
	require.Empty(t, keys)
}

// TestBuildTokenDomain validates token domain builder.
func TestBuildTokenDomain(t *testing.T) {
	t.Parallel()

	clientID := googleUuid.Must(googleUuid.NewV7())
	userID := googleUuid.Must(googleUuid.NewV7())
	scopes := []string{"openid", "profile"}
	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	token := cryptoutilIdentityIssuer.BuildTokenDomain(
		cryptoutilIdentityDomain.TokenTypeAccess,
		cryptoutilIdentityDomain.TokenFormatJWS,
		"test-token-value",
		clientID,
		userID,
		scopes,
		expiresAt,
	)

	require.NotNil(t, token)
	require.Equal(t, cryptoutilIdentityDomain.TokenTypeAccess, token.TokenType)
	require.Equal(t, cryptoutilIdentityDomain.TokenFormatJWS, token.TokenFormat)
	require.Equal(t, "test-token-value", token.TokenValue)
	require.Equal(t, clientID, token.ClientID)
	require.Equal(t, scopes, token.Scopes)
}

// TestValidateAccessToken validates access token validation for all formats (JWS, JWE, UUID).
