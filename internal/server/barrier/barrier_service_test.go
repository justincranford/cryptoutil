package barrierservice

import (
	"context"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"
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
	testDbType           = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
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
	testSqlRepository := cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSqlRepository.Shutdown()

	testOrmRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testJwkGenService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	originalUnsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	require.NoError(t, err)
	defer originalUnsealKeysService.Shutdown()

	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealKeysService, originalUnsealKeysService)
}

func Test_HappyPath_EncryptDecryptContent_Restart_DecryptAgain(t *testing.T) {
	testSqlRepository := cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
	defer testSqlRepository.Shutdown()

	testOrmRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testJwkGenService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	_, unsealJwk1, _, _, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, unsealJwk1)

	_, unsealJwk2, _, _, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	require.NotNil(t, unsealJwk2)

	originalUnsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJwk1, unsealJwk2})
	require.NoError(t, err)
	defer originalUnsealKeysService.Shutdown()

	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealKeysService, originalUnsealKeysService)

	restartedUnsealKeysService1a, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJwk1})
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealKeysService1a)
	defer restartedUnsealKeysService1a.Shutdown()

	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealKeysService, restartedUnsealKeysService1a)

	restartedUnsealKeysService1b, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJwk2})
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealKeysService1b)
	defer restartedUnsealKeysService1b.Shutdown()

	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealKeysService, restartedUnsealKeysService1b)

	_, invalidJwk, _, _, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	invalidUnsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{invalidJwk})
	require.NoError(t, err)
	require.NotNil(t, invalidUnsealKeysService)

	// retry previously successful unseal
	retryRestartedUnsealKeysService1a, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJwk1})
	require.NoError(t, err)
	require.NotNil(t, retryRestartedUnsealKeysService1a)
	defer retryRestartedUnsealKeysService1a.Shutdown()

	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealKeysService, retryRestartedUnsealKeysService1a)
}

func encryptDecryptContent_Restart_DecryptAgain(t *testing.T, testOrmRepository *cryptoutilOrmRepository.OrmRepository, originalUnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, restartedUnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) {
	const numEncryptsDecrypts = 11
	plaintext := []byte("hello, world!")

	// start service
	barrierService, err := NewBarrierService(testCtx, testTelemetryService, testJwkGenService, testOrmRepository, originalUnsealKeysService)
	require.NoError(t, err)
	defer barrierService.Shutdown()

	// encrypt N times
	encryptedBytesSlice := make([][]byte, 0, numEncryptsDecrypts)
	for i := range cap(encryptedBytesSlice) {
		t.Logf("Attempt: %d", i+1)
		var encryptedBytes []byte
		err := testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			var err error
			encryptedBytes, err = barrierService.EncryptContent(sqlTransaction, plaintext)
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
			decryptedBytes, err = barrierService.DecryptContent(sqlTransaction, encryptedBytes)
			return err
		})
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}

	// shutdown original service
	barrierService.Shutdown()
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		_, err = barrierService.EncryptContent(sqlTransaction, plaintext)
		return err
	})
	require.Error(t, err)
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		_, err = barrierService.DecryptContent(sqlTransaction, plaintext)
		return err
	})
	require.Error(t, err)

	// restart new service with same unseal key repository
	barrierService, err = NewBarrierService(testCtx, testTelemetryService, testJwkGenService, testOrmRepository, restartedUnsealKeysService)
	require.NoError(t, err)
	defer barrierService.Shutdown()

	// decrypt N times with restarted service
	for _, encryptedBytes := range encryptedBytesSlice {
		var decryptedBytes []byte
		err := testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
			var err error
			decryptedBytes, err = barrierService.DecryptContent(sqlTransaction, encryptedBytes)
			return err
		})
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}

	// shutdown restarted service
	barrierService.Shutdown()
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		_, err = barrierService.EncryptContent(sqlTransaction, plaintext)
		return err
	})
	require.Error(t, err)
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		_, err = barrierService.DecryptContent(sqlTransaction, plaintext)
		return err
	})
	require.Error(t, err)

	t.Log("Success")
}
