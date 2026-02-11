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

// TestNewJWEIssuer validates JWE issuer initialization.
func TestNewJWEIssuer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	db := repoFactory.DB()
	err = db.AutoMigrate(&cryptoutilIdentityDomain.Key{})
	require.NoError(t, err)

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockJWEKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)
	require.NotNil(t, jweIssuer)
}

// TestNewJWEIssuer_NilKeyRotationManager validates error when key rotation manager is nil.
func TestNewJWEIssuer_NilKeyRotationManager(t *testing.T) {
	t.Parallel()

	_, err := cryptoutilIdentityIssuer.NewJWEIssuer(nil)
	require.Error(t, err)
}

// TestEncryptToken validates token encryption.
func TestEncryptToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	db := repoFactory.DB()
	err = db.AutoMigrate(&cryptoutilIdentityDomain.Key{})
	require.NoError(t, err)

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockJWEKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	err = keyRotationMgr.RotateEncryptionKey(ctx)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	const testPlaintextToken = "test-plaintext-token"

	encrypted, err := jweIssuer.EncryptToken(ctx, testPlaintextToken)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
	require.NotEqual(t, testPlaintextToken, encrypted)
}

// TestDecryptToken validates token decryption.
func TestDecryptToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	db := repoFactory.DB()
	err = db.AutoMigrate(&cryptoutilIdentityDomain.Key{})
	require.NoError(t, err)

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockJWEKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	err = keyRotationMgr.RotateEncryptionKey(ctx)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	const testPlaintextToken = "test-plaintext-token"

	encrypted, err := jweIssuer.EncryptToken(ctx, testPlaintextToken)
	require.NoError(t, err)

	decrypted, err := jweIssuer.DecryptToken(ctx, encrypted)
	require.NoError(t, err)
	require.Equal(t, testPlaintextToken, decrypted)
}

// TestDecryptToken_InvalidFormat validates error for invalid encrypted token format.
func TestDecryptToken_InvalidFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	db := repoFactory.DB()
	err = db.AutoMigrate(&cryptoutilIdentityDomain.Key{})
	require.NoError(t, err)

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockJWEKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	err = keyRotationMgr.RotateEncryptionKey(ctx)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	_, err = jweIssuer.DecryptToken(ctx, "invalid-base64-token")
	require.Error(t, err)
}

// mockJWEKeyGenerator implements KeyGenerator for JWE testing.
type mockJWEKeyGenerator struct{}

func (m *mockJWEKeyGenerator) GenerateSigningKey(_ context.Context, algorithm string) (*cryptoutilIdentityIssuer.SigningKey, error) {
	return &cryptoutilIdentityIssuer.SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           []byte("mock-signing-key"),
		Algorithm:     algorithm,
		CreatedAt:     time.Now().UTC(),
		Active:        false,
		ValidForVerif: false,
	}, nil
}

func (m *mockJWEKeyGenerator) GenerateEncryptionKey(_ context.Context) (*cryptoutilIdentityIssuer.EncryptionKey, error) {
	return &cryptoutilIdentityIssuer.EncryptionKey{
		KeyID:        googleUuid.NewString(),
		Key:          []byte("0123456789abcdef0123456789abcdef"),
		CreatedAt:    time.Now().UTC(),
		Active:       false,
		ValidForDecr: false,
	}, nil
}
