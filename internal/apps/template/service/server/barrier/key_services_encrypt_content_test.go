// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

func TestContentKeysService_EncryptContent_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("intermediate_key_encryption_fails", func(t *testing.T) {
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

		// Delete all intermediate keys to cause encryption to fail.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Attempt to encrypt content - should fail because no intermediate key exists.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, _, encryptErr := contentKeysSvc.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get encrypted intermediate JWK")
	})

	t.Run("add_content_key_db_failure", func(t *testing.T) {
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

		// Create a content key first to establish UUID.
		var firstKeyID *googleUuid.UUID

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, keyID, encryptErr := contentKeysSvc.EncryptContent(tx, []byte("test data"))
			firstKeyID = keyID

			return encryptErr
		})
		require.NoError(t, err)

		// Try to manually insert a content key with the same UUID to cause UNIQUE constraint violation.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			return tx.AddContentKey(&cryptoutilAppsTemplateServiceServerBarrier.ContentKey{
				UUID:      *firstKeyID, // Duplicate UUID
				Encrypted: "fake_encrypted_jwk",
				KEKUUID:   googleUuid.New(),
			})
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "UNIQUE constraint failed") // SQLite error
	})
}

// TestContentKeysService_DecryptContent_ErrorPaths tests decryption error scenarios.
func TestContentKeysService_DecryptContent_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("invalid_jwe_format", func(t *testing.T) {
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

		// Attempt to decrypt with invalid JWE format.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, []byte("not-a-valid-jwe"))

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse JWE message")
	})

	t.Run("content_key_not_found", func(t *testing.T) {
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

		// First encrypt some content to get a valid JWE.
		var ciphertext []byte

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			var encryptErr error

			ciphertext, _, encryptErr = contentKeysSvc.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.NoError(t, err)

		// Delete all content keys.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_content_keys")
		require.NoError(t, err)

		// Attempt to decrypt - should fail because content key not found.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, ciphertext)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get encrypted content key")
	})

	t.Run("missing_kid_header", func(t *testing.T) {
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

		// Create a JWE without a kid header.
		// JWE compact format: header.encrypted_key.iv.ciphertext.tag
		// Header: {"alg":"A256KW","enc":"A256GCM"} - missing "kid" field.
		jweWithoutKid := []byte("eyJhbGciOiJBMjU2S1ciLCJlbmMiOiJBMjU2R0NNIn0.AAAA.AAAA.AAAA.AAAA")

		// Attempt to decrypt - should fail because kid header is missing.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, jweWithoutKid)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse JWE message kid")
	})

	t.Run("decrypt_content_key_failure", func(t *testing.T) {
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

		// Create first barrier with original unseal key.
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
		t.Cleanup(func() { intermediateKeysSvc1.Shutdown() })

		contentKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc1)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc1.Shutdown() })

		// Encrypt content with the first barrier.
		var ciphertext []byte

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			var encryptErr error

			ciphertext, _, encryptErr = contentKeysSvc1.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.NoError(t, err)

		// Delete all root and intermediate keys to simulate rotation/corruption.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Create second barrier with DIFFERENT unseal key.
		_, unsealJWK2, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK2})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc2.Shutdown() })

		rootKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc2.Shutdown() })

		intermediateKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { intermediateKeysSvc2.Shutdown() })

		contentKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc2.Shutdown() })

		// Attempt to decrypt with second barrier - should fail because intermediate key used to encrypt content key is missing.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc2.DecryptContent(tx, ciphertext)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt content key")
	})

	t.Run("decrypt_bytes_failure", func(t *testing.T) {
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

		// Encrypt content to get a valid JWE structure.
		var validCiphertext []byte

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			var encryptErr error

			validCiphertext, _, encryptErr = contentKeysSvc.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.NoError(t, err)

		// Convert to string, split into JWE parts.
		jweString := string(validCiphertext)
		parts := []byte(jweString)
		dotCount := 0
		thirdDotIdx := -1
		fourthDotIdx := -1

		for i := 0; i < len(parts); i++ {
			if parts[i] == '.' {
				dotCount++
				if dotCount == 3 {
					thirdDotIdx = i
				} else if dotCount == 4 {
					fourthDotIdx = i

					break
				}
			}
		}

		require.True(t, thirdDotIdx > 0 && fourthDotIdx > thirdDotIdx, "JWE compact serialization should have at least 4 dots")

		// Replace the ciphertext portion (between 3rd and 4th dot) with valid base64url but wrong length/content.
		// This will pass JWE parsing but fail during actual AES-GCM decryption.
		corruptedJWE := make([]byte, 0, len(parts))
		corruptedJWE = append(corruptedJWE, parts[:thirdDotIdx+1]...)
		corruptedJWE = append(corruptedJWE, []byte("AAAAAAAAAAAAAAAAAAAAAA")...) // Valid base64url, wrong ciphertext
		corruptedJWE = append(corruptedJWE, parts[fourthDotIdx:]...)

		// Attempt to decrypt - should fail at DecryptBytesWithContext.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, corruptedJWE)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt content with content key")
	})
}

// TestIntermediateKeysService_EncryptKey_ErrorPaths tests intermediate key encryption error scenarios.
