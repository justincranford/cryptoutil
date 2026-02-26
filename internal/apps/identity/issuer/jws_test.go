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

// TestNewJWSIssuer validates JWS issuer initialization.
func TestNewJWSIssuer(t *testing.T) {
	t.Parallel()

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

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockJWSKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)
	require.NotNil(t, jwsIssuer)
}

// TestNewJWSIssuer_MissingIssuer validates error when issuer is empty.
func TestNewJWSIssuer_MissingIssuer(t *testing.T) {
	t.Parallel()

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

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockJWSKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	_, err = cryptoutilIdentityIssuer.NewJWSIssuer(
		"",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.Error(t, err)
}

// TestNewJWSIssuer_MissingAlgorithm validates error when algorithm is empty.
func TestNewJWSIssuer_MissingAlgorithm(t *testing.T) {
	t.Parallel()

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

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockJWSKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	_, err = cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		"",
		1*time.Hour,
		1*time.Hour,
	)
	require.Error(t, err)
}

// TestNewJWSIssuer_NilKeyRotationManager validates error when key rotation manager is nil.
func TestNewJWSIssuer_NilKeyRotationManager(t *testing.T) {
	t.Parallel()

	_, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://localhost:8080",
		nil,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.Error(t, err)
}

// TestJWSIssueAccessToken validates JWS access token generation.
func TestJWSIssueAccessToken(t *testing.T) {
	t.Parallel()

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

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockJWSKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	err = keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
		cryptoutilSharedMagic.ClaimEmail: "test@example.com",
		cryptoutilSharedMagic.ClaimName:  "Test User",
	}

	token, err := jwsIssuer.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueIDToken validates ID token generation.
func TestIssueIDToken(t *testing.T) {
	t.Parallel()

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

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockJWSKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	err = keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub: googleUuid.Must(googleUuid.NewV7()).String(),
		cryptoutilSharedMagic.ClaimAud: "client123",
	}

	token, err := jwsIssuer.IssueIDToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueIDToken_MissingSub validates error when sub claim is missing.
func TestIssueIDToken_MissingSub(t *testing.T) {
	t.Parallel()

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

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockJWSKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	err = keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimAud: "client123",
	}

	_, err = jwsIssuer.IssueIDToken(ctx, claims)
	require.Error(t, err)
}

// TestIssueIDToken_MissingAud validates error when aud claim is missing.
func TestIssueIDToken_MissingAud(t *testing.T) {
	t.Parallel()

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

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockJWSKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	err = keyRotationMgr.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://localhost:8080",
		keyRotationMgr,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	claims := map[string]any{
		cryptoutilSharedMagic.ClaimSub: googleUuid.Must(googleUuid.NewV7()).String(),
	}

	_, err = jwsIssuer.IssueIDToken(ctx, claims)
	require.Error(t, err)
}

// mockJWSKeyGenerator implements KeyGenerator for JWS testing.
// Uses ProductionKeyGenerator to generate real RSA keys for signature testing.
type mockJWSKeyGenerator struct {
	productionGen *cryptoutilIdentityIssuer.ProductionKeyGenerator
}

func newMockJWSKeyGenerator() *mockJWSKeyGenerator {
	return &mockJWSKeyGenerator{
		productionGen: cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
	}
}

func (m *mockJWSKeyGenerator) GenerateSigningKey(ctx context.Context, algorithm string) (*cryptoutilIdentityIssuer.SigningKey, error) {
	return m.productionGen.GenerateSigningKey(ctx, algorithm) //nolint:wrapcheck // Test wrapper
}

func (m *mockJWSKeyGenerator) GenerateEncryptionKey(ctx context.Context) (*cryptoutilIdentityIssuer.EncryptionKey, error) {
	return m.productionGen.GenerateEncryptionKey(ctx) //nolint:wrapcheck // Test wrapper
}
