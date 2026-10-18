// Copyright (c) 2025-2026 Justin Cranford.

package orm

import (
	"testing"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Sequential: uses the package-shared SQLite fixture.
func TestUpdateElasticKeyMaterialKeyRevoke(t *testing.T) {
	db := testOrmRepository.GormDB()
	tenantID := googleUuid.New()
	elasticKeyID := googleUuid.New()
	materialKeyID := googleUuid.New()
	elasticKey, err := BuildElasticKey(
		tenantID,
		elasticKeyID,
		elasticKeyID.String(),
		"material revocation test",
		cryptoutilOpenapiModel.Internal,
		cryptoutilOpenapiModel.A256GCMDir,
		true,
		true,
		true,
		string(cryptoutilKmsServer.Active),
	)
	require.NoError(t, err)
	require.NoError(t, AddElasticKey(db, testTelemetryService.Slogger, elasticKey))

	materialKey := &MaterialKey{
		ElasticKeyID:                  elasticKeyID,
		MaterialKeyID:                 materialKeyID,
		MaterialKeyEncryptedNonPublic: []byte("encrypted-material"),
	}
	require.NoError(t, AddElasticKeyMaterialKey(db, testTelemetryService.Slogger, materialKey))

	invalidElasticKey := *materialKey
	invalidElasticKey.ElasticKeyID = googleUuid.UUID{}
	require.Error(t, UpdateElasticKeyMaterialKeyRevoke(db, testTelemetryService.Slogger, &invalidElasticKey))

	invalidMaterialKey := *materialKey
	invalidMaterialKey.MaterialKeyID = googleUuid.UUID{}
	require.Error(t, UpdateElasticKeyMaterialKeyRevoke(db, testTelemetryService.Slogger, &invalidMaterialKey))

	revokedAt := time.Now().UTC().UnixMilli()
	materialKey.MaterialKeyRevocationDate = &revokedAt
	require.NoError(t, UpdateElasticKeyMaterialKeyRevoke(db, testTelemetryService.Slogger, materialKey))

	storedMaterial := &MaterialKey{}
	require.NoError(t, db.First(storedMaterial, "elastic_key_id = ? AND material_key_id = ?", elasticKeyID, materialKeyID).Error)
	require.Equal(t, &revokedAt, storedMaterial.MaterialKeyRevocationDate)
}

// Sequential: uses the package-shared SQLite fixture.
func TestUpdateElasticKeyAndStatusValidation(t *testing.T) {
	db := testOrmRepository.GormDB()
	require.Error(t, UpdateElasticKey(db, testTelemetryService.Slogger, &ElasticKey{}))
	require.Error(t, UpdateElasticKeyStatus(db, testTelemetryService.Slogger, googleUuid.UUID{}, cryptoutilKmsServer.Inactive))

	tenantID := googleUuid.New()
	elasticKeyID := googleUuid.New()
	elasticKey, err := BuildElasticKey(
		tenantID,
		elasticKeyID,
		elasticKeyID.String(),
		"elastic key update test",
		cryptoutilOpenapiModel.Internal,
		cryptoutilOpenapiModel.A256GCMDir,
		true,
		true,
		true,
		string(cryptoutilKmsServer.Active),
	)
	require.NoError(t, err)
	require.NoError(t, AddElasticKey(db, testTelemetryService.Slogger, elasticKey))

	elasticKey.ElasticKeyDescription = "updated description"
	require.NoError(t, UpdateElasticKey(db, testTelemetryService.Slogger, elasticKey))
	require.NoError(t, UpdateElasticKeyStatus(db, testTelemetryService.Slogger, elasticKeyID, cryptoutilKmsServer.Inactive))

	storedElasticKey := &ElasticKey{}
	require.NoError(t, db.First(storedElasticKey, "elastic_key_id = ?", elasticKeyID).Error)
	require.Equal(t, "updated description", storedElasticKey.ElasticKeyDescription)
	require.Equal(t, cryptoutilKmsServer.Inactive, storedElasticKey.ElasticKeyStatus)
}

// Sequential: uses the package-shared SQLite fixture.
func TestGetMaterialKeysForElasticKey_InvalidID(t *testing.T) {
	_, err := GetMaterialKeysForElasticKey(
		testOrmRepository.GormDB(),
		testTelemetryService.Slogger,
		&googleUuid.UUID{},
		nil,
	)
	require.Error(t, err)
}
