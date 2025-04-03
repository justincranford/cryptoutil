package servicelogic

import (
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilUtil "cryptoutil/internal/util"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type serviceOrmMapper struct{}

func NewMapper() *serviceOrmMapper {
	return &serviceOrmMapper{}
}

// service => orm

func (m *serviceOrmMapper) toOrmAddKeyPool(keyPoolID uuid.UUID, serviceKeyPoolCreate *cryptoutilServiceModel.KeyPoolCreate) *cryptoutilOrmRepository.KeyPool {
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

func (m *serviceOrmMapper) toOrmKeyPoolProvider(serviceKeyPoolProvider *cryptoutilServiceModel.KeyPoolProvider) *cryptoutilOrmRepository.KeyPoolProvider {
	ormKeyPoolProvider := cryptoutilOrmRepository.KeyPoolProvider(*serviceKeyPoolProvider)
	return &ormKeyPoolProvider
}

func (m *serviceOrmMapper) toOrmKeyPoolAlgorithm(serviceKeyPoolProvider *cryptoutilServiceModel.KeyPoolAlgorithm) *cryptoutilOrmRepository.KeyPoolAlgorithm {
	ormKeyPoolAlgorithm := cryptoutilOrmRepository.KeyPoolAlgorithm(*serviceKeyPoolProvider)
	return &ormKeyPoolAlgorithm
}

func (m *serviceOrmMapper) toKeyPoolInitialStatus(serviceKeyPoolImportAllowed *cryptoutilServiceModel.KeyPoolImportAllowed) *cryptoutilOrmRepository.KeyPoolStatus {
	var ormKeyPoolStatus cryptoutilOrmRepository.KeyPoolStatus
	if *serviceKeyPoolImportAllowed {
		ormKeyPoolStatus = cryptoutilOrmRepository.KeyPoolStatus("pending_import")
	} else {
		ormKeyPoolStatus = cryptoutilOrmRepository.KeyPoolStatus("pending_generate")
	}
	return &ormKeyPoolStatus
}

// orm => service

func (m *serviceOrmMapper) toServiceKeyPools(ormKeyPools []cryptoutilOrmRepository.KeyPool) []cryptoutilServiceModel.KeyPool {
	serviceKeyPools := make([]cryptoutilServiceModel.KeyPool, len(ormKeyPools))
	for i, ormKeyPool := range ormKeyPools {
		serviceKeyPools[i] = *m.toServiceKeyPool(&ormKeyPool)
	}
	return serviceKeyPools
}

func (s *serviceOrmMapper) toServiceKeyPool(ormKeyPool *cryptoutilOrmRepository.KeyPool) *cryptoutilServiceModel.KeyPool {
	return &cryptoutilServiceModel.KeyPool{
		Id:                (*cryptoutilServiceModel.KeyPoolId)(&ormKeyPool.KeyPoolID),
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

func (m *serviceOrmMapper) toServiceKeyPoolAlgorithm(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) *cryptoutilServiceModel.KeyPoolAlgorithm {
	serviceKeyPoolAlgorithm := cryptoutilServiceModel.KeyPoolAlgorithm(*ormKeyPoolAlgorithm)
	return &serviceKeyPoolAlgorithm
}

func (m *serviceOrmMapper) toServiceKeyPoolProvider(ormKeyPoolProvider *cryptoutilOrmRepository.KeyPoolProvider) *cryptoutilServiceModel.KeyPoolProvider {
	serviceKeyPoolProvider := cryptoutilServiceModel.KeyPoolProvider(*ormKeyPoolProvider)
	return &serviceKeyPoolProvider
}

func (m *serviceOrmMapper) toServiceKeyPoolStatus(ormKeyPoolStatus *cryptoutilOrmRepository.KeyPoolStatus) *cryptoutilServiceModel.KeyPoolStatus {
	serviceKeyPoolStatus := cryptoutilServiceModel.KeyPoolStatus(*ormKeyPoolStatus)
	return &serviceKeyPoolStatus
}

func (m *serviceOrmMapper) toServiceKeys(ormKeys []cryptoutilOrmRepository.Key) []cryptoutilServiceModel.Key {
	serviceKeys := make([]cryptoutilServiceModel.Key, len(ormKeys))
	for i, ormKey := range ormKeys {
		serviceKeys[i] = *m.toServiceKey(&ormKey)
	}
	return serviceKeys
}

func (m *serviceOrmMapper) toServiceKey(ormKey *cryptoutilOrmRepository.Key) *cryptoutilServiceModel.Key {
	return &cryptoutilServiceModel.Key{
		Pool:         (*cryptoutilServiceModel.KeyPoolId)(&ormKey.KeyPoolID),
		Id:           &ormKey.KeyID,
		GenerateDate: (*cryptoutilServiceModel.KeyGenerateDate)(ormKey.KeyGenerateDate),
	}
}

func (m *serviceOrmMapper) toOrmGetKeyPoolsQueryParams(params *cryptoutilServiceModel.KeyPoolsQueryParams) (*cryptoutilOrmRepository.GetKeyPoolsFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyPoolIDs, err := m.toOrmUUIDs(params.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid Key Pool ID: %w", err))
	}
	names, err := m.toOrmStrings(params.Name)
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

func (m *serviceOrmMapper) toOrmGetKeyPoolKeysQueryParams(params *cryptoutilServiceModel.KeyPoolKeysQueryParams) (*cryptoutilOrmRepository.GetKeyPoolKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyIDs, err := m.toOrmUUIDs(params.Id)
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

func (m *serviceOrmMapper) toOrmGetKeysQueryParams(params *cryptoutilServiceModel.KeysQueryParams) (*cryptoutilOrmRepository.GetKeysFilters, error) {
	if params == nil {
		return nil, nil
	}
	var errs []error
	keyPoolIDs, err := m.toOrmUUIDs(params.Pool)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid KeyPoolID: %w", err))
	}
	keyIDs, err := m.toOrmUUIDs(params.Id)
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

func (*serviceOrmMapper) toOrmUUIDs(uuids *[]uuid.UUID) ([]uuid.UUID, error) {
	if uuids == nil || len(*uuids) == 0 {
		return nil, nil
	}
	for _, uuid := range *uuids {
		if uuid == cryptoutilUtil.ZeroUUID {
			return nil, fmt.Errorf("UUID must not be 00000000-0000-0000-0000-000000000000")
		}
	}
	return *uuids, nil
}

func (*serviceOrmMapper) toOrmStrings(strings *[]string) ([]string, error) {
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
			if maxDate.Compare(now) > 0 {
				errs = append(errs, fmt.Errorf("Max Date can't be in the future"))
			}
			if nonNullMinDate && minDate.Compare(*maxDate) > 0 {
				errs = append(errs, fmt.Errorf("Min Date must be before Max Date"))
			}
		}
	}
	return minDate, maxDate, errors.Join(errs...)
}

func (m *serviceOrmMapper) toOrmAlgorithms(algorithms *[]cryptoutilServiceModel.KeyPoolAlgorithm) ([]string, error) {
	newVar := toStrings(algorithms, func(algorithm cryptoutilServiceModel.KeyPoolAlgorithm) string { return string(algorithm) })
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmKeyPoolSorts(keyPoolSorts *[]cryptoutilServiceModel.KeyPoolSort) ([]string, error) {
	newVar := toStrings(keyPoolSorts, func(keyPoolSort cryptoutilServiceModel.KeyPoolSort) string { return string(keyPoolSort) })
	return newVar, nil
}

func (m *serviceOrmMapper) toOrmKeySorts(keySorts *[]cryptoutilServiceModel.KeySort) ([]string, error) {
	newVar := toStrings(keySorts, func(keySort cryptoutilServiceModel.KeySort) string { return string(keySort) })
	return newVar, nil
}

func (*serviceOrmMapper) toOrmPageNumber(pageNumber *cryptoutilServiceModel.PageNumber) (int, error) {
	if pageNumber == nil {
		return 0, nil
	} else if *pageNumber >= 0 {
		return *pageNumber, nil
	}
	return 0, fmt.Errorf("Page Number must be zero or higher")
}

func (*serviceOrmMapper) toOrmPageSize(pageSize *cryptoutilServiceModel.PageSize) (int, error) {
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
