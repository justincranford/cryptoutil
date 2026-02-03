// Copyright (c) 2025 Justin Cranford

//go:build ignore
// +build ignore

// TODO(v7-phase5): This test file is temporarily disabled because it imports
// cryptoutil/internal/kms/server/repository/sqlrepository which no longer exists.
// This will be fixed during Phase 5 (KMS Barrier Migration) when shared/barrier
// is merged INTO the template barrier.

//nolint:wrapcheck,thelper // Test code doesn't need to wrap errors or use t.Helper()
package rootkeysservice

import (
	"testing"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// TestRootKeysService_EncryptDecryptKey tests key encryption and decryption round-trip.
func TestRootKeysService_EncryptDecryptKey(t *testing.T) {
	// Setup test dependencies
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// Create unseal key service
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, unsealJWK)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceSimple)

	defer unsealKeysServiceSimple.Shutdown()

	// Create root keys service
	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)

	defer rootKeysService.Shutdown()

	// Generate a test intermediate key to encrypt
	_, clearIntermediateKey, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, clearIntermediateKey)

	// Test encryption and decryption in a transaction
	err = testOrmRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		// Encrypt the intermediate key
		encryptedKeyBytes, rootKeyKidUUID, err := rootKeysService.EncryptKey(sqlTransaction, clearIntermediateKey)
		require.NoError(t, err)
		require.NotNil(t, encryptedKeyBytes)
		require.Greater(t, len(encryptedKeyBytes), 0)
		require.NotNil(t, rootKeyKidUUID)

		// Decrypt the intermediate key
		decryptedKey, err := rootKeysService.DecryptKey(sqlTransaction, encryptedKeyBytes)
		require.NoError(t, err)
		require.NotNil(t, decryptedKey)

		return nil
	})
	require.NoError(t, err)
}

// TestRootKeysService_MultipleEncryptDecryptRounds tests multiple encryption/decryption cycles.
func TestRootKeysService_MultipleEncryptDecryptRounds(t *testing.T) {
	// Setup test dependencies
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// Create unseal key service
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	defer unsealKeysServiceSimple.Shutdown()

	// Create root keys service
	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.NoError(t, err)

	defer rootKeysService.Shutdown()

	// Test multiple rounds of encryption/decryption
	for i := 0; i < 3; i++ {
		err = testOrmRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			// Generate new intermediate key for each round
			_, clearIntermediateKey, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
			require.NoError(t, err)

			// Encrypt
			encryptedKeyBytes, rootKeyKidUUID, err := rootKeysService.EncryptKey(sqlTransaction, clearIntermediateKey)
			require.NoError(t, err, "round %d encryption failed", i)
			require.NotNil(t, encryptedKeyBytes)
			require.NotNil(t, rootKeyKidUUID)

			// Decrypt
			decryptedKey, err := rootKeysService.DecryptKey(sqlTransaction, encryptedKeyBytes)
			require.NoError(t, err, "round %d decryption failed", i)
			require.NotNil(t, decryptedKey)

			return nil
		})
		require.NoError(t, err)
	}
}

// TestRootKeysService_Shutdown tests service shutdown.
func TestRootKeysService_Shutdown(t *testing.T) {
	// Setup test dependencies
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// Create unseal key service
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	defer unsealKeysServiceSimple.Shutdown()

	// Create root keys service
	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.NoError(t, err)

	// Shutdown should not panic
	require.NotPanics(t, func() {
		rootKeysService.Shutdown()
	})
}

// TestRootKeysService_DecryptKey_InvalidData tests decryption with invalid encrypted data.
func TestRootKeysService_DecryptKey_InvalidData(t *testing.T) {
	// Setup test dependencies
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// Create unseal key service
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	defer unsealKeysServiceSimple.Shutdown()

	// Create root keys service
	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.NoError(t, err)

	defer rootKeysService.Shutdown()

	// Try to decrypt invalid data
	err = testOrmRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		invalidData := []byte("not valid encrypted key data")
		_, err := rootKeysService.DecryptKey(sqlTransaction, invalidData)
		require.Error(t, err, "decryption of invalid data should fail")

		return nil
	})
	require.NoError(t, err)
}

// TestRootKeysService_DecryptKey_EmptyData tests decryption with empty data.
func TestRootKeysService_DecryptKey_EmptyData(t *testing.T) {
	// Setup test dependencies
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// Create unseal key service
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	defer unsealKeysServiceSimple.Shutdown()

	// Create root keys service
	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.NoError(t, err)

	defer rootKeysService.Shutdown()

	// Try to decrypt empty data
	err = testOrmRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		_, err := rootKeysService.DecryptKey(sqlTransaction, []byte{})
		require.Error(t, err, "decryption of empty data should fail")

		return nil
	})
	require.NoError(t, err)
}

// TestNewRootKeysService_NilTelemetryService tests constructor with nil telemetry service.
func TestNewRootKeysService_NilTelemetryService(t *testing.T) {
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	defer unsealKeysServiceSimple.Shutdown()

	rootKeysService, err := NewRootKeysService(nil, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.Contains(t, err.Error(), "telemetryService must be non-nil")
}

// TestNewRootKeysService_NilJWKGenService tests constructor with nil JWK generation service.
func TestNewRootKeysService_NilJWKGenService(t *testing.T) {
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	defer unsealKeysServiceSimple.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, nil, testOrmRepository, unsealKeysServiceSimple)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.Contains(t, err.Error(), "jwkGenService must be non-nil")
}

// TestNewRootKeysService_NilOrmRepository tests constructor with nil ORM repository.
func TestNewRootKeysService_NilOrmRepository(t *testing.T) {
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	defer unsealKeysServiceSimple.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, nil, unsealKeysServiceSimple)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.Contains(t, err.Error(), "ormRepository must be non-nil")
}

// TestNewRootKeysService_NilUnsealKeysService tests constructor with nil unseal keys service.
func TestNewRootKeysService_NilUnsealKeysService(t *testing.T) {
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, nil)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.Contains(t, err.Error(), "unsealKeysService must be non-nil")
}
