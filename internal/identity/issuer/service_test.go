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

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// mockKeyGenerator implements KeyGenerator for testing.
type mockKeyGenerator struct{}

func (m *mockKeyGenerator) GenerateSigningKey(ctx context.Context, algorithm string) (*cryptoutilIdentityIssuer.SigningKey, error) {
	return &cryptoutilIdentityIssuer.SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           []byte("mock-signing-key"),
		Algorithm:     algorithm,
		CreatedAt:     time.Now(),
		Active:        false,
		ValidForVerif: false,
	}, nil
}

func (m *mockKeyGenerator) GenerateEncryptionKey(ctx context.Context) (*cryptoutilIdentityIssuer.EncryptionKey, error) {
	return &cryptoutilIdentityIssuer.EncryptionKey{
		KeyID:        googleUuid.NewString(),
		Key:          []byte("0123456789abcdef0123456789abcdef"),
		CreatedAt:    time.Now(),
		Active:       false,
		ValidForDecr: false,
	}, nil
}

// TestNewTokenService validates service initialization.
func TestNewTokenService(t *testing.T) {
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
		&mockKeyGenerator{},
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
	require.NotNil(t, jwsIssuer)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)
	require.NotNil(t, jweIssuer)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
	require.NotNil(t, uuidIssuer)

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)
	require.NotNil(t, service)
}

// TestIssueAccessTokenJWS validates JWS access token generation.
func TestIssueAccessTokenJWS(t *testing.T) {
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
		&mockKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	// Initialize signing key for token generation.
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

	claims := map[string]interface{}{
		"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
		"email": "test@example.com",
		"name":  "Test User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token, err := service.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}
