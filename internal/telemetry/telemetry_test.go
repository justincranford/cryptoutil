package telemetry

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"testing"
	"time"
)

var (
	testCtx              = context.Background()
	testTelemetryService *TelemetryService
)

func TestMain(m *testing.M) {
	testTelemetryService = RequireNewForTest(testCtx, "telemetry_test", false, false)
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func TestLogger(t *testing.T) {
	testTelemetryService.Slogger.Info("Initialized telemetry", "uptime", time.Since(testTelemetryService.StartTime).Seconds())
}

func TestMetric(t *testing.T) {
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
		exampleMetricHistogram.Record(testCtx, rand.Int64N(100))
		exampleMetricHistogram.Record(testCtx, rand.Int64N(100))
		exampleMetricHistogram.Record(testCtx, rand.Int64N(100))
	} else {
		testTelemetryService.Slogger.Error("metric failed", "error", fmt.Errorf("metric error: %w", err))
	}
}

func TestTrace(t *testing.T) {
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
