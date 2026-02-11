// Copyright (c) 2025 Justin Cranford
//
//

// Package elastickeytest provides testing utilities for elastic key generation pools.
package elastickeytest

import (
	"context"
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	rsa "crypto/rsa"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

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
	testSettings                 = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("keygenpool_test")
	testCtx                      = context.Background()
	testTelemetryService         *cryptoutilSharedTelemetry.TelemetryService
	happyPathWorkers             = []uint32{1, 2}
	happyPathSize                = []uint32{1, 3}
	happyPathMaxLifetimeKeys     = []uint64{1, cryptoutilSharedMagic.MaxPoolLifetimeValues}
	happyPathMaxLifetimeDuration = []time.Duration{cryptoutilSharedMagic.MaxPoolLifetimeDuration}
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
							name := fmt.Sprintf("workers[%d] size[%d] maxLifetimeKeys[%d] maxLifetimeDuration[%v] gets[%d]", workers, size, maxLifetimeKeys, maxLifetimeDuration, gets)
							testCases = append(testCases, &TestCase{name: name, workers: workers, size: size, maxLifetimeKeys: maxLifetimeKeys, maxLifetimeDuration: maxLifetimeDuration, gets: gets})
						}
					}
				}
			}
		}

		return testCases
	}()
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilSharedTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestPoolRSA(t *testing.T) {
	t.Parallel()

	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateRSAKeyPairFunction(2048), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, &rsa.PrivateKey{}, keyPair.Private)
				require.IsType(t, &rsa.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolECDSA(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P256()), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, &ecdsa.PrivateKey{}, keyPair.Private)
				require.IsType(t, &ecdsa.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolECDH(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateECDHKeyPairFunction(ecdh.P256()), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, &ecdh.PrivateKey{}, keyPair.Private)
				require.IsType(t, &ecdh.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolEdDSA(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPairFunction("Ed25519"), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				keyPair := keyGenPool.Get()
				require.IsType(t, ed25519.PrivateKey{}, keyPair.Private)
				require.IsType(t, ed25519.PublicKey{}, keyPair.Public)
			}
		})
	}
}

func TestPoolAES(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESKeyFunction(128), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				secretKey := keyGenPool.Get()
				require.IsType(t, cryptoutilSharedCryptoKeygen.SecretKey{}, secretKey)
				require.IsType(t, []byte{}, []byte(secretKey))
			}
		})
	}
}

func TestPoolAESHS(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateAESHSKeyFunction(256), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				secretKey := keyGenPool.Get()
				require.IsType(t, cryptoutilSharedCryptoKeygen.SecretKey{}, secretKey)
				require.IsType(t, []byte{}, []byte(secretKey))
			}
		})
	}
}

func TestPoolHMAC(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedCryptoKeygen.GenerateHMACKeyFunction(256), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				secretKey := keyGenPool.Get()
				require.IsType(t, cryptoutilSharedCryptoKeygen.SecretKey{}, secretKey)
				require.IsType(t, []byte{}, []byte(secretKey))
			}
		})
	}
}

func TestPoolUUIDv7(t *testing.T) {
	for _, tc := range happyPathTestCases {
		t.Run(tc.name, func(t *testing.T) {
			keyGenPool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(testCtx, testTelemetryService, tc.name, tc.workers, tc.size, tc.maxLifetimeKeys, tc.maxLifetimeDuration, cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false))
			require.NoError(t, err)
			require.NotNil(t, keyGenPool)

			defer keyGenPool.Cancel()

			for i := uint64(0); i < tc.gets; i++ {
				uuidv7 := keyGenPool.Get()
				require.IsType(t, googleUuid.UUID{}, *uuidv7)
			}
		})
	}
}
