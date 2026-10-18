// Copyright (c) 2025-2026 Justin Cranford.

package pool

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type panicErrorContext struct {
	context.Context
}

func (panicErrorContext) Err() error {
	panic("injected context panic")
}

func TestGenerateWorker_RecoversUnexpectedPanic(t *testing.T) {
	t.Parallel()

	poolInstance := &ValueGenPool[googleUuid.UUID]{
		cfg: &ValueGenPoolConfig[googleUuid.UUID]{
			telemetryService: testTelemetryService,
			poolName:         "worker-panic-recovery",
		},
		stopGeneratingCtx: panicErrorContext{Context: context.Background()},
	}

	require.NotPanics(t, func() {
		poolInstance.generateWorker(1)
	})
}

func TestGeneratePublishRelease_TimeLimitVerbose(t *testing.T) {
	t.Parallel()

	poolConfig, err := NewValueGenPoolConfig(
		context.Background(), testTelemetryService,
		"time-limit-verbose-direct", 1, 1, 1, time.Second,
		func() (googleUuid.UUID, error) { return googleUuid.NewV7() },
		true,
	)
	require.NoError(t, err)

	meter := testTelemetryService.MetricsProvider.Meter("cryptoutil.pool.time-limit-verbose-direct")
	generateCounterMetric, err := meter.Int64Counter("cryptoutil.pool.generate.count")
	require.NoError(t, err)

	poolInstance := &ValueGenPool[googleUuid.UUID]{
		poolStartTime:         time.Now().UTC().Add(-time.Hour),
		generateCounterMetric: generateCounterMetric,
		cfg:                   poolConfig,
		permissionChannel:     make(chan struct{}, 1),
	}
	poolInstance.permissionChannel <- struct{}{}

	err = poolInstance.generatePublishRelease(1, time.Now().UTC())
	require.ErrorContains(t, err, "reached max lifetime")
	require.Empty(t, poolInstance.permissionChannel)
}
