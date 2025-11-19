// Copyright (c) 2025 Justin Cranford
//
//

package elastickeytest

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"

	googleUuid "github.com/google/uuid"
)

func RequireNewRSA4096GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-4096", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(cryptoutilMagic.RSAKeySize4096), false))
}

func RequireNewRSA3072GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-3072", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(cryptoutilMagic.RSAKeySize3072), false))
}

func RequireNewRSA2048GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test RSA-2048", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(cryptoutilMagic.RSAKeySize2048), false))
}

func RequireNewECDSAP521GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P521", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P521()), false))
}

func RequireNewECDSAP384GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P384", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P384()), false))
}

func RequireNewECDSAP256GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDSA-P256", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256()), false))
}

func RequireNewECDHP521GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P521", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P521()), false))
}

func RequireNewECDHP384GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P384", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P384()), false))
}

func RequireNewECDHP256GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test ECDH-P256", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P256()), false))
}

func RequireNewEd25519GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*cryptoutilKeyGen.KeyPair] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test Ed25519", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateEDDSAKeyPairFunction("Ed25519"), false))
}

func RequireNewAES256GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256GCM", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(cryptoutilMagic.AESKeySize256), false))
}

func RequireNewA192GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192GCM", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(cryptoutilMagic.AESKeySize192), false))
}

func RequireNewA128GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128GCM", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(cryptoutilMagic.AESKeySize128), false))
}

func RequireNewAES256CbcHs512GcmGenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A256CBC-HS512", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(cryptoutilMagic.AESHSKeySize512), false))
}

func RequireNewA192CbcHs384GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A192CBC-HS384", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(cryptoutilMagic.AESHSKeySize384), false))
}

func RequireNewA128CbcHs256GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test A128CBC-HS256", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(cryptoutilMagic.AESHSKeySize256), false))
}

func RequireNewUUIDV7GenElasticKeyForTest(telemetryService *cryptoutilTelemetry.TelemetryService) *cryptoutilPool.ValueGenPool[*googleUuid.UUID] {
	return requireNewGenElasticKeyForTest(cryptoutilPool.NewValueGenPoolConfig(context.Background(), telemetryService, "Test UUIDv7", 1, cryptoutilMagic.TestPoolMaxSize, cryptoutilMagic.MaxPoolLifetimeValues, cryptoutilMagic.MaxPoolLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function(), false))
}

func requireNewGenElasticKeyForTest[T any](config *cryptoutilPool.ValueGenPoolConfig[T], err error) *cryptoutilPool.ValueGenPool[T] {
	keyGenPool, err := cryptoutilPool.NewValueGenPool(config, err)
	cryptoutilAppErr.RequireNoError(err, "failed to create key gen pool")

	return keyGenPool
}
