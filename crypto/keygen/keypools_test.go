package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"cryptoutil/crypto/asn1"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

const (
	exampleNumWorkersRsa   = 3
	exampleNumWorkersOther = 1
	exampleSize            = 3
	exampleMaxSize         = 3
	exampleMaxTime         = MaxTime
)

func TestPoolsExample(t *testing.T) {
	ctx := context.Background()
	telemetryService := cryptoutilTelemetry.NewService(ctx, "asn1_test", false, false)
	defer telemetryService.Shutdown(ctx)

	keys := generateKeys(ctx, telemetryService)
	writeKeys(telemetryService, keys)
	readKeys(telemetryService, keys)
}

func generateKeys(ctx context.Context, telemetryService *cryptoutilTelemetry.Service) []Key {
	rsaPool := NewKeyPool(ctx, telemetryService, "RSA 2048", exampleNumWorkersRsa, exampleSize, exampleMaxSize, exampleMaxTime, GenerateRSAKeyPairFunction(2048))
	ecdsaPool := NewKeyPool(ctx, telemetryService, "ECDSA P256", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateECDSAKeyPairFunction(elliptic.P256()))
	ecdhPool := NewKeyPool(ctx, telemetryService, "ECDH P256", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateECDHKeyPairFunction(ecdh.P256()))
	eddsaPool := NewKeyPool(ctx, telemetryService, "EdDSA Ed25519", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateEDKeyPairFunction("Ed25519"))
	aesPool := NewKeyPool(ctx, telemetryService, "AES 128", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateAESKeyFunction(128))
	hmacPool := NewKeyPool(ctx, telemetryService, "HMAC 256", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateHMACKeyFunction(256))

	defer rsaPool.Close()
	defer ecdsaPool.Close()
	defer ecdhPool.Close()
	defer eddsaPool.Close()
	defer aesPool.Close()
	defer hmacPool.Close()

	keys := make([]Key, 0, 6*exampleMaxSize) // 6 pools * K keys per pool
	for range exampleMaxSize {
		telemetryService.Slogger.Info("Getting keys")
		keys = append(keys, rsaPool.Get())
		keys = append(keys, ecdsaPool.Get())
		keys = append(keys, ecdhPool.Get())
		keys = append(keys, eddsaPool.Get())
		keys = append(keys, aesPool.Get())
		keys = append(keys, hmacPool.Get())
	}

	return keys
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

		err := asn1.PemWrite(key.Private, privatePemFilename)
		if err != nil {
			telemetryService.Slogger.Error("Write failed "+privatePemFilename, "error", err)
			os.Exit(-1)
		}

		err = asn1.DerWrite(key.Private, privateDerFilename)
		if err != nil {
			telemetryService.Slogger.Error("Write failed "+privateDerFilename, "error", err)
			os.Exit(-1)
		}

		if key.Public != nil {
			err = asn1.PemWrite(key.Public, publicPemFilename)
			if err != nil {
				telemetryService.Slogger.Error("Write failed "+baseFilename+"_pub.pem", "error", err)
				os.Exit(-1)
			}

			err = asn1.DerWrite(key.Public, publicDerFilename)
			if err != nil {
				telemetryService.Slogger.Error("Write failed "+baseFilename+"_pub.der", "error", err)
				os.Exit(-1)
			}
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

		_, err := asn1.PemRead(privatePemFilename)
		if err != nil {
			telemetryService.Slogger.Error("Write failed "+privatePemFilename, "error", err)
			os.Exit(-1)
		}

		_, _, err = asn1.DerRead(privateDerFilename)
		if err != nil {
			telemetryService.Slogger.Error("Read failed "+privateDerFilename, "error", err)
			os.Exit(-1)
		}

		if key.Public != nil {
			_, err = asn1.PemRead(publicPemFilename)
			if err != nil {
				telemetryService.Slogger.Error("Read failed "+publicPemFilename, "error", err)
				os.Exit(-1)
			}

			_, _, err = asn1.DerRead(publicDerFilename)
			if err != nil {
				telemetryService.Slogger.Error("Read failed "+publicDerFilename, "error", err)
				os.Exit(-1)
			}
		}
	}
}
