package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"cryptoutil/telemetry"

	"github.com/stretchr/testify/assert"
)

const (
	testNumWorkers = 2
	testSize       = 3
	testMaxSize    = 3
	testMaxTime    = 3 * time.Second
)

var (
	ctx              context.Context
	telemetryService *telemetry.Service
)

func TestMain(m *testing.M) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	telemetryService = telemetry.Init(ctx)
	telemetry.Shutdown(telemetryService)

	rc := m.Run()
	os.Exit(rc)
}

func TestPoolRSA(t *testing.T) {
	pool := NewKeyPool(ctx, telemetryService, "RSA", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateRSAKeyPair(2048))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, &rsa.PrivateKey{}, keyPair.Private)
		assert.IsType(t, &rsa.PublicKey{}, keyPair.Public)
	}
}

func TestPoolEcDSA(t *testing.T) {
	pool := NewKeyPool(ctx, telemetryService, "EC", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateECDSAKeyPair(elliptic.P256()))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, &ecdsa.PrivateKey{}, keyPair.Private)
		assert.IsType(t, &ecdsa.PublicKey{}, keyPair.Public)
	}
}

func TestPoolEcDH(t *testing.T) {
	pool := NewKeyPool(ctx, telemetryService, "EC", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateECDHKeyPair(ecdh.P256()))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, &ecdh.PrivateKey{}, keyPair.Private)
		assert.IsType(t, &ecdh.PublicKey{}, keyPair.Public)
	}
}

func TestPoolEdDSA(t *testing.T) {
	pool := NewKeyPool(ctx, telemetryService, "ED", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateEDKeyPair("Ed25519"))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, ed25519.PrivateKey{}, keyPair.Private)
		assert.IsType(t, ed25519.PublicKey{}, keyPair.Public)
	}
}

func TestPoolAES(t *testing.T) {
	pool := NewKeyPool(ctx, telemetryService, "AES", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateAESKey(128))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, []byte{}, keyPair.Private)
		assert.Nil(t, keyPair.Public)
	}
}

func TestPoolHMAC(t *testing.T) {
	pool := NewKeyPool(ctx, telemetryService, "AES", testNumWorkers, testSize, testMaxSize, testMaxTime, GenerateAESKey(128))
	defer pool.Close()

	for i := 0; i < testMaxSize; i++ {
		keyPair := pool.Get()
		assert.IsType(t, []byte{}, keyPair.Private)
		assert.Nil(t, keyPair.Public)
	}
}
