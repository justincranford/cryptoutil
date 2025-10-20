package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilClient "cryptoutil/internal/client"

	"github.com/stretchr/testify/require"
)

const (
	// Docker compose service names and ports.
	cryptoutilSqlitePort    = "8080"
	cryptoutilPostgres1Port = "8081"
	cryptoutilPostgres2Port = "8082"
	cryptoutilPrivatePort   = "9090"
	grafanaPort             = "3000"
	otelCollectorPort       = "8888"

	// Test timeouts.
	composeUpTimeout       = 5 * time.Minute
	dockerHealthTimeout    = 30 * time.Second // Docker services should be healthy in under 20s
	cryptoutilReadyTimeout = 2 * time.Minute  // Cryptoutil needs time to unseal
	testExecutionTimeout   = 10 * time.Minute
	httpClientTimeout      = 10 * time.Second
	serviceRetryInterval   = 2 * time.Second // Check more frequently
	httpRetryInterval      = 1 * time.Second

	// Test data.
	testElasticKeyName        = "e2e-test-key"
	testElasticKeyDescription = "E2E integration test key"
	testAlgorithm             = "RSA"
	testProvider              = "GO"
	testCleartext             = "Hello, World!"
)

var (
	// Public API URLs (ports 8080+).
	cryptoutilSqliteURL    = "https://127.0.0.1:" + cryptoutilSqlitePort
	cryptoutilPostgres1URL = "https://127.0.0.1:" + cryptoutilPostgres1Port
	cryptoutilPostgres2URL = "https://127.0.0.1:" + cryptoutilPostgres2Port

	// Private admin API URLs (port 9090 inside containers).
	cryptoutilSqliteAdminURL    = "https://127.0.0.1:" + cryptoutilPrivatePort
	cryptoutilPostgres1AdminURL = "https://127.0.0.1:" + cryptoutilPrivatePort
	cryptoutilPostgres2AdminURL = "https://127.0.0.1:" + cryptoutilPrivatePort

	grafanaURL       = "http://127.0.0.1:" + grafanaPort
	otelCollectorURL = "http://127.0.0.1:" + otelCollectorPort

	// Test data variables (so we can take their addresses).
	testElasticKeyNameVar        = testElasticKeyName
	testElasticKeyDescriptionVar = testElasticKeyDescription
	testAlgorithmVar             = testAlgorithm
	testProviderVar              = testProvider
	testCleartextVar             = testCleartext
)

// loadTestCertificates loads the test TLS certificates for HTTPS validation.
func loadTestCertificates(t *testing.T) *x509.CertPool {
	t.Helper()

	// Load the public server certificate as the root CA for testing
	certData, err := os.ReadFile("../tls_public_server_certificate_0.pem")
	require.NoError(t, err, "Failed to read test certificate file")

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(certData)
	require.True(t, ok, "Failed to parse test certificate")

	return certPool
}

// TestE2EIntegration performs end-to-end testing of all cryptoutil instances with telemetry verification.
func TestE2EIntegration(t *testing.T) {
	t.Parallel()

	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), testExecutionTimeout)
	defer cancel()

	// Load test certificates for TLS validation
	rootCAsPool := loadTestCertificates(t)

	// Start docker compose
	t.Logf("[%s] [%v] Starting docker compose services...", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	err := startDockerCompose(ctx)
	require.NoError(t, err, "Failed to start docker compose")

	// Give Docker Compose time to start services
	t.Log("Waiting for Docker Compose services to initialize...")
	time.Sleep(30 * time.Second)

	defer func() {
		t.Log("Stopping docker compose services...")

		err := stopDockerCompose(context.Background()) // Use background context for cleanup
		if err != nil {
			t.Logf("Warning: failed to stop docker compose: %v", err)
		}
	}()

	// Wait for all services to be ready (Docker health checks)
	waitForServicesReady(t, ctx, startTime)

	// Verify all services are reachable via public APIs
	verifyServicesAreReachable(t, ctx, rootCAsPool, startTime)

	// Test each cryptoutil instance
	testCryptoutilInstance(t, ctx, "cryptoutil_sqlite", &cryptoutilSqliteURL, &cryptoutilSqliteAdminURL, rootCAsPool)
	testCryptoutilInstance(t, ctx, "cryptoutil_postgres_1", &cryptoutilPostgres1URL, &cryptoutilPostgres1AdminURL, rootCAsPool)
	testCryptoutilInstance(t, ctx, "cryptoutil_postgres_2", &cryptoutilPostgres2URL, &cryptoutilPostgres2AdminURL, rootCAsPool)

	// Verify telemetry is flowing to Grafana
	verifyTelemetryFlow(t, ctx)
}

// startDockerCompose starts the docker compose services.
func startDockerCompose(ctx context.Context) error {
	fmt.Println("🔄 Stopping any existing Docker Compose services...")
	// Stop any existing services first to ensure clean state
	stopCmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../deployments/compose/compose.yml", "down", "-v", "--remove-orphans")

	stopOutput, stopErr := stopCmd.CombinedOutput()
	if stopErr != nil {
		// Log warning but don't fail - services might not be running
		fmt.Printf("⚠️  Warning: failed to stop existing services: %v, output: %s\n", stopErr, string(stopOutput))
	} else {
		fmt.Println("✅ Existing services stopped successfully")
	}

	fmt.Println("🚀 Starting fresh Docker Compose services...")
	// Start fresh services
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../deployments/compose/compose.yml", "up", "-d", "--force-recreate")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w, output: %s", err, string(output))
	}

	fmt.Printf("✅ Docker Compose services started successfully\n")

	return nil
}

// stopDockerCompose stops the docker compose services.
func stopDockerCompose(ctx context.Context) error {
	fmt.Println("🛑 Stopping Docker Compose services...")

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../deployments/compose/compose.yml", "down", "-v")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to stop Docker Compose services: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))

		return fmt.Errorf("docker compose down failed: %w, output: %s", err, string(output))
	}

	fmt.Println("✅ Docker Compose services stopped successfully")

	return nil
}

// waitForServicesReady waits for all services to report ready via Docker health checks.
func waitForServicesReady(t *testing.T, ctx context.Context, startTime time.Time) {
	t.Helper()
	t.Logf("[%s] [%v] Waiting for services to be ready...", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))

	// Wait for all Docker services to be healthy (they have their own health checks)
	waitForDockerServicesHealthy(t, ctx, startTime)

	t.Logf("[%s] [%v] All services are ready", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second))
}

// verifyServicesAreReachable verifies that all services are reachable via their public APIs.
func verifyServicesAreReachable(t *testing.T, ctx context.Context, rootCAsPool *x509.CertPool, startTime time.Time) {
	t.Helper()
	t.Log("Verifying services are reachable...")

	// Verify cryptoutil instances are accessible via public APIs (they should be unsealed now)
	waitForCryptoutilReady(t, ctx, &cryptoutilSqliteURL, rootCAsPool, startTime)
	waitForCryptoutilReady(t, ctx, &cryptoutilPostgres1URL, rootCAsPool, startTime)
	waitForCryptoutilReady(t, ctx, &cryptoutilPostgres2URL, rootCAsPool, startTime)

	// Wait for Grafana
	waitForHTTPReady(t, ctx, grafanaURL+"/api/health", cryptoutilReadyTimeout)

	// Wait for OTEL collector
	waitForHTTPReady(t, ctx, otelCollectorURL+"/metrics", cryptoutilReadyTimeout)

	t.Log("All services are reachable")
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

			for _, service := range services {
				healthy := isDockerServiceHealthy(service, startTime)
				status := "UNHEALTHY"

				if healthy {
					status = "HEALTHY"
				}

				t.Logf("  %s: %s", service, status)
			}

			t.Fatalf("Docker services not healthy after %v", dockerHealthTimeout)
		}

		checkCount++
		fmt.Printf("[%s] [%v] 🔍 Health check #%d: Checking %d services...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), checkCount, len(services))

		// Check if all services are healthy
		allHealthy := true
		unhealthyServices := []string{}

		for _, service := range services {
			healthy := isDockerServiceHealthy(service, startTime)
			status := "❌ UNHEALTHY"

			if healthy {
				status = "✅ HEALTHY"
			}

			fmt.Printf("[%s] [%v]    %s: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), service, status)

			if !healthy {
				allHealthy = false

				unhealthyServices = append(unhealthyServices, service)
			}
		}

		if allHealthy {
			fmt.Printf("[%s] [%v] ✅ All %d Docker services are healthy after %d checks\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), len(services), checkCount)

			return
		}

		fmt.Printf("[%s] [%v] ⏳ Waiting %v before next health check... (%d unhealthy: %v)\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceRetryInterval, len(unhealthyServices), unhealthyServices)
		time.Sleep(serviceRetryInterval)
	}
}

// isDockerServiceHealthy checks if a specific Docker service is healthy.
func isDockerServiceHealthy(serviceName string, startTime time.Time) bool {
	cmd := exec.Command("docker", "compose", "-f", "../deployments/compose/compose.yml", "ps", serviceName, "--format", "json")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[%s] [%v] ❌ Failed to check %s health: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceName, err)

		return false
	}

	// Parse the JSON output to check health status
	var service map[string]interface{}
	if err := json.Unmarshal(output, &service); err != nil {
		fmt.Printf("[%s] [%v] ❌ Failed to parse JSON for %s: %v\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceName, err)

		return false
	}

	// Debug: print the raw JSON for troubleshooting
	if serviceName == "opentelemetry-collector-contrib" {
		fmt.Printf("[%s] [%v] 🔍 OTEL service JSON: %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), string(output))
	}

	// Check health status - services with health checks will have "Health" field set to "healthy"
	if health, ok := service["Health"].(string); ok && health == "healthy" {
		fmt.Printf("[%s] [%v] 📊 %s health field: '%s' -> true\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceName, health)

		return true
	}

	// For services without health checks or with empty health, check if they're running
	if state, ok := service["State"].(string); ok && state == "running" {
		fmt.Printf("[%s] [%v] 📊 %s state field: '%s' -> true\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceName, state)

		return true
	}

	// If health is present but not healthy, or state is not running
	if health, ok := service["Health"].(string); ok {
		fmt.Printf("[%s] [%v] 📊 %s health field: '%s' -> false\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceName, health)
	} else {
		fmt.Printf("[%s] [%v] ❌ No Health or State field found for %s\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), serviceName)
	}

	return false
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

		// Check if the public server is responding (skip TLS verification for test)
		client := &http.Client{
			Timeout: httpClientTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // G402: TLS InsecureSkipVerify set true.
			},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, *baseURL+"/healthz", nil)
		if err != nil {
			fmt.Printf("[%s] [%v] ⏳ Cryptoutil at %s not ready yet (attempt %d, err: %v), waiting %v...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *baseURL, checkCount, err, serviceRetryInterval)
			time.Sleep(serviceRetryInterval)

			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[%s] [%v] ⏳ Cryptoutil at %s not ready yet (attempt %d, err: %v), waiting %v...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *baseURL, checkCount, err, serviceRetryInterval)
			time.Sleep(serviceRetryInterval)

			continue
		}

		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("[%s] [%v] ✅ Cryptoutil service ready at %s after %d checks\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *baseURL, checkCount)

			return
		}

		fmt.Printf("[%s] [%v] ⏳ Cryptoutil at %s not ready yet (attempt %d, status: %d), waiting %v...\n", time.Now().Format("15:04:05"), time.Since(startTime).Round(time.Second), *baseURL, checkCount, resp.StatusCode, serviceRetryInterval)
		time.Sleep(serviceRetryInterval)
	}
}

// waitForHTTPReady waits for an HTTP endpoint to return 200.
func waitForHTTPReady(t *testing.T, ctx context.Context, url string, timeout time.Duration) {
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
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
func testCryptoutilInstance(t *testing.T, ctx context.Context, instanceName string, publicBaseURL *string, privateAdminBaseURL *string, rootCAsPool *x509.CertPool) {
	t.Helper()
	t.Logf("Testing %s at %s", instanceName, *publicBaseURL)

	// Create OpenAPI client for public APIs
	client := cryptoutilClient.RequireClientWithResponses(t, publicBaseURL, rootCAsPool)

	// Test health check (liveness on public server)
	testHealthCheck(t, ctx, publicBaseURL, rootCAsPool)

	// Test service API - create elastic key
	elasticKey := testCreateElasticKey(t, ctx, client)

	// Test service API - generate material key
	testGenerateMaterialKey(t, ctx, client, elasticKey)

	// Test service API - encrypt/decrypt cycle
	testEncryptDecryptCycle(t, ctx, client, elasticKey)

	// Test service API - sign/verify cycle
	testSignVerifyCycle(t, ctx, client, elasticKey)

	t.Logf("✓ %s tests passed", instanceName)
}

// testHealthCheck verifies the health endpoints work.
func testHealthCheck(t *testing.T, ctx context.Context, publicBaseURL *string, rootCAsPool *x509.CertPool) {
	t.Helper()

	// Test liveness probe
	err := cryptoutilClient.CheckHealthz(publicBaseURL, rootCAsPool)
	require.NoError(t, err, "Health check failed for %s", *publicBaseURL)
	// Note: Readiness checks are performed by Docker healthchecks on private server (9090)
	// Public server (8080+) does not implement readiness endpoints
}

// testCreateElasticKey tests creating an elastic key.
func testCreateElasticKey(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses) *cryptoutilOpenapiModel.ElasticKey {
	t.Helper()

	importAllowed := false
	versioningAllowed := true

	elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
		t, &testElasticKeyNameVar, &testElasticKeyDescriptionVar,
		&testAlgorithmVar, &testProviderVar, &importAllowed, &versioningAllowed,
	)

	elasticKey := cryptoutilClient.RequireCreateElasticKeyResponse(t, ctx, client, elasticKeyCreate)
	require.NotNil(t, elasticKey.ElasticKeyID)

	return elasticKey
}

// testGenerateMaterialKey tests generating a material key.
func testGenerateMaterialKey(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	t.Helper()

	keyGenerate := cryptoutilClient.RequireMaterialKeyGenerateRequest(t)
	materialKey := cryptoutilClient.RequireMaterialKeyGenerateResponse(t, ctx, client, elasticKey.ElasticKeyID, keyGenerate)
	require.NotNil(t, materialKey.MaterialKeyID)
}

// testEncryptDecryptCycle tests a full encrypt/decrypt cycle.
func testEncryptDecryptCycle(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	t.Helper()

	// Encrypt
	encryptRequest := cryptoutilClient.RequireEncryptRequest(t, &testCleartextVar)
	encryptedText := cryptoutilClient.RequireEncryptResponse(t, ctx, client, elasticKey.ElasticKeyID, nil, encryptRequest)
	require.NotEmpty(t, *encryptedText)

	// Decrypt
	decryptRequest := cryptoutilClient.RequireDecryptRequest(t, encryptedText)
	decryptedText := cryptoutilClient.RequireDecryptResponse(t, ctx, client, elasticKey.ElasticKeyID, decryptRequest)
	require.Equal(t, testCleartext, *decryptedText)
}

// testSignVerifyCycle tests a full sign/verify cycle.
func testSignVerifyCycle(t *testing.T, ctx context.Context, client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	t.Helper()

	// Sign
	signRequest := cryptoutilClient.RequireSignRequest(t, &testCleartextVar)
	signedText := cryptoutilClient.RequireSignResponse(t, ctx, client, elasticKey.ElasticKeyID, nil, signRequest)
	require.NotEmpty(t, *signedText)

	// Verify
	verifyRequest := cryptoutilClient.RequireVerifyRequest(t, signedText)
	verifyResponse := cryptoutilClient.RequireVerifyResponse(t, ctx, client, elasticKey.ElasticKeyID, verifyRequest)
	require.Equal(t, "true", *verifyResponse)
}

// verifyTelemetryFlow verifies that telemetry is flowing to Grafana.
func verifyTelemetryFlow(t *testing.T, ctx context.Context) {
	t.Helper()
	t.Log("Verifying telemetry flow to Grafana...")

	client := &http.Client{Timeout: httpClientTimeout}

	// Check Grafana health
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafanaURL+"/api/health", nil)
	require.NoError(t, err, "Failed to create Grafana health request")

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to connect to Grafana")

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "Grafana health check failed")

	// Check OTEL collector metrics endpoint
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, otelCollectorURL+"/metrics", nil)
	require.NoError(t, err, "Failed to create OTEL collector metrics request")

	resp, err = client.Do(req)
	require.NoError(t, err, "Failed to connect to OTEL collector")

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "OTEL collector metrics check failed")

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

	t.Log("✓ Telemetry flow verification passed")
}
