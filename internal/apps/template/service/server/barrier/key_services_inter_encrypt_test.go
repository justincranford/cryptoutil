// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
)
func TestIntermediateKeysService_EncryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("no_intermediate_key_exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
		require.NoError(t, err)
		t.Cleanup(func() { telemetrySvc.Shutdown() })

		jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
		require.NoError(t, err)
		t.Cleanup(func() { jwkGenSvc.Shutdown() })

		db, cleanup := createKeyServiceTestDB(t)
		defer cleanup()

		repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
		require.NoError(t, err)
		t.Cleanup(func() { repo.Shutdown() })

		_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc.Shutdown() })

		rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc.Shutdown() })

		intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

		// Delete all intermediate keys.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Generate a test JWK to encrypt.
		_, testJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)

		// Attempt to encrypt - should fail because no intermediate key exists.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, _, encryptErr := intermediateKeysSvc.EncryptKey(tx, testJWK)

			return encryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get encrypted intermediate JWK latest from DB")
	})

	t.Run("decrypt_intermediate_key_failure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
		require.NoError(t, err)
		t.Cleanup(func() { telemetrySvc.Shutdown() })

		jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
		require.NoError(t, err)
		t.Cleanup(func() { jwkGenSvc.Shutdown() })

		db, cleanup := createKeyServiceTestDB(t)
		defer cleanup()

		repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
		require.NoError(t, err)
		t.Cleanup(func() { repo.Shutdown() })

		// Create first unseal key and initialize keys.
		_, unsealJWK1, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc1, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK1})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc1.Shutdown() })

		rootKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc1)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc1.Shutdown() })

		intermediateKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc1)
		require.NoError(t, err)
		intermediateKeysSvc1.Shutdown() // Shutdown after initialization.

		// Create DIFFERENT unseal key and new services.
		_, unsealJWK2, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK2})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc2.Shutdown() })

		rootKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc2.Shutdown() })

		// Create intermediate service with WRONG root service.
		intermediateKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { intermediateKeysSvc2.Shutdown() })

		// Generate content JWK to encrypt.
		_, contentJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)

		// Try to encrypt - should fail because wrong unseal key means can't decrypt intermediate key.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, _, encErr := intermediateKeysSvc2.EncryptKey(tx, contentJWK)

			return encErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt intermediate JWK latest")
	})
}

// TestNewService_NilParameters tests that NewService properly rejects nil parameters.
func TestNewService_NilParameters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	tests := []struct {
		name                string
		ctx                 context.Context
		telemetrySvc        *cryptoutilSharedTelemetry.TelemetryService
		jwkGenSvc           *cryptoutilSharedCryptoJose.JWKGenService
		repo                cryptoutilAppsTemplateServiceServerBarrier.Repository
		unsealSvc           cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrContains string
	}{
		{
			name:                "nil_context",
			ctx:                 nil,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           jwkGenSvc,
			repo:                repo,
			unsealSvc:           unsealSvc,
			expectedErrContains: "ctx must be non-nil",
		},
		{
			name:                "nil_telemetry_service",
			ctx:                 ctx,
			telemetrySvc:        nil,
			jwkGenSvc:           jwkGenSvc,
			repo:                repo,
			unsealSvc:           unsealSvc,
			expectedErrContains: "telemetryService must be non-nil",
		},
		{
			name:                "nil_jwkgen_service",
			ctx:                 ctx,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           nil,
			repo:                repo,
			unsealSvc:           unsealSvc,
			expectedErrContains: "jwkGenService must be non-nil",
		},
		{
			name:                "nil_repository",
			ctx:                 ctx,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           jwkGenSvc,
			repo:                nil,
			unsealSvc:           unsealSvc,
			expectedErrContains: "repository must be non-nil",
		},
		{
			name:                "nil_unseal_service",
			ctx:                 ctx,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           jwkGenSvc,
			repo:                repo,
			unsealSvc:           nil,
			expectedErrContains: "unsealKeysService must be non-nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, err := cryptoutilAppsTemplateServiceServerBarrier.NewService(tc.ctx, tc.telemetrySvc, tc.jwkGenSvc, tc.repo, tc.unsealSvc)
			require.Error(t, err)
			require.Nil(t, svc)
			require.Contains(t, err.Error(), tc.expectedErrContains)
		})
	}
}

// TestDecryptContent_InvalidKidFormat tests DecryptContent with invalid kid format in JWE.
func TestDecryptContent_InvalidKidFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	// Create a malformed JWE with an invalid kid format (not a UUID).
	// JWE compact format: header.encrypted_key.iv.ciphertext.tag
	// We create a valid-looking JWE header with invalid kid, rest can be garbage.
	malformedJWE := []byte("eyJhbGciOiJBMjU2S1ciLCJlbmMiOiJBMjU2R0NNIiwia2lkIjoibm90LWEtdXVpZCJ9.AAAA.AAAA.AAAA.AAAA")

	// Try to decrypt - should fail because kid is not a valid UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := contentKeysSvc.DecryptContent(tx, malformedJWE)

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse kid as uuid")
}

// TestIntermediateKeysService_DecryptKey_RootKeyMissing tests intermediate key decryption when root key is missing.
func TestIntermediateKeysService_DecryptKey_RootKeyMissing(t *testing.T) {
	t.Parallel()

	t.Run("intermediate_key_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
		require.NoError(t, err)
		t.Cleanup(func() { telemetrySvc.Shutdown() })

		jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
		require.NoError(t, err)
		t.Cleanup(func() { jwkGenSvc.Shutdown() })

		db, cleanup := createKeyServiceTestDB(t)
		defer cleanup()

		repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
		require.NoError(t, err)
		t.Cleanup(func() { repo.Shutdown() })

		_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc.Shutdown() })

		rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc.Shutdown() })

		intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

		// Create a JWE with a non-existent intermediate key kid.
		_, testKey, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)

		// Encrypt some data with this key - the resulting JWE will have a kid that doesn't exist in DB.
		_, jweBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{testKey}, []byte("test data"))
		require.NoError(t, err)

		// Try to decrypt - should fail because intermediate key not found.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := intermediateKeysSvc.DecryptKey(tx, jweBytes)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get intermediate key")
	})
}

// TestRootKeysService_EncryptKey_NoRootKey tests root key encryption when no root key exists.
func TestRootKeysService_EncryptKey_NoRootKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	// Delete the root key that was auto-created.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Generate a test JWK to encrypt.
	_, testJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Attempt to encrypt - should fail because no root key exists.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encryptErr := rootKeysSvc.EncryptKey(tx, testJWK)

		return encryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encrypted root JWK latest from DB")
}

// TestRotationService_RotateRootKey_NoExistingKey tests root key rotation when no key exists.
func TestRotationService_RotateRootKey_NoExistingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	// Create rotation service directly (without root keys service initialization).
	rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	require.NotNil(t, rotationSvc)

	// Try to rotate - should fail because no root key exists.
	_, err = rotationSvc.RotateRootKey(ctx, "test rotation")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get current root key")
}

// TestRotationService_RotateIntermediateKey_NoExistingKey tests intermediate key rotation when no key exists.
