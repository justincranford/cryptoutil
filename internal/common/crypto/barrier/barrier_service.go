package barrierservice

import (
	"context"
	"fmt"
	"sync"

	cryptoutilContentKeysService "cryptoutil/internal/common/crypto/barrier/contentkeysservice"
	cryptoutilIntermediateKeysService "cryptoutil/internal/common/crypto/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/common/crypto/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/common/crypto/barrier/unsealkeysservice"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
)

type BarrierService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	aes256KeyGenPool        *cryptoutilKeygen.KeyGenPool
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

	keyPoolConfig, err := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Barrier Service Keys AES-256-GCM", 3, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool config: %w", err)
	}
	aes256KeyGenPool, err := cryptoutilKeygen.NewGenKeyPool(keyPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool: %w", err)
	}

	rootKeysService, err := cryptoutilRootKeysService.NewRootKeysService(telemetryService, ormRepository, unsealKeysService, aes256KeyGenPool)
	if err != nil {
		aes256KeyGenPool.Close()
		return nil, fmt.Errorf("failed to create root keys service: %w", err)
	}

	intermediateKeysService, err := cryptoutilIntermediateKeysService.NewIntermediateKeysService(telemetryService, ormRepository, rootKeysService, aes256KeyGenPool)
	if err != nil {
		aes256KeyGenPool.Close()
		rootKeysService.Shutdown()
		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	contentKeysService, err := cryptoutilContentKeysService.NewContentKeysService(telemetryService, ormRepository, intermediateKeysService, aes256KeyGenPool)
	if err != nil {
		aes256KeyGenPool.Close()
		rootKeysService.Shutdown()
		intermediateKeysService.Shutdown()
		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	return &BarrierService{
		telemetryService:        telemetryService,
		ormRepository:           ormRepository,
		unsealKeysService:       unsealKeysService,
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
			d.aes256KeyGenPool.Close()
			d.aes256KeyGenPool = nil
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
