package barrier

import (
	"context"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"fmt"
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
	// testGivens             *orm.Givens
	skipReadOnlyTxTests bool
	testDbType          = cryptoutilSqlProvider.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	// testDbType = cryptoutilSqlProvider.DBTypePostgres
)

// var happyPathTestCases = []struct {
// 	alg jwa.KeyEncryptionAlgorithm
// }{
// 	{alg: jose.AlgA256GCMKW},
// 	{alg: jose.AlgDIRECT},
// }

func TestMain(m *testing.M) {
	rc := runAllTestsAndShutdown(m)
	os.Exit(rc)
}

func runAllTestsAndShutdown(m *testing.M) int {
	var err error

	testTelemetryService, err = cryptoutilTelemetry.NewService(testCtx, "servicelogic_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	defer testTelemetryService.Shutdown()

	switch testDbType {
	case cryptoutilSqlProvider.DBTypeSQLite:
		skipReadOnlyTxTests = true
		testSqlProvider, err = cryptoutilSqlProvider.NewSqlProvider(testCtx, testTelemetryService, cryptoutilSqlProvider.DBTypeSQLite, ":memory:", cryptoutilSqlProvider.ContainerModeDisabled)
	case cryptoutilSqlProvider.DBTypePostgres:
		skipReadOnlyTxTests = false
		testSqlProvider, err = cryptoutilSqlProvider.NewSqlProvider(testCtx, testTelemetryService, cryptoutilSqlProvider.DBTypePostgres, "", cryptoutilSqlProvider.ContainerModeRequired)
	default:
		err = fmt.Errorf("unsupported dbType %s", testDbType)
	}
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

	// testGivens = orm.NewGivens(testCtx, testTelemetryService)
	// defer testGivens.Shutdown()

	rc := m.Run()
	return rc
}

func Test_HappyPath_Bytes(t *testing.T) {
	plaintext := []byte("hello, world!")

	barrierService, err := NewBarrierService(testCtx, testTelemetryService, testRepositoryProvider)
	require.NoError(t, err)
	defer barrierService.Shutdown()

	encryptedBytes, err := barrierService.Encrypt(plaintext)
	require.NoError(t, err)
	t.Logf("JWE Headers: %s", string(encryptedBytes))

	decryptedBytes, err := barrierService.Decrypt(encryptedBytes)
	require.NoError(t, err)
	t.Logf("Decrypted: %s", string(decryptedBytes))

	require.Equal(t, plaintext, decryptedBytes)
	t.Log("Success")
}
