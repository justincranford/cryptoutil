//go:build e2e

// Package e2e provides end-to-end integration tests for identity services.
package e2e

import "time"

// E2E test configuration constants.
const (
	// Docker Compose configuration.
	composeFile    = "../../../../deployments/compose/identity-demo.yml"
	defaultProfile = "demo"

	// Health check timeouts - longer timeout for orchestration/failover tests.
	healthCheckTimeoutE2E     = 90 * time.Second
	healthCheckTimeout        = 90 * time.Second // Alias for compatibility.
	healthCheckTimeoutService = 5 * time.Second
	healthCheckRetry          = 5 * time.Second
)
