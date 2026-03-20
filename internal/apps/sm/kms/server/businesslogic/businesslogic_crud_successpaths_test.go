package businesslogic

import (
	"testing"

	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func seedImportableElasticKey(t *testing.T, stack *testStack, name string) googleUuid.UUID {
	t.Helper()

	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	ekID := googleUuid.New()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       "test-import",
		ElasticKeyProvider:          "Internal",
		ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMDir,
		ElasticKeyVersioningAllowed: false,
		ElasticKeyImportAllowed:     true,
		ElasticKeyStatus:            cryptoutilKmsServer.Active,
	}

	testify.NoError(t, stack.core.DB.Create(ek).Error)

	return ekID
}

func TestAddElasticKey_Success(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	desc := "add-ek-desc"
	result, err := stack.service.AddElasticKey(stack.ctx, &cryptoutilKmsServer.ElasticKeyCreate{
		Name:        "add-ek-success",
		Algorithm:   string(cryptoutilOpenapiModel.A256GCMDir),
		Provider:    providerInternal,
		Description: &desc,
	})

	testify.NoError(t, err)
	testify.NotNil(t, result)
	testify.Equal(t, "add-ek-success", *result.Name)
	testify.NotNil(t, result.ElasticKeyID)
}

func TestGenerateMaterialKeyInElasticKey_Success(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "gen-mk-success", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	result, err := stack.service.GenerateMaterialKeyInElasticKey(stack.ctx, &ekID, nil)

	testify.NoError(t, err)
	testify.NotNil(t, result)
	testify.NotNil(t, result.ElasticKeyID)
	testify.Equal(t, ekID, *result.ElasticKeyID)
}

func TestImportMaterialKey_Success(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	ekID := seedImportableElasticKey(t, stack, "import-mk-success")
	result, err := stack.service.ImportMaterialKey(stack.ctx, &ekID, &cryptoutilKmsServer.MaterialKeyImport{
		JWK: `{"kty":"oct","k":"dGVzdGtleWRhdGExMjM0NTY3ODkwMTIzNDU2"}`,
	})

	testify.NoError(t, err)
	testify.NotNil(t, result)
	testify.NotNil(t, result.ElasticKeyID)
	testify.Equal(t, ekID, *result.ElasticKeyID)
}
