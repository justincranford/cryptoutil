package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"certutil/asn1"
)

const (
	exampleNumWorkersRsa   = 3
	exampleNumWorkersOther = 1
	exampleSize            = 3
	exampleMaxSize         = 3
	exampleMaxTime         = MaxTime
)

func DoKeyPoolsExample(ctx context.Context, slogger *slog.Logger) {
	keys := generateKeys(ctx, slogger)

	writeKeys(slogger, keys)
	readKeys(slogger, keys)
}

func generateKeys(ctx context.Context, slogger *slog.Logger) []Key {
	rsaPool := NewKeyPool(ctx, slogger, "RSA 2048", exampleNumWorkersRsa, exampleSize, exampleMaxSize, exampleMaxTime, GenerateRSAKeyPair(2048))
	ecdsaPool := NewKeyPool(ctx, slogger, "ECDSA P256", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateECDSAKeyPair(elliptic.P256()))
	ecdhPool := NewKeyPool(ctx, slogger, "ECDH P256", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateECDHKeyPair(ecdh.P256()))
	eddsaPool := NewKeyPool(ctx, slogger, "EdDSA Ed25519", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateEDKeyPair("Ed25519"))
	aesPool := NewKeyPool(ctx, slogger, "AES 128", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateAESKey(128))
	hmacPool := NewKeyPool(ctx, slogger, "HMAC 256", exampleNumWorkersOther, exampleSize, exampleMaxSize, exampleMaxTime, GenerateHMACKey(256))

	defer rsaPool.Close()
	defer ecdsaPool.Close()
	defer ecdhPool.Close()
	defer eddsaPool.Close()
	defer aesPool.Close()
	defer hmacPool.Close()

	keys := make([]Key, 0, 6*exampleMaxSize) // 6 pools * K keys per pool
	for range exampleMaxSize {
		slogger.Info("Getting keys")
		keys = append(keys, rsaPool.Get())
		keys = append(keys, ecdsaPool.Get())
		keys = append(keys, ecdhPool.Get())
		keys = append(keys, eddsaPool.Get())
		keys = append(keys, aesPool.Get())
		keys = append(keys, hmacPool.Get())
	}

	return keys
}

func writeKeys(slogger *slog.Logger, keys []Key) {
	for i, key := range keys {
		baseFilename := filepath.Join("output", "key_"+strconv.Itoa(i+1))

		err := asn1.PemWrite(key.Private, baseFilename+"_pri.pem")
		if err != nil {
			slogger.Error("Write failed "+baseFilename+"_pri.pem", "error", err)
			os.Exit(-1)
		}

		err = asn1.DerWrite(key.Private, baseFilename+"_pri.der")
		if err != nil {
			slogger.Error("Write failed "+baseFilename+"_pri.der", "error", err)
			os.Exit(-1)
		}

		if key.Public != nil {
			err = asn1.PemWrite(key.Public, baseFilename+"_pub.pem")
			if err != nil {
				slogger.Error("Write failed "+baseFilename+"_pub.pem", "error", err)
				os.Exit(-1)
			}

			err = asn1.DerWrite(key.Public, baseFilename+"_pub.der")
			if err != nil {
				slogger.Error("Write failed "+baseFilename+"_pub.der", "error", err)
				os.Exit(-1)
			}
		}
	}
}

func readKeys(slogger *slog.Logger, keys []Key) {
	for i, key := range keys {
		baseFilename := filepath.Join("output", "key_"+strconv.Itoa(i+1))

		_, err := asn1.PemRead(baseFilename + "_pri.pem")
		if err != nil {
			slogger.Error("Write failed "+baseFilename+"_pri.pem", "error", err)
			os.Exit(-1)
		}

		_, _, err = asn1.DerRead(baseFilename + "_pri.der")
		if err != nil {
			slogger.Error("Read failed "+baseFilename+"_pri.der", "error", err)
			os.Exit(-1)
		}

		if key.Public != nil {
			_, err = asn1.PemRead(baseFilename + "_pub.pem")
			if err != nil {
				slogger.Error("Read failed "+baseFilename+"_pub.pem", "error", err)
				os.Exit(-1)
			}

			_, _, err = asn1.DerRead(baseFilename + "_pub.der")
			if err != nil {
				slogger.Error("Read failed "+baseFilename+"_pub.der", "error", err)
				os.Exit(-1)
			}
		}
	}
}
