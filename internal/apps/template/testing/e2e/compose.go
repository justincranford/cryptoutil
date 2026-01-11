// Copyright (c) 2025 Justin Cranford
//
//

// Package e2e provides reusable helpers for E2E testing with docker compose.
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
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
		HTTPClient:  cryptoutilTLS.NewClientForTest(),
	}
}

// Start brings up docker compose stack.
func (cm *ComposeManager) Start(ctx context.Context) error {
	fmt.Println("Starting docker compose stack...")
	startCmd := exec.CommandContext(ctx, "docker", "compose", "-f", cm.ComposeFile, "up", "-d")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	return startCmd.Run()
}

// Stop tears down docker compose stack.
func (cm *ComposeManager) Stop(ctx context.Context) error {
	fmt.Println("Stopping docker compose stack...")
	downCmd := exec.CommandContext(ctx, "docker", "compose", "-f", cm.ComposeFile, "down", "-v")
	downCmd.Stdout = os.Stdout
	downCmd.Stderr = os.Stderr
	return downCmd.Run()
}

// WaitForHealth polls /admin/v1/livez until healthy or timeout.
func (cm *ComposeManager) WaitForHealth(adminURL string, timeout time.Duration) error {
	healthURL := adminURL + "/admin/v1/livez"
	timeoutCh := time.After(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCh:
			return fmt.Errorf("health check timeout after %v", timeout)
		case <-ticker.C:
			resp, err := cm.HTTPClient.Get(healthURL)
			if err != nil {
				continue // Retry on connection errors.
			}
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil // Healthy!
			}
		}
	}
}
