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

	telemetryService := telemetry.Init(ctx, "main")
	telemetryService.Slogger.Info("Start", "uptime", time.Since(startTime).Seconds())
	defer func() {
		telemetryService.Slogger.Info("Stop", "uptime", time.Since(startTime).Seconds())
	}()
	defer telemetry.Shutdown(telemetryService)

	keygen.DoKeyPoolsExample(ctx, telemetryService)
	telemetry.DoMetricExample(ctx, telemetryService)
	telemetry.DoTraceExample(ctx, telemetryService)
}
