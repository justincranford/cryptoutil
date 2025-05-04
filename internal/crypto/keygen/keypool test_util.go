package keygen

import (
	"context"
	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

func RequireNewAes256GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A256GCM", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESKeyFunction(256))
	cryptoutilAppErr.RequireNoError(err, "failed to create A256GCM pool config")
	return requireNewGenKeyPoolForTest(keyPoolConfig)
}

func RequireNewA192GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A192GCM", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESKeyFunction(192))
	cryptoutilAppErr.RequireNoError(err, "failed to create A192GCM pool config")
	return requireNewGenKeyPoolForTest(keyPoolConfig)
}

func RequireNewA128GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A128GCM", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESKeyFunction(128))
	cryptoutilAppErr.RequireNoError(err, "failed to create A128GCM pool config")
	return requireNewGenKeyPoolForTest(keyPoolConfig)
}

func RequireNewAes256CbcHs512GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A256CBC-HS512", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESHSKeyFunction(512))
	cryptoutilAppErr.RequireNoError(err, "failed to create A256CBC-HS512 pool config")
	return requireNewGenKeyPoolForTest(keyPoolConfig)
}

func RequireNewA192CbcHs384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A192CBC-HS384", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESHSKeyFunction(384))
	cryptoutilAppErr.RequireNoError(err, "failed to create A192CBC-HS384 pool config")
	return requireNewGenKeyPoolForTest(keyPoolConfig)
}

func RequireNewA128CbcHs256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	keyPoolConfig, err := NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A128CBC-HS256", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESHSKeyFunction(256))
	cryptoutilAppErr.RequireNoError(err, "failed to create A128CBC-HS256 pool config")
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
