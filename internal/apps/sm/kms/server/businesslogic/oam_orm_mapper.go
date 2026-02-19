// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"fmt"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
)

// OamOrmMapper provides mapping between OpenAPI Model models and ORM objects.
type OamOrmMapper struct{}

// NewOamOrmMapper creates a new OpenAPI Model to ORM mapper.
func NewOamOrmMapper() *OamOrmMapper {
	return &OamOrmMapper{}
}

// ErrInvalidUUID is the error message for invalid UUID validation failures.
var ErrInvalidUUID = "invalid UUIDs"

// Default values for optional ElasticKey fields.
var (
	defaultElasticKeyProvider          = cryptoutilOpenapiModel.Internal
	defaultElasticKeyAlgorithm         = cryptoutilOpenapiModel.A256GCMA256KW
	defaultElasticKeyVersioningAllowed = true
	defaultElasticKeyImportAllowed     = false
)

// oam => orm

func (m *OamOrmMapper) toOrmAddElasticKey(elasticKeyID *googleUuid.UUID, tenantID googleUuid.UUID, oamElasticKeyCreate *cryptoutilKmsServer.ElasticKeyCreate) *cryptoutilOrmRepository.ElasticKey {
	// Apply defaults for optional fields
	provider := cryptoutilOpenapiModel.ElasticKeyProvider(oamElasticKeyCreate.Provider)
	if provider == "" {
		provider = defaultElasticKeyProvider
	}

	algorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm(oamElasticKeyCreate.Algorithm)
	if algorithm == "" {
		algorithm = defaultElasticKeyAlgorithm
	}

	versioningAllowed := defaultElasticKeyVersioningAllowed
	if oamElasticKeyCreate.VersioningAllowed != nil {
		versioningAllowed = *oamElasticKeyCreate.VersioningAllowed
	}

	importAllowed := defaultElasticKeyImportAllowed
	if oamElasticKeyCreate.ImportAllowed != nil {
		importAllowed = *oamElasticKeyCreate.ImportAllowed
	}

	description := ""
	if oamElasticKeyCreate.Description != nil {
		description = *oamElasticKeyCreate.Description
	}

	return &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                *elasticKeyID,
		TenantID:                    tenantID,
		ElasticKeyName:              oamElasticKeyCreate.Name,
		ElasticKeyDescription:       description,
		ElasticKeyProvider:          provider,
		ElasticKeyAlgorithm:         algorithm,
		ElasticKeyVersioningAllowed: versioningAllowed,
		ElasticKeyImportAllowed:     importAllowed,
		ElasticKeyStatus:            toElasticKeyStatusFromImportAllowed(importAllowed),
	}
}

func (*OamOrmMapper) toOrmAddMaterialKey(elasticKeyID, materialKeyID *googleUuid.UUID, materialKeyClearPublicJWKBytes, materialKeyEncryptedNonPublicJWKBytes []byte, materialKeyGenerateDate time.Time) *cryptoutilOrmRepository.MaterialKey {
	// Convert time.Time to Unix milliseconds for database storage
	generateDateMillis := materialKeyGenerateDate.UnixMilli()

	return &cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  *elasticKeyID,
		MaterialKeyID:                 *materialKeyID,
		MaterialKeyClearPublic:        materialKeyClearPublicJWKBytes,        // nil if repositoryElasticKey.ElasticKeyAlgorithm is Symmetric
		MaterialKeyEncryptedNonPublic: materialKeyEncryptedNonPublicJWKBytes, // nil if repositoryElasticKey.ElasticKeyImportAllowed=true
		MaterialKeyGenerateDate:       &generateDateMillis,                   // nil if repositoryElasticKey.ElasticKeyImportAllowed=true
	}
}

// toElasticKeyStatusFromImportAllowed returns the initial status based on import allowed flag.
func toElasticKeyStatusFromImportAllowed(isImportAllowed bool) cryptoutilKmsServer.ElasticKeyStatus {
	if isImportAllowed {
		return cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)
	}

	return cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate)
}

func toOamElasticKeyStatus(isImportAllowed *bool) *cryptoutilKmsServer.ElasticKeyStatus {
	var ormElasticKeyStatus cryptoutilKmsServer.ElasticKeyStatus
	if *isImportAllowed {
		ormElasticKeyStatus = cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)
	} else {
		ormElasticKeyStatus = cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate)
	}

	return &ormElasticKeyStatus
}

// orm => oam

func (m *OamOrmMapper) toOamElasticKeys(ormElasticKeys []cryptoutilOrmRepository.ElasticKey) []cryptoutilKmsServer.ElasticKey {
	oamElasticKeys := make([]cryptoutilKmsServer.ElasticKey, len(ormElasticKeys))
	for i, ormElasticKey := range ormElasticKeys {
		oamElasticKeys[i] = *m.toOamElasticKey(&ormElasticKey)
	}

	return oamElasticKeys
}

func (m *OamOrmMapper) toOamElasticKey(ormElasticKey *cryptoutilOrmRepository.ElasticKey) *cryptoutilKmsServer.ElasticKey {
	algorithm := string(ormElasticKey.ElasticKeyAlgorithm)
	provider := string(ormElasticKey.ElasticKeyProvider)

	return &cryptoutilKmsServer.ElasticKey{
		ElasticKeyID:      &ormElasticKey.ElasticKeyID,
		Name:              &ormElasticKey.ElasticKeyName,
		Description:       &ormElasticKey.ElasticKeyDescription,
		Algorithm:         &algorithm,
		Provider:          &provider,
		VersioningAllowed: &ormElasticKey.ElasticKeyVersioningAllowed,
		ImportAllowed:     &ormElasticKey.ElasticKeyImportAllowed,
		Status:            &ormElasticKey.ElasticKeyStatus,
	}
}

func (m *OamOrmMapper) toOamMaterialKeys(ormMaterialKeys []cryptoutilOrmRepository.MaterialKey) ([]cryptoutilKmsServer.MaterialKey, error) {
	oamMaterialKeys := make([]cryptoutilKmsServer.MaterialKey, len(ormMaterialKeys))

	var oamMaterialKey *cryptoutilKmsServer.MaterialKey

	var err error
	for i, ormMaterialKey := range ormMaterialKeys {
		oamMaterialKey, err = m.toOamMaterialKey(&ormMaterialKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get oam key: %w", err)
		}

		oamMaterialKeys[i] = *oamMaterialKey
	}

	return oamMaterialKeys, nil
}

func (m *OamOrmMapper) toOamMaterialKey(ormMaterialKey *cryptoutilOrmRepository.MaterialKey) (*cryptoutilKmsServer.MaterialKey, error) {
	if ormMaterialKey == nil {
		return nil, fmt.Errorf("material key cannot be nil")
	} else if ormMaterialKey.ElasticKeyID == (googleUuid.UUID{}) {
		return nil, fmt.Errorf("material key missing required elastic key ID")
	} else if ormMaterialKey.MaterialKeyID == (googleUuid.UUID{}) {
		return nil, fmt.Errorf("material key missing required material key ID")
	}

	var materialKeyClearPublic *string

	if ormMaterialKey.MaterialKeyClearPublic != nil {
		tmp := string(ormMaterialKey.MaterialKeyClearPublic)
		materialKeyClearPublic = &tmp
	}

	// Convert Unix milliseconds timestamps to time.Time for OpenAPI model
	var generateDate, importDate, expirationDate, revocationDate *time.Time

	if ormMaterialKey.MaterialKeyGenerateDate != nil {
		t := time.UnixMilli(*ormMaterialKey.MaterialKeyGenerateDate).UTC()
		generateDate = &t
	}

	if ormMaterialKey.MaterialKeyImportDate != nil {
		t := time.UnixMilli(*ormMaterialKey.MaterialKeyImportDate).UTC()
		importDate = &t
	}

	if ormMaterialKey.MaterialKeyExpirationDate != nil {
		t := time.UnixMilli(*ormMaterialKey.MaterialKeyExpirationDate).UTC()
		expirationDate = &t
	}

	if ormMaterialKey.MaterialKeyRevocationDate != nil {
		t := time.UnixMilli(*ormMaterialKey.MaterialKeyRevocationDate).UTC()
		revocationDate = &t
	}

	elasticKeyID := openapi_types.UUID(ormMaterialKey.ElasticKeyID)
	materialKeyID := openapi_types.UUID(ormMaterialKey.MaterialKeyID)

	return &cryptoutilKmsServer.MaterialKey{
		ElasticKeyID:   &elasticKeyID,
		MaterialKeyID:  &materialKeyID,
		GenerateDate:   generateDate,
		ImportDate:     importDate,
		ExpirationDate: expirationDate,
		RevocationDate: revocationDate,
		ClearPublic:    materialKeyClearPublic,
	}, nil
}

// Helper methods
