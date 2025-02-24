package main

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"certutil/asn1"
	"certutil/keygen"
	"certutil/telemetry"
)

const (
	numWorkersRsa   = 3
	numWorkersOther = 1
	size            = 3
	maxSize         = 3
	maxTime         = keygen.MaxTime
)

func main() {
	startTime := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slogger := telemetry.InitLogger(ctx, false)

	slogger.Info("Start", "uptime", time.Since(startTime).Seconds())
	defer func() {
		slogger.Info("Stop", "uptime", time.Since(startTime).Seconds())
	}()

	keys := generateKeys(ctx, slogger)

	writeKeys(slogger, keys)
	readKeys(slogger, keys)
}

func generateKeys(ctx context.Context, slogger *slog.Logger) []keygen.Key {
	rsaPool := keygen.NewKeyPool(ctx, slogger, "RSA 2048", numWorkersRsa, size, maxSize, maxTime, keygen.GenerateRSAKeyPair(2048))
	ecdsaPool := keygen.NewKeyPool(ctx, slogger, "ECDSA P256", numWorkersOther, size, maxSize, maxTime, keygen.GenerateECDSAKeyPair(elliptic.P256()))
	ecdhPool := keygen.NewKeyPool(ctx, slogger, "ECDH P256", numWorkersOther, size, maxSize, maxTime, keygen.GenerateECDHKeyPair(ecdh.P256()))
	eddsaPool := keygen.NewKeyPool(ctx, slogger, "EdDSA Ed25519", numWorkersOther, size, maxSize, maxTime, keygen.GenerateEDKeyPair("Ed25519"))
	aesPool := keygen.NewKeyPool(ctx, slogger, "AES 128", numWorkersOther, size, maxSize, maxTime, keygen.GenerateAESKey(128))
	hmacPool := keygen.NewKeyPool(ctx, slogger, "HMAC 256", numWorkersOther, size, maxSize, maxTime, keygen.GenerateHMACKey(256))

	defer rsaPool.Close()
	defer ecdsaPool.Close()
	defer ecdhPool.Close()
	defer eddsaPool.Close()
	defer aesPool.Close()
	defer hmacPool.Close()

	keys := make([]keygen.Key, 0, 6*maxSize) // 6 pools * K keys per pool
	for range maxSize {
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

func writeKeys(slogger *slog.Logger, keys []keygen.Key) {
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

func readKeys(slogger *slog.Logger, keys []keygen.Key) {
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
