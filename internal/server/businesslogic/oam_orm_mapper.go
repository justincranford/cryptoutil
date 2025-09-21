package businesslogic

import (
	"errors"
	"fmt"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
)

type oamOrmMapper struct{} // Mapper between OpenAPI Model models and ORM objects

func NewOamOrmMapper() *oamOrmMapper {
	return &oamOrmMapper{}
}

var ErrInvalidUuid = "invalid UUIDs"

// oam => orm

func (m *oamOrmMapper) toOrmAddElasticKey(elasticKeyID *googleUuid.UUID, oamElasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate) *cryptoutilOrmRepository.ElasticKey {
	return &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                *elasticKeyID,
		ElasticKeyName:              oamElasticKeyCreate.Name,
		ElasticKeyDescription:       oamElasticKeyCreate.Description,
		ElasticKeyProvider:          *oamElasticKeyCreate.Provider,
		ElasticKeyAlgorithm:         *oamElasticKeyCreate.Algorithm,
		ElasticKeyVersioningAllowed: *oamElasticKeyCreate.VersioningAllowed,
		ElasticKeyImportAllowed:     *oamElasticKeyCreate.ImportAllowed,
		ElasticKeyStatus:            *toOamElasticKeyStatus(oamElasticKeyCreate.ImportAllowed),
	}
}

func (*oamOrmMapper) toOrmAddMaterialKey(elasticKeyID *googleUuid.UUID, materialKeyID *googleUuid.UUID, materialKeyClearPublicJwkBytes []byte, materialKeyEncryptedNonPublicJwkBytes []byte, materialKeyGenerateDate time.Time) *cryptoutilOrmRepository.MaterialKey {
	return &cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  *elasticKeyID,
		MaterialKeyID:                 *materialKeyID,
		MaterialKeyClearPublic:        materialKeyClearPublicJwkBytes,        // nil if repositoryElasticKey.ElasticKeyAlgorithm is Symmetric
		MaterialKeyEncryptedNonPublic: materialKeyEncryptedNonPublicJwkBytes, // nil if repositoryElasticKey.ElasticKeyImportAllowed=true
		MaterialKeyGenerateDate:       &materialKeyGenerateDate,              // nil if repositoryElasticKey.ElasticKeyImportAllowed=true
	}
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

func (m *oamOrmMapper) toOamElasticKeys(ormElasticKeys []cryptoutilOrmRepository.ElasticKey) []cryptoutilOpenapiModel.ElasticKey {
	oamElasticKeys := make([]cryptoutilOpenapiModel.ElasticKey, len(ormElasticKeys))
	for i, ormElasticKey := range ormElasticKeys {
		oamElasticKeys[i] = *m.toOamElasticKey(&ormElasticKey)
	}
	return oamElasticKeys
}

func (s *oamOrmMapper) toOamElasticKey(ormElasticKey *cryptoutilOrmRepository.ElasticKey) *cryptoutilOpenapiModel.ElasticKey {
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

func (m *oamOrmMapper) toOamMaterialKeys(ormMaterialKeys []cryptoutilOrmRepository.MaterialKey) ([]cryptoutilOpenapiModel.MaterialKey, error) {
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

func (m *oamOrmMapper) toOamMaterialKey(ormMaterialKey *cryptoutilOrmRepository.MaterialKey) (*cryptoutilOpenapiModel.MaterialKey, error) {
	var materialKeyClearPublic *string
	if ormMaterialKey.MaterialKeyClearPublic != nil {
		tmp := string(ormMaterialKey.MaterialKeyClearPublic)
		materialKeyClearPublic = &tmp
	}
	return &cryptoutilOpenapiModel.MaterialKey{
		ElasticKeyID:   ormMaterialKey.ElasticKeyID,
		MaterialKeyID:  ormMaterialKey.MaterialKeyID,
		GenerateDate:   ormMaterialKey.MaterialKeyGenerateDate,
		ImportDate:     ormMaterialKey.MaterialKeyImportDate,
		ExpirationDate: ormMaterialKey.MaterialKeyExpirationDate,
		RevocationDate: ormMaterialKey.MaterialKeyRevocationDate,
		ClearPublic:    materialKeyClearPublic,
	}, nil
}

// Helper methods

func (m *oamOrmMapper) toOrmGetElasticKeysQueryParams(params *cryptoutilOpenapiModel.ElasticKeysQueryParams) (*cryptoutilOrmRepository.GetElasticKeysFilters, error) {
	if params == nil {
		return nil, nil
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

func (m *oamOrmMapper) toOrmGetMaterialKeysForElasticKeyQueryParams(params *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams) (*cryptoutilOrmRepository.GetElasticKeyMaterialKeysFilters, error) {
	if params == nil {
		return nil, nil
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

func (m *oamOrmMapper) toOrmGetMaterialKeysQueryParams(params *cryptoutilOpenapiModel.MaterialKeysQueryParams) (*cryptoutilOrmRepository.GetMaterialKeysFilters, error) {
	if params == nil {
		return nil, nil
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

func (*oamOrmMapper) toOptionalOrmUUIDs(uuids *[]googleUuid.UUID) ([]googleUuid.UUID, error) {
	if uuids == nil || len(*uuids) == 0 {
		return nil, nil
	}
	if err := cryptoutilUtil.ValidateUUIDs(*uuids, &ErrInvalidUuid); err != nil {
		return nil, err
	}
	return *uuids, nil
}

func (*oamOrmMapper) toOptionalOrmStrings(strings *[]string) ([]string, error) {
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

func (*oamOrmMapper) toOrmDateRange(minDate *time.Time, maxDate *time.Time) (*time.Time, *time.Time, error) {
	var errs []error
	nonNullMinDate := minDate != nil
	nonNullMaxDate := maxDate != nil
	if nonNullMinDate || nonNullMaxDate {
		now := time.Now().UTC()
		if nonNullMinDate && minDate.Compare(now) > 0 {
			errs = append(errs, fmt.Errorf("Min Date can't be in the future"))
		}
		if nonNullMaxDate {
			// if maxDate.Compare(now) > 0 {
			// 	errs = append(errs, fmt.Errorf("Max Date can't be in the future"))
			// }
			if nonNullMinDate && minDate.Compare(*maxDate) > 0 {
				errs = append(errs, fmt.Errorf("Min Date must be before Max Date"))
			}
		}
	}
	return minDate, maxDate, errors.Join(errs...)
}

func (m *oamOrmMapper) toOrmAlgorithms(algorithms *[]cryptoutilOpenapiModel.ElasticKeyAlgorithm) ([]string, error) {
	return toStrings(algorithms, func(algorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm) string {
		return string(algorithm)
	}), nil
}

func (m *oamOrmMapper) toOrmElasticKeySorts(elasticMaterialKeySorts *[]cryptoutilOpenapiModel.ElasticKeySort) ([]string, error) {
	return toStrings(elasticMaterialKeySorts, func(elasticMaterialKeySort cryptoutilOpenapiModel.ElasticKeySort) string {
		return string(elasticMaterialKeySort)
	}), nil
}

func (m *oamOrmMapper) toOrmMaterialKeySorts(keySorts *[]cryptoutilOpenapiModel.MaterialKeySort) ([]string, error) {
	return toStrings(keySorts, func(keySort cryptoutilOpenapiModel.MaterialKeySort) string {
		return string(keySort)
	}), nil
}

func (*oamOrmMapper) toOrmPageNumber(pageNumber *cryptoutilOpenapiModel.PageNumber) (int, error) {
	if pageNumber == nil {
		return 0, nil
	} else if *pageNumber >= 0 {
		return *pageNumber, nil
	}
	return 0, fmt.Errorf("Page Number must be zero or higher")
}

func (*oamOrmMapper) toOrmPageSize(pageSize *cryptoutilOpenapiModel.PageSize) (int, error) {
	if pageSize == nil {
		return 25, nil
	} else if *pageSize >= 1 {
		return *pageSize, nil
	}
	return 0, fmt.Errorf("Page Size must be one or higher")
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
