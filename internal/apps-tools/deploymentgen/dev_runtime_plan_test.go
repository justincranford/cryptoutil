// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"strings"
	"testing"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentRuntimePlanUsesOwnedGraphAndExactSuiteRoutes(t *testing.T) {
	t.Parallel()

	for _, selector := range []string{testServiceSMKMSMemSelector, "s:sm-kms:disk", testBootstrapSelect, ""} {
		t.Run("selector="+selector, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			plan, err := BuildDevelopmentRuntimePlan(model, lock, "p2sm1103", selector, lock.Images.Postgres.Reference)
			require.NoError(t, err)
			require.Equal(t, "cu-p2sm1103-cluster", plan.Cluster.Project)
			require.NotEqual(t, "r1", plan.Model.Revision)
			inventory, err := Build(plan.Model, selector)
			require.NoError(t, err)
			values, err := BuildLifecycleChartValues(plan.Model, inventory, plan.Lifecycle)
			require.NoError(t, err)

			for _, job := range values.Jobs {
				require.Equal(t, []string{pki.SuiteExecutable}, job.Command)
				require.False(t, job.APIToken)
				require.Empty(t, job.WriteSecretNames)
			}

			require.Equal(t, []string{testLifecycleLegacyProduct, testLifecycleLegacyService, cryptoutilSharedMagic.CMD_SERVER}, plan.RuntimeInput.Images[cryptoutilSharedMagic.IMJWTIssuer].Args)
			require.Equal(t, []string{
				pki.SuiteExecutable, testLifecycleLegacyProduct, testLifecycleLegacyService,
				strings.TrimPrefix(cryptoutilSharedMagic.PrivateAdminStartupzRequestPath, "/"),
			}, plan.RuntimeInput.Images[cryptoutilSharedMagic.IMJWTIssuer].Probes.Startup)
		})
	}
}

func TestDevelopmentRuntimePlanRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, run, selector, image string }{
		{"invalid development run", "shared-run", testServiceSMKMSMemSelector, "postgres@sha256:" + strings.Repeat("a", 64)},
		{"invalid development selector", testDevelopmentRun, "s:unknown:mem", "postgres@sha256:" + strings.Repeat("a", 64)},
		{"unpinned development image", testDevelopmentRun, testServiceSMKMSMemSelector, "suite:latest"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			_, err = BuildDevelopmentRuntimePlan(model, lock, test.run, test.selector, test.image)
			require.Error(t, err)
		})
	}
}

const (
	testDevelopmentInvalidGraph    = "invalid identity graph"
	testDevelopmentAbsentSource    = "absent verification source"
	testDevelopmentUnsupportedJob  = "unsupported database Job"
	testDevelopmentMissingMaterial = "missing owned material"
	testDevelopmentUnexpectedPool  = "unexpected workload pool"
)

func TestDevelopmentProfileRejectsGraphDriftBeforeRuntime(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testDevelopmentInvalidGraph, testDevelopmentAbsentSource, testDevelopmentUnsupportedJob, testDevelopmentMissingMaterial, testDevelopmentUnexpectedPool} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			model.Run = testDevelopmentRun
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			cluster, err := BuildLocalK3sPlan(model, lock)
			require.NoError(t, err)
			inventory, err := Build(model, testBootstrapSelect)
			require.NoError(t, err)

			switch test {
			case testDevelopmentInvalidGraph:
				inventory.Infra = append(inventory.Infra, inventory.Infra[0])
			case testDevelopmentAbsentSource:
				objects := []Object{}

				for _, object := range inventory.Infra {
					if object.Kind != testLifecycleSecretKind {
						objects = append(objects, object)
					}
				}

				inventory.Infra = objects
			case testDevelopmentUnsupportedJob:
				inventory.Releases[0].Objects = append(inventory.Releases[0].Objects, Object{Kind: lifecycleJobKind, Name: "unsupported-migrate-" + model.Revision})
			case testDevelopmentMissingMaterial:
				inventory.Infra = append(inventory.Infra, Object{Kind: testLifecycleSecretKind, Name: "unplanned-owned-material"})
			case testDevelopmentUnexpectedPool:
				inventory.Releases[0].Workloads[0].Pool = cryptoutilSharedMagic.STRATEGY_UNKNOWN
			}

			_, err = buildDevelopmentRuntimePlanForInventory(model, cluster, inventory, testBootstrapSelect, lock.Images.Postgres.Reference)
			require.Error(t, err)
		})
	}
}
