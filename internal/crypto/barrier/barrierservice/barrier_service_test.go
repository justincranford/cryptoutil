package barrierservice

import (
	"context"
	"os"
	"testing"

	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilUnsealService "cryptoutil/internal/crypto/barrier/unsealservice"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testDbType           = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "servicelogic_test", false, false)
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func Test_HappyPath_SameUnsealJwks(t *testing.T) {
	testSqlRepository := cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	originalUnsealRepository, err := cryptoutilUnsealRepository.NewUnsealRepositoryFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	require.NoError(t, err)

	originalUnsealService, err := cryptoutilUnsealService.NewUnsealService(testTelemetryService, testOrmRepository, originalUnsealRepository)
	require.NoError(t, err)

	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealService, originalUnsealService)
}

func Test_HappyPath_EncryptDecryptContent_Restart_DecryptAgain(t *testing.T) {
	testSqlRepository := cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlRepository.Shutdown()

	testOrmRepository := cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSqlRepository, true)
	defer testOrmRepository.Shutdown()

	originalUnsealRepository, unsealJwks, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 2)
	require.NoError(t, err)
	require.NotNil(t, unsealJwks)
	require.Len(t, unsealJwks, 2)
	originalUnsealService, err := cryptoutilUnsealService.NewUnsealService(testTelemetryService, testOrmRepository, originalUnsealRepository)
	require.NoError(t, err)

	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealService, originalUnsealService)

	unsealJwksCopy := make([]joseJwk.Key, len(unsealJwks))
	copy(unsealJwksCopy, unsealJwks)

	restartedUnsealRepository2, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwks)
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository2)
	restartedUnsealService2, err := cryptoutilUnsealService.NewUnsealService(testTelemetryService, testOrmRepository, restartedUnsealRepository2)
	require.NoError(t, err)
	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealService, restartedUnsealService2)

	restartedUnsealRepository1a, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwksCopy[:1])
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository1a)
	restartedUnsealService1a, err := cryptoutilUnsealService.NewUnsealService(testTelemetryService, testOrmRepository, restartedUnsealRepository1a)
	require.NoError(t, err)
	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealService, restartedUnsealService1a)

	restartedUnsealRepository1b, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwksCopy[1:])
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository1b)
	restartedUnsealService1b, err := cryptoutilUnsealService.NewUnsealService(testTelemetryService, testOrmRepository, restartedUnsealRepository1b)
	require.NoError(t, err)
	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealService, restartedUnsealService1b)

	invalidJwk, _, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
	require.NoError(t, err)
	invalidUnsealRepository, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple([]joseJwk.Key{invalidJwk})
	require.NoError(t, err)
	require.NotNil(t, invalidUnsealRepository)
	invalidUnsealService, err := cryptoutilUnsealService.NewUnsealService(testTelemetryService, testOrmRepository, invalidUnsealRepository)
	require.Error(t, err)
	require.Nil(t, invalidUnsealService)

	// retry previously successful unseal
	restartedUnsealRepository1a, err = cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwksCopy[:1])
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository1a)
	encryptDecryptContent_Restart_DecryptAgain(t, testOrmRepository, originalUnsealService, restartedUnsealService1b)
}

func encryptDecryptContent_Restart_DecryptAgain(t *testing.T, testOrmRepository *cryptoutilOrmRepository.OrmRepository, originalUnsealService *cryptoutilUnsealService.UnsealService, restartedUnsealService *cryptoutilUnsealService.UnsealService) {
	const numEncryptsDecrypts = 11
	plaintext := []byte("hello, world!")

	// start service
	barrierService, err := NewBarrierService(testCtx, testTelemetryService, testOrmRepository, originalUnsealService)
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
	barrierService, err = NewBarrierService(testCtx, testTelemetryService, testOrmRepository, restartedUnsealService)
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
