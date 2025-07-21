package elastickeytest

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
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

const (
	exampleNumWorkersRsa       = 3
	exampleNumWorkersOther     = 1
	examplePoolSize            = 3
	exampleMaxLifetimeKeys     = 3
	exampleMaxLifetimeDuration = cryptoutilPool.MaxLifetimeDuration
)

func TestPoolsExample(t *testing.T) {
	tempDir := t.TempDir()

	ctx := context.Background()

	testSettings.DevMode = true
	testSettings.Migrations = true
	testSettings.OTLPScope = "keygenpools_test"

	telemetryService := cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
	defer telemetryService.Shutdown()

	keys, err := generateKeys(ctx, telemetryService)
	if err != nil {
		slog.Error("failed to generate keys", "error", err)
		return
	}
	writeKeys(&tempDir, telemetryService, keys)
	readKeys(&tempDir, telemetryService, keys)
}

func generateKeys(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) ([]any, error) {
	rsaKeyGenPool, err1 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test RSA 2048", exampleNumWorkersRsa, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(2048)))
	ecdsaKeyGenPool, err2 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test ECDSA P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256())))
	ecdhKeyGenPool, err3 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test ECDH P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P256())))
	eddsaKeyGenPool, err4 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test EdDSA Ed25519", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateEDDSAKeyPairFunction("Ed25519")))
	aesKeyGenPool, err5 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test AES 128 GCM", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(128)))
	aesHsKeyGenPool, err6 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test AES HS 128", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(256)))
	hmacKeyGenPool, err7 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test HMAC 256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateHMACKeyFunction(256)))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7))
	}

	defer rsaKeyGenPool.Cancel()
	defer ecdsaKeyGenPool.Cancel()
	defer ecdhKeyGenPool.Cancel()
	defer eddsaKeyGenPool.Cancel()
	defer aesKeyGenPool.Cancel()
	defer aesHsKeyGenPool.Cancel()
	defer hmacKeyGenPool.Cancel()

	keys := make([]any, 0, 7*exampleMaxLifetimeKeys) // 7 pools * K keys per pool
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

func writeKeys(tempDir *string, telemetryService *cryptoutilTelemetry.TelemetryService, keys []any) {
	var err error
	for i, keyAny := range keys {
		baseFilename := filepath.Join(*tempDir, "key_"+strconv.Itoa(i+1))

		keyPair, ok := keyAny.(cryptoutilKeyGen.KeyPair)
		if ok {
			if keyPair.Private != nil {
				privatePemFilename := baseFilename + "_private.pem"
				privateDerFilename := baseFilename + "_private.der"

				err = cryptoutilAsn1.PemWrite(keyPair.Private, privatePemFilename)
				cryptoutilAppErr.RequireNoError(err, "Write failed "+privatePemFilename)

				err = cryptoutilAsn1.DerWrite(keyPair.Private, privateDerFilename)
				cryptoutilAppErr.RequireNoError(err, "Write failed "+privateDerFilename)
			}

			if keyPair.Public != nil {
				publicPemFilename := baseFilename + "_public.pem"
				publicDerFilename := baseFilename + "_public.der"

				err = cryptoutilAsn1.PemWrite(keyPair.Public, publicPemFilename)
				cryptoutilAppErr.RequireNoError(err, "Write failed "+baseFilename+"_pub.pem")

				err = cryptoutilAsn1.DerWrite(keyPair.Public, publicDerFilename)
				cryptoutilAppErr.RequireNoError(err, "Write failed "+baseFilename+"_pub.der")
			}
		}
		secretKey, ok := keyAny.([]byte)
		if ok {
			if secretKey != nil {
				secretPemFilename := baseFilename + "_secret.pem"
				secretDerFilename := baseFilename + "_secret.der"

				err = cryptoutilAsn1.PemWrite(secretKey, secretPemFilename)
				cryptoutilAppErr.RequireNoError(err, "Write failed "+secretPemFilename)

				err = cryptoutilAsn1.DerWrite(secretKey, secretDerFilename)
				cryptoutilAppErr.RequireNoError(err, "Write failed "+secretDerFilename)
			}
		}
	}
}

func readKeys(tempDir *string, telemetryService *cryptoutilTelemetry.TelemetryService, keys []any) {
	var err error
	for i, keyAny := range keys {
		baseFilename := filepath.Join(*tempDir, "key_"+strconv.Itoa(i+1))

		keyPair, ok := keyAny.(cryptoutilKeyGen.KeyPair)
		if ok {
			if keyPair.Private != nil {
				privatePemFilename := baseFilename + "_private.pem"
				privateDerFilename := baseFilename + "_private.der"

				_, err := cryptoutilAsn1.PemRead(privatePemFilename)
				cryptoutilAppErr.RequireNoError(err, "Read failed "+privatePemFilename)

				_, _, err = cryptoutilAsn1.DerRead(privateDerFilename)
				cryptoutilAppErr.RequireNoError(err, "Read failed "+privateDerFilename)

			}

			if keyPair.Public != nil {
				publicPemFilename := baseFilename + "_public.pem"
				publicDerFilename := baseFilename + "_public.der"

				_, err = cryptoutilAsn1.PemRead(publicPemFilename)
				cryptoutilAppErr.RequireNoError(err, "Read failed "+publicPemFilename)

				_, _, err = cryptoutilAsn1.DerRead(publicDerFilename)
				cryptoutilAppErr.RequireNoError(err, "Read failed "+publicDerFilename)
			}
		}

		secretKey, ok := keyAny.([]byte)
		if ok {
			if secretKey != nil {
				secretPemFilename := baseFilename + "_secret.pem"
				secretDerFilename := baseFilename + "_secret.der"

				_, err := cryptoutilAsn1.PemRead(secretPemFilename)
				cryptoutilAppErr.RequireNoError(err, "Read failed "+secretPemFilename)

				_, _, err = cryptoutilAsn1.DerRead(secretDerFilename)
				cryptoutilAppErr.RequireNoError(err, "Read failed "+secretDerFilename)
			}
		}
	}
}
