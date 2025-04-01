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

const (
	testNumWorkers = 2
	testSize       = 3
	testMaxSize    = 3
	testMaxTime    = 3 * time.Second
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
)

func TestMain(m *testing.M) {
	telemetryService, err := cryptoutilTelemetry.NewService(testCtx, "asn1_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	testTelemetryService = telemetryService
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func TestPoolRSA(t *testing.T) {
	pool := NewKeyPool(testCtx, testTelemetryService, "Test RSA 2048", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateRSAKeyPairFunction(2048))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, &rsa.PrivateKey{}, keyPair.Private)
		assert.IsType(t, &rsa.PublicKey{}, keyPair.Public)
	}
}

func TestPoolEcDSA(t *testing.T) {
	pool := NewKeyPool(testCtx, testTelemetryService, "Test ECDSA P256", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateECDSAKeyPairFunction(elliptic.P256()))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, &ecdsa.PrivateKey{}, keyPair.Private)
		assert.IsType(t, &ecdsa.PublicKey{}, keyPair.Public)
	}
}

func TestPoolEcDH(t *testing.T) {
	pool := NewKeyPool(testCtx, testTelemetryService, "Test ECDH P256", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateECDHKeyPairFunction(ecdh.P256()))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, &ecdh.PrivateKey{}, keyPair.Private)
		assert.IsType(t, &ecdh.PublicKey{}, keyPair.Public)
	}
}

func TestPoolEdDSA(t *testing.T) {
	pool := NewKeyPool(testCtx, testTelemetryService, "Test Ed25519", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateEDKeyPairFunction("Ed25519"))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, ed25519.PrivateKey{}, keyPair.Private)
		assert.IsType(t, ed25519.PublicKey{}, keyPair.Public)
	}
}

func TestPoolAES(t *testing.T) {
	pool := NewKeyPool(testCtx, testTelemetryService, "Test AES-128", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateAESKeyFunction(128))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, []byte{}, keyPair.Private)
		assert.Nil(t, keyPair.Public)
	}
}

func TestPoolHMAC(t *testing.T) {
	pool := NewKeyPool(testCtx, testTelemetryService, "Test HMAC-256", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateHMACKeyFunction(256))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, []byte{}, keyPair.Private)
		assert.Nil(t, keyPair.Public)
	}
}

func TestPoolUUIDv7(t *testing.T) {
	pool := NewKeyPool(testCtx, testTelemetryService, "Test UUIDv7", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateUUIDv7Function())
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, googleUuid.UUID{}, keyPair.Private)
		assert.Nil(t, keyPair.Public)
	}
}
