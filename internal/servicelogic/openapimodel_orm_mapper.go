package servicelogic

import (
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
)

type serviceOrmMapper struct{}

func NewMapper() *serviceOrmMapper {
	return &serviceOrmMapper{}
}

// service => orm

func (m *serviceOrmMapper) toOrmKeyPoolInsert(serviceKeyPoolCreate *cryptoutilServiceModel.KeyPoolCreate) *cryptoutilOrmRepository.KeyPool {
	return &cryptoutilOrmRepository.KeyPool{
		KeyPoolName:                serviceKeyPoolCreate.Name,
		KeyPoolDescription:         serviceKeyPoolCreate.Description,
		KeyPoolProvider:            *m.toOrmKeyPoolProvider(serviceKeyPoolCreate.Provider),
		KeyPoolAlgorithm:           *m.toOrmKeyPoolAlgorithm(serviceKeyPoolCreate.Algorithm),
		KeyPoolIsVersioningAllowed: *serviceKeyPoolCreate.VersioningAllowed,
		KeyPoolIsImportAllowed:     *serviceKeyPoolCreate.ImportAllowed,
		KeyPoolIsExportAllowed:     *serviceKeyPoolCreate.ExportAllowed,
		KeyPoolStatus:              *m.toKeyPoolInitialStatus(serviceKeyPoolCreate.ImportAllowed),
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

func (*serviceOrmMapper) toKeyPoolInitialStatus(serviceKeyPoolIsImportAllowed *cryptoutilServiceModel.KeyPoolIsImportAllowed) *cryptoutilOrmRepository.KeyPoolStatus {
	var ormKeyPoolStatus cryptoutilOrmRepository.KeyPoolStatus
	if *serviceKeyPoolIsImportAllowed {
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
		VersioningAllowed: &ormKeyPool.KeyPoolIsVersioningAllowed,
		ImportAllowed:     &ormKeyPool.KeyPoolIsImportAllowed,
		ExportAllowed:     &ormKeyPool.KeyPoolIsExportAllowed,
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
