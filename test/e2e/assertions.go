//go:build e2e

package test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	cryptoutilClient "cryptoutil/internal/client"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

// ServiceAssertions provides common assertions for service testing.
type ServiceAssertions struct {
	t         *testing.T
	startTime time.Time
	logFile   *os.File
}

// NewServiceAssertions creates a new service assertions helper.
func NewServiceAssertions(t *testing.T, startTime time.Time, logFile *os.File) *ServiceAssertions {
	t.Helper()

	return &ServiceAssertions{
		t:         t,
		startTime: startTime,
		logFile:   logFile,
	}
}

// AssertCryptoutilHealth checks that a cryptoutil instance is healthy.
func (a *ServiceAssertions) AssertCryptoutilHealth(baseURL string, rootCAsPool *x509.CertPool) {
	a.log("💚 Testing health check for %s", baseURL)
	err := cryptoutilClient.CheckHealthz(&baseURL, rootCAsPool)
	require.NoError(a.t, err, "Health check failed for %s", baseURL)
	a.log("✅ Health check passed for %s", baseURL)
}

// AssertCryptoutilReady waits for a cryptoutil instance to be ready.
func (a *ServiceAssertions) AssertCryptoutilReady(ctx context.Context, baseURL string, rootCAsPool *x509.CertPool) {
	a.log("⏳ Waiting for cryptoutil ready at %s", baseURL)

	giveUpTime := time.Now().Add(cryptoutilMagic.TestTimeoutCryptoutilReady)
	checkCount := 0

	for {
		require.False(a.t, time.Now().After(giveUpTime), "Cryptoutil service not ready after %v: %s", cryptoutilMagic.TestTimeoutCryptoutilReady, baseURL)

		checkCount++
		a.log("🔍 Cryptoutil readiness check #%d for %s", checkCount, baseURL)

		client := cryptoutilClient.RequireClientWithResponses(a.t, &baseURL, rootCAsPool)
		_, err := client.GetElastickeysWithResponse(ctx, nil)

		if err == nil {
			a.log("✅ Cryptoutil service ready at %s after %d checks", baseURL, checkCount)

			return
		}

		a.log("⏳ Cryptoutil at %s not ready yet (attempt %d), waiting %v...",
			baseURL, checkCount, cryptoutilMagic.TestTimeoutServiceRetry)
		time.Sleep(cryptoutilMagic.TestTimeoutServiceRetry)
	}
}

// AssertHTTPReady waits for an HTTP endpoint to return 200.
func (a *ServiceAssertions) AssertHTTPReady(ctx context.Context, url string, timeout time.Duration) {
	a.log("⏳ Waiting for HTTP endpoint ready: %s", url)

	giveUpTime := time.Now().Add(timeout)
	client := &http.Client{Timeout: cryptoutilMagic.TestTimeoutHTTPClient}

	for {
		require.False(a.t, time.Now().After(giveUpTime), "Service not ready after %v: %s", timeout, url)

		req, cancel := context.WithTimeout(ctx, cryptoutilMagic.TimeoutHTTPHealthRequest)
		httpReq, err := http.NewRequestWithContext(req, http.MethodGet, url, nil)
		require.NoError(a.t, err, "Failed to create request to %s", url)

		resp, err := client.Do(httpReq)

		cancel()

		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			a.log("✅ HTTP service ready at %s", url)

			return
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(cryptoutilMagic.TestTimeoutHTTPRetryInterval)
	}
}

// AssertTelemetryFlow verifies that telemetry is flowing to Grafana and OTEL collector.
func (a *ServiceAssertions) AssertTelemetryFlow(ctx context.Context, grafanaURL, otelURL string) {
	a.log("📊 Verifying telemetry flow")

	// Check Grafana health
	client := &http.Client{Timeout: cryptoutilMagic.TestTimeoutHTTPClient}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafanaURL+"/api/health", nil)
	require.NoError(a.t, err, "Failed to create Grafana health request")

	resp, err := client.Do(req)
	require.NoError(a.t, err, "Failed to connect to Grafana")

	defer resp.Body.Close()
	require.Equal(a.t, http.StatusOK, resp.StatusCode, "Grafana health check failed")

	a.log("✅ Grafana health check passed")

	// Check OTEL collector metrics
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, otelURL+"/metrics", nil)
	require.NoError(a.t, err, "Failed to create OTEL collector metrics request")

	resp, err = client.Do(req)
	require.NoError(a.t, err, "Failed to connect to OTEL collector")

	defer resp.Body.Close()
	require.Equal(a.t, http.StatusOK, resp.StatusCode, "OTEL collector metrics check failed")

	// Verify metrics contain cryptoutil service data
	body := make([]byte, 1024*1024) // 1MB buffer
	n, err := resp.Body.Read(body)
	require.NoError(a.t, err, "Failed to read OTEL metrics")

	metrics := string(body[:n])
	require.Contains(a.t, metrics, "cryptoutil", "No cryptoutil metrics found in OTEL collector")

	// Check for traces/logs/metrics indicators
	hasTraces := strings.Contains(metrics, "traces") || strings.Contains(metrics, "spans")
	hasLogs := strings.Contains(metrics, "logs") || strings.Contains(metrics, "log_records")
	hasMetrics := strings.Contains(metrics, "metrics") || strings.Contains(metrics, "data_points")

	require.True(a.t, hasTraces || hasLogs || hasMetrics, "No telemetry data found in OTEL collector")

	a.log("✅ Telemetry flow verification passed")
}

// AssertDockerServicesHealthy verifies all Docker services are healthy.
func (a *ServiceAssertions) AssertDockerServicesHealthy() {
	a.log("🔍 Verifying Docker services health")

	services := []string{
		"cryptoutil_sqlite",
		"cryptoutil_postgres_1",
		"cryptoutil_postgres_2",
		"postgres",
		"grafana-otel-lgtm",
		cryptoutilMagic.DockerServiceOtelCollectorHealthcheck,
	}

	cmd := exec.Command("docker", "compose", "-f", "../../deployments/compose/compose.yml", "ps", "--format", "json")
	output, err := cmd.Output()
	require.NoError(a.t, err, "Failed to check Docker services health")

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	var serviceList []map[string]interface{}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var service map[string]interface{}
		err := json.Unmarshal([]byte(line), &service)
		require.NoError(a.t, err, "Failed to parse Docker service JSON")

		serviceList = append(serviceList, service)
	}

	require.NotEmpty(a.t, serviceList, "No services found in Docker compose output")

	serviceMap := make(map[string]map[string]interface{})

	for _, service := range serviceList {
		if name, ok := service["Name"].(string); ok {
			if strings.Contains(name, "compose-") {
				parts := strings.Split(name, "-")
				if len(parts) >= cryptoutilMagic.DockerServiceNamePartsMin {
					serviceName := strings.Join(parts[1:len(parts)-1], "-")
					serviceMap[serviceName] = service
				}
			}
		}
	}

	for _, serviceName := range services {
		service, exists := serviceMap[serviceName]
		require.True(a.t, exists, "Service %s not found in Docker compose output", serviceName)

		if health, ok := service["Health"].(string); ok {
			require.Equal(a.t, "healthy", health, "Service %s is not healthy", serviceName)
		} else if serviceName == cryptoutilMagic.DockerServiceOtelCollectorHealthcheck {
			if state, ok := service["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateExited {
				if exitCode, ok := service["ExitCode"].(float64); ok && exitCode == 0 {
					// Health check passed
				} else {
					a.t.Errorf("Service %s health check failed with exit code %v", serviceName, exitCode)
				}
			} else if state == cryptoutilMagic.DockerServiceStateRunning {
				a.t.Errorf("Service %s health check still running", serviceName)
			} else {
				a.t.Errorf("Service %s health check failed", serviceName)
			}
		} else {
			if state, ok := service["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateRunning {
				// Service is running
			} else {
				a.t.Errorf("Service %s is not running", serviceName)
			}
		}

		a.log("✅ Service %s is healthy", serviceName)
	}

	a.log("✅ All Docker services are healthy")
}

// log provides structured logging for assertions.
func (a *ServiceAssertions) log(format string, args ...interface{}) {
	message := fmt.Sprintf("[%s] [%v] %s\n",
		time.Now().Format("15:04:05"),
		time.Since(a.startTime).Round(time.Second),
		fmt.Sprintf(format, args...))

	// Write to console
	fmt.Print(message)

	// Write to log file if available
	if a.logFile != nil {
		if _, err := a.logFile.WriteString(message); err != nil {
			// If we can't write to the log file, at least write to console
			fmt.Printf("⚠️ Failed to write to log file: %v\n", err)
		}
	}
}
