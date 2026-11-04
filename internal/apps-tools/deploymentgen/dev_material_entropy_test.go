// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"crypto/elliptic"
	crand "crypto/rand"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentMaterialEntropyFailuresNeverPublishPartialInputs(t *testing.T) {
	t.Parallel()

	for mode := range 4 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			plan, err := BuildDevelopmentRuntimePlan(model, lock, testDevelopmentRun, testServiceSMKMSMemSelector, lock.Images.Postgres.Reference)
			require.NoError(t, err)

			calls, last := 0, 1

			if mode == 1 {
				for _, secret := range plan.Lifecycle.Secrets {
					for key := range secret.Files {
						if key != pki.VerificationInputName && !strings.HasSuffix(key, cryptoutilSharedMagic.PKIInitPublicCertFileExtension) && !strings.HasSuffix(key, cryptoutilSharedMagic.PKIInitPrivateKeyFileExtension) {
							last++
						}
					}
				}
			}

			entropy := func(raw []byte) (int, error) {
				calls++
				if calls == last {
					if mode == 2 {
						return len(raw) - 1, nil
					}

					return 0, errors.New("PRIVATE")
				}

				return crand.Read(raw)
			}
			if mode == testMaterialInfraOffset {
				entropy = nil
			}

			material, err := generateDevelopmentMaterialWithEntropy(t.Context(), plan, DevelopmentMaterialPolicy{Curve: elliptic.P384(), Lifetime: time.Hour},
				cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair, entropy)
			require.Error(t, err)
			require.NotContains(t, err.Error(), "PRIVATE")
			require.Zero(t, len(material.Secrets))
			require.Zero(t, len(material.Token))
		})
	}
}
