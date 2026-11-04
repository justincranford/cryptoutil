// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"strings"
	"testing"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentPKIJobsVerifyEveryCertificateOnlySource(t *testing.T) {
	t.Parallel()

	for _, selector := range []string{testServiceSMKMSMemSelector, testBootstrapSelect, ""} {
		t.Run("certificate sources="+selector, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			plan, err := BuildDevelopmentRuntimePlan(model, lock, testDevelopmentRun, selector, lock.Images.Postgres.Reference)
			require.NoError(t, err)

			verifiedNamespaces := map[string]bool{}

			for _, job := range plan.Lifecycle.Jobs {
				if job.Args[0] != pki.VerifyCommand {
					continue
				}

				verifiedNamespaces[job.Namespace] = true
				want := []string{}

				for _, source := range plan.Lifecycle.Secrets {
					if source.Namespace != job.Namespace || source.Files[db.InputName] != "" {
						continue
					}

					for key := range source.Files {
						if strings.HasSuffix(key, cryptoutilSharedMagic.PKIInitPublicCertFileExtension) {
							want = append(want, source.Name)

							break
						}
					}
				}

				require.NotEmpty(t, want)
				require.ElementsMatch(t, want, job.SecretNames)
				require.Empty(t, job.WriteSecretNames)
				require.Len(t, job.Args, 2)
				require.True(t, strings.HasPrefix(job.Args[1], "--input=/run/cryptoutil/secrets/"))
				require.True(t, strings.HasSuffix(job.Args[1], "/"+pki.VerificationInputName))
			}

			require.True(t, verifiedNamespaces[plan.inventory.InfraNamespace])

			for _, release := range plan.inventory.Releases {
				require.True(t, verifiedNamespaces[release.Namespace])
			}
		})
	}
}
