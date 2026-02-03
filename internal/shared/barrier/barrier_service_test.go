// Copyright (c) 2025 Justin Cranford
//
//

//go:build ignore
// +build ignore

// TODO(v7-phase5): This test file is temporarily disabled because it imports
// cryptoutil/internal/kms/server/repository/sqlrepository which no longer exists.
// This will be fixed during Phase 5 (KMS Barrier Migration) when shared/barrier
// is merged INTO the template barrier.

package barrierservice

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

	_ "github.com/jackc/pgx/v5/stdlib"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("barrier_service_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
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

func TestNewService_ValidationErrors(t *testing.T) {
	// Create isolated resources for this test
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, nonPublicJWEJWK, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKForEncAndAlg(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK})
	require.NoError(t, err)

	defer unsealKeysService.Shutdown()

	tests := []struct {
		name              string
		ctx               context.Context
		telemetryService  *cryptoutilSharedTelemetry.TelemetryService
		jwkGenService     *cryptoutilSharedCryptoJose.JWKGenService
		ormRepository     *cryptoutilOrmRepository.OrmRepository
		unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
		expectedError     string
	}{
		{
			name:              "nil ctx",
			ctx:               nil,
			telemetryService:  testTelemetryService,
			jwkGenService:     testJWKGenService,
			ormRepository:     ormRepository,
			unsealKeysService: unsealKeysService,
			expectedError:     "ctx must be non-nil",
		},
		{
			name:              "nil telemetryService",
			ctx:               testCtx,
			telemetryService:  nil,
			jwkGenService:     testJWKGenService,
			ormRepository:     ormRepository,
			unsealKeysService: unsealKeysService,
			expectedError:     "telemetryService must be non-nil",
		},
		{
			name:              "nil jwkGenService",
			ctx:               testCtx,
			telemetryService:  testTelemetryService,
			jwkGenService:     nil,
			ormRepository:     ormRepository,
			unsealKeysService: unsealKeysService,
			expectedError:     "jwkGenService must be non-nil",
		},
		{
			name:              "nil ormRepository",
			ctx:               testCtx,
			telemetryService:  testTelemetryService,
			jwkGenService:     testJWKGenService,
			ormRepository:     nil,
			unsealKeysService: unsealKeysService,
			expectedError:     "ormRepository must be non-nil",
		},
		{
			name:              "nil unsealKeysService",
			ctx:               testCtx,
			telemetryService:  testTelemetryService,
			jwkGenService:     testJWKGenService,
			ormRepository:     ormRepository,
			unsealKeysService: nil,
			expectedError:     "unsealKeysService must be non-nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewService(tc.ctx, tc.telemetryService, tc.jwkGenService, tc.ormRepository, tc.unsealKeysService)
			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestNewService_Success(t *testing.T) {
	// Create isolated resources for this test
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, nonPublicJWEJWK, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKForEncAndAlg(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK})
	require.NoError(t, err)

	defer unsealKeysService.Shutdown()

	service, err := NewService(testCtx, testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, service)

	defer service.Shutdown()

	// Verify service is initialized correctly
	require.NotNil(t, service.telemetryService)
	require.NotNil(t, service.jwkGenService)
	require.NotNil(t, service.ormRepository)
	require.NotNil(t, service.unsealKeysService)
	require.NotNil(t, service.rootKeysService)
	require.NotNil(t, service.intermediateKeysService)
	require.NotNil(t, service.contentKeysService)
	require.False(t, service.closed)
}

func TestBarrierService_EncryptContent_ServiceClosed(t *testing.T) {
	// Create isolated resources for this test
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, nonPublicJWEJWK, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKForEncAndAlg(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK})
	require.NoError(t, err)

	defer unsealKeysService.Shutdown()

	service, err := NewService(testCtx, testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown the service first
	service.Shutdown()
	require.True(t, service.closed)

	// Try to encrypt after shutdown
	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		_, err := service.EncryptContent(sqlTransaction, []byte("test content"))

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "barrier service is closed")
}

func TestBarrierService_DecryptContent_ServiceClosed(t *testing.T) {
	// Create isolated resources for this test
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, nonPublicJWEJWK, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKForEncAndAlg(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK})
	require.NoError(t, err)

	defer unsealKeysService.Shutdown()

	service, err := NewService(testCtx, testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown the service first
	service.Shutdown()
	require.True(t, service.closed)

	// Try to decrypt after shutdown
	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		_, err := service.DecryptContent(sqlTransaction, []byte("encrypted content"))

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "barrier service is closed")
}

func TestBarrierService_Shutdown_MultipleTimesIdempotent(t *testing.T) {
	// Create isolated resources for this test
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, nonPublicJWEJWK, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKForEncAndAlg(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK})
	require.NoError(t, err)

	defer unsealKeysService.Shutdown()

	service, err := NewService(testCtx, testTelemetryService, testJWKGenService, ormRepository, unsealKeysService)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Shutdown multiple times - should be idempotent
	service.Shutdown()
	require.True(t, service.closed)
	require.Nil(t, service.contentKeysService)
	require.Nil(t, service.intermediateKeysService)
	require.Nil(t, service.rootKeysService)

	// Second shutdown should be safe (no panic)
	service.Shutdown()
	require.True(t, service.closed)

	// Third shutdown should also be safe
	service.Shutdown()
	require.True(t, service.closed)
}

func Test_HappyPath_SameUnsealJWKs(t *testing.T) {
	// initialize repositories, will be reused by original and restarted unseal service
	sqlRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJWKGenService, testSettings)
	defer ormRepository.Shutdown()

	_, nonPublicJWEJWK, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKForEncAndAlg(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWEJWK)

	originalUnsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK})
	require.NoError(t, err)

	defer originalUnsealKeysService.Shutdown()

	encryptDecryptContentRestartDecryptAgain(t, ormRepository, originalUnsealKeysService, originalUnsealKeysService)
}

func Test_HappyPath_EncryptDecryptContent_Restart_DecryptAgain(t *testing.T) {
	// initialize repositories, will be reused by original and restarted unseal service
	testSQLRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// generate three JWKs; 2 valid, 1 invalid
	_, nonPublicJWEJWK1, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWEJWK1)

	_, nonPublicJWEJWK2, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWEJWK2)

	_, nonPublicJWEJWKInvalid, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWEJWKInvalid)

	// unseal with both valid JWKs
	unsealKeysServiceJWKs12, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK1, nonPublicJWEJWK2})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceJWKs12)

	defer unsealKeysServiceJWKs12.Shutdown()

	// unseal with first valid JWK
	unsealKeysServiceJWK1, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK1})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceJWK1)

	defer unsealKeysServiceJWK1.Shutdown()

	// unseal with second valid JWK
	unsealKeysServiceJWK2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWK2})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceJWK2)

	defer unsealKeysServiceJWK2.Shutdown()

	// unseal with invalid JWK
	unsealKeysServiceInvalidJWK, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJWEJWKInvalid})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceInvalidJWK)

	// same repository will be used for all tests below
	// barrier encrypt uses the unseal service with both valid JWKs for all tests below

	// barrier decrypt using unseal service with both valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJWKs12, unsealKeysServiceJWKs12)

	// barrier decrypt using unseal service with first valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJWKs12, unsealKeysServiceJWK1)

	// barrier decrypt using unseal service with second valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJWKs12, unsealKeysServiceJWK2)

	// barrier decrypt using unseal service with both valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJWKs12, unsealKeysServiceJWKs12)
}

func encryptDecryptContentRestartDecryptAgain(t *testing.T, testOrmRepository *cryptoutilOrmRepository.OrmRepository, originalUnsealKeysService, restartedUnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) {
	t.Helper()

	const numEncryptsDecrypts = 11

	plaintext := []byte("hello, world!")

	// start barrier service
	barrierService1, err := NewService(testCtx, testTelemetryService, testJWKGenService, testOrmRepository, originalUnsealKeysService)
	require.NoError(t, err)

	defer barrierService1.Shutdown()

	// encrypt N times
	encryptedBytesSlice := make([][]byte, 0, numEncryptsDecrypts)
	for i := range cap(encryptedBytesSlice) {
		t.Logf("Attempt: %d", i+1)

		var encryptedBytes []byte

		err := testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			var err error

			encryptedBytes, err = barrierService1.EncryptContent(sqlTransaction, plaintext)

			return err
		})
		require.NoError(t, err)
		t.Logf("Encrypted Data > JWE Headers: %s", string(encryptedBytes))
		encryptedBytesSlice = append(encryptedBytesSlice, encryptedBytes)
	}

	// decrypt N times with original service
	for _, encryptedBytes := range encryptedBytesSlice {
		var decryptedBytes []byte

		err := testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			var err error

			decryptedBytes, err = barrierService1.DecryptContent(sqlTransaction, encryptedBytes)

			return err
		})
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}

	// shutdown original service
	barrierService1.Shutdown()

	// barrier encrypt with shut down service should fail
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		_, err = barrierService1.EncryptContent(sqlTransaction, plaintext)

		return err
	})
	require.Error(t, err)

	// barrier decrypt with shut down service should fail
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		_, err = barrierService1.DecryptContent(sqlTransaction, plaintext)

		return err
	})
	require.Error(t, err)

	// restart new service with same unseal key repository
	barrierService2, err := NewService(testCtx, testTelemetryService, testJWKGenService, testOrmRepository, restartedUnsealKeysService)
	require.NoError(t, err)

	defer barrierService2.Shutdown()

	// decrypt N times with restarted service
	for _, encryptedBytes := range encryptedBytesSlice {
		var decryptedBytes []byte

		err := testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			var err error

			decryptedBytes, err = barrierService2.DecryptContent(sqlTransaction, encryptedBytes)

			return err
		})
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}

	// shutdown restarted service
	barrierService2.Shutdown()

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		_, err = barrierService2.EncryptContent(sqlTransaction, plaintext)

		return err
	})
	require.Error(t, err)

	t.Log("Success")
}

func Test_ErrorCase_DecryptWithInvalidJWKs(t *testing.T) {
	// initialize repositories
	testSQLRepository := cryptoutilSQLRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJWKGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// generate valid JWKs for encryption
	_, validJWEJWK1, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, validJWEJWK1)

	_, validJWEJWK2, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, validJWEJWK2)

	// generate invalid JWK for decryption attempt
	_, invalidJWEJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, invalidJWEJWK)

	// create unseal service with valid JWKs for encryption
	validUnsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{validJWEJWK1, validJWEJWK2})
	require.NoError(t, err)
	require.NotNil(t, validUnsealKeysService)

	defer validUnsealKeysService.Shutdown()

	// create unseal service with invalid JWK for decryption attempt
	invalidUnsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{invalidJWEJWK})
	require.NoError(t, err)
	require.NotNil(t, invalidUnsealKeysService)

	defer invalidUnsealKeysService.Shutdown()

	plaintext := []byte("hello, world!")

	// encrypt content with valid JWKs
	var encryptedBytes []byte

	barrierService, err := NewService(testCtx, testTelemetryService, testJWKGenService, testOrmRepository, validUnsealKeysService)
	require.NoError(t, err)

	defer barrierService.Shutdown()

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		encryptedBytes, err = barrierService.EncryptContent(sqlTransaction, plaintext)

		return err
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptedBytes)

	// shutdown the valid service
	barrierService.Shutdown()

	// try to decrypt with invalid JWKs - this should fail
	barrierServiceInvalid, err := NewService(testCtx, testTelemetryService, testJWKGenService, testOrmRepository, invalidUnsealKeysService)
	require.NoError(t, err)

	defer barrierServiceInvalid.Shutdown()

	var decryptedBytes []byte

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		decryptedBytes, err = barrierServiceInvalid.DecryptContent(sqlTransaction, encryptedBytes)

		return err
	})
	require.Error(t, err, "Decryption should fail with invalid JWKs")
	require.Empty(t, decryptedBytes, "Decrypted content should be empty when decryption fails")

	t.Log("Success - decryption correctly failed with invalid JWKs")
}
