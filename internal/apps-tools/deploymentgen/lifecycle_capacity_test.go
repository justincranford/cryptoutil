// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	"errors"
	"math"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testLifecycleOldRevision = "r0"

func TestLifecycleAdmissionBudgetsAllConcurrentJobs(t *testing.T) {
	t.Parallel()

	for _, selector := range []string{testServiceSMKMSPGSelector, testServiceSMKMSMemSelector, ""} {
		t.Run(selector, func(t *testing.T) {
			t.Parallel()

			input, _ := bootstrapPreflightFixture(t, selector)
			base, err := EstimateCapacity(input.Inventory, input.RuntimeValues, input.InfrastructureValues)
			require.NoError(t, err)
			got, err := EstimateLifecycleCapacity(input)
			require.NoError(t, err)
			require.Equal(t, len(input.Lifecycle.Jobs), got.LifecyclePods)
			require.Positive(t, got.LifecycleRequests.CPUMilli)
			require.Positive(t, got.LifecycleRequests.MemoryBytes)
			require.Equal(t, base.TotalRequests.CPUMilli+got.LifecycleRequests.CPUMilli, got.TotalRequests.CPUMilli)
			require.Equal(t, base.TotalRequests.MemoryBytes+got.LifecycleRequests.MemoryBytes, got.TotalRequests.MemoryBytes)
			require.Equal(t, base.ApplicationPods, got.ApplicationPods)
		})
	}
}

func TestLifecycleAdmissionRejectsBeforePhysicalProbes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*testing.T, *BootstrapPreflightInput)
		want string
	}{
		{"missing contract", func(_ *testing.T, v *BootstrapPreflightInput) { v.Lifecycle = nil }, "explicit lifecycle"},
		{"mismatched revision", func(_ *testing.T, v *BootstrapPreflightInput) { v.Model.Revision = testLifecycleOldRevision }, "content-addressed model revision"},
		{"missing Jobs", func(t *testing.T, v *BootstrapPreflightInput) {
			v.Lifecycle.Jobs = nil

			var err error

			v.Model.Revision, err = LifecycleRevision(*v.Lifecycle)
			require.NoError(t, err)
			v.Inventory, err = build(v.Model, RegisteredScopes(), testServiceSMKMSMemSelector)
			require.NoError(t, err)
		}, "every graph"},
		{"runtime revision mismatch", func(_ *testing.T, v *BootstrapPreflightInput) { v.RuntimeValues[0].Revision = testLifecycleOldRevision }, lifecycleRevisionFormat},
		{"infra revision mismatch", func(_ *testing.T, v *BootstrapPreflightInput) {
			v.InfrastructureValues.Revision = testLifecycleOldRevision
		}, lifecycleRevisionFormat},
		{"request overflow", func(t *testing.T, v *BootstrapPreflightInput) {
			for index := range v.Lifecycle.Jobs {
				v.Lifecycle.Jobs[index].Resources.Requests.CPU = testCPUQuantityMaxMilli
				v.Lifecycle.Jobs[index].Resources.Limits.CPU = testCPUQuantityMaxMilli
			}

			var err error

			v.Model.Revision, err = LifecycleRevision(*v.Lifecycle)
			require.NoError(t, err)
			v.Inventory, err = build(v.Model, RegisteredScopes(), testServiceSMKMSMemSelector)
			require.NoError(t, err)

			for index := range v.RuntimeValues {
				v.RuntimeValues[index].Revision = v.Model.Revision
			}

			v.InfrastructureValues.Revision = v.Model.Revision
		}, "overflow"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			input, _ := bootstrapPreflightFixture(t, testServiceSMKMSMemSelector)
			test.edit(t, &input)

			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()

			_, err := CheckLocalK3sPreflight(ctx, input, func(context.Context, string, ...string) ([]byte, error) {
				t.Fatal("invalid lifecycle must fail before Docker")

				return nil, errors.New("unexpected Docker")
			}, func(string, string) (net.Listener, error) {
				t.Fatal("invalid lifecycle must fail before ports")

				return nil, errors.New("unexpected port")
			})
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestLifecycleAdmissionRejectsHostThatOnlyFitsApps(t *testing.T) {
	t.Parallel()

	input, _ := bootstrapPreflightFixture(t, testServiceSMKMSPGSelector)
	base, err := EstimateCapacity(input.Inventory, input.RuntimeValues, input.InfrastructureValues)
	require.NoError(t, err)
	got, err := EstimateLifecycleCapacity(input)
	require.NoError(t, err)

	host := ResourceCapacity{
		CPUMilli:    base.TotalRequests.CPUMilli + input.Reserve.CPUMilli,
		MemoryBytes: base.TotalRequests.MemoryBytes + input.Reserve.MemoryBytes,
	}
	require.NoError(t, CheckHostCapacity(base.TotalRequests, host, input.Reserve))
	require.Error(t, CheckHostCapacity(got.TotalRequests, host, input.Reserve))
}

func TestLifecycleAdmissionChecksCrossNamespaceOverflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, selector, want string
		divisor              int64
	}{
		{"Job aggregate overflow", testServiceSMKMSPGSelector, "sum admission Job", 4},
		{"host aggregate overflow", testServiceSMKMSMemSelector, "sum admission total", 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			input, _ := bootstrapPreflightFixture(t, test.selector)
			for index := range input.Lifecycle.Jobs {
				quantity := strconv.FormatInt(math.MaxInt64/test.divisor, 10) + "m"
				input.Lifecycle.Jobs[index].Resources.Requests.CPU = quantity
				input.Lifecycle.Jobs[index].Resources.Limits.CPU = quantity
			}

			var err error

			input.Model.Revision, err = LifecycleRevision(*input.Lifecycle)
			require.NoError(t, err)
			input.Inventory, err = build(input.Model, RegisteredScopes(), test.selector)
			require.NoError(t, err)

			for index := range input.RuntimeValues {
				input.RuntimeValues[index].Revision = input.Model.Revision
			}

			input.InfrastructureValues.Revision = input.Model.Revision
			_, err = EstimateLifecycleCapacity(input)
			require.ErrorContains(t, err, test.want)
		})
	}
}
