// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("telemetry_service_test")
	testCtx              = context.Background()
	testTelemetryService *TelemetryService
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestLogger(_ *testing.T) {
	testTelemetryService.Slogger.Info("Initialized telemetry", "uptime", time.Since(testTelemetryService.StartTime).Seconds())
}

func TestMetric(_ *testing.T) {
	exampleMetricsScope := testTelemetryService.MetricsProvider.Meter("example-scope")

	exampleMetricCounter, err := exampleMetricsScope.Float64UpDownCounter("example-counter")
	if err == nil {
		exampleMetricCounter.Add(testCtx, 1)
		exampleMetricCounter.Add(testCtx, -2)
		exampleMetricCounter.Add(testCtx, 4)
	} else {
		testTelemetryService.Slogger.Error("metric failed", "error", fmt.Errorf("metric error: %w", err))
	}

	exampleMetricHistogram, err := exampleMetricsScope.Int64Histogram("example-histogram")
	if err == nil {
		// Generate cryptographically secure random numbers for histogram test data
		val1, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			testTelemetryService.Slogger.Error("random generation failed", "error", err)

			return
		}

		val2, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			testTelemetryService.Slogger.Error("random generation failed", "error", err)

			return
		}

		val3, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			testTelemetryService.Slogger.Error("random generation failed", "error", err)

			return
		}

		exampleMetricHistogram.Record(testCtx, val1.Int64())
		exampleMetricHistogram.Record(testCtx, val2.Int64())
		exampleMetricHistogram.Record(testCtx, val3.Int64())
	} else {
		testTelemetryService.Slogger.Error("metric failed", "error", fmt.Errorf("metric error: %w", err))
	}
}

func TestTrace(_ *testing.T) {
	exampleTrace := testTelemetryService.TracesProvider.Tracer("example-trace")
	testTelemetryService.Slogger.Info("exampleTrace", "trace", exampleTrace)

	// simulate time spent in parent function, before calling child 1 function
	exampleParentSpanContext, exampleParentSpan := exampleTrace.Start(testCtx, "example-parent-span")

	time.Sleep(5 * time.Millisecond)
	exampleParentSpan.End()
	testTelemetryService.Slogger.Info("exampleParentSpan", "testCtx", exampleParentSpanContext, "span", exampleParentSpan)

	// simulate time spent in child 1 function
	exampleChildSpanContext1, exampleChildSpan1 := exampleTrace.Start(exampleParentSpanContext, "example-child-span-1")

	time.Sleep(10 * time.Millisecond)

	defer exampleChildSpan1.End()

	testTelemetryService.Slogger.Info("exampleChildSpan1", "testCtx", exampleChildSpanContext1, "span", exampleChildSpan1)

	// simulate time spent in parent function, before calling child 2 function
	time.Sleep(5 * time.Millisecond)

	// simulate time spent in child 2 function
	exampleChildSpanContext2, exampleChildSpan2 := exampleTrace.Start(exampleParentSpanContext, "example-child-span-2")

	time.Sleep(15 * time.Millisecond)

	defer exampleChildSpan2.End()

	testTelemetryService.Slogger.Info("exampleChildSpan2", "testCtx", exampleChildSpanContext2, "span", exampleChildSpan2)

	// simulate time spent in parent function, before returning
	time.Sleep(5 * time.Millisecond)
}
