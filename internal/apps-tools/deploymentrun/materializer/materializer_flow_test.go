// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestMaterializeUsesDefaultAPIClientWithoutNetworkOnCancellation(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	harness := newMaterializerHarness(t, fixture)
	ctx, cancel := context.WithCancel(context.Background())
	materializer := harness.materializer()
	materializer.httpDoer = nil
	readProjectedFile := materializer.readFile
	materializer.readFile = func(name string) ([]byte, error) {
		data, err := readProjectedFile(name)
		if name == cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath {
			cancel()
		}

		return data, err
	}

	err := materializer.Materialize(ctx, harness.configYAML, harness.inputs, harness.trust)
	require.ErrorContains(t, err, testCanceledError)
	require.Empty(t, harness.doer.requests)
	require.NoDirExists(t, harness.output)
}

func TestMaterializeStopsOnRoleSecretErrors(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)

	tests := []struct {
		name          string
		responseIndex int
		wantCalls     int
	}{
		{name: "health client Secret failure", responseIndex: 1, wantCalls: 2},
		{name: "OTLP client Secret failure", responseIndex: 2, wantCalls: testExpectedSecretRequestCount},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			harness := newMaterializerHarness(t, fixture)
			harness.doer.responses[testCase.responseIndex].status = 403
			err := harness.materializer().Materialize(context.Background(), harness.configYAML, harness.inputs, harness.trust)
			require.ErrorContains(t, err, "HTTP 403")
			require.Len(t, harness.doer.requests, testCase.wantCalls)
			require.NoDirExists(t, harness.output)
		})
	}
}

func TestMaterializeRejectsCancellationBeforePublishing(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	harness := newMaterializerHarness(t, fixture)
	ctx, cancel := context.WithCancel(context.Background())
	materializer := harness.materializer()
	materializer.now = func() time.Time {
		cancel()

		return fixture.now
	}

	err := materializer.Materialize(ctx, harness.configYAML, harness.inputs, harness.trust)
	require.ErrorContains(t, err, "canceled before publication")
	require.Len(t, harness.doer.requests, testExpectedSecretRequestCount)
	require.NoDirExists(t, harness.output)
}
