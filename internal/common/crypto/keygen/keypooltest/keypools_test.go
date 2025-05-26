package keypooltest

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

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilAsn1 "cryptoutil/internal/common/crypto/asn1"
	"cryptoutil/internal/common/crypto/keygen"
	"cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

const (
	exampleNumWorkersRsa       = 3
	exampleNumWorkersOther     = 1
	examplePoolSize            = 3
	exampleMaxLifetimeKeys     = 3
	exampleMaxLifetimeDuration = pool.MaxLifetimeDuration
)

func TestPoolsExample(t *testing.T) {
	tempDir := t.TempDir()

	ctx := context.Background()
	telemetryService := cryptoutilTelemetry.RequireNewForTest(testCtx, "keypools_test", false, false)
	defer telemetryService.Shutdown()

	keys, err := generateKeys(ctx, telemetryService)
	if err != nil {
		slog.Error("failed to generate keys", "error", err)
		return
	}
	writeKeys(&tempDir, telemetryService, keys)
	readKeys(&tempDir, telemetryService, keys)
}

func generateKeys(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) ([]keygen.Key, error) {
	rsaKeyGenPoolConfig, err1 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Test RSA 2048", exampleNumWorkersRsa, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, keygen.GenerateRSAKeyPairFunction(2048))
	ecdsaKeyGenPoolConfig, err2 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Test ECDSA P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, keygen.GenerateECDSAKeyPairFunction(elliptic.P256()))
	ecdhKeyGenPoolConfig, err3 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Test ECDH P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, keygen.GenerateECDHKeyPairFunction(ecdh.P256()))
	eddsaKeyGenPoolConfig, err4 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Test EdDSA Ed25519", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, keygen.GenerateEDDSAKeyPairFunction("Ed25519"))
	aesKeyGenPoolConfig, err5 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Test AES 128 GCM", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, keygen.GenerateAESKeyFunction(128))
	aesHsKeyGenPoolConfig, err6 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Test AES HS 128", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, keygen.GenerateAESHSKeyFunction(256))
	hmacKeyGenPoolConfig, err7 := pool.NewValueGenPoolConfig(ctx, telemetryService, "Test HMAC 256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, keygen.GenerateHMACKeyFunction(256))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7))
	}

	rsaKeyGenPool, err1 := pool.NewValueGenPool(rsaKeyGenPoolConfig)
	ecdsaKeyGenPool, err2 := pool.NewValueGenPool(ecdsaKeyGenPoolConfig)
	ecdhKeyGenPool, err3 := pool.NewValueGenPool(ecdhKeyGenPoolConfig)
	eddsaKeyGenPool, err4 := pool.NewValueGenPool(eddsaKeyGenPoolConfig)
	aesKeyGenPool, err5 := pool.NewValueGenPool(aesKeyGenPoolConfig)
	aesHsKeyGenPool, err6 := pool.NewValueGenPool(aesHsKeyGenPoolConfig)
	hmacKeyGenPool, err7 := pool.NewValueGenPool(hmacKeyGenPoolConfig)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7))
	}

	defer rsaKeyGenPool.Close()
	defer ecdsaKeyGenPool.Close()
	defer ecdhKeyGenPool.Close()
	defer eddsaKeyGenPool.Close()
	defer aesKeyGenPool.Close()
	defer aesHsKeyGenPool.Close()
	defer hmacKeyGenPool.Close()

	keys := make([]keygen.Key, 0, 7*exampleMaxLifetimeKeys) // 7 pools * K keys per pool
	for range exampleMaxLifetimeKeys {
		telemetryService.Slogger.Info("Getting keys")
		keys = append(keys, rsaKeyGenPool.Get())
		keys = append(keys, ecdsaKeyGenPool.Get())
		keys = append(keys, ecdhKeyGenPool.Get())
		keys = append(keys, eddsaKeyGenPool.Get())
		keys = append(keys, aesKeyGenPool.Get())
		keys = append(keys, aesHsKeyGenPool.Get())
		keys = append(keys, hmacKeyGenPool.Get())
	}

	return keys, nil
}

func writeKeys(tempDir *string, telemetryService *cryptoutilTelemetry.TelemetryService, keys []keygen.Key) {
	var err error
	for i, key := range keys {
		baseFilename := filepath.Join(*tempDir, "key_"+strconv.Itoa(i+1))

		if key.Private != nil {
			privatePemFilename := baseFilename + "_private.pem"
			privateDerFilename := baseFilename + "_private.der"

			err = cryptoutilAsn1.PemWrite(key.Private, privatePemFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+privatePemFilename)

			err = cryptoutilAsn1.DerWrite(key.Private, privateDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+privateDerFilename)
		}

		if key.Public != nil {
			publicPemFilename := baseFilename + "_public.pem"
			publicDerFilename := baseFilename + "_public.der"

			err = cryptoutilAsn1.PemWrite(key.Public, publicPemFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+baseFilename+"_pub.pem")

			err = cryptoutilAsn1.DerWrite(key.Public, publicDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+baseFilename+"_pub.der")
		}

		if key.Secret != nil {
			secretPemFilename := baseFilename + "_secret.pem"
			secretDerFilename := baseFilename + "_secret.der"

			err = cryptoutilAsn1.PemWrite(key.Secret, secretPemFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+secretPemFilename)

			err = cryptoutilAsn1.DerWrite(key.Secret, secretDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Write failed "+secretDerFilename)
		}
	}
}

func readKeys(tempDir *string, telemetryService *cryptoutilTelemetry.TelemetryService, keys []keygen.Key) {
	var err error
	for i, key := range keys {
		baseFilename := filepath.Join(*tempDir, "key_"+strconv.Itoa(i+1))

		if key.Private != nil {
			privatePemFilename := baseFilename + "_private.pem"
			privateDerFilename := baseFilename + "_private.der"

			_, err := cryptoutilAsn1.PemRead(privatePemFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+privatePemFilename)

			_, _, err = cryptoutilAsn1.DerRead(privateDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+privateDerFilename)

		}

		if key.Public != nil {
			publicPemFilename := baseFilename + "_public.pem"
			publicDerFilename := baseFilename + "_public.der"

			_, err = cryptoutilAsn1.PemRead(publicPemFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+publicPemFilename)

			_, _, err = cryptoutilAsn1.DerRead(publicDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+publicDerFilename)
		}

		if key.Secret != nil {
			secretPemFilename := baseFilename + "_secret.pem"
			secretDerFilename := baseFilename + "_secret.der"

			_, err := cryptoutilAsn1.PemRead(secretPemFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+secretPemFilename)

			_, _, err = cryptoutilAsn1.DerRead(secretDerFilename)
			cryptoutilAppErr.RequireNoError(err, "Read failed "+secretDerFilename)

		}
	}
}
