package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"log/slog"
	"os"
	"testing"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name    string
	workers int
	gets    int
	maxSize int
	maxTime time.Duration
}

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
)

func TestMain(m *testing.M) {
	telemetryService, err := cryptoutilTelemetry.NewService(testCtx, "keypool_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	testTelemetryService = telemetryService
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func TestPoolRSA(t *testing.T) {
	testCases := []TestCase{
		{name: "Finite RSA 2048", workers: 2, gets: 3, maxSize: 3, maxTime: 3 * time.Second},
		{name: "Indefinite RSA 2048", workers: 2, gets: 3, maxSize: MaxKeys, maxTime: MaxTime},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.gets, tc.maxSize, tc.maxTime, GenerateRSAKeyPairFunction(2048))
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, &rsa.PrivateKey{}, keyPair.Private)
				assert.IsType(t, &rsa.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolEcDSA(t *testing.T) {
	testCases := []TestCase{
		{name: "Finite ECDSA P256", workers: 2, gets: 3, maxSize: 3, maxTime: 3 * time.Second},
		{name: "Indefinite ECDSA P256", workers: 2, gets: 3, maxSize: MaxKeys, maxTime: MaxTime},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.gets, tc.maxSize, tc.maxTime, GenerateECDSAKeyPairFunction(elliptic.P256()))
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, &ecdsa.PrivateKey{}, keyPair.Private)
				assert.IsType(t, &ecdsa.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolEcDH(t *testing.T) {
	testCases := []TestCase{
		{name: "Finite ECDH P256", workers: 2, gets: 3, maxSize: 3, maxTime: 3 * time.Second},
		{name: "Indefinite ECDH P256", workers: 2, gets: 3, maxSize: MaxKeys, maxTime: MaxTime},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.gets, tc.maxSize, tc.maxTime, GenerateECDHKeyPairFunction(ecdh.P256()))
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, &ecdh.PrivateKey{}, keyPair.Private)
				assert.IsType(t, &ecdh.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolEdDSA(t *testing.T) {
	testCases := []TestCase{
		{name: "Finite Ed25519", workers: 2, gets: 3, maxSize: 3, maxTime: 3 * time.Second},
		{name: "Indefinite Ed25519", workers: 2, gets: 3, maxSize: MaxKeys, maxTime: MaxTime},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.gets, tc.maxSize, tc.maxTime, GenerateEDKeyPairFunction("Ed25519"))
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, ed25519.PrivateKey{}, keyPair.Private)
				assert.IsType(t, ed25519.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolAES(t *testing.T) {
	testCases := []TestCase{
		{name: "Finite AES 128", workers: 2, gets: 3, maxSize: 3, maxTime: 3 * time.Second},
		{name: "Indefinite AES 128", workers: 2, gets: 3, maxSize: MaxKeys, maxTime: MaxTime},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.gets, tc.maxSize, tc.maxTime, GenerateAESKeyFunction(128))
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, []byte{}, keyPair.Private)
				assert.Nil(t, keyPair.Public)
			}
		})
	}
}

func TestPoolHMAC(t *testing.T) {
	testCases := []TestCase{
		{name: "Finite HMAC 256", workers: 2, gets: 3, maxSize: 3, maxTime: 3 * time.Second},
		{name: "Indefinite HMAC 256", workers: 2, gets: 3, maxSize: MaxKeys, maxTime: MaxTime},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.gets, tc.maxSize, tc.maxTime, GenerateHMACKeyFunction(256))
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, []byte{}, keyPair.Private)
				assert.Nil(t, keyPair.Public)
			}
		})
	}
}

func TestPoolUUIDv7(t *testing.T) {
	testCases := []TestCase{
		{name: "Finite UUID V7", workers: 2, gets: 3, maxSize: 3, maxTime: 3 * time.Second},
		{name: "Indefinite UUID V7", workers: 2, gets: 3, maxSize: MaxKeys, maxTime: MaxTime},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.gets, tc.maxSize, tc.maxTime, GenerateUUIDv7Function())
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, googleUuid.UUID{}, keyPair.Private)
				assert.Nil(t, keyPair.Public)
			}
		})
	}
}
