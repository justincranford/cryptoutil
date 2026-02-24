package businesslogic

import (
	"context"
	"testing"
	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	testify "github.com/stretchr/testify/require"
	googleUuid "github.com/google/uuid"
)

func TestNoTenantContext(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	noTenant := context.Background()
	id := googleUuid.New()

	_, err := stack.service.GetElasticKeyByElasticKeyID(noTenant, &id)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "tenant context required")

	_, err = stack.service.GetElasticKeys(noTenant, nil)
	testify.Error(t, err)

	_, err = stack.service.GetMaterialKeysForElasticKey(noTenant, &id, nil)
	testify.Error(t, err)

	_, err = stack.service.GetMaterialKeys(noTenant, nil)
	testify.Error(t, err)

	_, err = stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(noTenant, &id, &id)
	testify.Error(t, err)

	_, err = stack.service.UpdateElasticKey(noTenant, &id, &cryptoutilKmsServer.ElasticKeyUpdate{Name: "x"})
	testify.Error(t, err)

	err = stack.service.DeleteElasticKey(noTenant, &id)
	testify.Error(t, err)

	_, err = stack.service.AddElasticKey(noTenant, &cryptoutilKmsServer.ElasticKeyCreate{
		Name: "x", Algorithm: string(cryptoutilOpenapiModel.A256GCMDir), Provider: providerInternal,
	})
	testify.Error(t, err)

	_, err = stack.service.GenerateMaterialKeyInElasticKey(noTenant, &id, nil)
	testify.Error(t, err)

	_, err = stack.service.PostEncryptByElasticKeyID(noTenant, &id, nil, []byte("x"))
	testify.Error(t, err)

	_, err = stack.service.PostSignByElasticKeyID(noTenant, &id, []byte("x"))
	testify.Error(t, err)

	_, err = stack.service.ImportMaterialKey(noTenant, &id, &cryptoutilKmsServer.MaterialKeyImport{JWK: "x"})
	testify.Error(t, err)
}

func TestAddElasticKey_ImportAllowed(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	_, err := stack.service.AddElasticKey(stack.ctx, &cryptoutilKmsServer.ElasticKeyCreate{
		Name:          "import-test",
		Algorithm:     string(cryptoutilOpenapiModel.A256GCMDir),
		Provider:      providerInternal,
		ImportAllowed: ptr(true),
	})
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "elasticKeyImportAllowed=true not supported yet")
}

func TestPostDecrypt_InvalidJWE(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := googleUuid.New()
	_, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, []byte("not-a-jwe"))
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to parse JWE message bytes")
}

func TestPostVerify_InvalidJWS(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := googleUuid.New()
	_, err := stack.service.PostVerifyByElasticKeyID(stack.ctx, &ekID, []byte("not-a-jws"))
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to parse JWS message bytes")
}

func TestPostGenerate_InvalidAlgorithm(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := googleUuid.New()
	badAlg := cryptoutilOpenapiModel.GenerateAlgorithm("BAD")
	_, _, _, err := stack.service.PostGenerateByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.GenerateParams{
		Alg: &badAlg,
	})
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to map generate algorithm")
}

func TestPostEncrypt_NoMaterialKey(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "enc-nomk", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	_, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, []byte("test"))
	testify.Error(t, err)
}

func TestPostSign_NoMaterialKey(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "sign-nomk", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	_, err := stack.service.PostSignByElasticKeyID(stack.ctx, &ekID, []byte("test"))
	testify.Error(t, err)
}

func TestPostGenerate_NoMaterialKey(t *testing.T) {
	t.Parallel()
	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "gen-nomk", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	validAlg := cryptoutilOpenapiModel.Oct256
	_, _, _, err := stack.service.PostGenerateByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.GenerateParams{Alg: &validAlg})
	testify.Error(t, err)
}

func ptr[T any](v T) *T { return &v }

