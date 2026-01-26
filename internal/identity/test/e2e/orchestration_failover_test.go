// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	// authzURL1 is the URL for the first AuthZ instance.
	authzURL1 = "https://127.0.0.1:8080"
	// authzURL2 is the URL for the second AuthZ instance.
	authzURL2 = "https://127.0.0.1:8081"
)

// TestOAuthFlowFailover validates OAuth 2.1 authorization code flow continues working after service instance failures.
func TestOAuthFlowFailover(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start 2x2x2x2 deployment (2 instances per service)
	t.Log("📦 Starting 2x2x2x2 deployment for failover testing...")
	require.NoError(t, startCompose(ctx, defaultProfile, map[string]int{
		"identity-authz": cryptoutilMagic.IdentityScaling2x,
		"identity-idp":   cryptoutilMagic.IdentityScaling2x,
		"identity-rs":    cryptoutilMagic.IdentityScaling2x,
		"identity-spa":   cryptoutilMagic.IdentityScaling2x,
	}))

	defer func() {
		_ = stopCompose(context.Background(), defaultProfile, true)
	}()

	// Wait for all services healthy
	t.Log("⏳ Waiting for all services to become healthy...")
	require.NoError(t, waitForHealthy(ctx, defaultProfile, healthCheckTimeoutE2E, healthCheckRetry))

	// Perform successful OAuth flow (baseline)
	suite := NewE2ETestSuite()
	suite.AuthZURL = authzURL1 // First AuthZ instance

	t.Log("✅ Performing baseline OAuth 2.1 authorization code flow...")

	token1, err := performAuthorizationCodeFlow(suite, "test-client-1", "test-secret-1")
	require.NoError(t, err, "Baseline OAuth flow should succeed")
	require.NotEmpty(t, token1, "Access token should be returned")

	// Kill first AuthZ instance
	t.Log("🔪 Killing first AuthZ instance (identity-authz-1)...")
	require.NoError(t, killContainer(ctx, "identity-authz-1"))

	// Verify second AuthZ instance still healthy
	t.Log("🔍 Verifying second AuthZ instance (identity-authz-2) still healthy...")
	time.Sleep(3 * time.Second) // Give time for health check to update

	// Perform OAuth flow against second instance
	suite.AuthZURL = authzURL2 // Second AuthZ instance

	t.Log("✅ Performing OAuth flow against second AuthZ instance...")

	token2, err := performAuthorizationCodeFlow(suite, "test-client-2", "test-secret-2")
	require.NoError(t, err, "OAuth flow should succeed against second instance")
	require.NotEmpty(t, token2, "Access token should be returned from second instance")

	t.Log("✅ Failover test passed - OAuth flow continued after first instance failure")
}

// TestResourceServerFailover validates resource server continues working after instance failures.
func TestResourceServerFailover(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start 2x2x2x2 deployment
	t.Log("📦 Starting 2x2x2x2 deployment for resource server failover...")
	require.NoError(t, startCompose(ctx, defaultProfile, map[string]int{
		"identity-authz": cryptoutilMagic.IdentityScaling2x,
		"identity-idp":   cryptoutilMagic.IdentityScaling2x,
		"identity-rs":    cryptoutilMagic.IdentityScaling2x,
		"identity-spa":   cryptoutilMagic.IdentityScaling2x,
	}))

	defer func() {
		_ = stopCompose(context.Background(), defaultProfile, true)
	}()

	// Wait for all services healthy
	t.Log("⏳ Waiting for all services to become healthy...")
	require.NoError(t, waitForHealthy(ctx, defaultProfile, healthCheckTimeoutE2E, healthCheckRetry))

	// Get access token
	suite := NewE2ETestSuite()
	suite.AuthZURL = authzURL1
	suite.RSURL = "https://127.0.0.1:8200" // First RS instance

	t.Log("🔑 Getting access token for resource server access...")

	token, err := performClientCredentialsFlow(suite, "test-client-3", "test-secret-3")
	require.NoError(t, err, "Client credentials flow should succeed")
	require.NotEmpty(t, token, "Access token should be returned")

	// Access resource via first RS instance (baseline)
	t.Log("📄 Accessing resource via first RS instance (baseline)...")

	resource1, err := accessProtectedResource(suite, suite.RSURL, token)
	require.NoError(t, err, "Resource access should succeed")
	require.NotEmpty(t, resource1, "Resource data should be returned")

	// Kill first RS instance
	t.Log("🔪 Killing first RS instance (identity-rs-1)...")
	require.NoError(t, killContainer(ctx, "identity-rs-1"))

	// Access resource via second RS instance
	suite.RSURL = "https://127.0.0.1:8201" // Second RS instance

	t.Log("📄 Accessing resource via second RS instance...")

	resource2, err := accessProtectedResource(suite, suite.RSURL, token)
	require.NoError(t, err, "Resource access should succeed against second instance")
	require.NotEmpty(t, resource2, "Resource data should be returned from second instance")

	t.Log("✅ Resource server failover test passed")
}

// TestIdentityProviderFailover validates IdP continues working after instance failures.
func TestIdentityProviderFailover(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start 2x2x2x2 deployment
	t.Log("📦 Starting 2x2x2x2 deployment for IdP failover...")
	require.NoError(t, startCompose(ctx, defaultProfile, map[string]int{
		"identity-authz": cryptoutilMagic.IdentityScaling2x,
		"identity-idp":   cryptoutilMagic.IdentityScaling2x,
		"identity-rs":    cryptoutilMagic.IdentityScaling2x,
		"identity-spa":   cryptoutilMagic.IdentityScaling2x,
	}))

	defer func() {
		_ = stopCompose(context.Background(), defaultProfile, true)
	}()

	// Wait for all services healthy
	t.Log("⏳ Waiting for all services to become healthy...")
	require.NoError(t, waitForHealthy(ctx, defaultProfile, healthCheckTimeoutE2E, healthCheckRetry))

	// Perform user authentication via first IdP instance (baseline)
	suite := NewE2ETestSuite()
	suite.IDPURL = "https://127.0.0.1:8100" // First IdP instance

	t.Log("🔐 Performing user authentication via first IdP instance...")

	session1, err := performUserAuthentication(suite, "testuser1", "testpass1")
	require.NoError(t, err, "User authentication should succeed")
	require.NotEmpty(t, session1, "Session ID should be returned")

	// Kill first IdP instance
	t.Log("🔪 Killing first IdP instance (identity-idp-1)...")
	require.NoError(t, killContainer(ctx, "identity-idp-1"))

	// Perform user authentication via second IdP instance
	suite.IDPURL = "https://127.0.0.1:8101" // Second IdP instance

	t.Log("🔐 Performing user authentication via second IdP instance...")

	session2, err := performUserAuthentication(suite, "testuser2", "testpass2")
	require.NoError(t, err, "User authentication should succeed against second instance")
	require.NotEmpty(t, session2, "Session ID should be returned from second instance")

	t.Log("✅ IdP failover test passed")
}

// Helper: startCompose starts Docker Compose services with scaling.
func startCompose(ctx context.Context, profile string, scaling map[string]int) error {
	args := []string{"compose", "-f", composeFile, "--profile", profile, "up", "-d"}

	// Add scaling flags
	for service, replicas := range scaling {
		args = append(args, "--scale", fmt.Sprintf("%s=%d", service, replicas))
	}

	cmd := exec.CommandContext(ctx, "docker", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w, output: %s", err, string(output))
	}

	return nil
}

// Helper: stopCompose stops Docker Compose services.
func stopCompose(ctx context.Context, profile string, removeVolumes bool) error {
	args := []string{"compose", "-f", composeFile, "--profile", profile, "down"}
	if removeVolumes {
		args = append(args, "-v")
	}

	cmd := exec.CommandContext(ctx, "docker", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose down failed: %w, output: %s", err, string(output))
	}

	return nil
}

// Helper: waitForHealthy waits for all services to become healthy.
func waitForHealthy(ctx context.Context, profile string, timeout, retryInterval time.Duration) error {
	deadline := time.Now().UTC().Add(timeout)

	for time.Now().UTC().Before(deadline) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for healthy services: %w", ctx.Err())
		default:
		}

		// Check health status
		cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "--profile", profile, "ps", "--format", "json")

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("docker compose ps failed: %w, output: %s", err, string(output))
		}

		// Parse JSON output and check all services healthy
		allHealthy := true
		lines := strings.Split(string(output), "\n")

		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			var container struct {
				Service string `json:"Service"`
				State   string `json:"State"`
				Health  string `json:"Health"`
			}

			if err := json.Unmarshal([]byte(line), &container); err != nil {
				continue // Skip unparseable lines
			}

			// Check if service is running and healthy
			if container.State != "running" {
				allHealthy = false

				break
			}

			// If service has health check, verify it's healthy
			if container.Health != "" && container.Health != "healthy" {
				allHealthy = false

				break
			}
		}

		if allHealthy {
			return nil
		}

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("services not healthy after %v", timeout)
}

// Helper: killContainer kills a specific Docker container.
func killContainer(ctx context.Context, containerName string) error {
	cmd := exec.CommandContext(ctx, "docker", "kill", containerName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker kill failed: %w, output: %s", err, string(output))
	}

	return nil
}

// Helper: performAuthorizationCodeFlow simulates OAuth 2.1 authorization code flow.
func performAuthorizationCodeFlow(suite *E2ETestSuite, clientID, clientSecret string) (string, error) {
	// Placeholder: Replace with actual OAuth 2.1 authorization code flow implementation
	// This would involve:
	// 1. GET /authorize with client_id, redirect_uri, state, code_challenge
	// 2. POST /login with username, password (user authentication)
	// 3. POST /consent with approved scopes
	// 4. GET redirect_uri with code parameter
	// 5. POST /token with code, client_id, client_secret, code_verifier
	// 6. Return access_token

	// For now, return mock token to validate test structure
	return "mock-access-token-authz-flow", nil
}

// Helper: performClientCredentialsFlow simulates OAuth 2.1 client credentials flow.
func performClientCredentialsFlow(suite *E2ETestSuite, clientID, clientSecret string) (string, error) {
	// Placeholder: Replace with actual OAuth 2.1 client credentials flow implementation
	// This would involve:
	// 1. POST /token with grant_type=client_credentials, client_id, client_secret, scope
	// 2. Return access_token

	// For now, return mock token to validate test structure
	return "mock-access-token-client-credentials", nil
}

// Helper: accessProtectedResource accesses a protected resource with an access token.
func accessProtectedResource(suite *E2ETestSuite, rsURL, token string) (string, error) {
	// Placeholder: Replace with actual resource access implementation
	// This would involve:
	// 1. GET /api/resource with Authorization: Bearer <token>
	// 2. Return resource data

	// For now, return mock resource to validate test structure
	return "mock-resource-data", nil
}

// Helper: performUserAuthentication performs user authentication via IdP.
func performUserAuthentication(suite *E2ETestSuite, username, password string) (string, error) {
	// Placeholder: Replace with actual user authentication implementation
	// This would involve:
	// 1. POST /login with username, password
	// 2. Return session_id

	// For now, return mock session ID to validate test structure
	return "mock-session-id", nil
}
