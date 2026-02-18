// Copyright (c) 2025 Justin Cranford
//
//

// Package e2e_infra provides reusable helpers for E2E testing with docker compose.
package e2e_infra

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"os/exec"
	"time"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// ComposeManager orchestrates docker compose lifecycle for E2E tests.
type ComposeManager struct {
	ComposeFile string
	HTTPClient  *http.Client
}

// NewComposeManager creates a compose manager with TLS-enabled HTTP client.
func NewComposeManager(composeFile string) *ComposeManager {
	return &ComposeManager{
		ComposeFile: composeFile,
		HTTPClient:  cryptoutilSharedCryptoTls.NewClientForTest(),
	}
}

// Start brings up docker compose stack.
//
// Note: Does NOT use --wait flag because Docker Compose --wait only works with containers that have
// native HEALTHCHECK instructions in their container image or Dockerfile. Many containers (like
// otel/opentelemetry-collector-contrib) don't have native healthchecks.
//
// Instead, this project uses three healthcheck strategies (implemented in docker_health.go):
//  1. Job-only healthchecks: Standalone jobs that must exit successfully (ExitCode=0)
//     Examples: healthcheck-secrets, builder-cryptoutil
//  2. Service-only healthchecks: Services with native HEALTHCHECK instructions
//     Examples: cryptoutil-sqlite, cryptoutil-postgres-1, postgres, grafana-otel-lgtm
//  3. Service with healthcheck job: Services use external sidecar job for health verification
//     Example: opentelemetry-collector-contrib with healthcheck-opentelemetry-collector-contrib
//
// Use WaitForMultipleServices() or WaitForServicesHealthy() after Start() to wait for services.
func (cm *ComposeManager) Start(ctx context.Context) error {
	fmt.Println("Starting docker compose stack...")

	startCmd := exec.CommandContext(ctx, "docker", "compose", "-f", cm.ComposeFile, "up", "-d")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to start docker compose: %w", err)
	}

	return nil
}

// Stop tears down docker compose stack.
func (cm *ComposeManager) Stop(ctx context.Context) error {
	fmt.Println("Stopping docker compose stack...")

	downCmd := exec.CommandContext(ctx, "docker", "compose", "-f", cm.ComposeFile, "down", "-v")
	downCmd.Stdout = os.Stdout
	downCmd.Stderr = os.Stderr

	if err := downCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop docker compose: %w", err)
	}

	return nil
}

// WaitForHealth polls an health endpoint until healthy or timeout.
// Supports both admin endpoints (/admin/api/v1/livez) and public endpoints (/health).
func (cm *ComposeManager) WaitForHealth(healthURL string, timeout time.Duration) error {
	ctx := context.Background()

	timeoutCh := time.After(timeout)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	fmt.Printf("[WaitForHealth] Starting health check for %s (timeout: %v)\n", healthURL, timeout)

	attempts := 0

	for {
		select {
		case <-timeoutCh:
			fmt.Printf("[WaitForHealth] TIMEOUT for %s after %d attempts\n", healthURL, attempts)

			return fmt.Errorf("health check timeout after %v", timeout)
		case <-ticker.C:
			attempts++

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			if err != nil {
				fmt.Printf("[WaitForHealth] Attempt %d for %s: request creation error: %v\n", attempts, healthURL, err)

				continue // Retry on request creation errors.
			}

			resp, err := cm.HTTPClient.Do(req)
			if err != nil {
				fmt.Printf("[WaitForHealth] Attempt %d for %s: connection error: %v\n", attempts, healthURL, err)

				continue // Retry on connection errors.
			}

			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Printf("[WaitForHealth] SUCCESS for %s after %d attempts\n", healthURL, attempts)

				return nil // Healthy!
			}

			fmt.Printf("[WaitForHealth] Attempt %d for %s: HTTP %d\n", attempts, healthURL, resp.StatusCode)
		}
	}
}

// WaitForMultipleServices waits for multiple services to be healthy concurrently.
// Returns error if any service fails health check within timeout.
func (cm *ComposeManager) WaitForMultipleServices(services map[string]string, timeout time.Duration) error {
	type result struct {
		name string
		err  error
	}

	resultsCh := make(chan result, len(services))

	// Start health checks for all services concurrently.
	for name, healthURL := range services {
		go func(serviceName, url string) {
			err := cm.WaitForHealth(url, timeout)
			resultsCh <- result{name: serviceName, err: err}
		}(name, healthURL)
	}

	// Collect results.
	for i := 0; i < len(services); i++ {
		res := <-resultsCh
		if res.err != nil {
			return fmt.Errorf("service %s health check failed: %w", res.name, res.err)
		}
	}

	return nil
}
