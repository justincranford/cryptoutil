// Copyright (c) 2025 Justin Cranford
//
//

package elastickeytest

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilPool "cryptoutil/internal/shared/pool"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
)

// RequireNewRSA4096GenElasticKeyForTest creates an RSA-4096 key pool for testing.
func RequireNewRSA4096GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-4096", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(cryptoutilMagic.RSAKeySize4096), false))
}

// RequireNewRSA3072GenElasticKeyForTest creates an RSA-3072 key pool for testing.
func RequireNewRSA3072GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-3072", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(cryptoutilMagic.RSAKeySize3072), false))
}

// RequireNewRSA2048GenElasticKeyForTest creates an RSA-2048 key pool for testing.
func RequireNewRSA2048GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-2048", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(cryptoutilMagic.RSAKeySize2048), false))
}

// RequireNewECDSAP521GenElasticKeyForTest creates an ECDSA-P521 key pool for testing.
func RequireNewECDSAP521GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P521", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P521()), false))
}

// RequireNewECDSAP384GenElasticKeyForTest creates an ECDSA-P384 key pool for testing.
func RequireNewECDSAP384GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P384", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P384()), false))
}

// RequireNewECDSAP256GenElasticKeyForTest creates an ECDSA-P256 key pool for testing.
func RequireNewECDSAP256GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P256", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256()), false))
}

// RequireNewECDHP521GenElasticKeyForTest creates an ECDH-P521 key pool for testing.
func RequireNewECDHP521GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P521", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P521()), false))
}

// RequireNewECDHP384GenElasticKeyForTest creates an ECDH-P384 key pool for testing.
func RequireNewECDHP384GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P384", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P384()), false))
}

// RequireNewECDHP256GenElasticKeyForTest creates an ECDH-P256 key pool for testing.
func RequireNewECDHP256GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P256", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P256()), false))
}

// RequireNewEd25519GenElasticKeyForTest creates an Ed25519 key pool for testing.
func RequireNewEd25519GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test Ed25519", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateEDDSAKeyPairFunction("Ed25519"), false))
}

// RequireNewAES256GcmGenElasticKeyForTest creates an AES-256-GCM key pool for testing.
func RequireNewAES256GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256GCM", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(cryptoutilMagic.AESKeySize256), false))
}

// RequireNewA192GcmGenElasticKeyForTest creates an AES-192-GCM key pool for testing.
func RequireNewA192GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192GCM", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(cryptoutilMagic.AESKeySize192), false))
}

// RequireNewA128GcmGenElasticKeyForTest creates an AES-128-GCM key pool for testing.
func RequireNewA128GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128GCM", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(cryptoutilMagic.AESKeySize128), false))
}

// RequireNewAES256CbcHs512GcmGenElasticKeyForTest creates an AES-256-CBC-HS512 key pool for testing.
func RequireNewAES256CbcHs512GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256CBC-HS512", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(cryptoutilMagic.AESHSKeySize512), false))
}

// RequireNewA192CbcHs384GenElasticKeyForTest creates an AES-192-CBC-HS384 key pool for testing.
func RequireNewA192CbcHs384GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192CBC-HS384", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(cryptoutilMagic.AESHSKeySize384), false))
}

// RequireNewA128CbcHs256GenElasticKeyForTest creates an AES-128-CBC-HS256 key pool for testing.
func RequireNewA128CbcHs256GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128CBC-HS256", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(cryptoutilMagic.AESHSKeySize256), false))
}

// RequireNewUUIDV7GenElasticKeyForTest creates a UUIDv7 generator pool for testing.
func RequireNewUUIDV7GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*googleUuid.UUID] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test UUIDv7", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilRandom.GenerateUUIDv7Function(), false))
}

func requireNewGenElasticKeyForTest[T any](config *cryptoutilPool.ValueGenPoolConfig[T], err error) *cryptoutilPool.ValueGenPool[T] {
	keyGenPool, err := cryptoutilPool.NewValueGenPool(config, err)
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool")

	return keyGenPool
}
