package keygen

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	googleUuid "github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name                string
	workers             uint32
	size                uint32
	maxLifetimeKeys     uint64
	maxLifetimeDuration time.Duration
	gets                uint64
}

var (
	testCtx                      = context.Background()
	testTelemetryService         *cryptoutilTelemetry.TelemetryService
	happyPathWorkers             = []uint32{1, 2}
	happyPathSize                = []uint32{1, 3}
	happyPathMaxLifetimeKeys     = []uint64{1, MaxLifetimeKeys}
	happyPathMaxLifetimeDuration = []time.Duration{MaxLifetimeDuration}
	happyPathGets                = []uint64{0, 1, 3}
	happyPathTestCases           = func() []*TestCase {
		testCases := make([]*TestCase, 0, len(happyPathWorkers)*len(happyPathSize)*len(happyPathMaxLifetimeKeys)*len(happyPathMaxLifetimeDuration)*len(happyPathGets))
		for _, workers := range happyPathWorkers {
			for _, size := range happyPathSize {
				if workers > size {
					continue
				}
				for _, maxLifetimeKeys := range happyPathMaxLifetimeKeys {
					if uint64(size) > maxLifetimeKeys {
						continue
					}
					for _, maxLifetimeDuration := range happyPathMaxLifetimeDuration {
						for _, gets := range happyPathGets {
							if gets > maxLifetimeKeys {
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
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "keypool_test", false, false)
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func TestPoolRSA(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateRSAKeyPairFunction(2048))
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, &rsa.PrivateKey{}, keyPair.Private)
				require.IsType(t, &rsa.PublicKey{}, keyPair.Public)
				require.Nil(t, keyPair.Secret)
			}
		})
	}
}

func TestPoolEcDSA(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateECDSAKeyPairFunction(elliptic.P256()))
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, &ecdsa.PrivateKey{}, keyPair.Private)
				require.IsType(t, &ecdsa.PublicKey{}, keyPair.Public)
				require.Nil(t, keyPair.Secret)
			}
		})
	}
}

func TestPoolEcDH(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateECDHKeyPairFunction(ecdh.P256()))
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, &ecdh.PrivateKey{}, keyPair.Private)
				require.IsType(t, &ecdh.PublicKey{}, keyPair.Public)
				require.Nil(t, keyPair.Secret)
			}
		})
	}
}

func TestPoolEdDSA(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateEDDSAKeyPairFunction("Ed25519"))
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, ed25519.PrivateKey{}, keyPair.Private)
				require.IsType(t, ed25519.PublicKey{}, keyPair.Public)
				require.Nil(t, keyPair.Secret)
			}
		})
	}
}

func TestPoolAES(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateAESKeyFunction(128))
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, []byte{}, keyPair.Secret)
				require.Nil(t, keyPair.Private)
				require.Nil(t, keyPair.Public)
			}
		})
	}
}

func TestPoolAESHS(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateAESHSKeyFunction(256))
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, []byte{}, keyPair.Secret)
				require.Nil(t, keyPair.Private)
				require.Nil(t, keyPair.Public)
			}
		})
	}
}

func TestPoolHMAC(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateHMACKeyFunction(256))
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, []byte{}, keyPair.Secret)
				require.Nil(t, keyPair.Private)
				require.Nil(t, keyPair.Public)
			}
		})
	}
}

func TestPoolUUIDv7(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPoolConfig, err := NewKeyGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, GenerateUUIDv7Function())
			require.NoError(t, err)
			require.NotNil(t, keyGenPoolConfig)
			keyGenPool, err := NewGenKeyPool(keyGenPoolConfig)
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)
			defer keyGenPool.Close()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, googleUuid.UUID{}, keyPair.Secret)
				require.Nil(t, keyPair.Private)
				require.Nil(t, keyPair.Public)
			}
		})
	}
}
