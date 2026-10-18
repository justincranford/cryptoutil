// Copyright (c) 2025-2026 Justin Cranford.
package pool

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

type errorHistogramMeter struct {
	metric.Meter
	err error
}

func (m errorHistogramMeter) Float64Histogram(string, ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return nil, m.err
}

type errorCounterMeter struct {
	metric.Meter
	err error
}

func (m errorCounterMeter) Int64Counter(string, ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return nil, m.err
}

func TestMetricCreationHelpers_PropagateErrors(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected metric creation error")
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "float histogram",
			run: func() error {
				_, err := newFloat64HistogramImpl(errorHistogramMeter{err: injectedErr}, "histogram")

				return err
			},
		},
		{
			name: "integer counter",
			run: func() error {
				_, err := newInt64CounterImpl(errorCounterMeter{err: injectedErr}, "counter")

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.ErrorIs(t, tc.run(), injectedErr)
		})
	}
}

func TestGetAndCancel_AfterCancellation(t *testing.T) {
	t.Parallel()

	poolConfig, err := NewValueGenPoolConfig(testCtx, testTelemetryService, "cancelled-get", 1, 1, 1, time.Second, func() (int, error) {
		return 1, nil
	}, true)
	require.NoError(t, err)

	poolInstance, err := NewValueGenPool(poolConfig, nil)
	require.NoError(t, err)

	poolInstance.Cancel()
	require.Zero(t, poolInstance.Get())
	poolInstance.Cancel()

	for range poolInstance.generateChannel {
	}
}
