// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"context"
	"fmt"
	"sync"

	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// BarrierService provides multi-layer encryption using unseal → root → intermediate → content key hierarchy.
// This version uses BarrierRepository interface to work with any database (KMS OrmRepository, gorm.DB, etc.)
type BarrierService struct {
	telemetryService        *cryptoutilTelemetry.TelemetryService
	jwkGenService           *cryptoutilJose.JWKGenService
	repository              BarrierRepository
	unsealKeysService       cryptoutilUnsealKeysService.UnsealKeysService
	rootKeysService         *RootKeysService
	intermediateKeysService *IntermediateKeysService
	contentKeysService      *ContentKeysService
	closed                  bool
	shutdownOnce            sync.Once
}

// NewBarrierService creates a new barrier service using the provided repository.
// The repository can be:
// - OrmBarrierRepository (wraps KMS OrmRepository for backward compatibility)
// - GormBarrierRepository (wraps gorm.DB for cipher-im and future services).
func NewBarrierService(
	ctx context.Context,
	telemetryService *cryptoutilTelemetry.TelemetryService,
	jwkGenService *cryptoutilJose.JWKGenService,
	repository BarrierRepository,
	unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService,
) (*BarrierService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	}

	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	}

	if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	}

	if repository == nil {
		return nil, fmt.Errorf("repository must be non-nil")
	}

	if unsealKeysService == nil {
		return nil, fmt.Errorf("unsealKeysService must be non-nil")
	}

	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, repository, unsealKeysService)
	if err != nil {
		return nil, fmt.Errorf("failed to create root keys service: %w", err)
	}

	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, jwkGenService, repository, rootKeysService)
	if err != nil {
		rootKeysService.Shutdown()

		return nil, fmt.Errorf("failed to create intermediate keys service: %w", err)
	}

	contentKeysService, err := NewContentKeysService(telemetryService, jwkGenService, repository, intermediateKeysService)
	if err != nil {
		rootKeysService.Shutdown()
		intermediateKeysService.Shutdown()

		return nil, fmt.Errorf("failed to create content keys service: %w", err)
	}

	return &BarrierService{
		telemetryService:        telemetryService,
		jwkGenService:           jwkGenService,
		repository:              repository,
		unsealKeysService:       unsealKeysService,
		rootKeysService:         rootKeysService,
		intermediateKeysService: intermediateKeysService,
		contentKeysService:      contentKeysService,
		closed:                  false,
	}, nil
}

// EncryptContentWithContext encrypts data using the content key (which is encrypted by intermediate key, which is encrypted by root key, which is encrypted by unseal key).
func (d *BarrierService) EncryptContentWithContext(ctx context.Context, clearBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}

	var encryptedBytes []byte

	err := d.repository.WithTransaction(ctx, func(tx BarrierTransaction) error {
		var err error

		encryptedBytes, _, err = d.contentKeysService.EncryptContent(tx, clearBytes)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content bytes: %w", err)
	}

	return encryptedBytes, nil
}

// DecryptContentWithContext decrypts data using the content key hierarchy.
func (d *BarrierService) DecryptContentWithContext(ctx context.Context, encryptedContentJWEMessageBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}

	var decryptedBytes []byte

	err := d.repository.WithTransaction(ctx, func(tx BarrierTransaction) error {
		var err error

		decryptedBytes, err = d.contentKeysService.DecryptContent(tx, encryptedContentJWEMessageBytes)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content bytes: %w", err)
	}

	return decryptedBytes, nil
}

// EncryptBytesWithContext is an alias for EncryptContentWithContext for API consistency.
func (d *BarrierService) EncryptBytesWithContext(ctx context.Context, clearBytes []byte) ([]byte, error) {
	return d.EncryptContentWithContext(ctx, clearBytes)
}

// DecryptBytesWithContext is an alias for DecryptContentWithContext for API consistency.
func (d *BarrierService) DecryptBytesWithContext(ctx context.Context, encryptedBytes []byte) ([]byte, error) {
	return d.DecryptContentWithContext(ctx, encryptedBytes)
}

// Shutdown releases all resources held by the barrier service.
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
		d.repository = nil
		d.telemetryService = nil
	})
}
