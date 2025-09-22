package barrierservice

import (
	"context"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("barrier_service_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJwkGenService    *cryptoutilJose.JwkGenService
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testJwkGenService = cryptoutilJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJwkGenService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func Test_HappyPath_SameUnsealJwks(t *testing.T) {
	// initialize repositories, will be reused by original and restarted unseal service
	sqlRepository := cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer sqlRepository.Shutdown()

	ormRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, sqlRepository, testJwkGenService, testSettings)
	defer ormRepository.Shutdown()

	_, nonPublicJweJwk, _, _, _, err := cryptoutilJose.GenerateJweJwkForEncAndAlg(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJweJwk)

	originalUnsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJweJwk})
	require.NoError(t, err)
	defer originalUnsealKeysService.Shutdown()

	encryptDecryptContentRestartDecryptAgain(t, ormRepository, originalUnsealKeysService, originalUnsealKeysService)
}

func Test_HappyPath_EncryptDecryptContent_Restart_DecryptAgain(t *testing.T) {
	// initialize repositories, will be reused by original and restarted unseal service
	testSQLRepository := cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSQLRepository.Shutdown()

	testOrmRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJwkGenService, testSettings)
	defer testOrmRepository.Shutdown()

	// generate three JWKs; 2 valid, 1 invalid
	_, nonPublicJweJwk1, _, _, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJweJwk1)

	_, nonPublicJweJwk2, _, _, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJweJwk2)

	_, nonPublicJweJwkInvalid, _, _, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJweJwkInvalid)

	// unseal with both valid JWKs
	unsealKeysServiceJwks12, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJweJwk1, nonPublicJweJwk2})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceJwks12)
	defer unsealKeysServiceJwks12.Shutdown()

	// unseal with first valid JWK
	unsealKeysServiceJwk1, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJweJwk1})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceJwk1)
	defer unsealKeysServiceJwk1.Shutdown()

	// unseal with second valid JWK
	unsealKeysServiceJwk2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJweJwk2})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceJwk2)
	defer unsealKeysServiceJwk2.Shutdown()

	// unseal with invalid JWK
	unsealKeysServiceInvalidJwk, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{nonPublicJweJwkInvalid})
	require.NoError(t, err)
	require.NotNil(t, unsealKeysServiceInvalidJwk)

	// same repository will be used for all tests below
	// barrier encrypt uses the unseal service with both valid JWKs for all tests below

	// barrier decrypt using unseal service with both valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJwks12, unsealKeysServiceJwks12)

	// barrier decrypt using unseal service with first valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJwks12, unsealKeysServiceJwk1)

	// barrier decrypt using unseal service with second valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJwks12, unsealKeysServiceJwk2)

	// TODO barrier decrypt using unseal service with invalid JWKs

	// barrier decrypt using unseal service with both valid JWKs
	encryptDecryptContentRestartDecryptAgain(t, testOrmRepository, unsealKeysServiceJwks12, unsealKeysServiceJwks12)
}

func encryptDecryptContentRestartDecryptAgain(t *testing.T, testOrmRepository *cryptoutilOrmRepository.OrmRepository, originalUnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, restartedUnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) {
	const numEncryptsDecrypts = 11
	plaintext := []byte("hello, world!")

	// start barrier service
	barrierService1, err := NewBarrierService(testCtx, testTelemetryService, testJwkGenService, testOrmRepository, originalUnsealKeysService)
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
	barrierService2, err := NewBarrierService(testCtx, testTelemetryService, testJwkGenService, testOrmRepository, restartedUnsealKeysService)
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
