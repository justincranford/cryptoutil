// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildRuntimeChartQuotaUsesMaximumContainerRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		workloads    []RuntimeChartWorkload
		materializer ResourceRequirements
		want         RuntimeChartQuota
	}{
		{
			name: "application container dominates",
			workloads: []RuntimeChartWorkload{{
				Name: testRuntimeQuotaWorkloadName, MaxReplicas: 2,
				Resources: ResourceRequirements{
					Requests: ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity128Mi},
					Limits:   ResourceSet{CPU: testCPUQuantity500m, Memory: testMemoryQuantity512Mi},
				},
			}},
			materializer: ResourceRequirements{
				Requests: ResourceSet{CPU: testCPUQuantity25m, Memory: testMemoryQuantity32Mi},
				Limits:   ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity64Mi},
			},
			want: RuntimeChartQuota{
				Pods: "2", RequestsCPU: testCPUQuantity200m, RequestsMemory: testMemoryQuantity256Mi,
				LimitsCPU: "1000m", LimitsMemory: testMemoryQuantity1Gi,
			},
		},
		{
			name: "sequential init container dominates",
			workloads: []RuntimeChartWorkload{{
				Name: testRuntimeQuotaWorkloadName, MaxReplicas: 2,
				Resources: ResourceRequirements{
					Requests: ResourceSet{CPU: "10m", Memory: "16Mi"},
					Limits:   ResourceSet{CPU: testCPUQuantity50m, Memory: testMemoryQuantity32Mi},
				},
			}},
			materializer: ResourceRequirements{
				Requests: ResourceSet{CPU: testCPUQuantity25m, Memory: testMemoryQuantity32Mi},
				Limits:   ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity64Mi},
			},
			want: RuntimeChartQuota{
				Pods: "2", RequestsCPU: testCPUQuantity50m, RequestsMemory: testMemoryQuantity64Mi,
				LimitsCPU: "200m", LimitsMemory: "128Mi",
			},
		},
		{
			name: "sums different workload replica capacities",
			workloads: []RuntimeChartWorkload{
				{
					Name: "worker-a", MaxReplicas: testLifecycleReservedOrdinals,
					Resources: ResourceRequirements{
						Requests: ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity128Mi},
						Limits:   ResourceSet{CPU: testCPUQuantity500m, Memory: testMemoryQuantity512Mi},
					},
				},
				{
					Name: "worker-b", MaxReplicas: 1,
					Resources: ResourceRequirements{
						Requests: ResourceSet{CPU: "250m", Memory: testMemoryQuantity256Mi},
						Limits:   ResourceSet{CPU: "1", Memory: testMemoryQuantity1Gi},
					},
				},
			},
			materializer: ResourceRequirements{
				Requests: ResourceSet{CPU: testCPUQuantity25m, Memory: testMemoryQuantity32Mi},
				Limits:   ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity64Mi},
			},
			want: RuntimeChartQuota{
				Pods: "4", RequestsCPU: "550m", RequestsMemory: "640Mi",
				LimitsCPU: "2500m", LimitsMemory: "2560Mi",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			quota, err := buildRuntimeChartQuota(test.workloads, PKIMaterializerImage{
				Resources: test.materializer,
			})
			require.NoError(t, err)
			require.Equal(t, test.want, quota)
		})
	}
}

func TestFormatRuntimeMemoryBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int64
		want  string
	}{
		{name: "Ti suffix", value: binaryQuantityTebi, want: "1" + runtimeMemorySuffixTebi},
		{name: "Gi suffix", value: binaryQuantityGibi, want: "1" + runtimeMemorySuffixGibi},
		{name: "Mi suffix", value: binaryQuantityMebi, want: "1" + runtimeMemorySuffixMebi},
		{name: "Ki suffix", value: binaryQuantityKibi, want: "1" + runtimeMemorySuffixKibi},
		{name: "plain bytes", value: 1, want: "1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, formatRuntimeMemoryBytes(test.value))
		})
	}
}

func TestBuildRuntimeChartQuotaRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	validResources := ResourceRequirements{
		Requests: ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity128Mi},
		Limits:   ResourceSet{CPU: testCPUQuantity500m, Memory: testMemoryQuantity512Mi},
	}
	tests := []struct {
		name         string
		workloads    []RuntimeChartWorkload
		materializer ResourceRequirements
		wantError    string
	}{
		{
			name: "missing workloads", materializer: validResources,
			wantError: "runtime quota requires at least one workload",
		},
		{
			name: "zero maximum replicas", materializer: validResources,
			workloads: []RuntimeChartWorkload{{
				Name: testRuntimeQuotaWorkloadName, MaxReplicas: 0, Resources: validResources,
			}},
			wantError: "runtime quota pod count exceeds limit",
		},
		{
			name: "invalid workload resources", materializer: validResources,
			workloads: []RuntimeChartWorkload{{Name: testRuntimeQuotaWorkloadName, MaxReplicas: 1}},
			wantError: "invalid runtime quota resources for " + testRuntimeQuotaWorkloadName,
		},
		{
			name: "invalid init resources",
			workloads: []RuntimeChartWorkload{{
				Name: testRuntimeQuotaWorkloadName, MaxReplicas: 1, Resources: validResources,
			}},
			wantError: "invalid PKI init resources for runtime quota",
		},
		{
			name: "request sum overflows", materializer: validResources,
			workloads: []RuntimeChartWorkload{{
				Name: testRuntimeQuotaWorkloadName, MaxReplicas: 2,
				Resources: ResourceRequirements{
					Requests: ResourceSet{CPU: testCPUQuantityMaxMilli, Memory: testMemoryQuantity128Mi},
					Limits:   ResourceSet{CPU: testCPUQuantityMaxMilli, Memory: testMemoryQuantity512Mi},
				},
			}},
			wantError: "sum runtime quota requests: resource request count or quantity overflows",
		},
		{
			name: "limit sum overflows", materializer: validResources,
			workloads: []RuntimeChartWorkload{{
				Name: testRuntimeQuotaWorkloadName, MaxReplicas: 2,
				Resources: ResourceRequirements{
					Requests: ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity128Mi},
					Limits:   ResourceSet{CPU: testCPUQuantityMaxMilli, Memory: testMemoryQuantity512Mi},
				},
			}},
			wantError: "sum runtime quota limits: resource request count or quantity overflows",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := buildRuntimeChartQuota(test.workloads, PKIMaterializerImage{
				Resources: test.materializer,
			})
			require.ErrorContains(t, err, test.wantError)
		})
	}
}
