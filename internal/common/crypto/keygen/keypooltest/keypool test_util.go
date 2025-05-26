package keypooltest

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	"cryptoutil/internal/common/crypto/keygen"
	"cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

func RequireNewRsa4096GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-4096", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateRSAKeyPairFunction(4096)))
}

func RequireNewRsa3072GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-3072", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateRSAKeyPairFunction(3072)))
}

func RequireNewRsa2048GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-2048", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateRSAKeyPairFunction(2048)))
}

func RequireNewECDSAP521GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P521", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateECDSAKeyPairFunction(elliptic.P521())))
}

func RequireNewECDSAP384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P384", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateECDSAKeyPairFunction(elliptic.P384())))
}

func RequireNewECDSAP256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P256", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateECDSAKeyPairFunction(elliptic.P256())))
}

func RequireNewECDHP521GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P521", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateECDHKeyPairFunction(ecdh.P521())))
}

func RequireNewECDHP384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P384", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateECDHKeyPairFunction(ecdh.P384())))
}

func RequireNewECDHP256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P256", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateECDHKeyPairFunction(ecdh.P256())))
}

func RequireNewEd25519GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test Ed25519", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateEDDSAKeyPairFunction("Ed25519")))
}

func RequireNewAes256GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256GCM", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateAESKeyFunction(256)))
}

func RequireNewA192GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192GCM", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateAESKeyFunction(192)))
}

func RequireNewA128GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128GCM", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateAESKeyFunction(128)))
}

func RequireNewAes256CbcHs512GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256CBC-HS512", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateAESHSKeyFunction(512)))
}

func RequireNewA192CbcHs384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192CBC-HS384", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateAESHSKeyFunction(384)))
}

func RequireNewA128CbcHs256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128CBC-HS256", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateAESHSKeyFunction(256)))
}

func RequireNewUuidV7GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *pool.ValueGenPool[keygen.Key] {
	return requireNewGenKeyPoolForTest(pool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test UUIDv7", 1, 3, pool.MaxLifetimeValues, pool.MaxLifetimeDuration, keygen.GenerateUUIDv7Function()))
}

func requireNewGenKeyPoolForTest(config *pool.ValueGenPoolConfig[keygen.Key], err error) *pool.ValueGenPool[keygen.Key] {
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool config")
	keyGenPool, err := pool.NewValueGenPool(config)
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool")
	return keyGenPool
}
