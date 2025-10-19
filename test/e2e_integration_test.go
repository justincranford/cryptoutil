package test

import (
	"context"
	"crypto/x509"
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
	grafanaPort             = "3000"
	otelCollectorPort       = "8888"

	// Test timeouts.
	composeUpTimeout     = 5 * time.Minute
	serviceReadyTimeout  = 2 * time.Minute
	testExecutionTimeout = 10 * time.Minute

	// Test data.
	testElasticKeyName        = "e2e-test-key"
	testElasticKeyDescription = "E2E integration test key"
	testAlgorithm             = "RSA"
	testProvider              = "GO"
	testCleartext             = "Hello, World!"
)

var (
	cryptoutilSqliteURL    = "https://localhost:" + cryptoutilSqlitePort
	cryptoutilPostgres1URL = "https://localhost:" + cryptoutilPostgres1Port
	cryptoutilPostgres2URL = "https://localhost:" + cryptoutilPostgres2Port
	grafanaURL             = "http://localhost:" + grafanaPort
	otelCollectorURL       = "http://localhost:" + otelCollectorPort

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

	ctx, cancel := context.WithTimeout(context.Background(), testExecutionTimeout)
	defer cancel()

	// Load test certificates for TLS validation
	rootCAsPool := loadTestCertificates(t)

	// Start docker compose
	t.Log("Starting docker compose services...")

	err := startDockerCompose(ctx)
	require.NoError(t, err, "Failed to start docker compose")

	defer func() {
		t.Log("Stopping docker compose services...")

		err := stopDockerCompose(context.Background()) // Use background context for cleanup
		if err != nil {
			t.Logf("Warning: failed to stop docker compose: %v", err)
		}
	}()

	// Wait for all services to be ready
	waitForServicesReady(t, ctx, rootCAsPool)

	// Test each cryptoutil instance
	testCryptoutilInstance(t, ctx, "cryptoutil_sqlite", &cryptoutilSqliteURL, rootCAsPool)
	testCryptoutilInstance(t, ctx, "cryptoutil_postgres_1", &cryptoutilPostgres1URL, rootCAsPool)
	testCryptoutilInstance(t, ctx, "cryptoutil_postgres_2", &cryptoutilPostgres2URL, rootCAsPool)

	// Verify telemetry is flowing to Grafana
	verifyTelemetryFlow(t, ctx)
}

// startDockerCompose starts the docker compose services.
func startDockerCompose(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../deployments/compose/compose.yml", "up", "-d")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w, output: %s", err, string(output))
	}

	return nil
}

// stopDockerCompose stops the docker compose services.
func stopDockerCompose(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../deployments/compose/compose.yml", "down", "-v")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose down failed: %w, output: %s", err, string(output))
	}

	return nil
}

// waitForServicesReady waits for all services to report ready.
func waitForServicesReady(t *testing.T, ctx context.Context, rootCAsPool *x509.CertPool) {
	t.Helper()
	t.Log("Waiting for services to be ready...")

	// Wait for cryptoutil instances
	cryptoutilClient.WaitUntilReady(&cryptoutilSqliteURL, serviceReadyTimeout, 5*time.Second, rootCAsPool)
	cryptoutilClient.WaitUntilReady(&cryptoutilPostgres1URL, serviceReadyTimeout, 5*time.Second, rootCAsPool)
	cryptoutilClient.WaitUntilReady(&cryptoutilPostgres2URL, serviceReadyTimeout, 5*time.Second, rootCAsPool)

	// Wait for Grafana
	waitForHTTPReady(t, ctx, grafanaURL+"/api/health", serviceReadyTimeout)

	// Wait for OTEL collector
	waitForHTTPReady(t, ctx, otelCollectorURL+"/metrics", serviceReadyTimeout)

	t.Log("All services are ready")
}

// waitForHTTPReady waits for an HTTP endpoint to return 200.
func waitForHTTPReady(t *testing.T, ctx context.Context, url string, timeout time.Duration) {
	t.Helper()

	giveUpTime := time.Now().Add(timeout)
	client := &http.Client{Timeout: 5 * time.Second}

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

		time.Sleep(2 * time.Second)
	}
}

// testCryptoutilInstance tests a single cryptoutil instance.
func testCryptoutilInstance(t *testing.T, ctx context.Context, instanceName string, baseURL *string, rootCAsPool *x509.CertPool) {
	t.Helper()
	t.Logf("Testing %s at %s", instanceName, *baseURL)

	// Create OpenAPI client
	client := cryptoutilClient.RequireClientWithResponses(t, baseURL, rootCAsPool)

	// Test health check
	testHealthCheck(t, ctx, baseURL, rootCAsPool)

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
func testHealthCheck(t *testing.T, ctx context.Context, baseURL *string, rootCAsPool *x509.CertPool) {
	t.Helper()

	// Test liveness probe
	err := cryptoutilClient.CheckHealthz(baseURL, rootCAsPool)
	require.NoError(t, err, "Health check failed for %s", *baseURL)

	// Test readiness probe
	err = cryptoutilClient.CheckReadyz(baseURL, rootCAsPool)
	require.NoError(t, err, "Readiness check failed for %s", *baseURL)
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

	client := &http.Client{Timeout: 10 * time.Second}

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
