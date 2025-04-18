package keygen

import (
	"context"
	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

func RequireNewAes256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test AES-256", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESKeyFunction(256))
	cryptoutilAppErr.RequireNoError(err, "failed to create AES-256 pool config")
	return requireNewGenKeyPoolForTest(keyPoolConfig)
}

func RequireNewUuidV7GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test UUIDv7", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateUUIDv7Function())
	cryptoutilAppErr.RequireNoError(err, "failed to create UUIDv7 pool config")
	return requireNewGenKeyPoolForTest(keyPoolConfig)
}

func requireNewGenKeyPoolForTest(config *KeyPoolConfig) *KeyGenPool {
	keyGenPool, err := NewGenKeyPool(config)
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool")
	return keyGenPool
}
