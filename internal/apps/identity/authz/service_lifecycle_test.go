// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// mockKeyGenerator implements KeyGenerator for testing.
type mockKeyGenerator struct{}

func (m *mockKeyGenerator) GenerateSigningKey(_ context.Context, algorithm string) (*cryptoutilIdentityIssuer.SigningKey, error) {
	return &cryptoutilIdentityIssuer.SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           []byte("mock-signing-key"),
		Algorithm:     algorithm,
		CreatedAt:     time.Now().UTC(),
		Active:        false,
		ValidForVerif: false,
	}, nil
}

func (m *mockKeyGenerator) GenerateEncryptionKey(_ context.Context) (*cryptoutilIdentityIssuer.EncryptionKey, error) {
	return &cryptoutilIdentityIssuer.EncryptionKey{
		KeyID:        googleUuid.NewString(),
		Key:          []byte("0123456789abcdef0123456789abcdef"),
		CreatedAt:    time.Now().UTC(),
		Active:       false,
		ValidForDecr: false,
	}, nil
}

func TestServiceStart(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test database.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := repoFactory.DB().DB() //nolint:errcheck // Test cleanup
		if sqlDB != nil {
			_ = sqlDB.Close() //nolint:errcheck // Test cleanup
		}
	})

	// Create token service.
	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:               "https://issuer.example.com",
		AccessTokenFormat:    cryptoutilSharedMagic.TokenFormatJWS,
		AccessTokenLifetime:  cryptoutilSharedMagic.DefaultAccessTokenLifetime,
		RefreshTokenLifetime: cryptoutilSharedMagic.DefaultRefreshTokenLifetime,
		IDTokenLifetime:      cryptoutilSharedMagic.DefaultIDTokenLifetime,
		SigningAlgorithm:     "RS256",
	}

	// Create key rotation manager.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	// Create JWS issuer.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		tokenConfig.AccessTokenLifetime,
		tokenConfig.IDTokenLifetime,
	)
	require.NoError(t, err)

	// Create JWE issuer.
	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	// Create UUID issuer.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	// Create authz service.
	cfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://issuer.example.com",
		},
	}
	authzSvc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, tokenSvc)

	// Test Start.
	err = authzSvc.Start(ctx)
	require.NoError(t, err, "Start should succeed with valid database connection")
}

func TestServiceStop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test database.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Create tokens table for DeleteExpired test.
	db := repoFactory.DB()
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tokens (
			id TEXT PRIMARY KEY,
			token_type TEXT NOT NULL,
			token_value TEXT NOT NULL,
			client_id TEXT NOT NULL,
			user_id TEXT,
			scope TEXT,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			revoked BOOLEAN DEFAULT false
		)
	`).Error
	require.NoError(t, err)

	// Create token service.
	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:               "https://issuer.example.com",
		AccessTokenFormat:    cryptoutilSharedMagic.TokenFormatJWS,
		AccessTokenLifetime:  cryptoutilSharedMagic.DefaultAccessTokenLifetime,
		RefreshTokenLifetime: cryptoutilSharedMagic.DefaultRefreshTokenLifetime,
		IDTokenLifetime:      cryptoutilSharedMagic.DefaultIDTokenLifetime,
		SigningAlgorithm:     "RS256",
	}

	// Create key rotation manager.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	// Create JWS issuer.
	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		tokenConfig.AccessTokenLifetime,
		tokenConfig.IDTokenLifetime,
	)
	require.NoError(t, err)

	// Create JWE issuer.
	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	// Create UUID issuer.
	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	// Create authz service.
	authzCfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://issuer.example.com",
		},
	}
	authzSvc := cryptoutilIdentityAuthz.NewService(authzCfg, repoFactory, tokenSvc)

	// Test Stop (should clean up expired tokens).
	// Note: Database connection is managed by RepositoryFactory, not by Service.
	err = authzSvc.Stop(ctx)
	require.NoError(t, err, "Stop should succeed with token cleanup")
}
