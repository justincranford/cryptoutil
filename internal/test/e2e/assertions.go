// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"context"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilClient "cryptoutil/internal/apps/sm/kms/client"

	"github.com/stretchr/testify/require"
)

// ServiceAssertions provides common assertions for service testing.
type ServiceAssertions struct {
	t      *testing.T
	logger *Logger
}

// NewServiceAssertions creates a new service assertions helper.
func NewServiceAssertions(t *testing.T, logger *Logger) *ServiceAssertions {
	t.Helper()

	return &ServiceAssertions{
		t:      t,
		logger: logger,
	}
}

// AssertCryptoutilHealth checks that a cryptoutil instance is healthy.
func (a *ServiceAssertions) AssertCryptoutilHealth(baseURL string, rootCAsPool *x509.CertPool) {
	Log(a.logger, "💚 Testing health check for %s", baseURL)
	err := cryptoutilClient.CheckHealthz(&baseURL, rootCAsPool)
	require.NoError(a.t, err, "Health check failed for %s", baseURL)
	Log(a.logger, "✅ Health check passed for %s", baseURL)
}

// AssertCryptoutilReady waits for a cryptoutil instance to be ready.
func (a *ServiceAssertions) AssertCryptoutilReady(ctx context.Context, baseURL string, rootCAsPool *x509.CertPool) {
	Log(a.logger, "⏳ Waiting for cryptoutil ready at %s", baseURL)

	giveUpTime := time.Now().UTC().Add(cryptoutilMagic.TestTimeoutCryptoutilReady)
	checkCount := 0

	for {
		require.False(a.t, time.Now().UTC().After(giveUpTime), "Cryptoutil service not ready after %v: %s", cryptoutilMagic.TestTimeoutCryptoutilReady, baseURL)

		checkCount++
		Log(a.logger, "🔍 Cryptoutil readiness check #%d for %s", checkCount, baseURL)

		client := cryptoutilClient.RequireClientWithResponses(a.t, &baseURL, rootCAsPool)

		_, err := client.GetElastickeysWithResponse(ctx, nil)
		if err == nil {
			Log(a.logger, "✅ Cryptoutil service ready at %s after %d checks", baseURL, checkCount)

			return
		}

		Log(a.logger, "⏳ Cryptoutil at %s not ready yet (attempt %d), waiting %v...",
			baseURL, checkCount, cryptoutilMagic.TestTimeoutServiceRetry)
		time.Sleep(cryptoutilMagic.TestTimeoutServiceRetry)
	}
}

// AssertHTTPReady waits for an HTTP endpoint to return 200.
func (a *ServiceAssertions) AssertHTTPReady(ctx context.Context, url string, timeout time.Duration) {
	Log(a.logger, "⏳ Waiting for HTTP endpoint ready: %s", url)

	giveUpTime := time.Now().UTC().Add(timeout)
	client := &http.Client{Timeout: cryptoutilMagic.TestTimeoutHTTPClient}

	for {
		require.False(a.t, time.Now().UTC().After(giveUpTime), "Service not ready after %v: %s", timeout, url)

		req, cancel := context.WithTimeout(ctx, cryptoutilMagic.TimeoutHTTPHealthRequest)
		httpReq, err := http.NewRequestWithContext(req, http.MethodGet, url, nil)
		require.NoError(a.t, err, "Failed to create request to %s", url)

		resp, err := client.Do(httpReq)

		cancel()

		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			Log(a.logger, "✅ HTTP service ready at %s", url)

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
	Log(a.logger, "📊 Verifying telemetry flow")

	// Check Grafana health
	client := &http.Client{Timeout: cryptoutilMagic.TestTimeoutHTTPClient}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafanaURL+"/api/health", nil)
	require.NoError(a.t, err, "Failed to create Grafana health request")

	grafanaResp, err := client.Do(req)
	require.NoError(a.t, err, "Failed to connect to Grafana")

	defer grafanaResp.Body.Close()

	require.Equal(a.t, http.StatusOK, grafanaResp.StatusCode, "Grafana health check failed")

	Log(a.logger, "✅ Grafana health check passed")

	// Check OTEL collector health
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, otelURL, nil)
	require.NoError(a.t, err, "Failed to create OTEL collector health request")

	otelResp, err := client.Do(req)
	require.NoError(a.t, err, "Failed to connect to OTEL collector health")

	defer otelResp.Body.Close()

	require.Equal(a.t, http.StatusOK, otelResp.StatusCode, "OTEL collector health check failed")

	Log(a.logger, "✅ OTEL collector health check passed")

	Log(a.logger, "✅ Telemetry infrastructure verification passed")
}

// AssertDockerServicesHealthy verifies all Docker services are healthy.
func (a *ServiceAssertions) AssertDockerServicesHealthy() {
	Log(a.logger, "🔍 Verifying Docker services health")

	output, err := runDockerComposeCommand(context.Background(), a.logger, dockerComposeDescBatchHealth, dockerComposeArgsPsServices)
	require.NoError(a.t, err, "Failed to check Docker services health")

	serviceMap, err := parseDockerComposePsOutput(output)
	require.NoError(a.t, err, "Failed to parse Docker compose output")

	healthStatus := determineServiceHealthStatus(serviceMap, dockerComposeServicesForHealthCheck)

	// Assert all services are healthy
	unhealthyServices := make([]string, 0)

	for serviceName, healthy := range healthStatus {
		if !healthy {
			unhealthyServices = append(unhealthyServices, serviceName)
		}
	}

	require.Empty(a.t, unhealthyServices, "The following services are not healthy: %v", unhealthyServices)

	Log(a.logger, "✅ All Docker services are healthy")
}
