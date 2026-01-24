// Copyright (c) 2025 Justin Cranford
//
//

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

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilPool "cryptoutil/internal/shared/pool"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

const (
	exampleNumWorkersRSA       = 3
	exampleNumWorkersOther     = 1
	examplePoolSize            = 3
	exampleMaxLifetimeKeys     = 3
	exampleMaxLifetimeDuration = cryptoutilMagic.MaxPoolLifetimeDuration
)

func TestPoolsExample(t *testing.T) {
	tempDir := t.TempDir()

	ctx := context.Background()

	telemetryService := cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
	defer telemetryService.Shutdown()

	keys, err := generateKeys(ctx, telemetryService)
	if err != nil {
		slog.Error("failed to generate keys", "error", err)

		return
	}

	writeKeys(&tempDir, keys)
	readKeys(&tempDir, keys)
}

func generateKeys(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) ([]any, error) {
	rsaKeyGenPool, err1 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test RSA 2048", exampleNumWorkersRSA, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateRSAKeyPairFunction(2048), false))
	ecdsaKeyGenPool, err2 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test ECDSA P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256()), false))
	ecdhKeyGenPool, err3 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test ECDH P256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P256()), false))
	eddsaKeyGenPool, err4 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test EdDSA Ed25519", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateEDDSAKeyPairFunction("Ed25519"), false))
	aesKeyGenPool, err5 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test AES 128 GCM", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(128), false))
	aesHsKeyGenPool, err6 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test AES HS 128", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateAESHSKeyFunction(256), false))
	hmacKeyGenPool, err7 := cryptoutilPool.NewValueGenPool(cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Test HMAC 256", exampleNumWorkersOther, examplePoolSize, exampleMaxLifetimeKeys, exampleMaxLifetimeDuration, cryptoutilKeyGen.GenerateHMACKeyFunction(256), false))

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

func writeKeys(tempDir *string, keys []any) {
	var err error

	for i, keyAny := range keys {
		baseFilename := filepath.Join(*tempDir, "key_"+strconv.Itoa(i+1))

		keyPair, ok := keyAny.(cryptoutilKeyGen.KeyPair)
		if ok {
			if keyPair.Private != nil {
				privatePEMFilename := baseFilename + "_private.pem"
				privateDERFilename := baseFilename + "_private.der"

				err = cryptoutilAsn1.PEMWrite(keyPair.Private, privatePEMFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Write failed "+privatePEMFilename)

				err = cryptoutilAsn1.DERWrite(keyPair.Private, privateDERFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Write failed "+privateDERFilename)
			}

			if keyPair.Public != nil {
				publicPEMFilename := baseFilename + "_public.pem"
				publicDERFilename := baseFilename + "_public.der"

				err = cryptoutilAsn1.PEMWrite(keyPair.Public, publicPEMFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Write failed "+baseFilename+"_pub.pem")

				err = cryptoutilAsn1.DERWrite(keyPair.Public, publicDERFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Write failed "+baseFilename+"_pub.der")
			}
		}

		secretKey, ok := keyAny.([]byte) // pragma: allowlist secret
		if ok {
			if secretKey != nil { // pragma: allowlist secret
				secretPEMFilename := baseFilename + "_secret.pem" // pragma: allowlist secret
				secretDERFilename := baseFilename + "_secret.der" // pragma: allowlist secret

				err = cryptoutilAsn1.PEMWrite(secretKey, secretPEMFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Write failed "+secretPEMFilename)

				err = cryptoutilAsn1.DERWrite(secretKey, secretDERFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Write failed "+secretDERFilename)
			}
		}
	}
}

func readKeys(tempDir *string, keys []any) {
	var err error

	for i, keyAny := range keys {
		baseFilename := filepath.Join(*tempDir, "key_"+strconv.Itoa(i+1))

		keyPair, ok := keyAny.(cryptoutilKeyGen.KeyPair)
		if ok {
			if keyPair.Private != nil {
				privatePEMFilename := baseFilename + "_private.pem"
				privateDERFilename := baseFilename + "_private.der"

				_, err := cryptoutilAsn1.PEMRead(privatePEMFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Read failed "+privatePEMFilename)

				_, _, err = cryptoutilAsn1.DERRead(privateDERFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Read failed "+privateDERFilename)
			}

			if keyPair.Public != nil {
				publicPEMFilename := baseFilename + "_public.pem"
				publicDERFilename := baseFilename + "_public.der"

				_, err = cryptoutilAsn1.PEMRead(publicPEMFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Read failed "+publicPEMFilename)

				_, _, err = cryptoutilAsn1.DERRead(publicDERFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Read failed "+publicDERFilename)
			}
		}

		secretKey, ok := keyAny.([]byte)
		if ok {
			if secretKey != nil { // pragma: allowlist secret
				secretPEMFilename := baseFilename + "_secret.pem" // pragma: allowlist secret
				secretDERFilename := baseFilename + "_secret.der" // pragma: allowlist secret

				_, err := cryptoutilAsn1.PEMRead(secretPEMFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Read failed "+secretPEMFilename)

				_, _, err = cryptoutilAsn1.DERRead(secretDERFilename)
				cryptoutilSharedApperr.RequireNoError(err, "Read failed "+secretDERFilename)
			}
		}
	}
}
