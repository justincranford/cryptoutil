// Copyright (c) 2025 Justin Cranford
//
//

package rootkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("root_keys_service_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testSQLRepository    *cryptoutilSQLRepository.SQLRepository
	testOrmRepository    *cryptoutilOrmRepository.OrmRepository
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testJWKGenService = cryptoutilSharedCryptoJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJWKGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestRootKeysService_HappyPath_OneUnsealJWKs(t *testing.T) {
	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, unsealJWK)

	unsealKeysServiceSimple, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceSimple)

	defer unsealKeysServiceSimple.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)

	defer rootKeysService.Shutdown()
}

func TestRootKeysService_SadPath_ZeroUnsealJWKs(t *testing.T) {
	unsealKeysServiceSimple := cryptoutilUnsealKeysService.RequireNewSimpleForTest([]joseJwk.Key{})

	require.NotNil(t, unsealKeysServiceSimple)
	defer unsealKeysServiceSimple.Shutdown()

	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK: invalid JWKs: jwks can't be empty")
}

func TestRootKeysService_SadPath_NilUnsealJWKs(t *testing.T) {
	unsealKeysServiceSimple := cryptoutilUnsealKeysService.RequireNewSimpleForTest(nil)

	require.NotNil(t, unsealKeysServiceSimple)
	defer unsealKeysServiceSimple.Shutdown()

	testSQLRepository = cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, testOrmRepository, unsealKeysServiceSimple)
	require.Error(t, err)
	require.Nil(t, rootKeysService)
	require.EqualError(t, err, "failed to initialize first root JWK: failed to encrypt first root JWK: failed to encrypt root JWK with unseal JWK: invalid JWKs: jwks can't be nil")
}

func TestNewRootKeysService_ValidationErrors(t *testing.T) {
	// NOTE: t.Parallel() removed due to shared in-memory SQLite database

	// Create isolated dependencies for this test.
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	defer unsealKeysService.Shutdown()

	tests := []struct {
		name           string
		telemetrySvc   *cryptoutilSharedTelemetry.TelemetryService
		jwkGenSvc      *cryptoutilSharedCryptoJose.JWKGenService
		ormRepo        *cryptoutilOrmRepository.OrmRepository
		unsealKeysSvc  cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrMsg string
	}{
		{
			name:           "nil telemetryService",
			telemetrySvc:   nil,
			jwkGenSvc:      testJWKGenService,
			ormRepo:        ormRepository,
			unsealKeysSvc:  unsealKeysService,
			expectedErrMsg: "telemetryService must be non-nil",
		},
		{
			name:           "nil jwkGenService",
			telemetrySvc:   testTelemetryService,
			jwkGenSvc:      nil,
			ormRepo:        ormRepository,
			unsealKeysSvc:  unsealKeysService,
			expectedErrMsg: "jwkGenService must be non-nil",
		},
		{
			name:           "nil ormRepository",
			telemetrySvc:   testTelemetryService,
			jwkGenSvc:      testJWKGenService,
			ormRepo:        nil,
			unsealKeysSvc:  unsealKeysService,
			expectedErrMsg: "ormRepository must be non-nil",
		},
		{
			name:           "nil unsealKeysService",
			telemetrySvc:   testTelemetryService,
			jwkGenSvc:      testJWKGenService,
			ormRepo:        ormRepository,
			unsealKeysSvc:  nil,
			expectedErrMsg: "unsealKeysService must be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE: t.Parallel() removed due to shared in-memory SQLite database

			rootKeysService, err := NewRootKeysService(tt.telemetrySvc, tt.jwkGenSvc, tt.ormRepo, tt.unsealKeysSvc)
			require.Error(t, err)
			require.Nil(t, rootKeysService)
			require.EqualError(t, err, tt.expectedErrMsg)
		})
	}
}

func TestRootKeysService_EncryptKey_HappyPath(t *testing.T) {
	// NOTE: t.Parallel() removed due to shared in-memory SQLite database

	// Create isolated dependencies for this test.
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	defer unsealKeysService.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)
	defer rootKeysService.Shutdown()

	// Generate an intermediate key to encrypt.
	_, intermediateKey, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	// Encrypt the intermediate key within a transaction.
	err = ormRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadWrite, func(tx *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedKeyBytes, rootKeyUUID, encryptErr := rootKeysService.EncryptKey(tx, intermediateKey)
		require.NoError(t, encryptErr)
		require.NotNil(t, encryptedKeyBytes)
		require.NotNil(t, rootKeyUUID)
		require.Greater(t, len(encryptedKeyBytes), 0)
		return nil
	})
	require.NoError(t, err)
}

func TestRootKeysService_EncryptAndDecryptKey_RoundTrip(t *testing.T) {
	// NOTE: t.Parallel() removed due to shared in-memory SQLite database

	// Create isolated dependencies for this test.
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	defer unsealKeysService.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)
	defer rootKeysService.Shutdown()

	// Generate an intermediate key to encrypt.
	originalKeyUUID, originalKey, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	var encryptedKeyBytes []byte

	// Encrypt the intermediate key.
	err = ormRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadWrite, func(tx *cryptoutilOrmRepository.OrmTransaction) error {
		var encryptErr error
		encryptedKeyBytes, _, encryptErr = rootKeysService.EncryptKey(tx, originalKey)
		return encryptErr
	})
	require.NoError(t, err)
	require.NotNil(t, encryptedKeyBytes)

	// Decrypt the intermediate key and verify it matches.
	err = ormRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadOnly, func(tx *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedKey, decryptErr := rootKeysService.DecryptKey(tx, encryptedKeyBytes)
		require.NoError(t, decryptErr)
		require.NotNil(t, decryptedKey)

		// Verify the decrypted key has the same kid.
		var originalKidString, decryptedKidString string
		require.NoError(t, originalKey.Get(joseJwk.KeyIDKey, &originalKidString))
		require.NoError(t, decryptedKey.Get(joseJwk.KeyIDKey, &decryptedKidString))
		require.Equal(t, originalKeyUUID.String(), originalKidString)
		require.Equal(t, originalKidString, decryptedKidString)

		return nil
	})
	require.NoError(t, err)
}

func TestRootKeysService_DecryptKey_InvalidFormat(t *testing.T) {
	// NOTE: t.Parallel() removed due to shared in-memory SQLite database

	// Create isolated dependencies for this test.
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	defer unsealKeysService.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)
	defer rootKeysService.Shutdown()

	// Try to decrypt invalid data.
	err = ormRepository.WithTransaction(testCtx, cryptoutilOrmRepository.ReadOnly, func(tx *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedKey, decryptErr := rootKeysService.DecryptKey(tx, []byte("not-a-valid-jwe"))
		require.Error(t, decryptErr)
		require.Nil(t, decryptedKey)
		require.Contains(t, decryptErr.Error(), "failed to parse encrypted intermediate key message")
		return nil
	})
	require.NoError(t, err)
}

func TestRootKeysService_Shutdown_Idempotent(t *testing.T) {
	// NOTE: t.Parallel() removed due to shared in-memory SQLite database

	// Create isolated dependencies for this test.
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	defer unsealKeysService.Shutdown()

	rootKeysService, err := NewRootKeysService(testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, rootKeysService)

	// Multiple shutdowns should be safe.
	rootKeysService.Shutdown()
	rootKeysService.Shutdown()
	rootKeysService.Shutdown()
}
