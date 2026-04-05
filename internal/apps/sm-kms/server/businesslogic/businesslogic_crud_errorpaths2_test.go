package businesslogic

import (
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm-kms/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm-kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func TestGenerateMaterialKeyInElasticKey_NotFound(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	nonExistentID := googleUuid.New()
	_, err := stack.service.GenerateMaterialKeyInElasticKey(stack.ctx, &nonExistentID, nil)

	testify.Error(t, err)
}

func TestGenerateMaterialKeyInElasticKey_InvalidStatus(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	// Use Disabled status which is valid in DB but not allowed for key generation.
	disabledStatus := cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled)
	ekID := seedElasticKey(t, stack, "gen-mk-invalid-status", cryptoutilOpenapiModel.A256GCMDir, disabledStatus)
	_, err := stack.service.GenerateMaterialKeyInElasticKey(stack.ctx, &ekID, nil)

	testify.Error(t, err)
}

func TestImportMaterialKey_NotFound(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	nonExistentID := googleUuid.New()
	_, err := stack.service.ImportMaterialKey(stack.ctx, &nonExistentID, &cryptoutilKmsServer.MaterialKeyImport{
		JWK: `{"kty":"oct","k":"dGVzdGtleWRhdGExMjM0NTY3ODkwMTIzNDU2"}`,
	})

	testify.Error(t, err)
}

// seedImportableElasticKeyWithStatus creates an ElasticKey with ImportAllowed=true and the given status.
func seedImportableElasticKeyWithStatus(t *testing.T, stack *testStack, name string, status cryptoutilKmsServer.ElasticKeyStatus) googleUuid.UUID {
	t.Helper()

	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	ekID := googleUuid.New()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       "test-import-status",
		ElasticKeyProvider:          "Internal",
		ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMDir,
		ElasticKeyVersioningAllowed: false,
		ElasticKeyImportAllowed:     true,
		ElasticKeyStatus:            status,
	}

	testify.NoError(t, stack.core.DB.Create(ek).Error)

	return ekID
}

func TestImportMaterialKey_InvalidStatus(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	// Disabled is valid in DB but not allowed for import (requires PendingImport or Active).
	disabledStatus := cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled)
	ekID := seedImportableElasticKeyWithStatus(t, stack, "import-mk-bad-status", disabledStatus)
	_, err := stack.service.ImportMaterialKey(stack.ctx, &ekID, &cryptoutilKmsServer.MaterialKeyImport{
		JWK: `{"kty":"oct","k":"dGVzdGtleWRhdGExMjM0NTY3ODkwMTIzNDU2"}`,
	})

	testify.Error(t, err)
	testify.Contains(t, err.Error(), "invalid")
}
