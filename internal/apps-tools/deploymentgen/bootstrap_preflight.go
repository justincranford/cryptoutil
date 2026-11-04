// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	bootstrapNetworkKind       = "network"
	bootstrapVolumeKind        = cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume
	bootstrapContainerKind     = "container"
	bootstrapAllContainersFlag = "--all"
	bootstrapJSONFormat        = "{{json .}}"
	bootstrapListCommand       = "ls"
)

// BootstrapPreflightInput supplies the selected projections and explicit host/K3s reserves.
type BootstrapPreflightInput struct {
	Model                Model
	Lock                 ToolchainLock
	Inventory            Inventory
	RuntimeValues        []RuntimeChartValues
	InfrastructureValues InfrastructureChartValues
	Reserve              ResourceCapacity
	Lifecycle            *LifecycleInput
}

// CheckLocalK3sPreflight admits a fresh Compose bootstrap without creating or deleting anything.
// The caller must provide a deadline and a runner honoring context cancellation. Existing
// run-owned resources are rejected too: this is not an adoption or restart authorization.
// Admission uses one physical host, not summed node capacities. It does not establish
// node placement, pod-slot, rollout headroom, or runtime readiness. All lifecycle
// Jobs are conservatively budgeted as concurrent with maximum application pods.
func CheckLocalK3sPreflight(
	ctx context.Context,
	input BootstrapPreflightInput,
	run DockerCommandRunner,
	listen func(string, string) (net.Listener, error),
) (CapacityEstimate, error) {
	return checkLocalK3sPreflightWithOwnership(ctx, input, run, listen, checkBootstrapDockerOwnership)
}

func checkLocalK3sPreflightWithOwnership(
	ctx context.Context,
	input BootstrapPreflightInput,
	run DockerCommandRunner,
	listen func(string, string) (net.Listener, error),
	ownership func(context.Context, LocalK3sPlan, DockerCommandRunner) error,
) (CapacityEstimate, error) {
	if ctx == nil || run == nil || listen == nil {
		return CapacityEstimate{}, fmt.Errorf("bootstrap preflight context, runner, and listener are required")
	}

	if _, bounded := ctx.Deadline(); !bounded {
		return CapacityEstimate{}, fmt.Errorf("bootstrap preflight requires a context deadline")
	}

	if err := ctx.Err(); err != nil {
		return CapacityEstimate{}, fmt.Errorf("bootstrap preflight canceled: %w", err)
	}

	plan, err := BuildLocalK3sPlan(input.Model, input.Lock)
	if err != nil {
		return CapacityEstimate{}, err
	}

	if input.Inventory.Run != plan.Run {
		return CapacityEstimate{}, fmt.Errorf("bootstrap and inventory run identities differ")
	}

	estimate, err := EstimateLifecycleCapacity(input)
	if err != nil {
		return CapacityEstimate{}, err
	}

	if input.Reserve.CPUMilli <= 0 || input.Reserve.MemoryBytes <= 0 {
		return CapacityEstimate{}, fmt.Errorf("explicit positive host and K3s resource reserves are required")
	}

	if err := ownership(ctx, plan, run); err != nil {
		return CapacityEstimate{}, err
	}

	host, err := ProbeDockerHostCapacity(ctx, run)
	if err != nil {
		return CapacityEstimate{}, err
	}

	if err := CheckHostCapacity(estimate.TotalRequests, host, input.Reserve); err != nil {
		return CapacityEstimate{}, err
	}

	if err := ctx.Err(); err != nil {
		return CapacityEstimate{}, fmt.Errorf("bootstrap preflight canceled: %w", err)
	}

	if err := CheckBootstrapPorts(input.Model.Bootstrap, listen); err != nil {
		return CapacityEstimate{}, err
	}

	if err := ctx.Err(); err != nil {
		return CapacityEstimate{}, fmt.Errorf("bootstrap preflight canceled: %w", err)
	}

	return estimate, nil
}

func checkBootstrapDockerOwnership(ctx context.Context, plan LocalK3sPlan, run DockerCommandRunner) error {
	nodes := append([]LocalK3sNode{plan.Server}, plan.Agents...)
	containers := make(map[string]bool, len(nodes))
	volumes := make(map[string]bool, len(nodes))

	for _, node := range nodes {
		containers[node.Container] = true
		volumes[node.Volume] = true
	}

	probes := []struct {
		kind  string
		args  []string
		names map[string]bool
	}{
		{kind: bootstrapContainerKind, args: []string{bootstrapContainerKind, bootstrapListCommand, bootstrapAllContainersFlag, dockerOutputFormatFlag, bootstrapJSONFormat}, names: containers},
		{kind: bootstrapNetworkKind, args: []string{bootstrapNetworkKind, bootstrapListCommand, dockerOutputFormatFlag, bootstrapJSONFormat}, names: map[string]bool{plan.Network: true}},
		{kind: bootstrapVolumeKind, args: []string{bootstrapVolumeKind, bootstrapListCommand, dockerOutputFormatFlag, bootstrapJSONFormat}, names: volumes},
	}

	for _, probe := range probes {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("bootstrap ownership probe canceled: %w", err)
		}

		output, err := run(ctx, "docker", probe.args...)
		if err != nil {
			return fmt.Errorf("read bootstrap %s ownership: %w", probe.kind, err)
		}

		decoder := json.NewDecoder(bytes.NewReader(output))

		for {
			var resource struct {
				Name   string
				Names  string
				Labels string
			}

			if err := decoder.Decode(&resource); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return fmt.Errorf("decode bootstrap %s ownership: %w", probe.kind, err)
			}

			name := resource.Name
			if probe.kind == bootstrapContainerKind {
				name = resource.Names
			}

			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("bootstrap %s ownership record has no name", probe.kind)
			}

			if probe.names[name] {
				return fmt.Errorf("bootstrap %s name already exists: %s", probe.kind, name)
			}

			for _, label := range strings.Split(resource.Labels, ",") {
				if label == "com.docker.compose.project="+plan.Project {
					return fmt.Errorf("bootstrap Compose project already owns %s: %s", probe.kind, name)
				}
			}
		}
	}

	return nil
}
