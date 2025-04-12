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
	rsaPoolConfig, err1 := NewKeyPoolConfig(ctx, telemetryService, "Test RSA 2048", exampleNumWorkersRsa, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateRSAKeyPairFunction(256))
	ecdsaPoolConfig, err2 := NewKeyPoolConfig(ctx, telemetryService, "Test ECDSA P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateECDSAKeyPairFunction(elliptic.P256()))
	ecdhPoolConfig, err3 := NewKeyPoolConfig(ctx, telemetryService, "Test ECDH P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateECDHKeyPairFunction(ecdh.P256()))
	eddsaPoolConfig, err4 := NewKeyPoolConfig(ctx, telemetryService, "Test EdDSA Ed25519", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateEDKeyPairFunction("Ed25519"))
	aesPoolConfig, err5 := NewKeyPoolConfig(ctx, telemetryService, "Test AES 128", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateAESKeyFunction(128))
	hmacPoolConfig, err6 := NewKeyPoolConfig(ctx, telemetryService, "Test HMAC 256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, GenerateHMACKeyFunction(256))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2, err3, err4, err5, err6))
	}

	rsaPool, err1 := NewKeyPool(rsaPoolConfig)
	ecdsaPool, err2 := NewKeyPool(ecdsaPoolConfig)
	ecdhPool, err3 := NewKeyPool(ecdhPoolConfig)
	eddsaPool, err4 := NewKeyPool(eddsaPoolConfig)
	aesPool, err5 := NewKeyPool(aesPoolConfig)
	hmacPool, err6 := NewKeyPool(hmacPoolConfig)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6))
	}

	defer rsaPool.Close()
	defer ecdsaPool.Close()
	defer ecdhPool.Close()
	defer eddsaPool.Close()
	defer aesPool.Close()
	defer hmacPool.Close()

	keys := make([]Key, 0, 6*exampleMaxLifetimeKeys) // 6 pools * K keys per pool
	for range exampleMaxLifetimeKeys {
		telemetryService.Slogger.Info("Getting keys")
		keys = append(keys, rsaPool.Get())
		keys = append(keys, ecdsaPool.Get())
		keys = append(keys, ecdhPool.Get())
		keys = append(keys, eddsaPool.Get())
		keys = append(keys, aesPool.Get())
		keys = append(keys, hmacPool.Get())
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
