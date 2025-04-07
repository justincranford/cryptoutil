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
	name    string
	workers int
	size    int
	gets    int
	maxSize int
	maxTime time.Duration
}

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
	happyPathWorkers     = []int{1, 2}
	happyPathSize        = []int{1, 3}
	happyPathMaxSize     = []int{1, 3, MaxKeys}
	happyPathMaxTime     = []time.Duration{MaxTime}
	happyPathGets        = []int{0, 1, 3, 4}
	happyPathTestCases   = func() []TestCase {
		testCases := make([]TestCase, len(happyPathWorkers)*len(happyPathSize)*len(happyPathMaxSize)*len(happyPathMaxTime)*len(happyPathGets))
		for _, workers := range happyPathWorkers {
			for _, size := range happyPathSize {
				if workers > size {
					continue
				}
				for _, maxSize := range happyPathMaxSize {
					if size > maxSize {
						continue
					}
					for _, maxTime := range happyPathMaxTime {
						for _, gets := range happyPathGets {
							if gets > maxSize {
								continue
							}
							name := fmt.Sprintf("workers[%d] size[%d] maxSize[%d] maxTime[%v] gets[%d]", workers, size, maxSize, time.Duration(maxTime), gets)
							testCases = append(testCases, TestCase{name: name, workers: workers, size: size, maxSize: maxSize, maxTime: time.Duration(maxTime), gets: gets})
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
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxSize, tc.maxTime, GenerateRSAKeyPairFunction(512))
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
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxSize, tc.maxTime, GenerateECDSAKeyPairFunction(elliptic.P256()))
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
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxSize, tc.maxTime, GenerateECDHKeyPairFunction(ecdh.P256()))
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
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxSize, tc.maxTime, GenerateEDKeyPairFunction("Ed25519"))
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
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxSize, tc.maxTime, GenerateAESKeyFunction(128))
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
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxSize, tc.maxTime, GenerateHMACKeyFunction(256))
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
			pool, err := NewKeyPool(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxSize, tc.maxTime, GenerateUUIDv7Function())
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
