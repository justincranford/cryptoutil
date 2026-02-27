// Copyright (c) 2025 Justin Cranford
//
//

package pool

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"
	"time"

	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

// TestNewValueGenPool_MetricCreationErrors covers the metric creation error paths
// in NewValueGenPool (lines 84-106). NOT parallel — modifies package-level vars.
func TestNewValueGenPool_MetricCreationErrors(t *testing.T) {
	tests := []struct {
		name          string
		histFailAt    int
		counterFailAt int
		wantErr       string
	}{
		{name: "get histogram error", histFailAt: 1, wantErr: "failed to create get metric"},
		{name: "permission histogram error", histFailAt: 2, wantErr: "failed to create permission metric"},
		{name: "generate histogram error", histFailAt: 3, wantErr: "failed to create generate metric"},
		{name: "get counter error", counterFailAt: 1, wantErr: "failed to create get counter metric"},
		{name: "generate counter error", counterFailAt: 2, wantErr: "failed to create generate counter metric"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			histCallCount := 0
			origHist := newFloat64HistogramFn

			if tc.histFailAt > 0 {
				newFloat64HistogramFn = func(m metric.Meter, name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
					histCallCount++
					if histCallCount == tc.histFailAt {
						return nil, fmt.Errorf("injected histogram error")
					}

					return origHist(m, name, opts...)
				}
			}

			defer func() { newFloat64HistogramFn = origHist }()

			counterCallCount := 0
			origCounter := newInt64CounterFn

			if tc.counterFailAt > 0 {
				newInt64CounterFn = func(m metric.Meter, name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
					counterCallCount++
					if counterCallCount == tc.counterFailAt {
						return nil, fmt.Errorf("injected counter error")
					}

					return origCounter(m, name, opts...)
				}
			}

			defer func() { newInt64CounterFn = origCounter }()

			_, err := NewValueGenPool(NewValueGenPoolConfig(
				context.Background(), testTelemetryService,
				"metric-error", 1, 1, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, time.Second,
				cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false,
			))
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
