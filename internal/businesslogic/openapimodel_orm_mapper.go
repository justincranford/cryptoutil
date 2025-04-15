package businesslogic

import (
	"errors"
	"fmt"
	"time"

	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilUtil "cryptoutil/internal/util"

	googleUuid "github.com/google/uuid"
)

type serviceOrmMapper struct{}

func NewMapper() *serviceOrmMapper {
	return &serviceOrmMapper{}
}

// service => orm

func (m *serviceOrmMapper) toOrmAddKeyPool(keyPoolID googleUuid.UUID, serviceKeyPoolCreate *cryptoutilBusinessLogicModel.KeyPoolCreate) *cryptoutilOrmRepository.KeyPool {
	return &cryptoutilOrmRepository.KeyPool{
		KeyPoolID:                keyPoolID,
		KeyPoolName:              serviceKeyPoolCreate.Name,
		KeyPoolDescription:       serviceKeyPoolCreate.Description,
		KeyPoolProvider:          *m.toOrmKeyPoolProvider(serviceKeyPoolCreate.Provider),
		KeyPoolAlgorithm:         *m.toOrmKeyPoolAlgorithm(serviceKeyPoolCreate.Algorithm),
		KeyPoolVersioningAllowed: *serviceKeyPoolCreate.VersioningAllowed,
		KeyPoolImportAllowed:     *serviceKeyPoolCreate.ImportAllowed,
		KeyPoolExportAllowed:     *serviceKeyPoolCreate.ExportAllowed,
		KeyPoolStatus:            *m.toKeyPoolInitialStatus(serviceKeyPoolCreate.ImportAllowed),
	}
}

func (m *serviceOrmMapper) toOrmKeyPoolProvider(serviceKeyPoolProvider *cryptoutilBusinessLogicModel.KeyPoolProvider) *cryptoutilOrmRepository.KeyPoolProvider {
	ormKeyPoolProvider := cryptoutilOrmRepository.KeyPoolProvider(*serviceKeyPoolProvider)
	return &ormKeyPoolProvider
}

func (m *serviceOrmMapper) toOrmKeyPoolAlgorithm(serviceKeyPoolProvider *cryptoutilBusinessLogicModel.KeyPoolAlgorithm) *cryptoutilOrmRepository.KeyPoolAlgorithm {
	ormKeyPoolAlgorithm := cryptoutilOrmRepository.KeyPoolAlgorithm(*serviceKeyPoolProvider)
	return &ormKeyPoolAlgorithm
}

func (m *serviceOrmMapper) toKeyPoolInitialStatus(serviceKeyPoolImportAllowed *cryptoutilBusinessLogicModel.KeyPoolImportAllowed) *cryptoutilOrmRepository.KeyPoolStatus {
	var ormKeyPoolStatus cryptoutilOrmRepository.KeyPoolStatus
	if *serviceKeyPoolImportAllowed {
		ormKeyPoolStatus = cryptoutilOrmRepository.KeyPoolStatus("pending_import")
	} else {
		ormKeyPoolStatus = cryptoutilOrmRepository.KeyPoolStatus("pending_generate")
	}
	return &ormKeyPoolStatus
}

// orm => service

func (m *serviceOrmMapper) toServiceKeyPools(ormKeyPools []cryptoutilOrmRepository.KeyPool) []cryptoutilBusinessLogicModel.KeyPool {
	serviceKeyPools := make([]cryptoutilBusinessLogicModel.KeyPool, len(ormKeyPools))
	for i, ormKeyPool := range ormKeyPools {
		serviceKeyPools[i] = *m.toServiceKeyPool(&ormKeyPool)
	}
	return serviceKeyPools
}

func (s *serviceOrmMapper) toServiceKeyPool(ormKeyPool *cryptoutilOrmRepository.KeyPool) *cryptoutilBusinessLogicModel.KeyPool {
	return &cryptoutilBusinessLogicModel.KeyPool{
		Id:                (*cryptoutilBusinessLogicModel.KeyPoolId)(&ormKeyPool.KeyPoolID),
		Name:              &ormKeyPool.KeyPoolName,
		Description:       &ormKeyPool.KeyPoolDescription,
		Algorithm:         s.toServiceKeyPoolAlgorithm(&ormKeyPool.KeyPoolAlgorithm),
		Provider:          s.toServiceKeyPoolProvider(&ormKeyPool.KeyPoolProvider),
		VersioningAllowed: &ormKeyPool.KeyPoolVersioningAllowed,
		ImportAllowed:     &ormKeyPool.KeyPoolImportAllowed,
		ExportAllowed:     &ormKeyPool.KeyPoolExportAllowed,
		Status:            s.toServiceKeyPoolStatus(&ormKeyPool.KeyPoolStatus),
	}
}

func (m *serviceOrmMapper) toServiceKeyPoolAlgorithm(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) *cryptoutilBusinessLogicModel.KeyPoolAlgorithm {
	serviceKeyPoolAlgorithm := cryptoutilBusinessLogicModel.KeyPoolAlgorithm(*ormKeyPoolAlgorithm)
	return &serviceKeyPoolAlgorithm
}

func (m *serviceOrmMapper) toServiceKeyPoolProvider(ormKeyPoolProvider *cryptoutilOrmRepository.KeyPoolProvider) *cryptoutilBusinessLogicModel.KeyPoolProvider {
	serviceKeyPoolProvider := cryptoutilBusinessLogicModel.KeyPoolProvider(*ormKeyPoolProvider)
	return &serviceKeyPoolProvider
}

func (m *serviceOrmMapper) toServiceKeyPoolStatus(ormKeyPoolStatus *cryptoutilOrmRepository.KeyPoolStatus) *cryptoutilBusinessLogicModel.KeyPoolStatus {
	serviceKeyPoolStatus := cryptoutilBusinessLogicModel.KeyPoolStatus(*ormKeyPoolStatus)
	return &serviceKeyPoolStatus
}

func (m *serviceOrmMapper) toServiceKeys(ormKeys []cryptoutilOrmRepository.Key) []cryptoutilBusinessLogicModel.Key {
	serviceKeys := make([]cryptoutilBusinessLogicModel.Key, len(ormKeys))
	for i, ormKey := range ormKeys {
		serviceKeys[i] = *m.toServiceKey(&ormKey)
	}
	return serviceKeys
}

func (m *serviceOrmMapper) toServiceKey(ormKey *cryptoutilOrmRepository.Key) *cryptoutilBusinessLogicModel.Key {
	return &cryptoutilBusinessLogicModel.Key{
		Pool:         (*cryptoutilBusinessLogicModel.KeyPoolId)(&ormKey.KeyPoolID),
		Id:           &ormKey.KeyID,
		GenerateDate: (*cryptoutilBusinessLogicModel.KeyGenerateDate)(ormKey.KeyGenerateDate),
	}
}

func (m *serviceOrmMapper) toOrmGetKeyPoolsQueryParams(params *cryptoutilBusinessLogicModel.KeyPoolsQueryParams) (*cryptoutilOrmRepository.GetKeyPoolsFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyPoolIDs, err := m.toOptionalOrmUUIDs(params.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool ID: %w", err))
	}
	names, err := m.toOptionalOrmStrings(params.Name)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool Name: %w", err))
	}
	algorithms, err := m.toOrmAlgorithms(params.Algorithm)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool Algorithm: %w", err))
	}
	sorts, err := m.toOrmKeyPoolSorts(params.Sort)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool Sort: %w", err))
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
		return nil, fmt.Errorf("invalid Get Key Pools parameters: %w", errors.Join(errs...))
	}

	return &cryptoutilOrmRepository.GetKeyPoolsFilters{
		ID:                keyPoolIDs,
		Name:              names,
		Algorithm:         algorithms,
		VersioningAllowed: params.VersioningAllowed,
		ImportAllowed:     params.ImportAllowed,
		ExportAllowed:     params.ExportAllowed,
		Sort:              sorts,
		PageNumber:        pageNumber,
		PageSize:          pageSize,
	}, nil
}

func (m *serviceOrmMapper) toOrmGetKeyPoolKeysQueryParams(params *cryptoutilBusinessLogicModel.KeyPoolKeysQueryParams) (*cryptoutilOrmRepository.GetKeyPoolKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyIDs, err := m.toOptionalOrmUUIDs(params.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid KeyID: %w", err))
	}
	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}
	sorts, err := m.toOrmKeySorts(params.Sort)
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
		return nil, fmt.Errorf("invalid Get Key Pool Keys parameters: %w", errors.Join(errs...))
	}
	return &cryptoutilOrmRepository.GetKeyPoolKeysFilters{
		ID:                  keyIDs,
		MinimumGenerateDate: minGenerateDate,
		MaximumGenerateDate: maxGenerateDate,
		Sort:                sorts,
		PageNumber:          pageNumber,
		PageSize:            pageSize,
	}, nil
}

func (m *serviceOrmMapper) toOrmGetKeysQueryParams(params *cryptoutilBusinessLogicModel.KeysQueryParams) (*cryptoutilOrmRepository.GetKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyPoolIDs, err := m.toOptionalOrmUUIDs(params.Pool)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid KeyPoolID: %w", err))
	}
	keyIDs, err := m.toOptionalOrmUUIDs(params.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid KeyID: %w", err))
	}
	minGenerateDate, maxGenerateDate, err := m.toOrmDateRange(params.MinGenerateDate, params.MaxGenerateDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Generate Date range: %w", err))
	}
	sorts, err := m.toOrmKeySorts(params.Sort)
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

	return &cryptoutilOrmRepository.GetKeysFilters{
		Pool:                keyPoolIDs,
		ID:                  keyIDs,
		MinimumGenerateDate: minGenerateDate,
		MaximumGenerateDate: maxGenerateDate,
		Sort:                sorts,
		PageNumber:          pageNumber,
		PageSize:            pageSize,
	}, nil
}

// Helper methods

func (*serviceOrmMapper) toOptionalOrmUUIDs(uuids *[]googleUuid.UUID) ([]googleUuid.UUID, error) {
	if uuids == nil || len(*uuids) == 0 {
		return nil, nil
	}
	if err := cryptoutilUtil.ValidateUUIDs(*uuids, "invalid UUIDs"); err != nil {
		return nil, err
	}
	return *uuids, nil
}

func (*serviceOrmMapper) toOptionalOrmStrings(strings *[]string) ([]string, error) {
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

func (*serviceOrmMapper) toOrmDateRange(minDate *time.Time, maxDate *time.Time) (*time.Time, *time.Time, error) {
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

func (m *serviceOrmMapper) toOrmAlgorithms(algorithms *[]cryptoutilBusinessLogicModel.KeyPoolAlgorithm) ([]string, error) {
	newVar := toStrings(algorithms, func(algorithm cryptoutilBusinessLogicModel.KeyPoolAlgorithm) string { return string(algorithm) })
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmKeyPoolSorts(keyPoolSorts *[]cryptoutilBusinessLogicModel.KeyPoolSort) ([]string, error) {
	newVar := toStrings(keyPoolSorts, func(keyPoolSort cryptoutilBusinessLogicModel.KeyPoolSort) string { return string(keyPoolSort) })
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmKeySorts(keySorts *[]cryptoutilBusinessLogicModel.KeySort) ([]string, error) {
	newVar := toStrings(keySorts, func(keySort cryptoutilBusinessLogicModel.KeySort) string { return string(keySort) })
	return newVar, nil
}

func (*serviceOrmMapper) toOrmPageNumber(pageNumber *cryptoutilBusinessLogicModel.PageNumber) (int, error) {
	if pageNumber == nil {
		return 0, nil
	} else if *pageNumber >= 0 {
		return *pageNumber, nil
	}
	return 0, fmt.Errorf("Page Number must be zero or higher")
}

func (*serviceOrmMapper) toOrmPageSize(pageSize *cryptoutilBusinessLogicModel.PageSize) (int, error) {
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
