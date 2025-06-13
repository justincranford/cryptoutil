package intermediatekeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
	testJwkGenService    *cryptoutilJose.JwkGenService
	testSqlRepository    *cryptoutilSqlRepository.SqlRepository
	testOrmRepository    *cryptoutilOrmRepository.OrmRepository
	testDbType           = cryptoutilSqlRepository.DBTypeSQLite // Caution: modernc.org/sqlite doesn't support read-only transactions, but PostgreSQL does
	testRootKeysService  *cryptoutilRootKeysService.RootKeysService
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "intermediate_keys_service_test", false, false)
		defer testTelemetryService.Shutdown()

		testJwkGenService = cryptoutilJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJwkGenService.Shutdown()

		testSqlRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testDbType)
		defer testSqlRepository.Shutdown()

		testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testJwkGenService, testSqlRepository, true)
		defer testOrmRepository.Shutdown()

		unsealKeysService := cryptoutilUnsealKeysService.RequireNewFromSysInfoForTest()
		defer unsealKeysService.Shutdown()

		testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testJwkGenService, testOrmRepository, unsealKeysService)
		defer testRootKeysService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestIntermediateKeysService_HappyPath(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testJwkGenService, testOrmRepository, testRootKeysService)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)
	defer intermediateKeysService.Shutdown()

	_, clearContentKey, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir)
	require.NoError(t, err)
	require.NotNil(t, clearContentKey)

	var encryptedContentKeyBytes []byte
	var intermediateKeyKidUuid *googleUuid.UUID
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedContentKeyBytes, intermediateKeyKidUuid, err = intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
		require.NoError(t, err)
		require.NotNil(t, encryptedContentKeyBytes)
		require.NotNil(t, intermediateKeyKidUuid)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, encryptedContentKeyBytes)
	require.NotNil(t, intermediateKeyKidUuid)

	var decryptedContentKey joseJwk.Key
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedContentKey, err = intermediateKeysService.DecryptKey(sqlTransaction, encryptedContentKeyBytes)
		require.NoError(t, err)
		require.NotNil(t, decryptedContentKey)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, decryptedContentKey)

	require.ElementsMatch(t, clearContentKey.Keys(), decryptedContentKey.Keys())
}
