// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	dockerOutputFormatFlag = "--format"
	dockerInfoCommand      = "info"
	dockerHostFormat       = "{{.NCPU}} {{.MemTotal}}"
)

// ResourceCapacity contains physical CPU in millicores and memory in bytes.
type ResourceCapacity struct {
	CPUMilli    int64
	MemoryBytes int64
}

// CapacityEstimate is a conservative maximum-replica request estimate.
// It is a host-total admission preflight, not a per-node scheduler simulation.
type CapacityEstimate struct {
	ApplicationPods        int
	InfrastructurePods     int
	ApplicationRequests    ResourceCapacity
	InfrastructureRequests ResourceCapacity
	TotalRequests          ResourceCapacity
	LifecyclePods          int
	LifecycleRequests      ResourceCapacity
}

// DockerCommandRunner executes one bounded command and returns its combined output.
type DockerCommandRunner func(context.Context, string, ...string) ([]byte, error)

// EstimateCapacity validates projection completeness and totals the maximum app-pod and
// enabled infrastructure requests. For each app Pod, Kubernetes schedules the greater of
// the app-container and sequential PKI-init-container requests, not their sum.
func EstimateCapacity(
	inventory Inventory,
	runtimeValues []RuntimeChartValues,
	infrastructureValues InfrastructureChartValues,
) (CapacityEstimate, error) {
	if err := validateInventory(inventory); err != nil {
		return CapacityEstimate{}, fmt.Errorf("invalid expected inventory: %w", err)
	}

	if len(runtimeValues) != len(inventory.Releases) || len(infrastructureValues.Clients) != len(inventory.Releases) {
		return CapacityEstimate{}, fmt.Errorf("runtime and infrastructure projections must cover every selected release")
	}

	if infrastructureValues.Run != inventory.Run ||
		infrastructureValues.Namespace != inventory.InfraNamespace ||
		infrastructureValues.HelmRelease != "cu-"+inventory.Run+"-infra" {
		return CapacityEstimate{}, fmt.Errorf("infrastructure projection identity differs from expected inventory")
	}

	runtimeByKey := make(map[string]RuntimeChartValues, len(runtimeValues))
	for _, values := range runtimeValues {
		runtimeByKey[values.ReleaseKey] = values
	}

	clients := make(map[string]InfrastructureClient, len(infrastructureValues.Clients))
	for _, client := range infrastructureValues.Clients {
		key := client.Namespace + "/" + client.Release
		clients[key] = client
	}

	estimate := CapacityEstimate{}
	hasPostgres := false

	for _, release := range inventory.Releases {
		values, exists := runtimeByKey[release.Key]
		if !exists {
			return CapacityEstimate{}, fmt.Errorf("runtime projection missing: %s", release.Key)
		}

		releaseHasPostgres := len(release.Databases) > 0
		if releaseHasPostgres {
			hasPostgres = true
		}

		if values.Run != inventory.Run || values.ReleaseKey != release.Key ||
			values.Namespace != release.Namespace || values.HelmRelease != release.Helm ||
			values.HasPostgres != releaseHasPostgres {
			return CapacityEstimate{}, fmt.Errorf("runtime projection identity or PostgreSQL setting differs: %s", release.Key)
		}

		clientKey := release.Namespace + "/" + release.Helm

		client, exists := clients[clientKey]
		if !exists || client.Postgres != releaseHasPostgres {
			return CapacityEstimate{}, fmt.Errorf("infrastructure client projection differs: %s", clientKey)
		}

		initRequests, _, err := parseResourceRequirements(values.PKI.Resources)
		if err != nil {
			return CapacityEstimate{}, fmt.Errorf("invalid PKI init resources for %s: %w", release.Key, err)
		}

		workloads := make(map[string]Workload, len(release.Workloads))
		for _, workload := range release.Workloads {
			workloads[workload.Name] = workload
		}

		if len(values.Workloads) != len(workloads) {
			return CapacityEstimate{}, fmt.Errorf("runtime workload projection count differs: %s", release.Key)
		}

		seenWorkloads := make(map[string]bool, len(values.Workloads))
		for _, workloadValues := range values.Workloads {
			if seenWorkloads[workloadValues.Name] {
				return CapacityEstimate{}, fmt.Errorf("duplicate runtime workload projection: %s/%s", release.Key, workloadValues.Name)
			}

			seenWorkloads[workloadValues.Name] = true

			workload, exists := workloads[workloadValues.Name]
			if !exists || workload.PSID != workloadValues.PSID || workload.Pool != workloadValues.Pool ||
				workload.Bounds.Min != workloadValues.MinReplicas || workload.Bounds.Max != workloadValues.MaxReplicas {
				return CapacityEstimate{}, fmt.Errorf("runtime workload projection differs: %s/%s", release.Key, workloadValues.Name)
			}

			appRequests, _, err := parseResourceRequirements(workloadValues.Resources)
			if err != nil {
				return CapacityEstimate{}, fmt.Errorf("invalid runtime resources for %s/%s: %w", release.Key, workloadValues.Name, err)
			}

			podRequests := ResourceCapacity{
				CPUMilli:    max(appRequests.CPUMilli, initRequests.CPUMilli),
				MemoryBytes: max(appRequests.MemoryBytes, initRequests.MemoryBytes),
			}
			if err := addResourceCapacity(&estimate.ApplicationRequests, podRequests, workloadValues.MaxReplicas); err != nil {
				return CapacityEstimate{}, fmt.Errorf("sum runtime requests: %w", err)
			}

			estimate.ApplicationPods += workloadValues.MaxReplicas
		}
	}

	components := []struct {
		name        string
		component   InfrastructureComponent
		mustBeOn    bool
		expectedRep int
	}{
		{
			name: "postgres leader", component: infrastructureValues.Components.PostgresLeader,
			mustBeOn:    hasPostgres,
			expectedRep: 1,
		},
		{
			name: "postgres follower", component: infrastructureValues.Components.PostgresFollower,
			mustBeOn:    hasPostgres,
			expectedRep: 1,
		},
		{
			name: "collector", component: infrastructureValues.Components.Collector,
			mustBeOn: true, expectedRep: 2,
		},
		{
			name: "LGTM", component: infrastructureValues.Components.LGTM,
			mustBeOn: true, expectedRep: 1,
		},
	}

	for _, entry := range components {
		if entry.component.Enabled != entry.mustBeOn {
			return CapacityEstimate{}, fmt.Errorf("%s enabled state differs from selected inventory", entry.name)
		}

		if !entry.mustBeOn {
			continue
		}

		if entry.component.Replicas != entry.expectedRep {
			return CapacityEstimate{}, fmt.Errorf("%s replicas must equal %d", entry.name, entry.expectedRep)
		}

		requests, _, err := parseResourceRequirements(entry.component.Resources)
		if err != nil {
			return CapacityEstimate{}, fmt.Errorf("invalid %s resources: %w", entry.name, err)
		}

		if err := addResourceCapacity(&estimate.InfrastructureRequests, requests, entry.component.Replicas); err != nil {
			return CapacityEstimate{}, fmt.Errorf("sum infrastructure requests: %w", err)
		}

		estimate.InfrastructurePods += entry.component.Replicas
	}

	estimate.TotalRequests = estimate.ApplicationRequests
	if err := addResourceCapacity(&estimate.TotalRequests, estimate.InfrastructureRequests, 1); err != nil {
		return CapacityEstimate{}, fmt.Errorf("sum total resource requests: %w", err)
	}

	return estimate, nil
}

// CheckHostCapacity rejects a request budget that cannot fit on the measured Docker host after
// explicit host/K3s reserves. The reserve must be supplied by the caller; zero is not inferred.
func CheckHostCapacity(required, host, reserve ResourceCapacity) error {
	if host.CPUMilli <= 0 || host.MemoryBytes <= 0 {
		return fmt.Errorf("measured Docker host CPU and memory capacity are required")
	}

	if reserve.CPUMilli <= 0 || reserve.MemoryBytes <= 0 {
		return fmt.Errorf("explicit positive host and K3s resource reserves are required")
	}

	if required.CPUMilli < 0 || required.MemoryBytes < 0 ||
		reserve.CPUMilli >= host.CPUMilli || reserve.MemoryBytes >= host.MemoryBytes {
		return fmt.Errorf("required resources and explicit reserves must be nonnegative and below host capacity")
	}

	available := ResourceCapacity{
		CPUMilli: host.CPUMilli - reserve.CPUMilli, MemoryBytes: host.MemoryBytes - reserve.MemoryBytes,
	}
	if required.CPUMilli > available.CPUMilli || required.MemoryBytes > available.MemoryBytes {
		return fmt.Errorf(
			"maximum requested resources exceed Docker host capacity after reserves: required=%dm/%d bytes available=%dm/%d bytes",
			required.CPUMilli, required.MemoryBytes, available.CPUMilli, available.MemoryBytes,
		)
	}

	return nil
}

// ProbeDockerHostCapacity reads the Docker daemon's single physical-host CPU and memory values.
// K3s nodes are not summed as independent copies of this host capacity.
func ProbeDockerHostCapacity(ctx context.Context, run DockerCommandRunner) (ResourceCapacity, error) {
	if ctx == nil || run == nil {
		return ResourceCapacity{}, fmt.Errorf("docker capacity probe context and command runner are required")
	}

	output, err := run(ctx, "docker", dockerInfoCommand, dockerOutputFormatFlag, dockerHostFormat)
	if err != nil {
		return ResourceCapacity{}, fmt.Errorf("read Docker host capacity: %w", err)
	}

	fields := strings.Fields(string(output))
	if len(fields) != 2 {
		return ResourceCapacity{}, fmt.Errorf("docker host capacity output must contain CPU and memory values")
	}

	cpuCount, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || cpuCount <= 0 || cpuCount > math.MaxInt64/cpuMilliUnitsPerCore {
		return ResourceCapacity{}, fmt.Errorf("docker host CPU capacity is invalid: %q", fields[0])
	}

	memoryBytes, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || memoryBytes <= 0 {
		return ResourceCapacity{}, fmt.Errorf("docker host memory capacity is invalid: %q", fields[1])
	}

	return ResourceCapacity{CPUMilli: cpuCount * cpuMilliUnitsPerCore, MemoryBytes: memoryBytes}, nil
}

func addResourceCapacity(total *ResourceCapacity, value ResourceCapacity, count int) error {
	if count < 0 || value.CPUMilli < 0 || value.MemoryBytes < 0 ||
		int64(count) > 0 && value.CPUMilli > math.MaxInt64/int64(count) ||
		int64(count) > 0 && value.MemoryBytes > math.MaxInt64/int64(count) {
		return fmt.Errorf("resource request count or quantity overflows")
	}

	cpuTotal := value.CPUMilli * int64(count)

	memoryTotal := value.MemoryBytes * int64(count)
	if total.CPUMilli > math.MaxInt64-cpuTotal || total.MemoryBytes > math.MaxInt64-memoryTotal {
		return fmt.Errorf("resource request sum overflows")
	}

	total.CPUMilli += cpuTotal
	total.MemoryBytes += memoryTotal

	return nil
}
