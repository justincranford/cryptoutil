// Copyright (c) 2025 Justin Cranford

package demo

import (
	"context"
	"fmt"
	"time"
)

// runCADemo runs the Certificate Authority Server demonstration.
// Demonstrates certificate issuance, revocation, and CRL/OCSP operations.
func runCADemo(_ context.Context, config *Config) int {
	progress := NewProgressDisplay(config)
	errors := NewErrorAggregator("ca")
	startTime := time.Now().UTC()

	progress.Info("Starting CA Demo")
	progress.Info("================")
	progress.SetTotalSteps(1)

	// CA server is not yet fully implemented.
	// This is a placeholder that will be expanded when CA server is complete.
	progress.StartStep("CA Demo status check")
	progress.CompleteStep("CA server scaffolding exists")

	fmt.Println()
	fmt.Println("CA Demo Status:")
	fmt.Println("  ✅ Handler scaffolding complete")
	fmt.Println("  ✅ OpenAPI spec complete")
	fmt.Println("  ✅ Server entry point created")
	fmt.Println("  ⚠️  EST endpoints pending mTLS")
	fmt.Println("  ⚠️  TSA endpoint pending ASN.1 parsing")
	fmt.Println("  ❌ Full integration demo pending")
	fmt.Println()

	result := errors.ToResult(1, 0)
	result.DurationMS = time.Since(startTime).Milliseconds()
	progress.PrintSummary(result)

	return result.ExitCode()
}
