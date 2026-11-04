// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleOwnershipError = "ownership"
	testLifecyclePodCountError  = "pod count"
	testLifecycleBadQuantity    = "bad-quantity"
	testLifecycleOverflowError  = "overflow"
)

func TestLifecycleJobQuotaHeadroom(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
	lifecycle, err := BuildLifecycleChartValues(model, inventory, input)
	require.NoError(t, err)

	release := inventory.Releases[0]
	runtime, err := BuildRuntimeChartValues(
		model, inventory, release.Key, chartImageFixture(RegisteredScopes()), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(inventory, release.Key),
	)
	require.NoError(t, err)
	adjusted, err := AddLifecycleRuntimeQuota(runtime, lifecycle)
	require.NoError(t, err)
	require.Equal(t, "4", adjusted.Quota.Pods)
	require.Equal(t, "250m", adjusted.Quota.RequestsCPU)
	require.Equal(t, "320Mi", adjusted.Quota.RequestsMemory)
	require.Equal(t, "1200m", adjusted.Quota.LimitsCPU)
	require.Equal(t, "1152Mi", adjusted.Quota.LimitsMemory)
	require.Equal(t, "2", runtime.Quota.Pods, "input projection must remain unchanged")

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", infrastructureChartDefaultsPath))
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(raw)
	require.NoError(t, err)
	infra, err := BuildInfrastructureChartValues(model, inventory, defaults)
	require.NoError(t, err)
	adjustedInfra, err := AddLifecycleInfrastructureQuota(infra, lifecycle)
	require.NoError(t, err)
	originalPods, err := strconv.Atoi(infra.Quota.Pods)
	require.NoError(t, err)
	require.Equal(t, strconv.Itoa(originalPods+3), adjustedInfra.Quota.Pods)
}

func TestLifecycleQuotaRejectsInvalidOwnershipAndResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*RuntimeChartValues, *LifecycleChartValues)
		want string
	}{
		{"wrong run", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Run = testDifferentIdentity }, testLifecycleOwnershipError},
		{"wrong revision", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Revision = "old" }, lifecycleRevisionFormat},
		{"wrong namespace", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Namespace = testDifferentIdentity }, testLifecycleOwnershipError},
		{"wrong release", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.HelmRelease = testDifferentIdentity }, testLifecycleOwnershipError},
		{"invalid pod count", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Quota.Pods = testLifecycleBadQuantity }, testLifecyclePodCountError},
		{"negative pod count", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Quota.Pods = "-1" }, testLifecyclePodCountError},
		{"capacity overflow", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Quota.Pods = strconv.Itoa(maxExpandedPods) }, testLifecyclePodCountError},
		{"invalid base resources", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Quota.RequestsCPU = testLifecycleBadQuantity }, "quota"},
		{"invalid job resources", func(_ *RuntimeChartValues, v *LifecycleChartValues) { v.Jobs[0].Resources = ResourceRequirements{} }, "Job resources"},
		{"request sum overflow", func(r *RuntimeChartValues, _ *LifecycleChartValues) {
			r.Quota.RequestsCPU = testCPUQuantityMaxMilli
			r.Quota.LimitsCPU = testCPUQuantityMaxMilli
		}, testLifecycleOverflowError},
		{"limit sum overflow", func(r *RuntimeChartValues, _ *LifecycleChartValues) { r.Quota.LimitsCPU = testCPUQuantityMaxMilli }, testLifecycleOverflowError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
			lifecycle, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)

			runtime := RuntimeChartValues{
				Run: model.Run, Revision: model.Revision, Namespace: lifecycle.Jobs[0].Namespace, HelmRelease: lifecycle.Jobs[0].Release,
				Quota: RuntimeChartQuota{
					Pods: "2", RequestsCPU: "100m", RequestsMemory: testMemoryQuantity128Mi, LimitsCPU: testCPUQuantity200m, LimitsMemory: testMemoryQuantity256Mi,
				},
			}
			test.edit(&runtime, &lifecycle)
			_, err = AddLifecycleRuntimeQuota(runtime, lifecycle)
			require.ErrorContains(t, err, test.want)
		})
	}

	_, err := AddLifecycleInfrastructureQuota(InfrastructureChartValues{Run: testDifferentIdentity}, LifecycleChartValues{Run: "review"})
	require.ErrorContains(t, err, testLifecycleOwnershipError)
	_, err = AddLifecycleInfrastructureQuota(InfrastructureChartValues{}, LifecycleChartValues{})
	require.Error(t, err)
}
