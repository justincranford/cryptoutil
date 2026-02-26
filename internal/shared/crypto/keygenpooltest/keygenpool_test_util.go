// Copyright (c) 2025 Justin Cranford
//
//

package elastickeytest

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
)

// RequireNewRSA4096GenElasticKeyForTest creates an RSA-4096 key pool for testing.
func RequireNewRSA4096GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-4096", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateRSAKeyPairFunction(cryptoutilSharedMagic.RSAKeySize4096), false))
}

// RequireNewRSA3072GenElasticKeyForTest creates an RSA-3072 key pool for testing.
func RequireNewRSA3072GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-3072", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateRSAKeyPairFunction(cryptoutilSharedMagic.RSAKeySize3072), false))
}

// RequireNewRSA2048GenElasticKeyForTest creates an RSA-2048 key pool for testing.
func RequireNewRSA2048GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-2048", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateRSAKeyPairFunction(cryptoutilSharedMagic.RSAKeySize2048), false))
}

// RequireNewECDSAP521GenElasticKeyForTest creates an ECDSA-P521 key pool for testing.
func RequireNewECDSAP521GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P521", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P521()), false))
}

// RequireNewECDSAP384GenElasticKeyForTest creates an ECDSA-P384 key pool for testing.
func RequireNewECDSAP384GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P384", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P384()), false))
}

// RequireNewECDSAP256GenElasticKeyForTest creates an ECDSA-P256 key pool for testing.
func RequireNewECDSAP256GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P256", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P256()), false))
}

// RequireNewECDHP521GenElasticKeyForTest creates an ECDH-P521 key pool for testing.
func RequireNewECDHP521GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P521", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDHKeyPairFunction(ecdh.P521()), false))
}

// RequireNewECDHP384GenElasticKeyForTest creates an ECDH-P384 key pool for testing.
func RequireNewECDHP384GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P384", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDHKeyPairFunction(ecdh.P384()), false))
}

// RequireNewECDHP256GenElasticKeyForTest creates an ECDH-P256 key pool for testing.
func RequireNewECDHP256GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P256", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDHKeyPairFunction(ecdh.P256()), false))
}

// RequireNewEd25519GenElasticKeyForTest creates an Ed25519 key pool for testing.
func RequireNewEd25519GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test Ed25519", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPairFunction(cryptoutilSharedMagic.EdCurveEd25519), false))
}

// RequireNewAES256GcmGenElasticKeyForTest creates an AES-256-GCM key pool for testing.
func RequireNewAES256GcmGenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256GCM", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(cryptoutilSharedMagic.AESKeySize256), false))
}

// RequireNewA192GcmGenElasticKeyForTest creates an AES-192-GCM key pool for testing.
func RequireNewA192GcmGenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192GCM", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(cryptoutilSharedMagic.AESKeySize192), false))
}

// RequireNewA128GcmGenElasticKeyForTest creates an AES-128-GCM key pool for testing.
func RequireNewA128GcmGenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128GCM", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(cryptoutilSharedMagic.AESKeySize128), false))
}

// RequireNewAES256CbcHs512GcmGenElasticKeyForTest creates an AES-256-CBC-HS512 key pool for testing.
func RequireNewAES256CbcHs512GcmGenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256CBC-HS512", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESHSKeyFunction(cryptoutilSharedMagic.AESHSKeySize512), false))
}

// RequireNewA192CbcHs384GenElasticKeyForTest creates an AES-192-CBC-HS384 key pool for testing.
func RequireNewA192CbcHs384GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192CBC-HS384", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESHSKeyFunction(cryptoutilSharedMagic.AESHSKeySize384), false))
}

// RequireNewA128CbcHs256GenElasticKeyForTest creates an AES-128-CBC-HS256 key pool for testing.
func RequireNewA128CbcHs256GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[cryptoutilSharedCryptoKeygen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128CBC-HS256", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESHSKeyFunction(cryptoutilSharedMagic.AESHSKeySize256), false))
}

// RequireNewUUIDV7GenElasticKeyForTest creates a UUIDv7 generator pool for testing.
func RequireNewUUIDV7GenElasticKeyForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService) *cryptoutilSharedPool.ValueGenPool[*googleUuid.UUID] {
	return requireNewGenElasticKeyForTest(cryptoutilSharedPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test UUIDv7", 1, cryptoutilSharedMagic.TestPoolMaxSize, cryptoutilSharedMagic.MaxPoolLifetimeValues, cryptoutilSharedMagic.MaxPoolLifetimeDuration, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
}

func requireNewGenElasticKeyForTest[T any](config *cryptoutilSharedPool.ValueGenPoolConfig[T], err error) *cryptoutilSharedPool.ValueGenPool[T] {
	keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(config, err)
	cryptoutilSharedApperr.RequireNoError(err, "failed to create key gen pool")

	return keyGenPool
}
