package barrierservice

import (
	"context"
	"os"
	"testing"

	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
	testDbType           = cryptoutilSqlProvider.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "servicelogic_test", false, false)
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func Test_HappyPath_SameUnsealJwks(t *testing.T) {
	testSqlProvider := cryptoutilSqlProvider.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err := cryptoutilOrmRepository.NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	originalUnsealRepository, err := cryptoutilUnsealRepository.NewUnsealRepositoryFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	require.NoError(t, err)

	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealRepository, originalUnsealRepository)
}

func Test_HappyPath_EncryptDecryptContent_Restart_DecryptAgain(t *testing.T) {
	testSqlProvider := cryptoutilSqlProvider.RequireNewForTest(testCtx, testTelemetryService, testDbType)
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err := cryptoutilOrmRepository.NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	originalUnsealRepository, unsealJwks, err := cryptoutilUnsealRepository.NewUnsealRepositoryMock(t, 2)
	require.NoError(t, err)
	require.NotNil(t, unsealJwks)
	require.Len(t, unsealJwks, 2)

	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealRepository, originalUnsealRepository)

	unsealJwksCopy := make([]joseJwk.Key, len(unsealJwks))
	copy(unsealJwksCopy, unsealJwks)

	restartedUnsealRepository2, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwks)
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository2)
	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealRepository, restartedUnsealRepository2)

	restartedUnsealRepository1a, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwksCopy[:1])
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository1a)
	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealRepository, restartedUnsealRepository1a)

	restartedUnsealRepository1b, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwksCopy[1:])
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository1b)
	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealRepository, restartedUnsealRepository1b)

	invalidJwk, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
	require.NoError(t, err)
	restartedUnsealRepository0, err := cryptoutilUnsealRepository.NewUnsealRepositorySimple([]joseJwk.Key{invalidJwk})
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository0)
	barrierService, err := NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider, restartedUnsealRepository0)
	require.Error(t, err)
	require.Nil(t, barrierService)

	// retry previously successful unseal
	restartedUnsealRepository1a, err = cryptoutilUnsealRepository.NewUnsealRepositorySimple(unsealJwksCopy[:1])
	require.NoError(t, err)
	require.NotNil(t, restartedUnsealRepository1a)
	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealRepository, restartedUnsealRepository1a)
}

func encryptDecryptContent_Restart_DecryptAgain(t *testing.T, testRepositoryProvider *cryptoutilOrmRepository.RepositoryProvider, originalUnsealRepository cryptoutilUnsealRepository.UnsealRepository, restartedUnsealRepository cryptoutilUnsealRepository.UnsealRepository) {
	const numEncryptsDecrypts = 11
	plaintext := []byte("hello, world!")

	// start service
	barrierService, err := NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider, originalUnsealRepository)
	require.NoError(t, err)
	defer barrierService.Shutdown()

	// encrypt N times
	encryptedBytesSlice := make([][]byte, 0, numEncryptsDecrypts)
	for i := range cap(encryptedBytesSlice) {
		t.Logf("Attempt: %d", i+1)
		encryptedBytes, err := barrierService.EncryptContent(plaintext)
		require.NoError(t, err)
		t.Logf("Encrypted Data > JWE Headers: %s", string(encryptedBytes))
		encryptedBytesSlice = append(encryptedBytesSlice, encryptedBytes)
	}

	// decrypt N times with original service
	for _, encryptedBytes := range encryptedBytesSlice {
		decryptedBytes, err := barrierService.DecryptContent(encryptedBytes)
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}

	// shutdown original service
	barrierService.Shutdown()
	_, err = barrierService.EncryptContent(plaintext)
	require.Error(t, err)
	_, err = barrierService.DecryptContent(plaintext)
	require.Error(t, err)

	// restart new service with same unseal key repository
	barrierService, err = NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider, restartedUnsealRepository)
	require.NoError(t, err)
	defer barrierService.Shutdown()

	// decrypt N times with restarted service
	for _, encryptedBytes := range encryptedBytesSlice {
		decryptedBytes, err := barrierService.DecryptContent(encryptedBytes)
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}

	// shutdown restarted service
	barrierService.Shutdown()
	_, err = barrierService.EncryptContent(plaintext)
	require.Error(t, err)
	_, err = barrierService.DecryptContent(plaintext)
	require.Error(t, err)

	t.Log("Success")
}
