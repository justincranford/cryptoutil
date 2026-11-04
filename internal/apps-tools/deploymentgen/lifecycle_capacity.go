// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import "fmt"

// EstimateLifecycleCapacity validates explicit lifecycle ownership and sums all
// selected Job requests alongside maximum app/init and infrastructure requests.
// Quota headroom and host headroom are distinct; quotas alone do not reserve host
// capacity. No DAG scheduling assumption is used to discount concurrent Jobs.
func EstimateLifecycleCapacity(input BootstrapPreflightInput) (CapacityEstimate, error) {
	if input.Lifecycle == nil {
		return CapacityEstimate{}, fmt.Errorf("explicit lifecycle input is required for admission")
	}

	lifecycle, err := BuildLifecycleChartValues(input.Model, input.Inventory, *input.Lifecycle)
	if err != nil {
		return CapacityEstimate{}, fmt.Errorf("admission lifecycle contract: %w", err)
	}

	for _, runtime := range input.RuntimeValues {
		if _, err := AddLifecycleRuntimeQuota(runtime, lifecycle); err != nil {
			return CapacityEstimate{}, fmt.Errorf("admission runtime lifecycle quota: %w", err)
		}
	}

	if _, err := AddLifecycleInfrastructureQuota(input.InfrastructureValues, lifecycle); err != nil {
		return CapacityEstimate{}, fmt.Errorf("admission infrastructure lifecycle quota: %w", err)
	}

	estimate, err := EstimateCapacity(input.Inventory, input.RuntimeValues, input.InfrastructureValues)
	if err != nil {
		return CapacityEstimate{}, err
	}

	for _, job := range lifecycle.Jobs {
		requests, _, err := parseResourceRequirements(job.Resources)
		if err != nil {
			return CapacityEstimate{}, fmt.Errorf("admission Job resources: %w", err)
		}

		if err := addResourceCapacity(&estimate.LifecycleRequests, requests, 1); err != nil {
			return CapacityEstimate{}, fmt.Errorf("sum admission Job requests: %w", err)
		}

		estimate.LifecyclePods++
	}

	if err := addResourceCapacity(&estimate.TotalRequests, estimate.LifecycleRequests, 1); err != nil {
		return CapacityEstimate{}, fmt.Errorf("sum admission total requests: %w", err)
	}

	return estimate, nil
}
