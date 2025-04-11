package barrierservice

import (
	"context"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/util/sysinfo"
	"log/slog"
	"os"
	"testing"

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
	var err error

	testTelemetryService, err = cryptoutilTelemetry.NewService(testCtx, "servicelogic_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	defer testTelemetryService.Shutdown()

	os.Exit(m.Run())
}

func Test_HappyPath_SameUnsealJwks(t *testing.T) {
	testSqlProvider, err := cryptoutilSqlProvider.NewSqlProviderForTest(testCtx, testTelemetryService, testDbType)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err := cryptoutilOrmRepository.NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	originalUnsealKeyRepository, err := cryptoutilUnsealRepository.NewUnsealKeyRepositoryFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	require.NoError(t, err)

	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealKeyRepository, originalUnsealKeyRepository)
}

func Test_HappyPath_EncryptDecryptContent_Restart_DecryptAgain(t *testing.T) {
	testSqlProvider, err := cryptoutilSqlProvider.NewSqlProviderForTest(testCtx, testTelemetryService, testDbType)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err := cryptoutilOrmRepository.NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	originalUnsealKeyRepository, unsealJwks, err := cryptoutilUnsealRepository.NewUnsealKeyRepositoryMock(t, 2)
	require.NoError(t, err)
	require.NotNil(t, unsealJwks)
	require.Len(t, unsealJwks, 2)

	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealKeyRepository, originalUnsealKeyRepository)

	unsealJwksCopy := make([]joseJwk.Key, len(unsealJwks))
	copy(unsealJwksCopy, unsealJwks)

	// all same keys
	restartedUnsealKeyRepository2, err := cryptoutilUnsealRepository.NewUnsealKeyRepositorySimple(unsealJwksCopy)
	require.NoError(t, err)
	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealKeyRepository, restartedUnsealKeyRepository2)

	// remove first key, 2nd key remains
	// unsealJwksCopy = unsealJwksCopy[1:]
	restartedUnsealKeyRepository1, err := cryptoutilUnsealRepository.NewUnsealKeyRepositorySimple(unsealJwksCopy)
	require.NoError(t, err)
	encryptDecryptContent_Restart_DecryptAgain(t, testRepositoryProvider, originalUnsealKeyRepository, restartedUnsealKeyRepository1)
}

func encryptDecryptContent_Restart_DecryptAgain(t *testing.T, testRepositoryProvider *cryptoutilOrmRepository.RepositoryProvider, originalUnsealKeyRepository cryptoutilUnsealRepository.UnsealKeyRepository, restartedUnsealKeyRepository cryptoutilUnsealRepository.UnsealKeyRepository) {
	const numEncryptsDecrypts = 11
	plaintext := []byte("hello, world!")

	// start service
	barrierService, err := NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider, originalUnsealKeyRepository)
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
	barrierService, err = NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider, restartedUnsealKeyRepository)
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
