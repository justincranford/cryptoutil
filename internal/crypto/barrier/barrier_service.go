package barrierservice

import (
	"context"
	"fmt"
	"sync"

	cryptoutilIntermediateKeysService "cryptoutil/internal/crypto/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

type BarrierService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	aes256KeyGenPool        *cryptoutilKeygen.KeyGenPool
	unsealRepository        cryptoutilUnsealRepository.UnsealRepository
	rootKeysService         *cryptoutilRootKeysService.RootKeysService
	intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
	closed                  bool
	shutdownOnce            sync.Once
}

func NewBarrierService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealRepository cryptoutilUnsealRepository.UnsealRepository) (*BarrierService, error) {
	keyPoolConfig, err := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Crypto Service AES-256", 3, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool config: %w", err)
	}
	aes256KeyGenPool, err := cryptoutilKeygen.NewGenKeyPool(keyPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool: %w", err)
	}

	rootKeysService, err := cryptoutilRootKeysService.NewRootKeysService(telemetryService, ormRepository, unsealRepository)
	if err != nil {
		aes256KeyGenPool.Close()
		return nil, fmt.Errorf("failed to create root keys service: %w", err)
	}

	// TODO add intermediateKeysService to intermediateKeysService
	intermediateKeysService, err := cryptoutilIntermediateKeysService.NewIntermediateKeysService(telemetryService, ormRepository, rootKeysService)
	if err != nil {
		rootKeysService.Shutdown()
		aes256KeyGenPool.Close()
		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	return &BarrierService{
		telemetryService:        telemetryService,
		ormRepository:           ormRepository,
		aes256KeyGenPool:        aes256KeyGenPool,
		unsealRepository:        unsealRepository,
		rootKeysService:         rootKeysService,
		intermediateKeysService: intermediateKeysService,
		closed:                  false,
	}, nil
}

func (d *BarrierService) Shutdown() {
	d.shutdownOnce.Do(func() {
		d.closed = true
		if d.aes256KeyGenPool != nil {
			d.aes256KeyGenPool.Close()
			d.aes256KeyGenPool = nil
		}
		if d.intermediateKeysService != nil {
			d.intermediateKeysService.Shutdown()
			d.intermediateKeysService = nil
		}
		if d.rootKeysService != nil {
			d.rootKeysService.Shutdown()
			d.rootKeysService = nil
		}
		d.unsealRepository = nil
		d.ormRepository = nil
		d.telemetryService = nil
	})
}
