// Copyright (c) 2025 Justin Cranford
//
//

// Package barrierservice provides hierarchical encryption-at-rest using a multi-layer key hierarchy.
package barrierservice

import (
	"context"
	"fmt"
	"sync"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilContentKeysService "cryptoutil/internal/shared/barrier/contentkeysservice"
	cryptoutilIntermediateKeysService "cryptoutil/internal/shared/barrier/intermediatekeysservice"
	cryptoutilRootKeysService "cryptoutil/internal/shared/barrier/rootkeysservice"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// BarrierService provides hierarchical encryption-at-rest using unseal, root, intermediate, and content keys.
type BarrierService struct {
	telemetryService        *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService           *cryptoutilSharedCryptoJose.JWKGenService
	ormRepository           *cryptoutilOrmRepository.OrmRepository
	unsealKeysService       cryptoutilUnsealKeysService.UnsealKeysService
	rootKeysService         *cryptoutilRootKeysService.RootKeysService
	intermediateKeysService *cryptoutilIntermediateKeysService.IntermediateKeysService
	contentKeysService      *cryptoutilContentKeysService.ContentKeysService
	closed                  bool
	shutdownOnce            sync.Once
}

// NewService creates a new BarrierService with the specified dependencies.
func NewService(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) (*BarrierService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if unsealKeysService == nil {
		return nil, fmt.Errorf("unsealKeysService must be non-nil")
	}

	rootKeysService, err := cryptoutilRootKeysService.NewRootKeysService(telemetryService, jwkGenService, ormRepository, unsealKeysService)
	if err != nil {
		return nil, fmt.Errorf("failed to create root keys service: %w", err)
	}

	intermediateKeysService, err := cryptoutilIntermediateKeysService.NewIntermediateKeysService(telemetryService, jwkGenService, ormRepository, rootKeysService)
	if err != nil {
		rootKeysService.Shutdown()

		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	contentKeysService, err := cryptoutilContentKeysService.NewContentKeysService(telemetryService, jwkGenService, ormRepository, intermediateKeysService)
	if err != nil {
		rootKeysService.Shutdown()
		intermediateKeysService.Shutdown()

		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	return &BarrierService{
		telemetryService:        telemetryService,
		jwkGenService:           jwkGenService,
		ormRepository:           ormRepository,
		unsealKeysService:       unsealKeysService,
		rootKeysService:         rootKeysService,
		intermediateKeysService: intermediateKeysService,
		contentKeysService:      contentKeysService,
		closed:                  false,
	}, nil
}

// EncryptContent encrypts the provided clear bytes using the content key hierarchy.
func (d *BarrierService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}

	encryptedContentJWEMessageBytes, _, err := d.contentKeysService.EncryptContent(sqlTransaction, clearBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content bytes: %w", err)
	}

	return encryptedContentJWEMessageBytes, nil
}

// DecryptContent decrypts the provided encrypted JWE message bytes using the content key hierarchy.
func (d *BarrierService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encryptedContentJWEMessageBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}

	decryptedBytes, err := d.contentKeysService.DecryptContent(sqlTransaction, encryptedContentJWEMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content bytes: %w", err)
	}

	return decryptedBytes, nil
}

// Shutdown releases all resources held by the BarrierService.
func (d *BarrierService) Shutdown() {
	d.shutdownOnce.Do(func() {
		d.closed = true
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
