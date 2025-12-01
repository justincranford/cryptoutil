// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides Identity demo implementation.
package demo

import (
	"context"
	"time"
)

// Identity demo step counts.
const (
	identityStepCount  = 5
	identityPassedInit = 1
	identitySkipped    = 4
)

// runIdentityDemo executes the Identity demo.
func runIdentityDemo(ctx context.Context, config *Config) int {
	progress := NewProgressDisplay(config)
	errors := NewErrorAggregator("identity")
	startTime := time.Now()

	progress.Info("Starting Identity Demo")
	progress.Info("=======================")
	progress.SetTotalSteps(identityStepCount)

	// Step 1: Parse configuration.
	progress.StartStep("Parsing configuration")

	// TODO: Implement Identity demo configuration parsing.
	progress.CompleteStep("Parsed configuration")

	// Step 2: Start server.
	progress.StartStep("Starting Identity server")

	// TODO: Implement Identity server startup.
	progress.SkipStep("Starting Identity server", "not yet implemented")

	// Step 3: Wait for health checks.
	progress.StartStep("Waiting for health checks")

	// TODO: Implement Identity health checks.
	progress.SkipStep("Health checks", "not yet implemented")

	// Step 4: Register demo client.
	progress.StartStep("Registering demo client")

	// TODO: Implement demo client registration.
	progress.SkipStep("Registering demo client", "not yet implemented")

	// Step 5: Demonstrate OAuth flows.
	progress.StartStep("Demonstrating OAuth 2.1 flows")

	// TODO: Implement OAuth flow demonstration.
	progress.SkipStep("OAuth 2.1 flows", "not yet implemented")

	// Calculate final result.
	result := errors.ToResult(identityPassedInit, identitySkipped)
	result.DurationMS = time.Since(startTime).Milliseconds()

	progress.PrintSummary(result)

	return result.ExitCode()
}
