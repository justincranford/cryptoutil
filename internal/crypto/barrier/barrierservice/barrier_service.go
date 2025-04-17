package barrierservice

import (
	"context"
	"fmt"
	"sync"

	cryptoutilIntermediateKeysService "cryptoutil/internal/crypto/barrier/intermediatekeysservice"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

type BarrierService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	aes256KeyGenPool        *cryptoutilKeygen.KeyGenPool
	intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
	closed                  bool
	shutdownOnce            sync.Once
}

func NewBarrierService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService) (*BarrierService, error) {
	keyPoolConfig, err := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Crypto Service AES-256", 3, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool config: %w", err)
	}
	aes256KeyGenPool, err := cryptoutilKeygen.NewGenKeyPool(keyPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool: %w", err)
	}

	return &BarrierService{
		telemetryService:        telemetryService,
		ormRepository:           ormRepository,
		intermediateKeysService: intermediateKeysService,
		aes256KeyGenPool:        aes256KeyGenPool,
		closed:                  false,
	}, nil
}

func (d *BarrierService) Shutdown() {
	d.shutdownOnce.Do(func() {
		if d.aes256KeyGenPool != nil {
			d.aes256KeyGenPool.Close()
			d.aes256KeyGenPool = nil
		}
		d.ormRepository = nil
		d.telemetryService = nil
		d.closed = true
	})
}
