package barrierservice

import (
	"context"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"log/slog"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	testCtx                = context.Background()
	testTelemetryService   *cryptoutilTelemetry.Service
	testSqlProvider        *cryptoutilSqlProvider.SqlProvider
	testRepositoryProvider *cryptoutilOrmRepository.RepositoryProvider
	testDbType             = cryptoutilSqlProvider.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
)

func TestMain(m *testing.M) {
	var err error

	testTelemetryService, err = cryptoutilTelemetry.NewService(testCtx, "servicelogic_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	defer testTelemetryService.Shutdown()

	testSqlProvider, err = cryptoutilSqlProvider.NewSqlProviderForTest(testCtx, testTelemetryService, testDbType)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize sqlProvider", "error", err)
		os.Exit(-1)
	}
	defer testSqlProvider.Shutdown()

	testRepositoryProvider, err = cryptoutilOrmRepository.NewRepositoryOrm(testCtx, testTelemetryService, testSqlProvider, true)
	if err != nil {
		testTelemetryService.Slogger.Error("failed to initailize repositoryProvider", "error", err)
		os.Exit(-1)
	}
	defer testRepositoryProvider.Shutdown()

	os.Exit(m.Run())
}

func Test_HappyPath_Bytes(t *testing.T) {
	plaintext := []byte("hello, world!")

	// start service
	barrierService, err := NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider)
	require.NoError(t, err)
	defer barrierService.Shutdown()

	// encrypt N times
	encryptedBytesSlice := make([][]byte, 0, 11)
	for i := range cap(encryptedBytesSlice) {
		t.Logf("Attempt: %d", i+1)
		encryptedBytes, err := barrierService.EncryptContent(plaintext)
		require.NoError(t, err)
		t.Logf("Encrypted Data > JWE Headers: %s", string(encryptedBytes))
		encryptedBytesSlice = append(encryptedBytesSlice, encryptedBytes)
	}

	// decrypt N times with service
	for _, encryptedBytes := range encryptedBytesSlice {
		decryptedBytes, err := barrierService.DecryptContent(encryptedBytes)
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}

	// shutdown service
	barrierService.Shutdown()
	_, err = barrierService.EncryptContent(plaintext)
	require.Error(t, err)
	_, err = barrierService.DecryptContent(plaintext)
	require.Error(t, err)

	// restart service
	barrierService, err = NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider)
	require.NoError(t, err)
	defer barrierService.Shutdown()

	// decrypt N times with restarted service
	for _, encryptedBytes := range encryptedBytesSlice {
		decryptedBytes, err := barrierService.DecryptContent(encryptedBytes)
		require.NoError(t, err)
		require.Equal(t, plaintext, decryptedBytes)
	}
	t.Log("Success")
}
