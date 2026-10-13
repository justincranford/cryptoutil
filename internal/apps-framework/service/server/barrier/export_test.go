// Copyright (c) 2025-2026 Justin Cranford.

package barrier

import (
	"context"
	"fmt"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

const (
	barrierTestAES256PoolName  = "barrier test AES-256"
	barrierTestEd25519PoolName = "barrier test Ed25519"
	barrierTestUUIDv7PoolName  = "barrier test UUIDv7"
)

func newBarrierTestValueGenPool[T any](
	ctx context.Context,
	telemetryService *cryptoutilSharedTelemetry.TelemetryService,
	poolName string,
	generateFn func() (T, error),
) (*cryptoutilSharedPool.ValueGenPool[T], error) {
	poolConfig, err := cryptoutilSharedPool.NewValueGenPoolConfig(
		ctx,
		telemetryService,
		poolName,
		cryptoutilSharedMagic.DefaultPoolConfigTest.NumWorkers,
		cryptoutilSharedMagic.DefaultPoolConfigTest.MaxSize,
		cryptoutilSharedMagic.MaxPoolLifetimeValues,
		cryptoutilSharedMagic.MaxPoolLifetimeDuration,
		generateFn,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create barrier test pool config %s: %w", poolName, err)
	}

	valuePool, err := cryptoutilSharedPool.NewValueGenPool(poolConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create barrier test pool %s: %w", poolName, err)
	}

	return valuePool, nil
}

func newBarrierTestJWKGenService(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
	switch {
	case ctx == nil:
		return nil, fmt.Errorf("context must be non-nil")
	case telemetryService == nil:
		return nil, fmt.Errorf("telemetry service must be non-nil")
	}

	aes256Pool, err := newBarrierTestValueGenPool(ctx, telemetryService, barrierTestAES256PoolName, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(cryptoutilSharedMagic.AESKeySize256))
	if err != nil {
		return nil, err
	}

	ed25519Pool, err := newBarrierTestValueGenPool(ctx, telemetryService, barrierTestEd25519PoolName, cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPairFunction(cryptoutilSharedMagic.EdCurveEd25519))
	if err != nil {
		aes256Pool.Cancel()

		return nil, err
	}

	uuidPool, err := newBarrierTestValueGenPool(ctx, telemetryService, barrierTestUUIDv7PoolName, cryptoutilSharedUtilRandom.GenerateUUIDv7Function())
	if err != nil {
		aes256Pool.Cancel()
		ed25519Pool.Cancel()

		return nil, err
	}

	return &cryptoutilSharedCryptoJose.JWKGenService{
		AES256KeyGenPool:  aes256Pool,
		ED25519KeyGenPool: ed25519Pool,
		UUIDv7KeyGenPool:  uuidPool,
	}, nil
}

func shutdownBarrierTestJWKGenService(service *cryptoutilSharedCryptoJose.JWKGenService) {
	if service == nil {
		return
	}

	if service.AES256KeyGenPool != nil {
		service.AES256KeyGenPool.Cancel()
	}

	if service.ED25519KeyGenPool != nil {
		service.ED25519KeyGenPool.Cancel()
	}

	if service.UUIDv7KeyGenPool != nil {
		service.UUIDv7KeyGenPool.Cancel()
	}
}

func ExportedNewBarrierTestJWKGenService(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
	return newBarrierTestJWKGenService(ctx, telemetryService)
}

func ExportedShutdownBarrierTestJWKGenService(service *cryptoutilSharedCryptoJose.JWKGenService) {
	shutdownBarrierTestJWKGenService(service)
}
