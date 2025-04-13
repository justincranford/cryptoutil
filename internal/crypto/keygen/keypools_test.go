package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilAsn1 "cryptoutil/internal/crypto/asn1"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

const (
	exampleNumWorkersRsa       = 3
	exampleNumWorkersOther     = 1
	examplePoolSize            = 3
	exampleMaxLifetimeKeys     = 3
	exampleMaxLifetimeDuration = MaxLifetimeDuration
)

func TestPoolsExample(t *testing.T) {
	ctx := context.Background()
	telemetryService := cryptoutilTelemetry.RequireNewForTest(testCtx, "keypools_test", false, false)
	defer telemetryService.Shutdown()

	keys, err := generateKeys(ctx, telemetryService)
	if err != nil {
		slog.Error("failed to generate keys", "error", err)
		return
	}
	writeKeys(telemetryService, keys)
	readKeys(telemetryService, keys)
}

func generateKeys(ctx context.Context, telemetryService *cryptoutilTelemetry.Service) ([]Key, error) {
	rsaKeyGenPoolConfig, err1 := NewKeyGenPoolConfig(ctx, telemetryService, "Test RSA 2048", exampleNumWorkersRsa, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateRSAKeyPairFunction(256))
	ecdsaKeyGenPoolConfig, err2 := NewKeyGenPoolConfig(ctx, telemetryService, "Test ECDSA P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateECDSAKeyPairFunction(elliptic.P256()))
	ecdhKeyGenPoolConfig, err3 := NewKeyGenPoolConfig(ctx, telemetryService, "Test ECDH P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateECDHKeyPairFunction(ecdh.P256()))
	eddsaKeyGenPoolConfig, err4 := NewKeyGenPoolConfig(ctx, telemetryService, "Test EdDSA Ed25519", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateEDKeyPairFunction("Ed25519"))
	aesKeyGenPoolConfig, err5 := NewKeyGenPoolConfig(ctx, telemetryService, "Test AES 128", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateAESKeyFunction(128))
	hmacKeyGenPoolConfig, err6 := NewKeyGenPoolConfig(ctx, telemetryService, "Test HMAC 256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateHMACKeyFunction(256))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2, err3, err4, err5, err6))
	}

	rsaKeyGenPool, err1 := NewGenKeyPool(rsaKeyGenPoolConfig)
	ecdsaKeyGenPool, err2 := NewGenKeyPool(ecdsaKeyGenPoolConfig)
	ecdhKeyGenPool, err3 := NewGenKeyPool(ecdhKeyGenPoolConfig)
	eddsaKeyGenPool, err4 := NewGenKeyPool(eddsaKeyGenPoolConfig)
	aesKeyGenPool, err5 := NewGenKeyPool(aesKeyGenPoolConfig)
	hmacKeyGenPool, err6 := NewGenKeyPool(hmacKeyGenPoolConfig)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6))
	}

	defer rsaKeyGenPool.Close()
	defer ecdsaKeyGenPool.Close()
	defer ecdhKeyGenPool.Close()
	defer eddsaKeyGenPool.Close()
	defer aesKeyGenPool.Close()
	defer hmacKeyGenPool.Close()

	keys := make([]Key, 0, 6*exampleMaxLifetimeKeys) // 6 pools * K keys per pool
	for range exampleMaxLifetimeKeys {
		telemetryService.Slogger.Info("Getting keys")
		keys = append(keys, rsaKeyGenPool.Get())
		keys = append(keys, ecdsaKeyGenPool.Get())
		keys = append(keys, ecdhKeyGenPool.Get())
		keys = append(keys, eddsaKeyGenPool.Get())
		keys = append(keys, aesKeyGenPool.Get())
		keys = append(keys, hmacKeyGenPool.Get())
	}

	return keys, nil
}

func writeKeys(telemetryService *cryptoutilTelemetry.Service, keys []Key) {
	for i, key := range keys {
		baseFilename := filepath.Join("output", "key_"+strconv.Itoa(i+1))
		privatePemFilename := baseFilename + "_private.pem"
		privateDerFilename := baseFilename + "_private.der"
		publicPemFilename := baseFilename + "_public.pem"
		publicDerFilename := baseFilename + "_public.der"

		if key.Public == nil {
			privatePemFilename = baseFilename + "_secret.pem"
			privateDerFilename = baseFilename + "_secret.der"
		}

		err := cryptoutilAsn1.PemWrite(key.Private, privatePemFilename)
		cryptoutilAppErr.RequireNoError(err, "Write failed "+privatePemFilename)

		err = cryptoutilAsn1.DerWrite(key.Private, privateDerFilename)
		cryptoutilAppErr.RequireNoError(err, "Write failed "+privateDerFilename)

		if key.Public != nil {
			err = cryptoutilAsn1.PemWrite(key.Public, publicPemFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+baseFilename+"_pub.pem")

			err = cryptoutilAsn1.DerWrite(key.Public, publicDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+baseFilename+"_pub.der")
		}
	}
}

func readKeys(telemetryService *cryptoutilTelemetry.Service, keys []Key) {
	for i, key := range keys {
		baseFilename := filepath.Join("output", "key_"+strconv.Itoa(i+1))
		privatePemFilename := baseFilename + "_private.pem"
		privateDerFilename := baseFilename + "_private.der"
		publicPemFilename := baseFilename + "_public.pem"
		publicDerFilename := baseFilename + "_public.der"

		if key.Public == nil {
			privatePemFilename = baseFilename + "_secret.pem"
			privateDerFilename = baseFilename + "_secret.der"
		}

		_, err := cryptoutilAsn1.PemRead(privatePemFilename)
		cryptoutilAppErr.RequireNoError(err, "Write failed "+privatePemFilename)

		_, _, err = cryptoutilAsn1.DerRead(privateDerFilename)
		cryptoutilAppErr.RequireNoError(err, "Read failed "+privateDerFilename)

		if key.Public != nil {
			_, err = cryptoutilAsn1.PemRead(publicPemFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+publicPemFilename)

			_, _, err = cryptoutilAsn1.DerRead(publicDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+publicDerFilename)
		}
	}
}
