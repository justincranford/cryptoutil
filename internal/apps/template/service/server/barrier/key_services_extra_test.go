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
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

func TestEncryptContent_InvalidInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
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

	// Test with empty content (should fail - empty is invalid).
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := contentKeysSvc.EncryptContent(tx, []byte{})
		require.Error(t, encErr)
		require.Contains(t, encErr.Error(), "clearBytes")

		return nil
	})
	require.NoError(t, err)

	// Test with nil content (should also fail - nil is invalid).
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := contentKeysSvc.EncryptContent(tx, nil)
		require.Error(t, encErr)
		require.Contains(t, encErr.Error(), "clearBytes")

		return nil
	})
	require.NoError(t, err)

	// Test with large content.
	largeContent := make([]byte, cryptoutilSharedMagic.DefaultLogsBatchSize*cryptoutilSharedMagic.DefaultLogsBatchSize) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	}

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		encryptedLarge, _, encErr := contentKeysSvc.EncryptContent(tx, largeContent)
		if encErr != nil {
			return encErr
		}

		decrypted, decErr := contentKeysSvc.DecryptContent(tx, encryptedLarge)
		if decErr != nil {
			return decErr
		}

		require.Equal(t, largeContent, decrypted)

		return nil
	})
	require.NoError(t, err)
}

// TestIntermediateKeysService_EncryptKey_NoIntermediateKey tests EncryptKey when no intermediate key exists.
func TestIntermediateKeysService_EncryptKey_NoIntermediateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
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

	// Create services to initialize keys.
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

	// Try to encrypt key - should fail because no intermediate key exists.
	_, testContentJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := intermediateKeysSvc.EncryptKey(tx, testContentJWK)

		return encErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encrypted intermediate JWK")
}

// TestRootKeysService_EncryptKey_NoRootKey_DeletedKey tests EncryptKey when no root key exists.
func TestRootKeysService_EncryptKey_NoRootKey_DeletedKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
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

	// Create root keys service to initialize keys.
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	// Delete all root keys.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Try to encrypt key - should fail because no root key exists.
	_, testIntermediateJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := rootKeysSvc.EncryptKey(tx, testIntermediateJWK)

		return encErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encrypted root JWK")
}

// TestIntermediateKeysService_DecryptKey_NoRootKey tests DecryptKey when root key is missing.
func TestIntermediateKeysService_DecryptKey_NoRootKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
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

	// Create services to initialize keys.
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	// Encrypt a key first to get valid encrypted data.
	var encryptedKeyBytes []byte

	_, testContentJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		var encErr error

		encryptedKeyBytes, _, encErr = intermediateKeysSvc.EncryptKey(tx, testContentJWK)

		return encErr
	})
	require.NoError(t, err)

	// Delete all root keys.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Try to decrypt - should fail because root key is missing.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decErr := intermediateKeysSvc.DecryptKey(tx, encryptedKeyBytes)

		return decErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get root key")
}

// TestRepositoryAddKey_NilInput tests that Add* methods reject nil key inputs.
func TestRepositoryAddKey_NilInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, cleanup := createKeyServiceTestDB(t)
	t.Cleanup(cleanup)

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	// Test AddRootKey with nil key.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddRootKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test AddIntermediateKey with nil key.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddIntermediateKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test AddContentKey with nil key.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddContentKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")
}

// TestRepositoryGetKey_NilUUID tests that Get*Key methods reject nil UUID inputs.
func TestRepositoryGetKey_NilUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, cleanup := createKeyServiceTestDB(t)
	t.Cleanup(cleanup)

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	// Test GetRootKey with nil UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, getErr := tx.GetRootKey(nil)

		return getErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test GetIntermediateKey with nil UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, getErr := tx.GetIntermediateKey(nil)

		return getErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test GetContentKey with nil UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, getErr := tx.GetContentKey(nil)

		return getErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")
}

// TestRootKeysService_EncryptKey_ErrorPaths tests EncryptKey error scenarios.
func TestRootKeysService_EncryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(*testing.T) (*cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, cryptoutilAppsTemplateServiceServerBarrier.Repository, *cryptoutilSharedCryptoJose.JWKGenService)
		wantErr string
	}{
		{
			name: "decrypt_root_key_failure",
			setup: func(t *testing.T) (*cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, cryptoutilAppsTemplateServiceServerBarrier.Repository, *cryptoutilSharedCryptoJose.JWKGenService) {
				// Create two different unseal keys.
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
				require.NoError(t, err)
				t.Cleanup(func() { telemetrySvc.Shutdown() })

				jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
				require.NoError(t, err)
				t.Cleanup(func() { jwkGenSvc.Shutdown() })

				db, cleanup := createKeyServiceTestDB(t)
				t.Cleanup(cleanup)

				repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
				require.NoError(t, err)
				t.Cleanup(func() { repo.Shutdown() })

				// Create first unseal key and initialize root key.
				_, unsealJWK1, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
				require.NoError(t, err)
				unsealSvc1, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK1})
				require.NoError(t, err)
				// Don't cleanup unsealSvc1 yet.

				rootKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc1)
				require.NoError(t, err)
				// Don't cleanup rootKeysSvc1 yet - still need DB populated.

				// Create DIFFERENT unseal key and new service.
				_, unsealJWK2, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
				require.NoError(t, err)
				unsealSvc2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK2})
				require.NoError(t, err)
				t.Cleanup(func() {
					rootKeysSvc1.Shutdown()
					unsealSvc1.Shutdown()
					unsealSvc2.Shutdown()
				})

				// Create service with WRONG unseal key.
				rootKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc2)
				require.NoError(t, err)
				t.Cleanup(func() { rootKeysSvc2.Shutdown() })

				return rootKeysSvc2, repo, jwkGenSvc
			},
			wantErr: "failed to decrypt root JWK latest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootKeysSvc, repo, jwkGenSvc := tc.setup(t)

			var testJWK joseJwk.Key

			if tc.name == "encrypt_bytes_failure" {
				// Create signing key (Ed25519) which cannot be used for encryption.
				_, _, signKey, _, _, err := jwkGenSvc.GenerateJWSJWK(cryptoutilSharedCryptoJose.AlgEdDSA)
				require.NoError(t, err)

				testJWK = signKey
			} else {
				// Generate intermediate JWK to encrypt.
				_, intermediateJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
				require.NoError(t, err)

				testJWK = intermediateJWK
			}

			// Try to encrypt - should fail.
			err := repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
				_, _, encErr := rootKeysSvc.EncryptKey(tx, testJWK)

				return encErr
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
