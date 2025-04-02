package servicelogic

import (
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"

	"github.com/google/uuid"
)

type serviceOrmMapper struct{}

func NewMapper() *serviceOrmMapper {
	return &serviceOrmMapper{}
}

// service => orm

func (m *serviceOrmMapper) toOrmKeyPoolInsert(keyPoolID uuid.UUID, serviceKeyPoolCreate *cryptoutilServiceModel.KeyPoolCreate) *cryptoutilOrmRepository.KeyPool {
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

func (*serviceOrmMapper) toOrmKeyPoolProvider(serviceKeyPoolProvider *cryptoutilServiceModel.KeyPoolProvider) *cryptoutilOrmRepository.KeyPoolProvider {
	ormKeyPoolProvider := cryptoutilOrmRepository.KeyPoolProvider(*serviceKeyPoolProvider)
	return &ormKeyPoolProvider
}

func (*serviceOrmMapper) toOrmKeyPoolAlgorithm(serviceKeyPoolProvider *cryptoutilServiceModel.KeyPoolAlgorithm) *cryptoutilOrmRepository.KeyPoolAlgorithm {
	ormKeyPoolAlgorithm := cryptoutilOrmRepository.KeyPoolAlgorithm(*serviceKeyPoolProvider)
	return &ormKeyPoolAlgorithm
}

func (*serviceOrmMapper) toKeyPoolInitialStatus(serviceKeyPoolImportAllowed *cryptoutilServiceModel.KeyPoolImportAllowed) *cryptoutilOrmRepository.KeyPoolStatus {
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

func (*serviceOrmMapper) toServiceKeyPoolAlgorithm(ormKeyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) *cryptoutilServiceModel.KeyPoolAlgorithm {
	serviceKeyPoolAlgorithm := cryptoutilServiceModel.KeyPoolAlgorithm(*ormKeyPoolAlgorithm)
	return &serviceKeyPoolAlgorithm
}

func (*serviceOrmMapper) toServiceKeyPoolProvider(ormKeyPoolProvider *cryptoutilOrmRepository.KeyPoolProvider) *cryptoutilServiceModel.KeyPoolProvider {
	serviceKeyPoolProvider := cryptoutilServiceModel.KeyPoolProvider(*ormKeyPoolProvider)
	return &serviceKeyPoolProvider
}

func (*serviceOrmMapper) toServiceKeyPoolStatus(ormKeyPoolStatus *cryptoutilOrmRepository.KeyPoolStatus) *cryptoutilServiceModel.KeyPoolStatus {
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

func (*serviceOrmMapper) toServiceKey(ormKey *cryptoutilOrmRepository.Key) *cryptoutilServiceModel.Key {
	return &cryptoutilServiceModel.Key{
		Pool:         (*cryptoutilServiceModel.KeyPoolId)(&ormKey.KeyPoolID),
		Id:           &ormKey.KeyID,
		GenerateDate: (*cryptoutilServiceModel.KeyGenerateDate)(ormKey.KeyGenerateDate),
	}
}
