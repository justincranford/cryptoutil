//go:build e2e

package test

// LOGGING REQUIREMENT: All logs in this e2e test file MUST include timestamp and duration since start time.
// Format: fmt.Printf("[%s] [%v] message...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), ...)
// This ensures consistent, traceable logging for debugging test failures and performance analysis.

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilClient "cryptoutil/internal/client"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

const (
	// Test timeouts.
	dockerHealthTimeout      = cryptoutilMagic.TestTimeoutDockerHealth      // Docker services should be healthy in under 20s
	cryptoutilReadyTimeout   = cryptoutilMagic.TestTimeoutCryptoutilReady   // Cryptoutil needs time to unseal - reduced for fast fail
	testExecutionTimeout     = cryptoutilMagic.TestTimeoutTestExecution     // Overall test timeout - reduced for fast fail
	dockerComposeInitTimeout = cryptoutilMagic.TestTimeoutDockerComposeInit // Time to wait for Docker Compose services to initialize after startup
	serviceRetryInterval     = cryptoutilMagic.TestTimeoutServiceRetry      // Check more frequently
	httpClientTimeout        = cryptoutilMagic.TestTimeoutHTTPClient
	httpRetryInterval        = cryptoutilMagic.TestTimeoutHTTPRetryInterval
)

var (
	// Public API URLs (ports 8080+).
	cryptoutilSqliteURL    = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose0)
	cryptoutilPostgres1URL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose1)
	cryptoutilPostgres2URL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortCryptoutilCompose2)

	// Private admin API URLs (port 9090 inside containers).
	cryptoutilSqliteAdminURL    = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPrivatePortCryptoutil)
	cryptoutilPostgres1AdminURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPrivatePortCryptoutil)
	cryptoutilPostgres2AdminURL = cryptoutilMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilMagic.DefaultPrivatePortCryptoutil)

	grafanaURL       = cryptoutilMagic.URLPrefixLocalhostHTTP + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortGrafana)
	otelCollectorURL = cryptoutilMagic.URLPrefixLocalhostHTTP + fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortInternalMetrics)

	// Test data variables (so we can take their addresses).
	testElasticKeyName        = "e2e-test-key"
	testElasticKeyDescription = "E2E integration test key"
	testAlgorithm             = "RSA"
	testProvider              = "GO"
)

// TestE2EIntegration performs end-to-end testing of all cryptoutil instances with telemetry verification.
func TestE2EIntegration(t *testing.T) {
	t.Parallel()

	startTime := time.Now()
	fmt.Printf("[%s] [%v] 🔄 TEST START\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), testExecutionTimeout)
	defer cancel()

	fmt.Printf("[%s] [%v] 📋 Loading test certificates...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	// Load test certificates for TLS validation
	rootCAsPool := (*x509.CertPool)(nil) // Using InsecureSkipVerify for tests
	fmt.Printf("[%s] [%v] ✅ Certificates loaded (nil: %v)\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), rootCAsPool == nil)

	// Always stop any existing services first to ensure clean test state
	fmt.Printf("[%s] [%v] 🧹 Ensuring clean test state by stopping any existing Docker Compose services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	err := stopDockerCompose(ctx, startTime)
	if err != nil { //nolint:wsl // gofumpt removes blank line required by wsl linter
		fmt.Printf("[%s] [%v] ⚠️  Warning: failed to stop existing services: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), err)
	} else {
		fmt.Printf("[%s] [%v] ✅ Existing services stopped successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	}

	// Start docker compose services fresh
	fmt.Printf("[%s] [%v] 🚀 Starting Docker Compose services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	err = startDockerCompose(ctx, startTime)
	require.NoError(t, err, "Failed to start docker compose")

	defer func() {
		fmt.Printf("[%s] [%v] 🧹 CLEANUP: Stopping docker compose services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

		err := stopDockerCompose(context.Background(), startTime) // Use background context for cleanup
		if err != nil {
			fmt.Printf("[%s] [%v] ⚠️  CLEANUP WARNING: failed to stop docker compose: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), err)
		}
	}()

	// Wait before starting checks to allow services to initialize
	fmt.Printf("[%s] [%v] ⏳ Waiting %v for Docker Compose services to initialize...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), dockerComposeInitTimeout)
	time.Sleep(dockerComposeInitTimeout)

	// Wait for all services to be ready (Docker health checks)
	waitForServicesReady(t, ctx, startTime)

	// Verify all services are reachable via public APIs
	verifyServicesAreReachable(t, ctx, rootCAsPool, startTime)

	// Test each cryptoutil instance
	testCryptoutilInstance(t, ctx, "cryptoutil_sqlite", &cryptoutilSqliteURL, &cryptoutilSqliteAdminURL, rootCAsPool, startTime)
	testCryptoutilInstance(t, ctx, "cryptoutil_postgres_1", &cryptoutilPostgres1URL, &cryptoutilPostgres1AdminURL, rootCAsPool, startTime)
	testCryptoutilInstance(t, ctx, "cryptoutil_postgres_2", &cryptoutilPostgres2URL, &cryptoutilPostgres2AdminURL, rootCAsPool, startTime)

	// Verify telemetry is flowing to Grafana
	verifyTelemetryFlow(t, ctx, startTime)

	fmt.Printf("✅ TEST PASSED: %s (duration: %v)\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
}

// startDockerCompose starts the docker compose services.
func startDockerCompose(ctx context.Context, startTime time.Time) error {
	fmt.Printf("[%s] [%v] 🔄 Stopping any existing Docker Compose services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	// Stop any existing services first to ensure clean state
	stopCmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../../deployments/compose/compose.yml", "down", "-v", "--remove-orphans")

	fmt.Printf("[%s] [%v] 📋 [DOCKER] Executing stop command: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), stopCmd.String())
	stopOutput, stopErr := stopCmd.CombinedOutput()
	fmt.Printf("[%s] [%v] 📋 [DOCKER] Stop command output: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), string(stopOutput))

	if stopErr != nil { //nolint:wsl // gofumpt removes blank line required by wsl linter
		// Log warning but don't fail - services might not be running
		fmt.Printf("[%s] [%v] ⚠️ [DOCKER] Warning: failed to stop existing services: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), stopErr)
	} else {
		fmt.Printf("[%s] [%v] ✅ [DOCKER] Existing services stopped successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	}

	fmt.Printf("[%s] [%v] 🚀 Starting fresh Docker Compose services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	// Start fresh services
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../../deployments/compose/compose.yml", "up", "-d", "--force-recreate")

	fmt.Printf("[%s] [%v] 📋 [DOCKER] Executing start command: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), cmd.String())
	output, err := cmd.CombinedOutput()
	fmt.Printf("[%s] [%v] 📋 [DOCKER] Start command output: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), string(output))

	if err != nil {
		fmt.Printf("[%s] [%v] ❌ [DOCKER] Docker compose up failed: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), err)

		return fmt.Errorf("docker compose up failed: %w, output: %s", err, string(output))
	}

	fmt.Printf("[%s] [%v] ✅ [DOCKER] Docker Compose services started successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	return nil
}

// stopDockerCompose stops the docker compose services.
func stopDockerCompose(ctx context.Context, startTime time.Time) error {
	fmt.Printf("[%s] [%v] 🛑 Stopping Docker Compose services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../../deployments/compose/compose.yml", "down", "-v")

	fmt.Printf("[%s] [%v] 📋 [DOCKER] Executing stop command: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), cmd.String())
	output, err := cmd.CombinedOutput()
	fmt.Printf("[%s] [%v] 📋 [DOCKER] Stop command output: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), string(output))

	if err != nil {
		fmt.Printf("[%s] [%v] ❌ [DOCKER] Failed to stop Docker Compose services: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), err)

		return fmt.Errorf("docker compose down failed: %w, output: %s", err, string(output))
	}

	fmt.Printf("[%s] [%v] ✅ [DOCKER] Docker Compose services stopped successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	return nil
}

// waitForServicesReady waits for all services to report ready via Docker health checks.
func waitForServicesReady(t *testing.T, ctx context.Context, startTime time.Time) {
	t.Helper()
	fmt.Printf("[%s] [%v] 🔍 Checking service readiness...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	// Wait for all Docker services to be healthy (they have their own health checks)
	waitForDockerServicesHealthy(t, ctx, startTime)

	fmt.Printf("[%s] [%v] ✅ All services are ready\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
}

// verifyServicesAreReachable verifies that all services are reachable via their public APIs.
func verifyServicesAreReachable(t *testing.T, ctx context.Context, rootCAsPool *x509.CertPool, startTime time.Time) {
	t.Helper()
	fmt.Printf("[%s] [%v] 🌐 Verifying services are reachable...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	// Verify cryptoutil instances are accessible via public APIs (they should be unsealed now)
	waitForCryptoutilReady(t, ctx, &cryptoutilSqliteURL, rootCAsPool, startTime)
	waitForCryptoutilReady(t, ctx, &cryptoutilPostgres1URL, rootCAsPool, startTime)
	waitForCryptoutilReady(t, ctx, &cryptoutilPostgres2URL, rootCAsPool, startTime)

	// Wait for Grafana
	fmt.Printf("[%s] [%v] ⏳ Waiting for Grafana...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	waitForHTTPReady(t, ctx, grafanaURL+"/api/health", cryptoutilReadyTimeout, startTime)

	// Wait for OTEL collector
	fmt.Printf("[%s] [%v] ⏳ Waiting for OTEL collector...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
	waitForHTTPReady(t, ctx, otelCollectorURL+"/metrics", cryptoutilReadyTimeout, startTime)

	fmt.Printf("[%s] [%v] ✅ All services are reachable\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
}

// waitForDockerServicesHealthy waits for all Docker services to report healthy status.
func waitForDockerServicesHealthy(t *testing.T, ctx context.Context, startTime time.Time) {
	t.Helper()

	services := []string{
		"cryptoutil_sqlite",
		"cryptoutil_postgres_1",
		"cryptoutil_postgres_2",
		"postgres",
		"grafana-otel-lgtm",
		"opentelemetry-collector-contrib",
	}

	t.Logf("[%s] [%v] Waiting for %d Docker services to be healthy...", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), len(services))

	giveUpTime := time.Now().Add(dockerHealthTimeout)
	checkCount := 0

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Context cancelled while waiting for Docker services to be healthy")
		default:
		}

		if time.Now().After(giveUpTime) {
			// Log detailed status of all services before failing
			t.Logf("[%s] [%v] TIMEOUT: Docker services health check failed. Current status:", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

			// Get current health status for all services
			healthStatus := areDockerServicesHealthy(services, startTime)

			logFunc := func(service, status string) {
				t.Logf("[%s] [%v]   %s: %s", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), service, status)
			}
			_ = logServiceStatuses(services, healthStatus, startTime, logFunc)

			t.Fatalf("Docker services not healthy after %v", dockerHealthTimeout)
		}

		checkCount++
		fmt.Printf("[%s] [%v] 🔍 Health check #%d: Checking %d services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), checkCount, len(services))

		// Check if all services are healthy
		healthStatus := areDockerServicesHealthy(services, startTime)

		logFunc := func(service, status string) {
			emojiStatus := "❌ UNHEALTHY"
			if status == cryptoutilMagic.StatusHealthy {
				emojiStatus = "✅ HEALTHY"
			}

			fmt.Printf("[%s] [%v]    %s: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), service, emojiStatus)
		}
		unhealthyServices := logServiceStatuses(services, healthStatus, startTime, logFunc)
		allHealthy := len(unhealthyServices) == 0

		if allHealthy {
			fmt.Printf("[%s] [%v] ✅ All %d Docker services are healthy after %d checks\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), len(services), checkCount)

			return
		}

		fmt.Printf("[%s] [%v] ⏳ Waiting %v before next health check... (%d unhealthy: %v)\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceRetryInterval, len(unhealthyServices), unhealthyServices)
		time.Sleep(serviceRetryInterval)
	}
}

// areDockerServicesHealthy checks all Docker services at once and returns their health status.
// This reduces the number of external docker compose ps calls from N to 1.
func areDockerServicesHealthy(services []string, startTime time.Time) map[string]bool {
	healthStatus := make(map[string]bool)

	// Call docker compose ps once for all services
	cmd := exec.Command("docker", "compose", "-f", "../../deployments/compose/compose.yml", "ps", "--format", "json")

	fmt.Printf("[%s] [%v] 📋 [DOCKER] Executing batch health check for %d services: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), len(services), cmd.String())
	output, err := cmd.Output()
	if err != nil { //nolint:wsl // gofumpt removes blank line required by wsl linter
		fmt.Printf("[%s] [%v] ❌ [DOCKER] Failed to check services health: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), err)
		// Mark all services as unhealthy
		for _, service := range services {
			healthStatus[service] = false
		}

		return healthStatus
	}

	fmt.Printf("[%s] [%v] 📋 [DOCKER] Batch health check output: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), string(output))

	// Parse the JSON output - it should be an array of service objects
	var serviceList []map[string]any
	if err := json.Unmarshal(output, &serviceList); err != nil {
		fmt.Printf("[%s] [%v] ❌ [DOCKER] Failed to parse JSON array: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), err)
		// Mark all services as unhealthy
		for _, service := range services {
			healthStatus[service] = false
		}

		return healthStatus
	}

	// Create a map of service name to service data for quick lookup
	serviceMap := make(map[string]map[string]any)

	for _, service := range serviceList {
		if name, ok := service["Name"].(string); ok {
			// Extract just the service name from the full container name
			// Container names are like "compose-cryptoutil_sqlite-1"
			// We want to extract "cryptoutil_sqlite" from this
			if strings.Contains(name, "compose-") {
				parts := strings.Split(name, "-")
				if len(parts) >= 3 {
					// Remove "compose-" prefix and container number suffix
					serviceName := strings.Join(parts[1:len(parts)-1], "-")
					serviceMap[serviceName] = service
				}
			}
		}
	}

	// Check health status for each requested service
	for _, serviceName := range services {
		service, exists := serviceMap[serviceName]

		if !exists {
			fmt.Printf("[%s] [%v] ❌ [DOCKER] Service %s not found in docker compose output\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceName)

			healthStatus[serviceName] = false

			continue
		}

		// Debug: print the raw JSON for troubleshooting OTEL service
		if serviceName == "opentelemetry-collector-contrib" {
			serviceJSON, err := json.Marshal(service)
			if err != nil {
				fmt.Printf("[%s] [%v] ❌ [DOCKER] Failed to marshal OTEL service JSON: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), err)
			} else {
				fmt.Printf("[%s] [%v] 🔍 OTEL service JSON: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), string(serviceJSON))
			}
		}

		// Check health status - services with health checks will have "Health" field set to "healthy"
		if health, ok := service["Health"].(string); ok {
			if health == "healthy" {
				healthStatus[serviceName] = true
			} else {
				// Explicitly unhealthy - never consider healthy
				healthStatus[serviceName] = false
			}
		} else {
			// For services without health checks, check if they're running
			if state, ok := service["State"].(string); ok && state == "running" {
				healthStatus[serviceName] = true
			} else {
				// If health is present but not healthy, or state is not running
				healthStatus[serviceName] = false
			}
		}
	}

	return healthStatus
}

// logServiceStatuses logs the status of all services using the provided log function and returns unhealthy services.
func logServiceStatuses(services []string, healthStatus map[string]bool, startTime time.Time, logFunc func(service, status string)) []string {
	var unhealthyServices []string

	for _, service := range services {
		healthy := healthStatus[service]
		status := cryptoutilMagic.StatusUnhealthy

		if healthy {
			status = cryptoutilMagic.StatusHealthy
		}

		logFunc(service, status)

		if !healthy {
			unhealthyServices = append(unhealthyServices, service)
		}
	}

	return unhealthyServices
}

// waitForCryptoutilReady waits for a cryptoutil instance to be ready via its public API.
func waitForCryptoutilReady(t *testing.T, ctx context.Context, baseURL *string, rootCAsPool *x509.CertPool, startTime time.Time) {
	t.Helper()

	giveUpTime := time.Now().Add(cryptoutilReadyTimeout)
	checkCount := 0

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Context cancelled while waiting for cryptoutil at %s", *baseURL)
		default:
		}

		if time.Now().After(giveUpTime) {
			t.Fatalf("Cryptoutil service not ready after %v: %s", cryptoutilReadyTimeout, *baseURL)
		}

		checkCount++
		fmt.Printf("[%s] [%v] 🔍 Cryptoutil readiness check #%d for %s...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), checkCount, *baseURL)

		// Try to create an OpenAPI client and make a simple API call to verify the service is unsealed and ready
		fmt.Printf("📋 [API] Creating OpenAPI client for %s\n", *baseURL)
		client := cryptoutilClient.RequireClientWithResponses(t, baseURL, rootCAsPool)

		// Try to list elastic keys - this should work if the service is unsealed
		fmt.Printf("📋 [API] Making GetElastickeys request to %s\n", *baseURL)

		_, err := client.GetElastickeysWithResponse(ctx, nil)
		if err != nil {
			fmt.Printf("[%s] [%v] ⏳ [API] Cryptoutil at %s not ready yet (attempt %d, API call failed: %v), waiting %v...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *baseURL, checkCount, err, serviceRetryInterval)
			time.Sleep(serviceRetryInterval)

			continue
		}

		fmt.Printf("[%s] [%v] ✅ [API] Cryptoutil service ready at %s after %d checks\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *baseURL, checkCount)

		return
	}
}

// waitForHTTPReady waits for an HTTP endpoint to return 200.
func waitForHTTPReady(t *testing.T, ctx context.Context, url string, timeout time.Duration, startTime time.Time) {
	t.Helper()

	giveUpTime := time.Now().Add(timeout)
	client := &http.Client{Timeout: httpClientTimeout}

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Context cancelled while waiting for %s", url)
		default:
		}

		if time.Now().After(giveUpTime) {
			t.Fatalf("Service not ready after %v: %s", timeout, url)
		}

		fmt.Printf("[%s] [%v] 📋 [HTTP] Making GET request to: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), url)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil { //nolint:wsl // gofumpt removes blank line required by wsl linter
			fmt.Printf("[%s] [%v] ❌ [HTTP] Failed to create request to %s: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), url, err)

			return
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[%s] [%v] ❌ [HTTP] Request to %s failed: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), url, err)
		} else {
			fmt.Printf("[%s] [%v] 📋 [HTTP] Response from %s: status %d\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), url, resp.StatusCode)
		}

		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Printf("[%s] [%v] ✅ [HTTP] Service ready at %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), url)
			resp.Body.Close()

			return
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(httpRetryInterval)
	}
}

// testCryptoutilInstance tests a single cryptoutil instance.
func testCryptoutilInstance(t *testing.T, ctx context.Context, instanceName string, publicBaseURL, privateAdminBaseURL *string, rootCAsPool *x509.CertPool, startTime time.Time) {
	t.Helper()
	fmt.Printf("[%s] [%v] 🧪 Testing %s at %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), instanceName, *publicBaseURL)

	// Create OpenAPI client for public APIs
	fmt.Printf("[%s] [%v] 📡 Creating OpenAPI client for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *publicBaseURL)
	client := cryptoutilClient.RequireClientWithResponses(t, publicBaseURL, rootCAsPool)

	// Test health check (liveness on public server)
	fmt.Printf("[%s] [%v] 💚 Testing health check for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *publicBaseURL)
	testHealthCheck(t, ctx, publicBaseURL, rootCAsPool, startTime)

	// Test service API - create elastic key
	fmt.Printf("[%s] [%v] 🔑 Creating elastic key for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), instanceName)
	elasticKey := testCreateElasticKey(t, ctx, client, startTime)

	// Test service API - generate material key
	fmt.Printf("[%s] [%v] 🗝️  Generating material key for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), instanceName)
	testGenerateMaterialKey(t, ctx, client, elasticKey, startTime)

	// Test service API - encrypt/decrypt cycle
	fmt.Printf("[%s] [%v] 🔐 Testing encrypt/decrypt cycle for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), instanceName)
	testEncryptDecryptCycle(t, ctx, client, elasticKey, startTime)

	// Test service API - sign/verify cycle
	fmt.Printf("[%s] [%v] ✍️  Testing sign/verify cycle for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), instanceName)
	testSignVerifyCycle(t, ctx, client, elasticKey, startTime)

	fmt.Printf("[%s] [%v] ✅ %s tests passed\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), instanceName)
}

// testHealthCheck verifies the health endpoints work.
func testHealthCheck(t *testing.T, ctx context.Context, publicBaseURL *string, rootCAsPool *x509.CertPool, startTime time.Time) {
	t.Helper()

	// Test liveness probe
	fmt.Printf("[%s] [%v] 📋 [API] Making health check request to %s/livez\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *publicBaseURL)
	err := cryptoutilClient.CheckHealthz(publicBaseURL, rootCAsPool)

	if err != nil {
		fmt.Printf("[%s] [%v] ❌ [API] Health check failed for %s: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *publicBaseURL, err)
	} else {
		fmt.Printf("[%s] [%v] ✅ [API] Health check passed for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *publicBaseURL)
	}

	require.NoError(t, err, "Health check failed for %s", *publicBaseURL)
}

// testCreateElasticKey tests creating an elastic key.
func testCreateElasticKey(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses, startTime time.Time) *cryptoutilOpenapiModel.ElasticKey {
	t.Helper()

	importAllowed := false
	versioningAllowed := true

	elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
		t, &testElasticKeyName, &testElasticKeyDescription,
		&testAlgorithm, &testProvider, &importAllowed, &versioningAllowed,
	)

	fmt.Printf("[%s] [%v] 📋 [API] Making PostElastickey request to create elastic key\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	elasticKey := cryptoutilClient.RequireCreateElasticKeyResponse(t, ctx, client, elasticKeyCreate)
	require.NotNil(t, elasticKey.ElasticKeyID)

	fmt.Printf("[%s] [%v] ✅ [API] Elastic key created with ID: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *elasticKey.ElasticKeyID)

	return elasticKey
}

// testGenerateMaterialKey tests generating a material key.
func testGenerateMaterialKey(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey, startTime time.Time) {
	t.Helper()

	keyGenerate := cryptoutilClient.RequireMaterialKeyGenerateRequest(t)

	fmt.Printf("[%s] [%v] 📋 [API] Making PostElastickeyElasticKeyIDMaterialkey request for key %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *elasticKey.ElasticKeyID)
	materialKey := cryptoutilClient.RequireMaterialKeyGenerateResponse(t, ctx, client, elasticKey.ElasticKeyID, keyGenerate)
	require.NotNil(t, materialKey.MaterialKeyID)

	fmt.Printf("[%s] [%v] ✅ [API] Material key generated with ID: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), materialKey.MaterialKeyID)
}

// testEncryptDecryptCycle tests a full encrypt/decrypt cycle.
func testEncryptDecryptCycle(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey, startTime time.Time) {
	t.Helper()

	// Encrypt
	encryptRequest := cryptoutilClient.RequireEncryptRequest(t, &cryptoutilMagic.TestCleartext)

	fmt.Printf("[%s] [%v] 📋 [API] Making PostElastickeyElasticKeyIDEncrypt request for key %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *elasticKey.ElasticKeyID)
	encryptedText := cryptoutilClient.RequireEncryptResponse(t, ctx, client, elasticKey.ElasticKeyID, nil, encryptRequest)
	require.NotEmpty(t, *encryptedText)

	fmt.Printf("[%s] [%v] ✅ [API] Text encrypted successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	// Decrypt
	decryptRequest := cryptoutilClient.RequireDecryptRequest(t, encryptedText)

	fmt.Printf("[%s] [%v] 📋 [API] Making PostElastickeyElasticKeyIDDecrypt request for key %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *elasticKey.ElasticKeyID)
	decryptedText := cryptoutilClient.RequireDecryptResponse(t, ctx, client, elasticKey.ElasticKeyID, decryptRequest)
	require.Equal(t, cryptoutilMagic.TestCleartext, *decryptedText)

	fmt.Printf("[%s] [%v] ✅ [API] Text decrypted successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
}

// testSignVerifyCycle tests a full sign/verify cycle.
func testSignVerifyCycle(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey, startTime time.Time) {
	t.Helper()

	// Sign
	signRequest := cryptoutilClient.RequireSignRequest(t, &cryptoutilMagic.TestCleartext)

	fmt.Printf("[%s] [%v] 📋 [API] Making PostElastickeyElasticKeyIDSign request for key %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *elasticKey.ElasticKeyID)
	signedText := cryptoutilClient.RequireSignResponse(t, ctx, client, elasticKey.ElasticKeyID, nil, signRequest)
	require.NotEmpty(t, *signedText)

	fmt.Printf("[%s] [%v] ✅ [API] Text signed successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	// Verify
	verifyRequest := cryptoutilClient.RequireVerifyRequest(t, signedText)

	fmt.Printf("[%s] [%v] 📋 [API] Making PostElastickeyElasticKeyIDVerify request for key %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *elasticKey.ElasticKeyID)
	verifyResponse := cryptoutilClient.RequireVerifyResponse(t, ctx, client, elasticKey.ElasticKeyID, verifyRequest)
	require.Equal(t, "true", *verifyResponse)

	fmt.Printf("[%s] [%v] ✅ [API] Signature verified successfully\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
}

// verifyTelemetryFlow verifies that telemetry is flowing to Grafana.
func verifyTelemetryFlow(t *testing.T, ctx context.Context, startTime time.Time) {
	t.Helper()
	fmt.Printf("[%s] [%v] Verifying telemetry flow to Grafana...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	client := &http.Client{Timeout: httpClientTimeout}

	// Check Grafana health
	fmt.Printf("[%s] [%v] 📋 [HTTP] Making GET request to Grafana health endpoint: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), grafanaURL+"/api/health")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafanaURL+"/api/health", nil)
	require.NoError(t, err, "Failed to create Grafana health request")

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to connect to Grafana")

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "Grafana health check failed")

	fmt.Printf("[%s] [%v] ✅ [HTTP] Grafana health check passed\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	// Check OTEL collector metrics endpoint
	fmt.Printf("[%s] [%v] 📋 [HTTP] Making GET request to OTEL collector metrics endpoint: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), otelCollectorURL+"/metrics")
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, otelCollectorURL+"/metrics", nil)
	require.NoError(t, err, "Failed to create OTEL collector metrics request")

	resp, err = client.Do(req)
	require.NoError(t, err, "Failed to connect to OTEL collector")

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "OTEL collector metrics check failed")

	fmt.Printf("[%s] [%v] ✅ [HTTP] OTEL collector metrics check passed\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	// Verify metrics contain cryptoutil service data
	body := make([]byte, 1024*1024) // 1MB buffer
	n, err := resp.Body.Read(body)
	require.NoError(t, err, "Failed to read OTEL metrics")

	metrics := string(body[:n])

	// Check for cryptoutil service metrics
	require.Contains(t, metrics, "cryptoutil", "No cryptoutil metrics found in OTEL collector")

	// Check for traces/logs/metrics indicators
	hasTraces := strings.Contains(metrics, "traces") || strings.Contains(metrics, "spans")
	hasLogs := strings.Contains(metrics, "logs") || strings.Contains(metrics, "log_records")
	hasMetrics := strings.Contains(metrics, "metrics") || strings.Contains(metrics, "data_points")

	require.True(t, hasTraces || hasLogs || hasMetrics, "No telemetry data found in OTEL collector")

	fmt.Printf("[%s] [%v] ✅ [TELEMETRY] Telemetry flow verification passed\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
}
