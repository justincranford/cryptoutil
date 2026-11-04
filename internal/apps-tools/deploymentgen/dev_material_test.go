// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	"crypto/elliptic"
	"errors"
	"testing"
	"time"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentMaterialRealPolicyAndNegativeGuards(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testDevRealSmoke, testDevMaterialWrongRun, testDevWrongSelector, testDevNilCurve, testDevLifetime, testDevCanceled, testDevEntropy} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			plan, err := BuildDevelopmentRuntimePlan(model, lock, testDevelopmentRun, testServiceSMKMSMemSelector, lock.Images.Postgres.Reference)
			require.NoError(t, err)

			policy := DevelopmentMaterialPolicy{Curve: elliptic.P384(), Lifetime: 24 * time.Hour}
			ctx := context.Background()
			generate := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair

			switch test {
			case testDevMaterialWrongRun:
				plan.Model.Run = testDifferentIdentity
			case testDevWrongSelector:
				plan.Selector = testBootstrapSelect
			case testDevNilCurve:
				policy.Curve = nil
			case testDevLifetime:
				policy.Lifetime = 30 * 24 * time.Hour
			case testDevCanceled:
				canceled, cancel := context.WithCancel(ctx)
				cancel()

				ctx = canceled
			case testDevEntropy:
				generate = func(elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) { return nil, errors.New("PRIVATE") }
			}

			var material DevelopmentMaterial
			if test == testDevRealSmoke {
				material, err = GenerateDevelopmentMaterial(ctx, plan, policy)
			} else {
				material, err = generateDevelopmentMaterial(ctx, plan, policy, generate)
			}

			if test != testDevRealSmoke {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")
				require.Empty(t, material.Secrets)

				return
			}

			require.NoError(t, err)
			require.Len(t, material.Token, 64)
			require.Len(t, material.Secrets, len(plan.Lifecycle.Secrets))

			for _, secret := range material.Secrets {
				require.Equal(t, testDevelopmentRun, secret.Metadata.Labels[cryptoutilSharedMagic.KubernetesLabelRun])

				for _, raw := range secret.Data {
					require.NotEmpty(t, raw)
				}
			}
		})
	}
}
