// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentConfigurationUsesExplicitProtectedMounts(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testDevCompleteDomain, testDevAbsentDomain, testDevMalformedDomain, testDevConfigWrongRun} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			plan, err := BuildDevelopmentRuntimePlan(model, lock, testDevelopmentRun, testServiceSMKMSMemSelector, lock.Images.Postgres.Reference)
			require.NoError(t, err)
			raw, err := os.ReadFile(filepath.Join("..", "..", "..", cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.IMJWTIssuer, "sm-kms-domain.yml"))
			require.NoError(t, err)

			switch test {
			case testDevAbsentDomain:
				raw = nil
			case testDevMalformedDomain:
				raw = []byte("{")
			case testDevConfigWrongRun:
				plan.Model.Run = testDifferentIdentity
			}

			configured, err := ConfigureDevelopmentSmoke(plan, raw)
			if test != testDevCompleteDomain {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			for _, input := range configured.RuntimeInput.Runtime.Workloads {
				require.Contains(t, input.FrameworkConfig, "server-admin-tls-client-policy: require-and-verify")
				require.Contains(t, input.FrameworkConfig, "/run/cryptoutil/credentials/0/unseal-5of5.secret")
				require.Contains(t, input.FrameworkConfig, "/run/cryptoutil/pki/output/admin-server/tls.key")
				require.Equal(t, string(raw), input.DomainConfig)
			}

			require.NotEqual(t, plan.RuntimeInput.Images[cryptoutilSharedMagic.IMJWTIssuer].Probes, configured.RuntimeInput.Images[cryptoutilSharedMagic.IMJWTIssuer].Probes)
		})
	}
}
