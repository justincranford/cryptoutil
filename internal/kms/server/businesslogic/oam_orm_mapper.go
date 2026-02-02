// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"errors"
	"fmt"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

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

func (m *OamOrmMapper) toOrmAddElasticKey(elasticKeyID *googleUuid.UUID, oamElasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate) *cryptoutilOrmRepository.ElasticKey {
	// Apply defaults for optional fields
	provider := defaultElasticKeyProvider
	if oamElasticKeyCreate.Provider != nil {
		provider = *oamElasticKeyCreate.Provider
	}

	algorithm := defaultElasticKeyAlgorithm
	if oamElasticKeyCreate.Algorithm != nil {
		algorithm = *oamElasticKeyCreate.Algorithm
	}

	versioningAllowed := defaultElasticKeyVersioningAllowed
	if oamElasticKeyCreate.VersioningAllowed != nil {
		versioningAllowed = *oamElasticKeyCreate.VersioningAllowed
	}

	importAllowed := defaultElasticKeyImportAllowed
	if oamElasticKeyCreate.ImportAllowed != nil {
		importAllowed = *oamElasticKeyCreate.ImportAllowed
	}

	return &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                *elasticKeyID,
		ElasticKeyName:              oamElasticKeyCreate.Name,
		ElasticKeyDescription:       oamElasticKeyCreate.Description,
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
func toElasticKeyStatusFromImportAllowed(isImportAllowed bool) cryptoutilOpenapiModel.ElasticKeyStatus {
	if isImportAllowed {
		return cryptoutilOpenapiModel.PendingImport
	}

	return cryptoutilOpenapiModel.PendingGenerate
}

func toOamElasticKeyStatus(isImportAllowed *bool) *cryptoutilOpenapiModel.ElasticKeyStatus {
	var ormElasticKeyStatus cryptoutilOpenapiModel.ElasticKeyStatus
	if *isImportAllowed {
		ormElasticKeyStatus = cryptoutilOpenapiModel.PendingImport
	} else {
		ormElasticKeyStatus = cryptoutilOpenapiModel.PendingGenerate
	}

	return &ormElasticKeyStatus
}

// orm => oam

func (m *OamOrmMapper) toOamElasticKeys(ormElasticKeys []cryptoutilOrmRepository.ElasticKey) []cryptoutilOpenapiModel.ElasticKey {
	oamElasticKeys := make([]cryptoutilOpenapiModel.ElasticKey, len(ormElasticKeys))
	for i, ormElasticKey := range ormElasticKeys {
		oamElasticKeys[i] = *m.toOamElasticKey(&ormElasticKey)
	}

	return oamElasticKeys
}

func (m *OamOrmMapper) toOamElasticKey(ormElasticKey *cryptoutilOrmRepository.ElasticKey) *cryptoutilOpenapiModel.ElasticKey {
	return &cryptoutilOpenapiModel.ElasticKey{
		ElasticKeyID:      &ormElasticKey.ElasticKeyID,
		Name:              &ormElasticKey.ElasticKeyName,
		Description:       &ormElasticKey.ElasticKeyDescription,
		Algorithm:         &ormElasticKey.ElasticKeyAlgorithm,
		Provider:          &ormElasticKey.ElasticKeyProvider,
		VersioningAllowed: &ormElasticKey.ElasticKeyVersioningAllowed,
		ImportAllowed:     &ormElasticKey.ElasticKeyImportAllowed,
		Status:            &ormElasticKey.ElasticKeyStatus,
	}
}

func (m *OamOrmMapper) toOamMaterialKeys(ormMaterialKeys []cryptoutilOrmRepository.MaterialKey) ([]cryptoutilOpenapiModel.MaterialKey, error) {
	oamMaterialKeys := make([]cryptoutilOpenapiModel.MaterialKey, len(ormMaterialKeys))

	var oamMaterialKey *cryptoutilOpenapiModel.MaterialKey

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

func (m *OamOrmMapper) toOamMaterialKey(ormMaterialKey *cryptoutilOrmRepository.MaterialKey) (*cryptoutilOpenapiModel.MaterialKey, error) {
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

	return &cryptoutilOpenapiModel.MaterialKey{
		ElasticKeyID:   ormMaterialKey.ElasticKeyID,
		MaterialKeyID:  ormMaterialKey.MaterialKeyID,
		GenerateDate:   generateDate,
		ImportDate:     importDate,
		ExpirationDate: expirationDate,
		RevocationDate: revocationDate,
		ClearPublic:    materialKeyClearPublic,
	}, nil
}

// Helper methods

func (m *OamOrmMapper) toOrmGetElasticKeysQueryParams(params *cryptoutilOpenapiModel.ElasticKeysQueryParams) (*cryptoutilOrmRepository.GetElasticKeysFilters, error) {
	if params == nil {
		return &cryptoutilOrmRepository.GetElasticKeysFilters{}, nil
	}

	var errs []error

	elasticKeyIDs, err := m.toOptionalOrmUUIDs(params.ElasticKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key ID: %w", err))
	}

	names, err := m.toOptionalOrmStrings(params.Name)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key Name: %w", err))
	}

	algorithms, err := m.toOrmAlgorithms(params.Algorithm)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key Algorithm: %w", err))
	}

	sorts, err := m.toOrmElasticKeySorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Elastic Key Sort: %w", err))
	}

	pageNumber, err := m.toOrmPageNumber(params.Page)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Number: %w", err))
	}

	pageSize, err := m.toOrmPageSize(params.Size)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Size: %w", err))
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid Get Elastic Keys parameters: %w", errors.Join(errs...))
	}

	return &cryptoutilOrmRepository.GetElasticKeysFilters{
		ElasticKeyID:      elasticKeyIDs,
		Name:              names,
		Algorithm:         algorithms,
		VersioningAllowed: params.VersioningAllowed,
		ImportAllowed:     params.ImportAllowed,
		Sort:              sorts,
		PageNumber:        pageNumber,
		PageSize:          pageSize,
	}, nil
}

func (m *OamOrmMapper) toOrmGetMaterialKeysForElasticKeyQueryParams(params *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams) (*cryptoutilOrmRepository.GetElasticKeyMaterialKeysFilters, error) {
	if params == nil {
		return &cryptoutilOrmRepository.GetElasticKeyMaterialKeysFilters{}, nil
	}

	var errs []error

	materialKeyIDs, err := m.toOptionalOrmUUIDs(params.MaterialKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid MaterialKeyID: %w", err))
	}

	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}

	sorts, err := m.toOrmMaterialKeySorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Sort: %w", err))
	}

	pageNumber, err := m.toOrmPageNumber(params.Page)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Number: %w", err))
	}

	pageSize, err := m.toOrmPageSize(params.Size)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Size: %w", err))
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid Get Elastic Key Keys parameters: %w", errors.Join(errs...))
	}

	return &cryptoutilOrmRepository.GetElasticKeyMaterialKeysFilters{
		ElasticKeyID:        materialKeyIDs,
		MinimumGenerateDate: minGenerateDate,
		MaximumGenerateDate: maxGenerateDate,
		Sort:                sorts,
		PageNumber:          pageNumber,
		PageSize:            pageSize,
	}, nil
}

func (m *OamOrmMapper) toOrmGetMaterialKeysQueryParams(params *cryptoutilOpenapiModel.MaterialKeysQueryParams) (*cryptoutilOrmRepository.GetMaterialKeysFilters, error) {
	if params == nil {
		return &cryptoutilOrmRepository.GetMaterialKeysFilters{}, nil
	}

	var errs []error

	elasticKeyIDs, err := m.toOptionalOrmUUIDs(params.ElasticKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid ElasticKeyID: %w", err))
	}

	materialKeyIDs, err := m.toOptionalOrmUUIDs(params.MaterialKeyID)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid MaterialKeyID: %w", err))
	}

	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}

	sorts, err := m.toOrmMaterialKeySorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Sort: %w", err))
	}

	pageNumber, err := m.toOrmPageNumber(params.Page)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Number: %w", err))
	}

	pageSize, err := m.toOrmPageSize(params.Size)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Page Size: %w", err))
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid Get Keys parameters: %w", errors.Join(errs...))
	}

	return &cryptoutilOrmRepository.GetMaterialKeysFilters{
		ElasticKeyID:        elasticKeyIDs,
		MaterialKeyID:       materialKeyIDs,
		MinimumGenerateDate: minGenerateDate,
		MaximumGenerateDate: maxGenerateDate,
		Sort:                sorts,
		PageNumber:          pageNumber,
		PageSize:            pageSize,
	}, nil
}

func (*OamOrmMapper) toOptionalOrmUUIDs(uuids *[]googleUuid.UUID) ([]googleUuid.UUID, error) {
	if uuids == nil || len(*uuids) == 0 {
		return nil, nil
	}

	if err := cryptoutilSharedUtilRandom.ValidateUUIDs(*uuids, &ErrInvalidUUID); err != nil {
		return nil, fmt.Errorf("failed to validate UUIDs: %w", err)
	}

	return *uuids, nil
}

func (*OamOrmMapper) toOptionalOrmStrings(strings *[]string) ([]string, error) {
	if strings == nil || len(*strings) == 0 {
		return nil, nil
	}

	for _, value := range *strings {
		if len(value) == 0 {
			return nil, fmt.Errorf("value must not be empty string")
		}
	}

	return *strings, nil
}

func (*OamOrmMapper) toOrmDateRange(minDate, maxDate *time.Time) (*time.Time, *time.Time, error) {
	var errs []error

	nonNullMinDate := minDate != nil
	nonNullMaxDate := maxDate != nil

	if nonNullMinDate || nonNullMaxDate {
		now := time.Now().UTC()
		if nonNullMinDate && minDate.Compare(now) > 0 {
			errs = append(errs, fmt.Errorf("min date can't be in the future"))
		}

		if nonNullMaxDate {
			// if maxDate.Compare(now) > 0 {
			// 	errs = append(errs, fmt.Errorf("max Date can't be in the future"))
			// }
			if nonNullMinDate && minDate.Compare(*maxDate) > 0 {
				errs = append(errs, fmt.Errorf("min date must be before max date"))
			}
		}
	}

	if len(errs) > 0 {
		return minDate, maxDate, fmt.Errorf("invalid date range: %w", errors.Join(errs...))
	}

	return minDate, maxDate, nil
}

func (m *OamOrmMapper) toOrmAlgorithms(algorithms *[]cryptoutilOpenapiModel.ElasticKeyAlgorithm) ([]string, error) {
	if algorithms != nil {
		// Validate algorithm values
		for _, algorithm := range *algorithms {
			if string(algorithm) == "" {
				return nil, fmt.Errorf("algorithm cannot be empty")
			}
		}
	}

	return toStrings(algorithms, func(algorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm) string {
		return string(algorithm)
	}), nil
}

func (m *OamOrmMapper) toOrmElasticKeySorts(elasticMaterialKeySorts *[]cryptoutilOpenapiModel.ElasticKeySort) ([]string, error) {
	if elasticMaterialKeySorts != nil {
		// Validate sort values
		for _, sort := range *elasticMaterialKeySorts {
			if string(sort) == "" {
				return nil, fmt.Errorf("elastic key sort cannot be empty")
			}
		}
	}

	return toStrings(elasticMaterialKeySorts, func(elasticMaterialKeySort cryptoutilOpenapiModel.ElasticKeySort) string {
		return string(elasticMaterialKeySort)
	}), nil
}

func (m *OamOrmMapper) toOrmMaterialKeySorts(keySorts *[]cryptoutilOpenapiModel.MaterialKeySort) ([]string, error) {
	if keySorts != nil {
		// Validate sort values against allowed enum values
		allowedSorts := map[string]bool{
			"elastic_key_id":       true,
			"elastic_key_id:ASC":   true,
			"elastic_key_id:DESC":  true,
			"material_key_id":      true,
			"material_key_id:ASC":  true,
			"material_key_id:DESC": true,
			"generate_date":        true,
			"generate_date:ASC":    true,
			"generate_date:DESC":   true,
			"import_date":          true,
			"import_date:ASC":      true,
			"import_date:DESC":     true,
			"expiration_date":      true,
			"expiration_date:ASC":  true,
			"expiration_date:DESC": true,
			"revocation_date":      true,
			"revocation_date:ASC":  true,
			"revocation_date:DESC": true,
		}

		for _, keySort := range *keySorts {
			sortStr := string(keySort)
			if sortStr == "" {
				return nil, fmt.Errorf("material key sort cannot be empty")
			}

			if !allowedSorts[sortStr] {
				return nil, fmt.Errorf("invalid material key sort value: %s", sortStr)
			}
		}
	}

	return toStrings(keySorts, func(keySort cryptoutilOpenapiModel.MaterialKeySort) string {
		return string(keySort)
	}), nil
}

func (*OamOrmMapper) toOrmPageNumber(pageNumber *cryptoutilOpenapiModel.PageNumber) (int, error) {
	if pageNumber == nil {
		return 0, nil
	} else if *pageNumber >= 0 {
		return *pageNumber, nil
	}

	return 0, fmt.Errorf("page number must be zero or higher")
}

func (*OamOrmMapper) toOrmPageSize(pageSize *cryptoutilOpenapiModel.PageSize) (int, error) {
	if pageSize == nil {
		return cryptoutilSharedMagic.DefaultPageSize, nil
	} else if *pageSize >= 1 {
		return *pageSize, nil
	}

	return 0, fmt.Errorf("page size must be one or higher")
}

func toStrings[T any](items *[]T, toString func(T) string) []string {
	if items == nil || len(*items) == 0 {
		return nil
	}

	converted := make([]string, 0, len(*items))
	for _, item := range *items {
		converted = append(converted, toString(item))
	}

	return converted
}
