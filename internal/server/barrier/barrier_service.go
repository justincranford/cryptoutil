package barrierservice

import (
	"context"
	"fmt"
	"sync"

	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilContentKeysService "cryptoutil/internal/server/barrier/contentkeysservice"
	cryptoutilIntermediateKeysService "cryptoutil/internal/server/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
)

type BarrierService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	uuidV7KeyGenPool        *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
	aes256KeyGenPool        *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey]
	unsealKeysService       cryptoutilUnsealKeysService.UnsealKeysService
	rootKeysService         *cryptoutilRootKeysService.RootKeysService
	intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
	contentKeysService      *cryptoutilContentKeysService.ContentKeysService
	closed                  bool
	shutdownOnce            sync.Once
}

func NewBarrierService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) (*BarrierService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if unsealKeysService == nil {
		return nil, fmt.Errorf("unsealKeysService must be non-nil")
	}

	uuidV7KeyGenPool, err := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Barrier Service UUIDv7", 2, 2, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function()))
	if err != nil {
		return nil, fmt.Errorf("failed to create UUID pool: %w", err)
	}

	aes256KeyGenPool, err := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Barrier Service Keys AES-256-GCM", 3, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256)))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool: %w", err)
	}

	rootKeysService, err := cryptoutilRootKeysService.NewRootKeysService(telemetryService, ormRepository, unsealKeysService, uuidV7KeyGenPool, aes256KeyGenPool)
	if err != nil {
		aes256KeyGenPool.Cancel()
		return nil, fmt.Errorf("failed to create root keys service: %w", err)
	}

	intermediateKeysService, err := cryptoutilIntermediateKeysService.NewIntermediateKeysService(telemetryService, ormRepository, rootKeysService, uuidV7KeyGenPool, aes256KeyGenPool)
	if err != nil {
		aes256KeyGenPool.Cancel()
		rootKeysService.Shutdown()
		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	contentKeysService, err := cryptoutilContentKeysService.NewContentKeysService(telemetryService, ormRepository, intermediateKeysService, uuidV7KeyGenPool, aes256KeyGenPool)
	if err != nil {
		aes256KeyGenPool.Cancel()
		rootKeysService.Shutdown()
		intermediateKeysService.Shutdown()
		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	return &BarrierService{
		telemetryService:        telemetryService,
		ormRepository:           ormRepository,
		unsealKeysService:       unsealKeysService,
		uuidV7KeyGenPool:        uuidV7KeyGenPool,
		aes256KeyGenPool:        aes256KeyGenPool,
		rootKeysService:         rootKeysService,
		intermediateKeysService: intermediateKeysService,
		contentKeysService:      contentKeysService,
		closed:                  false,
	}, nil
}

func (d *BarrierService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}
	encryptedContentJweMessageBytes, _, err := d.contentKeysService.EncryptContent(sqlTransaction, clearBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content bytes: %w", err)
	}
	return encryptedContentJweMessageBytes, nil
}

func (d *BarrierService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentJweMessageBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}
	decryptedBytes, err := d.contentKeysService.DecryptContent(sqlTransaction, encryptedContentJweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content bytes: %w", err)
	}
	return decryptedBytes, nil
}

func (d *BarrierService) Shutdown() {
	d.shutdownOnce.Do(func() {
		d.closed = true
		if d.aes256KeyGenPool != nil {
			d.aes256KeyGenPool.Cancel()
			d.aes256KeyGenPool = nil
		}
		if d.uuidV7KeyGenPool != nil {
			d.uuidV7KeyGenPool.Cancel()
			d.uuidV7KeyGenPool = nil
		}
		if d.contentKeysService != nil {
			d.contentKeysService.Shutdown()
			d.contentKeysService = nil
		}
		if d.intermediateKeysService != nil {
			d.intermediateKeysService.Shutdown()
			d.intermediateKeysService = nil
		}
		if d.rootKeysService != nil {
			d.rootKeysService.Shutdown()
			d.rootKeysService = nil
		}
		d.unsealKeysService = nil
		d.ormRepository = nil
		d.telemetryService = nil
	})
}
