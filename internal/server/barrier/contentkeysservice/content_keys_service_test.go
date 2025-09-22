package contentkeysservice

import (
	"context"
	"os"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilIntermediateKeysService "cryptoutil/internal/server/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testSettings                = cryptoutilConfig.RequireNewForTest("content_keys_service_test")
	testCtx                     = context.Background()
	testTelemetryService        *cryptoutilTelemetry.TelemetryService
	testJwkGenService           *cryptoutilJose.JwkGenService
	testSQLRepository           *cryptoutilSqlRepository.SqlRepository
	testOrmRepository           *cryptoutilOrmRepository.OrmRepository
	testRootKeysService         *cryptoutilRootKeysService.RootKeysService
	testIntermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		testJwkGenService = cryptoutilJose.RequireNewForTest(testCtx, testTelemetryService)
		defer testJwkGenService.Shutdown()

		testSQLRepository = cryptoutilSqlRepository.RequireNewForTest(testCtx, testTelemetryService, testSettings)
		defer testSQLRepository.Shutdown()

		testOrmRepository = cryptoutilOrmRepository.RequireNewForTest(testCtx, testTelemetryService, testSQLRepository, testJwkGenService, testSettings)
		defer testOrmRepository.Shutdown()

		_, unsealJwk, _, _, _, err := testJwkGenService.GenerateJweJwk(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
		cryptoutilAppErr.RequireNoError(err, "failed to generate unseal JWK for test")

		unsealKeysService := cryptoutilUnsealKeysService.RequireNewSimpleForTest([]joseJwk.Key{unsealJwk})
		defer unsealKeysService.Shutdown()

		testRootKeysService = cryptoutilRootKeysService.RequireNewForTest(testTelemetryService, testJwkGenService, testOrmRepository, unsealKeysService)
		defer testRootKeysService.Shutdown()

		testIntermediateKeysService = cryptoutilIntermediateKeysService.RequireNewForTest(testTelemetryService, testJwkGenService, testOrmRepository, testRootKeysService)
		defer testIntermediateKeysService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestContentKeysService_HappyPath(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testJwkGenService, testOrmRepository, testIntermediateKeysService)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)
	defer contentKeysService.Shutdown()

	clearBytes := []byte("Hello World")

	var encrypted []byte
	var contentKeyKidUuid *googleUuid.UUID
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encrypted, contentKeyKidUuid, err = contentKeysService.EncryptContent(sqlTransaction, clearBytes)
		require.NoError(t, err)
		require.NotNil(t, encrypted)
		require.NotNil(t, contentKeyKidUuid)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, encrypted)
	require.NotNil(t, contentKeyKidUuid)

	var decrypted []byte
	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decrypted, err = contentKeysService.DecryptContent(sqlTransaction, encrypted)
		require.NoError(t, err)
		require.NotNil(t, decrypted)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, decrypted)

	require.Equal(t, clearBytes, decrypted)
}
