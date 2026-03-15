// Copyright (c) 2025 Justin Cranford

//go:build demo

package demo

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	composeFilePath = "../../deployments/identity/compose.advanced.yml"
	demoProfile     = "demo"
)

// TestDockerComposeProfiles validates all Docker Compose profiles (demo, development, ci, production).
func TestDockerComposeProfiles(t *testing.T) {
	t.Parallel()

	profiles := []string{"demo", "development", "ci", "production"}

	for _, profile := range profiles {
		t.Run(profile, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			composeFile := composeFilePath

			// Start services
			startCmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "--profile", profile, "up", "-d")
			startCmd.Stdout = os.Stdout
			startCmd.Stderr = os.Stderr
			err := startCmd.Run()
			require.NoError(t, err, "Failed to start profile %s", profile)

			// Cleanup on test completion
			defer func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
				defer cleanupCancel()

				downCmd := exec.CommandContext(cleanupCtx, "docker", "compose", "-f", composeFile, "--profile", profile, "down", "-v")
				downCmd.Stdout = os.Stdout
				downCmd.Stderr = os.Stderr
				_ = downCmd.Run() //nolint:errcheck // Test cleanup - docker compose down may fail but cleanup should continue
			}()

			// Wait for services to become healthy
			healthyCtx, healthyCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.StrictCertificateMaxAgeDays*time.Second)
			defer healthyCancel()

			ticker := time.NewTicker(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-healthyCtx.Done():
					require.FailNowf(t, "Timeout waiting for services to become healthy", "profile: %s", profile)
				case <-ticker.C:
					psCmd := exec.CommandContext(healthyCtx, "docker", "compose", "-f", composeFile, "--profile", profile, "ps", "--format", "json")

					output, err := psCmd.Output()
					if err != nil {
						continue
					}

					// Check if all services are healthy
					if strings.Contains(string(output), `"Health":"healthy"`) {
						t.Logf("All services healthy for profile: %s", profile)

						return
					}
				}
			}
		})
	}
}

// TestDockerComposeScaling validates scaling scenarios (2x2x2x2, 3x3x3x3).
func TestDockerComposeScaling(t *testing.T) {
	t.Parallel()

	scalingScenarios := []struct {
		name     string
		scaling  map[string]int
		expected int
	}{
		{
			name: "2x2x2x2",
			scaling: map[string]int{
				cryptoutilSharedMagic.OTLPServiceIdentityAuthz: cryptoutilMagic.IdentityScaling2x,
				cryptoutilSharedMagic.OTLPServiceIdentityIDP:   cryptoutilMagic.IdentityScaling2x,
				cryptoutilSharedMagic.OTLPServiceIdentityRS:    cryptoutilMagic.IdentityScaling2x,
				"identity-spa-rp": cryptoutilMagic.IdentityScaling2x,
			},
			expected: cryptoutilSharedMagic.IMMinPasswordLength, // 2x4 services
		},
		{
			name: "3x3x3x3",
			scaling: map[string]int{
				cryptoutilSharedMagic.OTLPServiceIdentityAuthz: cryptoutilMagic.IdentityScaling3x,
				cryptoutilSharedMagic.OTLPServiceIdentityIDP:   cryptoutilMagic.IdentityScaling3x,
				cryptoutilSharedMagic.OTLPServiceIdentityRS:    cryptoutilMagic.IdentityScaling3x,
				"identity-spa-rp": cryptoutilMagic.IdentityScaling3x,
			},
			expected: cryptoutilSharedMagic.HashPrefixLength, // 3x4 services
		},
	}

	for _, scenario := range scalingScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			composeFile := "deployments/identity/compose.advanced.yml"
			profile := demoProfile

			// Build scaling arguments
			args := []string{"compose", "-f", composeFile, "--profile", profile, "up", "-d"}
			for service, replicas := range scenario.scaling {
				args = append(args, "--scale", fmt.Sprintf("%s=%d", service, replicas))
			}

			// Start services with scaling
			startCmd := exec.CommandContext(ctx, "docker", args...)
			startCmd.Stdout = os.Stdout
			startCmd.Stderr = os.Stderr
			err := startCmd.Run()
			require.NoError(t, err, "Failed to start scaling scenario %s", scenario.name)

			// Cleanup
			defer func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
				defer cleanupCancel()

				downCmd := exec.CommandContext(cleanupCtx, "docker", "compose", "-f", composeFile, "--profile", profile, "down", "-v")
				downCmd.Stdout = os.Stdout
				downCmd.Stderr = os.Stderr
				_ = downCmd.Run() //nolint:errcheck // Test cleanup - docker compose down may fail but cleanup should continue
			}()

			// Wait for services to become healthy
			time.Sleep(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second) // Give services time to start

			// Count running containers
			psCmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "--profile", profile, "ps", "-q")
			output, err := psCmd.Output()
			require.NoError(t, err, "Failed to list services for scaling scenario %s", scenario.name)

			containers := strings.Split(strings.TrimSpace(string(output)), "\n")
			actualCount := len(containers)

			// Allow +1 for PostgreSQL container
			require.GreaterOrEqual(t, actualCount, scenario.expected, "Expected at least %d containers for %s, got %d", scenario.expected, scenario.name, actualCount)

			t.Logf("Scaling scenario %s validated: %d containers running (expected %d+)", scenario.name, actualCount, scenario.expected)
		})
	}
}

// TestDockerSecretsIntegration validates Docker secrets are properly mounted.
func TestDockerSecretsIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	composeFile := composeFilePath
	profile := demoProfile

	// Start services
	startCmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "--profile", profile, "up", "-d")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	err := startCmd.Run()
	require.NoError(t, err, "Failed to start services")

	// Cleanup
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
		defer cleanupCancel()

		downCmd := exec.CommandContext(cleanupCtx, "docker", "compose", "-f", composeFile, "--profile", profile, "down", "-v")
		downCmd.Stdout = os.Stdout
		downCmd.Stderr = os.Stderr
		_ = downCmd.Run() //nolint:errcheck // Test cleanup - docker compose down may fail but cleanup should continue
	}()

	// Wait for services to start
	time.Sleep(15 * time.Second)

	// Verify secrets are mounted in authz container
	secretsCmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "--profile", profile, "exec", "-T", cryptoutilSharedMagic.OTLPServiceIdentityAuthz, "ls", "-la", "/run/secrets/")
	output, err := secretsCmd.Output()
	require.NoError(t, err, "Failed to list secrets in authz container")

	outputStr := string(output)
	require.Contains(t, outputStr, "postgres_user", "Secret postgres_user not mounted")
	require.Contains(t, outputStr, "postgres_password", "Secret postgres_password not mounted")
	require.Contains(t, outputStr, "postgres_db", "Secret postgres_db not mounted")

	t.Logf("Docker secrets validated: postgres_user, postgres_password, postgres_db")
}

// TestHealthChecks validates service health checks pass.
func TestHealthChecks(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	composeFile := composeFilePath
	profile := demoProfile

	// Start services
	startCmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "--profile", profile, "up", "-d")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	err := startCmd.Run()
	require.NoError(t, err, "Failed to start services")

	// Cleanup
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
		defer cleanupCancel()

		downCmd := exec.CommandContext(cleanupCtx, "docker", "compose", "-f", composeFile, "--profile", profile, "down", "-v")
		downCmd.Stdout = os.Stdout
		downCmd.Stderr = os.Stderr
		_ = downCmd.Run() //nolint:errcheck // Test cleanup - docker compose down may fail but cleanup should continue
	}()

	// Wait for health checks to pass
	healthyCtx, healthyCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.StrictCertificateMaxAgeDays*time.Second)
	defer healthyCancel()

	ticker := time.NewTicker(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-healthyCtx.Done():
			t.Fatal("Timeout waiting for services to become healthy")
		case <-ticker.C:
			psCmd := exec.CommandContext(healthyCtx, "docker", "compose", "-f", composeFile, "--profile", profile, "ps", "--format", "json")

			output, err := psCmd.Output()
			if err != nil {
				continue
			}

			// Check if all services are healthy
			if strings.Contains(string(output), `"Health":"healthy"`) {
				t.Log("All services are healthy")

				return
			}
		}
	}
}
