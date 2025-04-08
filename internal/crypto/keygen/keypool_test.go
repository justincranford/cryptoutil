package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name                string
	workers             int
	size                int
	maxLifetimeKeys     int64
	maxLifetimeDuration time.Duration
	gets                int
}

var (
	testCtx                      = context.Background()
	testTelemetryService         *cryptoutilTelemetry.Service
	happyPathWorkers             = []int{1, 2}
	happyPathSize                = []int{1, 3}
	happyPathMaxLifetimeKeys     = []int64{1, 3, MaxLifetimeKeys}
	happyPathMaxLifetimeDuration = []time.Duration{MaxLifetimeDuration}
	happyPathGets                = []int{0, 1, 3, 4}
	happyPathTestCases           = func() []*TestCase {
		testCases := make([]*TestCase, 0, len(happyPathWorkers)*len(happyPathSize)*len(happyPathMaxLifetimeKeys)*len(happyPathMaxLifetimeDuration)*len(happyPathGets))
		for _, workers := range happyPathWorkers {
			for _, size := range happyPathSize {
				if workers > size {
					continue
				}
				for _, maxLifetimeKeys := range happyPathMaxLifetimeKeys {
					if int64(size) > maxLifetimeKeys {
						continue
					}
					for _, maxLifetimeDuration := range happyPathMaxLifetimeDuration {
						for _, gets := range happyPathGets {
							if int64(gets) > maxLifetimeKeys {
								continue
							}
							name := fmt.Sprintf("workers[%d] size[%d] maxLifetimeKeys[%d] maxLifetimeDuration[%v] gets[%d]", workers, size, maxLifetimeKeys, time.Duration(maxLifetimeDuration), gets)
							testCases = append(testCases, &TestCase{name: name, workers: workers, size: size, maxLifetimeKeys: maxLifetimeKeys, maxLifetimeDuration: time.Duration(maxLifetimeDuration), gets: gets})
						}
					}
				}
			}
		}
		return testCases
	}()
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
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateRSAKeyPairFunction(256))
			require.NoError(t, err)
			require.NotNil(t, pool)
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
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateECDSAKeyPairFunction(elliptic.P256()))
			require.NoError(t, err)
			require.NotNil(t, pool)
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
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateECDHKeyPairFunction(ecdh.P256()))
			require.NoError(t, err)
			require.NotNil(t, pool)
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
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateEDKeyPairFunction("Ed25519"))
			require.NoError(t, err)
			require.NotNil(t, pool)
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
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateAESKeyFunction(128))
			require.NoError(t, err)
			require.NotNil(t, pool)
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
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateHMACKeyFunction(256))
			require.NoError(t, err)
			require.NotNil(t, pool)
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
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateUUIDv7Function())
			require.NoError(t, err)
			require.NotNil(t, pool)
			defer pool.Close()

			for i := 0; i < tc.gets; i++ {
				keyPair := pool.Get()
				assert.IsType(t, googleUuid.UUID{}, keyPair.Private)
				assert.Nil(t, keyPair.Public)
			}
		})
	}
}
