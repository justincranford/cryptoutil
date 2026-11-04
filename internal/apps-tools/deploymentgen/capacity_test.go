// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestEstimateCapacityFullMatrixRejectsInsufficientDockerHost(t *testing.T) {
	t.Parallel()

	inventory, runtimeValues, infrastructureValues := capacityFixture(t, "")
	estimate, err := EstimateCapacity(inventory, runtimeValues, infrastructureValues)
	require.NoError(t, err)
	require.Equal(t, 294, estimate.ApplicationPods)
	require.Equal(t, 5, estimate.InfrastructurePods)
	require.Equal(t, int64(29_400), estimate.ApplicationRequests.CPUMilli)
	require.Equal(t, int64(294*128*1024*1024), estimate.ApplicationRequests.MemoryBytes)

	err = CheckHostCapacity(
		estimate.TotalRequests,
		ResourceCapacity{CPUMilli: 14_000, MemoryBytes: 8_326_410_240},
		ResourceCapacity{CPUMilli: 1_000, MemoryBytes: 1 << 30},
	)
	require.ErrorContains(t, err, "maximum requested resources exceed Docker host capacity")
}

func TestEstimateCapacitySelectedPGReleaseFitsMeasuredHost(t *testing.T) {
	t.Parallel()

	inventory, runtimeValues, infrastructureValues := capacityFixture(t, "s:sm-kms:pg")
	estimate, err := EstimateCapacity(inventory, runtimeValues, infrastructureValues)
	require.NoError(t, err)
	require.Equal(t, 2, estimate.ApplicationPods)
	require.Equal(t, 5, estimate.InfrastructurePods)

	require.NoError(t, CheckHostCapacity(
		estimate.TotalRequests,
		ResourceCapacity{CPUMilli: 14_000, MemoryBytes: 8_326_410_240},
		ResourceCapacity{CPUMilli: 1_000, MemoryBytes: 3 << 30},
	))
}

func TestEstimateCapacityUsesMaximumReplicaBounds(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	model.Overrides["sm-kms/disk"] = Bounds{Min: 1, Max: testThree}
	inventory, err := build(model, scopes, "s:sm-kms:disk")
	require.NoError(t, err)
	runtimeValues, infrastructureValues := capacityProjections(t, model, scopes, inventory)

	estimate, err := EstimateCapacity(inventory, runtimeValues, infrastructureValues)
	require.NoError(t, err)
	require.Equal(t, 3, estimate.ApplicationPods)
	require.Equal(t, int64(3*128*1024*1024), estimate.ApplicationRequests.MemoryBytes)
	require.Equal(t, 3, estimate.InfrastructurePods)
}

func TestEstimateCapacityRejectsInconsistentProjections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		selector string
		edit     func(*Inventory, *[]RuntimeChartValues, *InfrastructureChartValues)
		wantErr  string
	}{
		{
			name: testInvalidExpectedInventoryCaseName,
			edit: func(inventory *Inventory, _ *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				inventory.Releases[0].Helm = ""
			},
			wantErr: testInvalidExpectedInventoryCaseName,
		},
		{
			name: "missing runtime release",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				*runtimeValues = (*runtimeValues)[:0]
			},
			wantErr: "must cover every selected release",
		},
		{
			name: "runtime projection does not cover expected release",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].ReleaseKey = testUnknownReleaseKey
			},
			wantErr: "runtime projection missing",
		},
		{
			name: "runtime projection identity mismatch",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].Run = "wrong-run"
			},
			wantErr: "runtime projection identity or PostgreSQL setting differs",
		},
		{
			name: "infrastructure projection omits selected client",
			edit: func(_ *Inventory, _ *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				infra.Clients[0].Release = testUnknownReleaseKey
			},
			wantErr: testInfrastructureClientProjectionMismatch,
		},
		{
			name: "workload replica mismatch",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].Workloads[0].MaxReplicas++
			},
			wantErr: "runtime workload projection differs",
		},
		{
			name:     "duplicate workload identity",
			selector: testMixedMemorySelector,
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].Workloads[1].Name = (*runtimeValues)[0].Workloads[0].Name
			},
			wantErr: "duplicate runtime workload projection",
		},
		{
			name: "workload projection count mismatch",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].Workloads = append((*runtimeValues)[0].Workloads, (*runtimeValues)[0].Workloads[0])
			},
			wantErr: "runtime workload projection count differs",
		},
		{
			name: "stale postgres component for sqlite-only release",
			edit: func(_ *Inventory, _ *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				infra.Components.PostgresLeader.Enabled = true
			},
			wantErr: "postgres leader enabled state differs",
		},
		{
			name: "infrastructure replica mismatch",
			edit: func(_ *Inventory, _ *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				infra.Components.Collector.Replicas++
			},
			wantErr: "collector replicas must equal 2",
		},
		{
			name: "wrong infrastructure client backend",
			edit: func(_ *Inventory, _ *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				infra.Clients[0].Postgres = true
			},
			wantErr: testInfrastructureClientProjectionMismatch,
		},
		{
			name: "wrong infrastructure identity",
			edit: func(_ *Inventory, _ *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				infra.Namespace = "wrong"
			},
			wantErr: "infrastructure projection identity differs",
		},
		{
			name: "invalid runtime request",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].Workloads[0].Resources.Requests.Memory = cryptoutilSharedMagic.DevSetupTestInvalidName
			},
			wantErr: testInvalidRuntimeResourcesError,
		},
		{
			name: "invalid runtime limit",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].Workloads[0].Resources.Limits.CPU = cryptoutilSharedMagic.DevSetupTestInvalidName
			},
			wantErr: testInvalidRuntimeResourcesError,
		},
		{
			name: testInvalidPKIInitResourcesError,
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				(*runtimeValues)[0].PKI.Resources.Requests.CPU = cryptoutilSharedMagic.DevSetupTestInvalidName
			},
			wantErr: testInvalidPKIInitResourcesError,
		},
		{
			name: "invalid infrastructure resources",
			edit: func(_ *Inventory, _ *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				infra.Components.Collector.Resources.Requests.CPU = cryptoutilSharedMagic.DevSetupTestInvalidName
			},
			wantErr: "invalid collector resources",
		},
		{
			name: "application request multiplication overflow",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, _ *InfrastructureChartValues) {
				workload := &(*runtimeValues)[0].Workloads[0]
				workload.Resources.Requests.Memory = "5000000000000000000"
				workload.Resources.Limits.Memory = "5000000000000000000"
			},
			wantErr: "sum runtime requests: resource request count or quantity overflows",
		},
		{
			name: "infrastructure request multiplication overflow",
			edit: func(_ *Inventory, _ *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				infra.Components.Collector.Resources.Requests.Memory = "5000000000000000000"
				infra.Components.Collector.Resources.Limits.Memory = "5000000000000000000"
			},
			wantErr: "sum infrastructure requests: resource request count or quantity overflows",
		},
		{
			name: "total request addition overflow",
			edit: func(_ *Inventory, runtimeValues *[]RuntimeChartValues, infra *InfrastructureChartValues) {
				workload := &(*runtimeValues)[0].Workloads[0]
				workload.Resources.Requests.Memory = "4600000000000000000"
				workload.Resources.Limits.Memory = "4600000000000000000"
				infra.Components.Collector.Resources.Requests.Memory = "100000000000000000"
				infra.Components.Collector.Resources.Limits.Memory = "100000000000000000"
			},
			wantErr: "sum total resource requests: resource request sum overflows",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			selector := test.selector
			if selector == "" {
				selector = testServiceSMKMSMemSelector
			}

			inventory, runtimeValues, infrastructureValues := capacityFixture(t, selector)
			test.edit(&inventory, &runtimeValues, &infrastructureValues)

			_, err := EstimateCapacity(inventory, runtimeValues, infrastructureValues)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestCheckHostCapacityRejectsMissingOrInsufficientMeasurements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		required ResourceCapacity
		host     ResourceCapacity
		reserve  ResourceCapacity
		wantErr  string
	}{
		{
			name:     "missing physical host measurement",
			required: ResourceCapacity{CPUMilli: 1, MemoryBytes: 1},
			reserve:  ResourceCapacity{CPUMilli: 1, MemoryBytes: 1},
			wantErr:  "measured Docker host",
		},
		{
			name:     "missing explicit reserve",
			required: ResourceCapacity{CPUMilli: 1, MemoryBytes: 1},
			host:     ResourceCapacity{CPUMilli: 1_000, MemoryBytes: 1 << 30},
			wantErr:  "explicit positive",
		},
		{
			name:     "reserve consumes host",
			required: ResourceCapacity{CPUMilli: 1, MemoryBytes: 1},
			host:     ResourceCapacity{CPUMilli: 1_000, MemoryBytes: 1 << 30},
			reserve:  ResourceCapacity{CPUMilli: 1_000, MemoryBytes: 1},
			wantErr:  "below host capacity",
		},
		{
			name:     "request fits reserved budget",
			required: ResourceCapacity{CPUMilli: 100, MemoryBytes: 1 << 20},
			host:     ResourceCapacity{CPUMilli: 2_000, MemoryBytes: 2 << 30},
			reserve:  ResourceCapacity{CPUMilli: 500, MemoryBytes: 1 << 30},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := CheckHostCapacity(test.required, test.host, test.reserve)
			if test.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestProbeDockerHostCapacity(t *testing.T) {
	t.Parallel()

	injected := errors.New("injected Docker info failure")
	tests := []struct {
		name      string
		output    string
		runError  error
		nilRunner bool
		want      ResourceCapacity
		wantErr   string
	}{
		{
			name:   "physical host values",
			output: "14 8326410240\n",
			want:   ResourceCapacity{CPUMilli: 14_000, MemoryBytes: 8_326_410_240},
		},
		{name: "daemon error", runError: injected, wantErr: "read Docker host capacity"},
		{name: "malformed output", output: "fourteen", wantErr: "must contain CPU and memory"},
		{name: "zero CPU", output: "0 1024", wantErr: "CPU capacity is invalid"},
		{name: "invalid memory", output: "4 invalid", wantErr: "memory capacity is invalid"},
		{name: "missing runner", nilRunner: true, wantErr: "context and command runner are required"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var runner DockerCommandRunner = func(_ context.Context, name string, args ...string) ([]byte, error) {
				require.Equal(t, "docker", name)
				require.Equal(t, []string{dockerInfoCommand, dockerOutputFormatFlag, dockerHostFormat}, args)

				return []byte(test.output), test.runError
			}
			if test.nilRunner {
				runner = nil
			}

			got, err := ProbeDockerHostCapacity(context.Background(), runner)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)

				if test.runError != nil {
					require.ErrorIs(t, err, test.runError)
				}

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}

	var nilContext context.Context

	_, err := ProbeDockerHostCapacity(nilContext, func(context.Context, string, ...string) ([]byte, error) {
		return nil, nil
	})
	require.ErrorContains(t, err, "context and command runner are required")
}

func TestAddResourceCapacityRejectsInvalidOrOverflowingInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		total   ResourceCapacity
		value   ResourceCapacity
		count   int
		wantErr string
	}{
		{name: "negative count", count: -1, wantErr: testResourceCountOrQuantityOverflowError},
		{name: "negative quantity", value: ResourceCapacity{CPUMilli: -1}, count: 1, wantErr: testResourceCountOrQuantityOverflowError},
		{
			name:  "multiplication overflow",
			value: ResourceCapacity{CPUMilli: math.MaxInt64},
			count: 2, wantErr: testResourceCountOrQuantityOverflowError,
		},
		{
			name:  "addition overflow",
			total: ResourceCapacity{CPUMilli: math.MaxInt64},
			value: ResourceCapacity{CPUMilli: 1},
			count: 1, wantErr: "resource request sum overflows",
		},
		{
			name:  "valid accumulation",
			total: ResourceCapacity{CPUMilli: 100},
			value: ResourceCapacity{CPUMilli: 200, MemoryBytes: 1024},
			count: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			total := test.total

			err := addResourceCapacity(&total, test.value, test.count)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, ResourceCapacity{CPUMilli: 500, MemoryBytes: 2048}, total)
		})
	}
}

func TestParseCPUResourceRejectsMilliCoreOverflow(t *testing.T) {
	t.Parallel()

	_, err := parseCPUResource("9223372036854776")
	require.ErrorContains(t, err, "CPU quantity")
}

func capacityFixture(t *testing.T, selector string) (Inventory, []RuntimeChartValues, InfrastructureChartValues) {
	t.Helper()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, selector)
	require.NoError(t, err)
	runtimeValues, infrastructureValues := capacityProjections(t, model, scopes, inventory)

	return inventory, runtimeValues, infrastructureValues
}

func capacityProjections(
	t *testing.T,
	model Model,
	scopes []Scope,
	inventory Inventory,
) ([]RuntimeChartValues, InfrastructureChartValues) {
	t.Helper()

	runtimeValues := make([]RuntimeChartValues, 0, len(inventory.Releases))
	images := chartImageFixture(scopes)

	for _, release := range inventory.Releases {
		selected := inventoryWithRelease(inventory, release.Key)
		values, err := BuildRuntimeChartValues(
			model, selected, release.Key, images, chartMaterializerFixture(),
			ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
			chartRuntimeInputFixture(selected, release.Key),
		)
		require.NoError(t, err, release.Key)

		runtimeValues = append(runtimeValues, values)
	}

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)
	infrastructureValues, err := BuildInfrastructureChartValues(model, inventory, defaults)
	require.NoError(t, err)

	return runtimeValues, infrastructureValues
}
