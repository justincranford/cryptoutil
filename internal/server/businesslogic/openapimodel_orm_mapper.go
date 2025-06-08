package businesslogic

import (
	"errors"
	"fmt"
	"time"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
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
	newVar := toStrings(algorithms, func(algorithm cryptoutilBusinessLogicModel.KeyPoolAlgorithm) string {
		return string(algorithm)
	})
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

func (*BusinessLogicService) toEncAndAlg(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) (*joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	switch *ormKeyPoolAlgorithm {
	case cryptoutilOrmRepository.A256GCM_A256KW:
		return &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW, nil
	case cryptoutilOrmRepository.A192GCM_A256KW:
		return &cryptoutilJose.EncA192GCM, &cryptoutilJose.AlgA256KW, nil
	case cryptoutilOrmRepository.A128GCM_A256KW:
		return &cryptoutilJose.EncA128GCM, &cryptoutilJose.AlgA256KW, nil
	case cryptoutilOrmRepository.A256GCM_A192KW:
		return &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA192KW, nil
	case cryptoutilOrmRepository.A192GCM_A192KW:
		return &cryptoutilJose.EncA192GCM, &cryptoutilJose.AlgA192KW, nil
	case cryptoutilOrmRepository.A128GCM_A192KW:
		return &cryptoutilJose.EncA128GCM, &cryptoutilJose.AlgA192KW, nil
	case cryptoutilOrmRepository.A256GCM_A128KW:
		return &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA128KW, nil
	case cryptoutilOrmRepository.A192GCM_A128KW:
		return &cryptoutilJose.EncA192GCM, &cryptoutilJose.AlgA128KW, nil
	case cryptoutilOrmRepository.A128GCM_A128KW:
		return &cryptoutilJose.EncA128GCM, &cryptoutilJose.AlgA128KW, nil

	case cryptoutilOrmRepository.A256GCM_A256GCMKW:
		return &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256GCMKW, nil
	case cryptoutilOrmRepository.A192GCM_A256GCMKW:
		return &cryptoutilJose.EncA192GCM, &cryptoutilJose.AlgA256GCMKW, nil
	case cryptoutilOrmRepository.A128GCM_A256GCMKW:
		return &cryptoutilJose.EncA128GCM, &cryptoutilJose.AlgA256GCMKW, nil
	case cryptoutilOrmRepository.A256GCM_A192GCMKW:
		return &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA192GCMKW, nil
	case cryptoutilOrmRepository.A192GCM_A192GCMKW:
		return &cryptoutilJose.EncA192GCM, &cryptoutilJose.AlgA192GCMKW, nil
	case cryptoutilOrmRepository.A128GCM_A192GCMKW:
		return &cryptoutilJose.EncA128GCM, &cryptoutilJose.AlgA192GCMKW, nil
	case cryptoutilOrmRepository.A256GCM_A128GCMKW:
		return &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA128GCMKW, nil
	case cryptoutilOrmRepository.A192GCM_A128GCMKW:
		return &cryptoutilJose.EncA192GCM, &cryptoutilJose.AlgA128GCMKW, nil
	case cryptoutilOrmRepository.A128GCM_A128GCMKW:
		return &cryptoutilJose.EncA128GCM, &cryptoutilJose.AlgA128GCMKW, nil

	case cryptoutilOrmRepository.A256GCM_dir:
		return &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgDir, nil
	case cryptoutilOrmRepository.A192GCM_dir:
		return &cryptoutilJose.EncA192GCM, &cryptoutilJose.AlgDir, nil
	case cryptoutilOrmRepository.A128GCM_dir:
		return &cryptoutilJose.EncA128GCM, &cryptoutilJose.AlgDir, nil

	case cryptoutilOrmRepository.A256CBCHS512_A256KW:
		return &cryptoutilJose.EncA256CBC_HS512, &cryptoutilJose.AlgA256KW, nil
	case cryptoutilOrmRepository.A192CBCHS384_A256KW:
		return &cryptoutilJose.EncA192CBC_HS384, &cryptoutilJose.AlgA256KW, nil
	case cryptoutilOrmRepository.A128CBCHS256_A256KW:
		return &cryptoutilJose.EncA128CBC_HS256, &cryptoutilJose.AlgA256KW, nil
	case cryptoutilOrmRepository.A256CBCHS512_A192KW:
		return &cryptoutilJose.EncA256CBC_HS512, &cryptoutilJose.AlgA192KW, nil
	case cryptoutilOrmRepository.A192CBCHS384_A192KW:
		return &cryptoutilJose.EncA192CBC_HS384, &cryptoutilJose.AlgA192KW, nil
	case cryptoutilOrmRepository.A128CBCHS256_A192KW:
		return &cryptoutilJose.EncA128CBC_HS256, &cryptoutilJose.AlgA192KW, nil
	case cryptoutilOrmRepository.A256CBCHS512_A128KW:
		return &cryptoutilJose.EncA256CBC_HS512, &cryptoutilJose.AlgA128KW, nil
	case cryptoutilOrmRepository.A192CBCHS384_A128KW:
		return &cryptoutilJose.EncA192CBC_HS384, &cryptoutilJose.AlgA128KW, nil
	case cryptoutilOrmRepository.A128CBCHS256_A128KW:
		return &cryptoutilJose.EncA128CBC_HS256, &cryptoutilJose.AlgA128KW, nil

	case cryptoutilOrmRepository.A256CBCHS512_A256GCMKW:
		return &cryptoutilJose.EncA256CBC_HS512, &cryptoutilJose.AlgA256GCMKW, nil
	case cryptoutilOrmRepository.A192CBCHS384_A256GCMKW:
		return &cryptoutilJose.EncA192CBC_HS384, &cryptoutilJose.AlgA256GCMKW, nil
	case cryptoutilOrmRepository.A128CBCHS256_A256GCMKW:
		return &cryptoutilJose.EncA128CBC_HS256, &cryptoutilJose.AlgA256GCMKW, nil
	case cryptoutilOrmRepository.A256CBCHS512_A192GCMKW:
		return &cryptoutilJose.EncA256CBC_HS512, &cryptoutilJose.AlgA192GCMKW, nil
	case cryptoutilOrmRepository.A192CBCHS384_A192GCMKW:
		return &cryptoutilJose.EncA192CBC_HS384, &cryptoutilJose.AlgA192GCMKW, nil
	case cryptoutilOrmRepository.A128CBCHS256_A192GCMKW:
		return &cryptoutilJose.EncA128CBC_HS256, &cryptoutilJose.AlgA192GCMKW, nil
	case cryptoutilOrmRepository.A256CBCHS512_A128GCMKW:
		return &cryptoutilJose.EncA256CBC_HS512, &cryptoutilJose.AlgA128GCMKW, nil
	case cryptoutilOrmRepository.A192CBCHS384_A128GCMKW:
		return &cryptoutilJose.EncA192CBC_HS384, &cryptoutilJose.AlgA128GCMKW, nil
	case cryptoutilOrmRepository.A128CBCHS256_A128GCMKW:
		return &cryptoutilJose.EncA128CBC_HS256, &cryptoutilJose.AlgA128GCMKW, nil

	case cryptoutilOrmRepository.A256CBCHS512_dir:
		return &cryptoutilJose.EncA256CBC_HS512, &cryptoutilJose.AlgDir, nil
	case cryptoutilOrmRepository.A192CBCHS384_dir:
		return &cryptoutilJose.EncA192CBC_HS384, &cryptoutilJose.AlgDir, nil
	case cryptoutilOrmRepository.A128CBCHS256_dir:
		return &cryptoutilJose.EncA128CBC_HS256, &cryptoutilJose.AlgDir, nil
	default:
		return nil, nil, fmt.Errorf("unsupported keyPool encryption algorithm '%s'", *ormKeyPoolAlgorithm)
	}
}

func (*BusinessLogicService) toAlg(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) (*joseJwa.SignatureAlgorithm, error) {
	switch *ormKeyPoolAlgorithm {
	case cryptoutilOrmRepository.RS512:
		return &cryptoutilJose.AlgRS512, nil
	case cryptoutilOrmRepository.RS384:
		return &cryptoutilJose.AlgRS384, nil
	case cryptoutilOrmRepository.RS256:
		return &cryptoutilJose.AlgRS256, nil

	case cryptoutilOrmRepository.PS512:
		return &cryptoutilJose.AlgPS512, nil
	case cryptoutilOrmRepository.PS384:
		return &cryptoutilJose.AlgPS384, nil
	case cryptoutilOrmRepository.PS256:
		return &cryptoutilJose.AlgPS256, nil

	case cryptoutilOrmRepository.ES512:
		return &cryptoutilJose.AlgES512, nil
	case cryptoutilOrmRepository.ES384:
		return &cryptoutilJose.AlgES384, nil
	case cryptoutilOrmRepository.ES256:
		return &cryptoutilJose.AlgES256, nil

	case cryptoutilOrmRepository.HS512:
		return &cryptoutilJose.AlgHS512, nil
	case cryptoutilOrmRepository.HS384:
		return &cryptoutilJose.AlgHS384, nil
	case cryptoutilOrmRepository.HS256:
		return &cryptoutilJose.AlgHS256, nil

	case cryptoutilOrmRepository.EdDSA:
		return &cryptoutilJose.AlgEdDSA, nil
	default:
		return nil, fmt.Errorf("unsupported keyPool signature algorithm '%s'", *ormKeyPoolAlgorithm)
	}
}
