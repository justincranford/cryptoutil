// Copyright (c) 2025-2026 Justin Cranford.
//

//go:build e2e

package test_orch_e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// ComposeManager manages Docker Compose stack lifecycle for PS-ID TLS E2E tests.
// It starts two compose files (main + test port-expose override) so Go tests
// can directly dial OTel gRPC/HTTP from the host.
type ComposeManager struct {
	spec TLSPSIDSpec
}

// NewComposeManager returns a ComposeManager for the given PS-ID spec.
func NewComposeManager(spec TLSPSIDSpec) *ComposeManager {
	return &ComposeManager{spec: spec}
}

// Start brings up the compose stack services needed for TLS E2E tests.
// It runs pki-init, OTel Collector, Grafana LGTM, and all app variants.
func (m *ComposeManager) Start(ctx context.Context) error {
	args := []string{
		"compose",
		"-f", m.spec.ComposeFile,
		"-f", m.spec.ComposeOverrideFile,
		"up", "-d", "--build",
	}
	args = append(args, m.spec.StartupServices()...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compose up failed for %s: %w", m.spec.PSID, err)
	}

	return nil
}

// Stop tears down the compose stack and removes volumes.
func (m *ComposeManager) Stop(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", m.spec.ComposeFile,
		"-f", m.spec.ComposeOverrideFile,
		"down", "-v",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compose down failed for %s: %w", m.spec.PSID, err)
	}

	return nil
}
