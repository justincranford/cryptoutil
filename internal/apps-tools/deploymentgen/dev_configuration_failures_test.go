// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentConfigurationRejectsInvalidGraphMissingImagesAndMisnamedDiskMode(t *testing.T) {
	t.Parallel()

	for mode, name := range []string{"configuration graph invalid", "missing projected images", "misnamed memory mode backed by disk"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)

			if mode == 2 {
				for index := range model.Modes {
					if model.Modes[index].Name == cryptoutilSharedMagic.DeploymentPoolMem {
						model.Modes[index].Pools = []string{cryptoutilSharedMagic.DeploymentPoolDisk}
					}
				}
			}

			plan, err := BuildDevelopmentRuntimePlan(model, lock, testDevelopmentRun, testServiceSMKMSMemSelector, lock.Images.Postgres.Reference)
			require.NoError(t, err)
			raw, err := os.ReadFile(filepath.Join("..", "..", "..", cryptoutilSharedMagic.CICDConfigsDir, cryptoutilSharedMagic.IMJWTIssuer, "sm-kms-domain.yml"))
			require.NoError(t, err)

			if mode == 0 {
				plan.Model.ClusterDomain = ""
			}

			if mode == 1 {
				plan.RuntimeInput.Images = nil
			}

			_, err = ConfigureDevelopmentSmoke(plan, raw)
			require.Error(t, err)
		})
	}
}
