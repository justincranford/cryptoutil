package main

import (
	"context"
	"time"

	"cryptoutil/keygen"
	"cryptoutil/telemetry"
)

func main() {
	startTime := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slogger, logsProvider, metricsProvider, tracesProvider := telemetry.Init(ctx)
	slogger.Info("Start", "uptime", time.Since(startTime).Seconds())
	defer func() {
		slogger.Info("Stop", "uptime", time.Since(startTime).Seconds())
	}()
	defer telemetry.Shutdown(slogger, tracesProvider, metricsProvider, logsProvider)

	keygen.DoKeyPoolsExample(ctx, slogger)
	telemetry.DoMetricExample(ctx, slogger, metricsProvider)
	telemetry.DoTraceExample(ctx, slogger, tracesProvider)
}
