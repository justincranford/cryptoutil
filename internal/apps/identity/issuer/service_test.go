// Copyright (c) 2025 Justin Cranford
//
//

package issuer_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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

// mockKeyGenerator implements KeyGenerator for testing using real RSA keys.
type mockKeyGenerator struct {
	productionGen *cryptoutilIdentityIssuer.ProductionKeyGenerator
}

func newMockKeyGenerator() *mockKeyGenerator {
	return &mockKeyGenerator{
		productionGen: cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
	}
}

func (m *mockKeyGenerator) GenerateSigningKey(ctx context.Context, algorithm string) (*cryptoutilIdentityIssuer.SigningKey, error) {
	return m.productionGen.GenerateSigningKey(ctx, algorithm) //nolint:wrapcheck // Test wrapper
}

func (m *mockKeyGenerator) GenerateEncryptionKey(ctx context.Context) (*cryptoutilIdentityIssuer.EncryptionKey, error) {
	return m.productionGen.GenerateEncryptionKey(ctx) //nolint:wrapcheck // Test wrapper
}

// setupTestService creates a token service with initialized keys for testing.
func setupTestService(t *testing.T, tokenFormat string) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	db := repoFactory.DB()
	err = db.AutoMigrate(&cryptoutilIdentityDomain.Key{})
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		AccessTokenFormat: tokenFormat,
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing key for token generation.
	err = keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	require.NoError(t, err)

	// Initialize encryption key if using JWE format.
	if tokenFormat == cryptoutilSharedMagic.IdentityTokenFormatJWE {
		err = keyRotationMgr.RotateEncryptionKey(ctx)
		require.NoError(t, err)
	}

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

	return cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)
}

// TestNewTokenService validates service initialization.
func TestNewTokenService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
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
		SigningAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		AccessTokenFormat: cryptoutilSharedMagic.DefaultBrowserSessionCookie,
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
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
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
		SigningAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		AccessTokenFormat: cryptoutilSharedMagic.DefaultBrowserSessionCookie,
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing key for token generation.
	err = keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
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
		cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
		cryptoutilSharedMagic.ClaimEmail: "test@example.com",
		cryptoutilSharedMagic.ClaimName:  "Test User",
		cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Unix(),
		cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(1 * time.Hour).Unix(),
	}

	token, err := service.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueAccessTokenInvalidFormat validates error handling for unsupported token format.
func TestIssueAccessTokenInvalidFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		AccessTokenFormat: "invalid-format",
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

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
		cryptoutilSharedMagic.ClaimEmail: "test@example.com",
	}

	_, err = service.IssueAccessToken(ctx, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported access token format")
}

// TestIssueAccessTokenJWE validates JWE access token generation.
func TestIssueAccessTokenJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
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
		SigningAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		AccessTokenFormat: cryptoutilSharedMagic.IdentityTokenFormatJWE,
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing and encryption keys.
	err = keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	require.NoError(t, err)

	err = keyRotationMgr.RotateEncryptionKey(ctx)
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
		cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
		cryptoutilSharedMagic.ClaimEmail: "test@example.com",
		cryptoutilSharedMagic.ClaimName:  "Test User",
		cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Unix(),
		cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(1 * time.Hour).Unix(),
	}

	token, err := service.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueAccessTokenUUID validates UUID access token generation.
func TestIssueAccessTokenUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
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
		SigningAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		AccessTokenFormat: cryptoutilSharedMagic.IdentityTokenFormatUUID,
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

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
		cryptoutilSharedMagic.ClaimEmail: "test@example.com",
	}

	token, err := service.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUID.
	_, err = googleUuid.Parse(token)
	require.NoError(t, err)
}

// TestIssueIDTokenService validates ID token generation via service.
