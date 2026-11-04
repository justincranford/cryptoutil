// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"strconv"
)

// AddLifecycleRuntimeQuota reserves peak concurrent Jobs alongside maximum app
// replicas. The returned runtime values replace (not supplement) the same budget
// ResourceQuota. Call before rendering or installing either chart.
func AddLifecycleRuntimeQuota(runtime RuntimeChartValues, lifecycle LifecycleChartValues) (RuntimeChartValues, error) {
	if runtime.Run != lifecycle.Run || runtime.Revision != lifecycle.Revision {
		return RuntimeChartValues{}, fmt.Errorf("runtime and lifecycle ownership or revision differs")
	}

	owned := false

	for _, boundary := range lifecycle.Namespaces {
		if boundary.Name == runtime.Namespace && boundary.Release == runtime.HelmRelease {
			owned = true

			break
		}
	}

	if !owned {
		return RuntimeChartValues{}, fmt.Errorf("runtime namespace or release is outside lifecycle ownership")
	}

	if err := validateMaterializerTrustSources(runtime, lifecycle); err != nil {
		return RuntimeChartValues{}, err
	}

	pods, requests, limits, err := lifecycleQuotaAddition(runtime.Namespace, lifecycle)
	if err != nil {
		return RuntimeChartValues{}, err
	}

	basePods, err := strconv.Atoi(runtime.Quota.Pods)
	if err != nil || basePods < 1 || basePods > maxExpandedPods-pods {
		return RuntimeChartValues{}, fmt.Errorf("invalid runtime lifecycle quota pod count")
	}

	baseRequests, baseLimits, err := parseResourceRequirements(ResourceRequirements{
		Requests: ResourceSet{CPU: runtime.Quota.RequestsCPU, Memory: runtime.Quota.RequestsMemory},
		Limits:   ResourceSet{CPU: runtime.Quota.LimitsCPU, Memory: runtime.Quota.LimitsMemory},
	})
	if err != nil {
		return RuntimeChartValues{}, fmt.Errorf("runtime lifecycle quota: %w", err)
	}

	if err := addResourceCapacity(&baseRequests, requests, 1); err != nil {
		return RuntimeChartValues{}, err
	}

	if err := addResourceCapacity(&baseLimits, limits, 1); err != nil {
		return RuntimeChartValues{}, err
	}

	runtime.Quota = RuntimeChartQuota{
		Pods:        strconv.Itoa(basePods + pods),
		RequestsCPU: formatRuntimeCPUMilli(baseRequests.CPUMilli), RequestsMemory: formatRuntimeMemoryBytes(baseRequests.MemoryBytes),
		LimitsCPU: formatRuntimeCPUMilli(baseLimits.CPUMilli), LimitsMemory: formatRuntimeMemoryBytes(baseLimits.MemoryBytes),
	}

	return runtime, nil
}

// AddLifecycleInfrastructureQuota reserves explicit Job resources in the
// existing infrastructure budget without changing replicas or creating another
// quota. It is conservative: all selected Jobs may overlap.
func AddLifecycleInfrastructureQuota(infra InfrastructureChartValues, lifecycle LifecycleChartValues) (InfrastructureChartValues, error) {
	if infra.Run != lifecycle.Run || infra.Revision != lifecycle.Revision {
		return InfrastructureChartValues{}, fmt.Errorf("infrastructure and lifecycle ownership or revision differs")
	}

	runtime, err := AddLifecycleRuntimeQuota(RuntimeChartValues{
		Run: infra.Run, Revision: infra.Revision, Namespace: infra.Namespace, HelmRelease: infra.HelmRelease,
		Quota: RuntimeChartQuota(infra.Quota),
	}, lifecycle)
	if err != nil {
		return InfrastructureChartValues{}, err
	}

	infra.Quota = InfrastructureQuota(runtime.Quota)

	return infra, nil
}

func lifecycleQuotaAddition(namespace string, lifecycle LifecycleChartValues) (int, ResourceCapacity, ResourceCapacity, error) {
	pods := 0

	var requests, limits ResourceCapacity

	for _, job := range lifecycle.Jobs {
		if job.Namespace != namespace {
			continue
		}

		jobRequests, jobLimits, err := parseResourceRequirements(job.Resources)
		if err != nil {
			return 0, ResourceCapacity{}, ResourceCapacity{}, fmt.Errorf("lifecycle Job resources: %w", err)
		}

		if err := addResourceCapacity(&requests, jobRequests, 1); err != nil {
			return 0, ResourceCapacity{}, ResourceCapacity{}, err
		}

		if err := addResourceCapacity(&limits, jobLimits, 1); err != nil {
			return 0, ResourceCapacity{}, ResourceCapacity{}, err
		}

		pods++
	}

	return pods, requests, limits, nil
}
