// Copyright (c) 2025-2026 Justin Cranford.
//

package businesslogic

import (
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm-kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

const (
	testInvalidMaterialKeyBytes = "invalid-material-key"
	testInvalidPublicJWKBytes   = "invalid-public-jwk"
	testInvalidJWKJSON          = "not-json"
)

// Sequential: mutates the package-shared SQLite database and barrier service.
func TestGetAndDecryptMaterialKey_InvalidStoredKeyMaterial(t *testing.T) {
	tests := []struct {
		name                    string
		algorithm               cryptoutilOpenapiModel.ElasticKeyAlgorithm
		encryptedMaterialBytes  []byte
		clearPublicMaterial     []byte
		encryptInvalidJWKBefore bool
		wantError               string
	}{
		{
			name:                   "barrier ciphertext rejected",
			algorithm:              cryptoutilOpenapiModel.A256GCMDir,
			encryptedMaterialBytes: []byte(testInvalidMaterialKeyBytes),
			wantError:              "failed to decrypt MaterialKeyEncryptedNonPublic",
		},
		{
			name:                    "decrypted private key is invalid",
			algorithm:               cryptoutilOpenapiModel.A256GCMDir,
			encryptInvalidJWKBefore: true,
			wantError:               "failed to parse decrypted MaterialKeyEncryptedNonPublic",
		},
		{
			name:                "clear public key is invalid",
			algorithm:           cryptoutilOpenapiModel.A256GCMECDHESA256KW,
			clearPublicMaterial: []byte(testInvalidPublicJWKBytes),
			wantError:           "failed to parse MaterialKeyClearPublic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stack := setupCryptoTestStack(t)
			nameID, err := googleUuid.NewV7()
			testify.NoError(t, err)
			elasticKeyID := seedBarrierEncryptedElasticKey(t, stack, "invalid-material-"+nameID.String(), tc.algorithm)

			encryptedMaterial := tc.encryptedMaterialBytes
			if tc.encryptInvalidJWKBefore {
				encryptedMaterial, err = stack.service.barrierService.EncryptContentWithContext(stack.ctx, []byte(testInvalidJWKJSON))
				testify.NoError(t, err)
			}

			updates := map[string]any{}
			if encryptedMaterial != nil {
				updates["material_key_encrypted_non_public"] = encryptedMaterial
			}

			if tc.clearPublicMaterial != nil {
				updates["material_key_clear_public"] = tc.clearPublicMaterial
			}

			testify.NoError(
				t,
				stack.core.DB.Model(&cryptoutilOrmRepository.MaterialKey{}).
					Where("elastic_key_id = ?", elasticKeyID).
					Updates(updates).
					Error,
			)

			_, err = stack.service.PostEncryptByElasticKeyID(stack.ctx, &elasticKeyID, &cryptoutilOpenapiModel.EncryptParams{}, []byte("payload"))
			testify.ErrorContains(t, err, tc.wantError)
		})
	}
}

// Sequential: shuts down a package-local barrier service to exercise ImportMaterialKey's error path.
func TestImportMaterialKey_BarrierClosed(t *testing.T) {
	stack := setupCryptoTestStack(t)
	useClosedBarrier(stack)

	elasticKeyID, err := googleUuid.NewV7()
	testify.NoError(t, err)
	_, err = stack.service.ImportMaterialKey(stack.ctx, &elasticKeyID, &cryptoutilKmsServer.MaterialKeyImport{JWK: testInvalidJWKJSON})
	testify.ErrorContains(t, err, "failed to encrypt imported MaterialKey")
}

// Sequential: mutates the package-shared SQLite database.
func TestGenerateMaterialKey_BarrierClosed(t *testing.T) {
	stack := setupCryptoTestStack(t)
	elasticKeyID := seedBarrierEncryptedElasticKey(t, stack, "generate-material-barrier-closed", cryptoutilOpenapiModel.A256GCMDir)
	useClosedBarrier(stack)

	_, err := stack.service.GenerateMaterialKeyInElasticKey(stack.ctx, &elasticKeyID, nil)
	testify.ErrorContains(t, err, "failed to encrypt new MaterialKey for ElasticKey")
}

// Sequential: exercises shared-database service setup with isolated barrier failures.
func TestAddElasticKey_GenerationErrors(t *testing.T) {
	tests := []struct {
		name         string
		algorithm    cryptoutilOpenapiModel.ElasticKeyAlgorithm
		closeBarrier bool
		wantError    string
	}{
		{
			name:      "unsupported algorithm",
			algorithm: cryptoutilOpenapiModel.ElasticKeyAlgorithm("UNSUPPORTED"),
			wantError: "failed to generate first MaterialKey for ElasticKey",
		},
		{
			name:         "barrier encryption fails",
			algorithm:    cryptoutilOpenapiModel.A256GCMDir,
			closeBarrier: true,
			wantError:    "failed to encrypt first MaterialKey for ElasticKey",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stack := setupCryptoTestStack(t)
			if tc.closeBarrier {
				useClosedBarrier(stack)
			}

			description := testCryptoDescription
			_, err := stack.service.AddElasticKey(stack.ctx, &cryptoutilKmsServer.ElasticKeyCreate{
				Algorithm:   string(tc.algorithm),
				Description: &description,
				Name:        "add-material-" + tc.name,
			})
			testify.ErrorContains(t, err, tc.wantError)
		})
	}
}

// Sequential: mutates the package-shared SQLite database.
func TestPostEncrypt_ElasticKeyNotFound(t *testing.T) {
	stack := setupCryptoTestStack(t)
	elasticKeyID, err := googleUuid.NewV7()
	testify.NoError(t, err)

	_, err = stack.service.PostEncryptByElasticKeyID(stack.ctx, &elasticKeyID, &cryptoutilOpenapiModel.EncryptParams{}, []byte("payload"))
	testify.ErrorContains(t, err, "failed to get and decrypt latest MaterialKey for ElasticKey")
}

// Sequential: mutates the package-shared SQLite database.
func TestGetMaterialKeysForElasticKey_ElasticKeyNotFound(t *testing.T) {
	stack := setupCryptoTestStack(t)
	elasticKeyID, err := googleUuid.NewV7()
	testify.NoError(t, err)

	_, err = stack.service.GetMaterialKeysForElasticKey(stack.ctx, &elasticKeyID, nil)
	testify.ErrorContains(t, err, "failed to get MaterialKey for ElasticKey")
}

// Sequential: mutates the package-shared SQLite database.
func TestGetMaterialKeysForElasticKey_InvalidStoredMaterialKey(t *testing.T) {
	stack := setupCryptoTestStack(t)
	elasticKeyID := seedBarrierEncryptedElasticKey(t, stack, "invalid-stored-material-key", cryptoutilOpenapiModel.A256GCMDir)
	testify.NoError(t, stack.core.DB.Create(&cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  elasticKeyID,
		MaterialKeyID:                 googleUuid.UUID{},
		MaterialKeyEncryptedNonPublic: []byte(testInvalidMaterialKeyBytes),
	}).Error)

	_, err := stack.service.GetMaterialKeysForElasticKey(stack.ctx, &elasticKeyID, nil)
	testify.ErrorContains(t, err, "failed to map MaterialKey for ElasticKey")
}

func useClosedBarrier(stack *testStack) {
	closedBarrier := &cryptoutilAppsFrameworkServiceServerBarrier.Service{}
	closedBarrier.Shutdown()
	stack.service.barrierService = closedBarrier
}
