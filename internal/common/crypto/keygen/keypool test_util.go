package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

func RequireNewRsa4096GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test RSA-4096", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateRSAKeyPairFunction(4096)))
}

func RequireNewRsa3072GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test RSA-3072", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateRSAKeyPairFunction(3072)))
}

func RequireNewRsa2048GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test RSA-2048", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateRSAKeyPairFunction(2048)))
}

func RequireNewECDSAP521GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P521", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateECDSAKeyPairFunction(elliptic.P521())))
}

func RequireNewECDSAP384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P384", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateECDSAKeyPairFunction(elliptic.P384())))
}

func RequireNewECDSAP256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P256", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateECDSAKeyPairFunction(elliptic.P256())))
}

func RequireNewECDHP521GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P521", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateECDHKeyPairFunction(ecdh.P521())))
}

func RequireNewECDHP384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P384", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateECDHKeyPairFunction(ecdh.P384())))
}

func RequireNewECDHP256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P256", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateECDHKeyPairFunction(ecdh.P256())))
}

func RequireNewEd25519GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test Ed25519", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateEDDSAKeyPairFunction("Ed25519")))
}

func RequireNewAes256GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A256GCM", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESKeyFunction(256)))
}

func RequireNewA192GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A192GCM", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESKeyFunction(192)))
}

func RequireNewA128GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A128GCM", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESKeyFunction(128)))
}

func RequireNewAes256CbcHs512GcmGenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A256CBC-HS512", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESHSKeyFunction(512)))
}

func RequireNewA192CbcHs384GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A192CBC-HS384", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESHSKeyFunction(384)))
}

func RequireNewA128CbcHs256GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test A128CBC-HS256", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateAESHSKeyFunction(256)))
}

func RequireNewUuidV7GenKeyPoolForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *KeyGenPool {
	return requireNewGenKeyPoolForTest(NewKeyGenPoolConfig(context.Background(), telemetryService, "Test UUIDv7", 1, 3, MaxLifetimeKeys, MaxLifetimeDuration, GenerateUUIDv7Function()))
}

func requireNewGenKeyPoolForTest(config *KeyPoolConfig, err error) *KeyGenPool {
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool config")
	keyGenPool, err := NewGenKeyPool(config)
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool")
	return keyGenPool
}
