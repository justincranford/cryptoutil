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

	telemetryService := telemetry.Init(ctx, startTime, "main", false, true)
	defer telemetryService.Shutdown()

	keygen.DoKeyPoolsExample(ctx, telemetryService)
	telemetry.DoMetricExample(ctx, telemetryService)
	telemetry.DoTraceExample(ctx, telemetryService)
}
