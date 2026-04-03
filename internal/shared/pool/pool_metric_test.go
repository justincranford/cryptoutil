// Copyright (c) 2025 Justin Cranford
//
//

package pool

import (
	"context"
	"fmt"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

// TestNewValueGenPool_MetricCreationErrors covers the metric creation error paths
// in NewValueGenPool (lines 84-106). Uses fn-param injection via newValueGenPoolInternal.
func TestNewValueGenPool_MetricCreationErrors(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			histCallCount := 0
			stubHistFn := func(m metric.Meter, name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
				histCallCount++
				if tc.histFailAt > 0 && histCallCount == tc.histFailAt {
					return nil, fmt.Errorf("injected histogram error")
				}

				return newFloat64HistogramImpl(m, name, opts...)
			}

			counterCallCount := 0
			stubCounterFn := func(m metric.Meter, name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
				counterCallCount++
				if tc.counterFailAt > 0 && counterCallCount == tc.counterFailAt {
					return nil, fmt.Errorf("injected counter error")
				}

				return newInt64CounterImpl(m, name, opts...)
			}

			cfg, cfgErr := NewValueGenPoolConfig(
				context.Background(), testTelemetryService,
				"metric-error", 1, 1, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, time.Second,
				cryptoutilSharedUtilRandom.GenerateUUIDv7Function(), false,
			)
			_, err := newValueGenPoolInternal(cfg, cfgErr, stubHistFn, stubCounterFn)
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
