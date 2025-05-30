package keypooltest

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"

	googleUuid "github.com/google/uuid"
)

func RequireNewRsa4096GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-4096", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(4096)))
}

func RequireNewRsa3072GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-3072", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(3072)))
}

func RequireNewRsa2048GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-2048", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(2048)))
}

func RequireNewECDSAP521GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P521", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P521())))
}

func RequireNewECDSAP384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P384", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P384())))
}

func RequireNewECDSAP256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P256", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256())))
}

func RequireNewECDHP521GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P521", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P521())))
}

func RequireNewECDHP384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P384", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P384())))
}

func RequireNewECDHP256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P256", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P256())))
}

func RequireNewEd25519GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test Ed25519", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateEDDSAKeyPairFunction("Ed25519")))
}

func RequireNewAes256GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256GCM", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(256)))
}

func RequireNewA192GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192GCM", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(192)))
}

func RequireNewA128GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128GCM", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(128)))
}

func RequireNewAes256CbcHs512GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256CBC-HS512", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(512)))
}

func RequireNewA192CbcHs384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192CBC-HS384", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(384)))
}

func RequireNewA128CbcHs256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128CBC-HS256", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(256)))
}

func RequireNewUuidV7GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*googleUuid.UUID] {
	return requireNewGenKeyPoolForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test UUIDv7", 1, 3, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function()))
}

func requireNewGenKeyPoolForTest[T any](config *cryptoutilPool.ValueGenPoolConfig[T], err error) *cryptoutilPool.ValueGenPool[T] {
	keyGenPool, err := cryptoutilPool.NewValueGenPool(config, err)
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool")
	return keyGenPool
}
