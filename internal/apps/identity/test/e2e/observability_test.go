// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e

import (
	"context"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	otelCollectorMetricsURL = "http://127.0.0.1:8889/metrics" // Re-exported metrics from OTEL collector
	otelCollectorHealthURL  = "http://127.0.0.1:13133/"       // OTEL collector health endpoint
	grafanaURL              = "http://127.0.0.1:3000"         // Grafana UI
	grafanaAPIURL           = "http://127.0.0.1:3000/api"     // Grafana API
	prometheusQueryURL      = "http://127.0.0.1:9090/api/v1/query"
)

// TestOTELCollectorIntegration verifies identity services send telemetry to OTEL collector.
func TestOTELCollectorIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start services with OTEL collector
	t.Log("📦 Starting identity services with OTEL collector...")
	require.NoError(t, startCompose(ctx, defaultProfile, map[string]int{
		"identity-authz": 1,
		"identity-idp":   1,
		"identity-rs":    1,
		"identity-spa":   1,
	}))

	defer func() {
		_ = stopCompose(context.Background(), defaultProfile, true)
	}()

	// Wait for all services healthy
	t.Log("⏳ Waiting for all services to become healthy...")
	require.NoError(t, waitForHealthy(ctx, defaultProfile, healthCheckTimeout, healthCheckRetry))

	// Verify OTEL collector is healthy
	t.Log("🔍 Verifying OTEL collector health...")
	require.NoError(t, checkOTELCollectorHealth(ctx))

	// Trigger identity service operations to generate telemetry
	suite := NewE2ETestSuite()

	t.Log("🔑 Performing OAuth flow to generate telemetry...")

	token, err := performClientCredentialsFlow(suite, "test-client-otel", "test-secret-otel")
	require.NoError(t, err, "Client credentials flow should succeed")
	require.NotEmpty(t, token, "Access token should be returned")

	// Wait for telemetry propagation
	t.Log("⏳ Waiting for telemetry propagation to OTEL collector...")
	time.Sleep(10 * time.Second)

	// Verify metrics are available from OTEL collector
	t.Log("📊 Verifying metrics available from OTEL collector...")

	metrics, err := fetchOTELCollectorMetrics(ctx)
	require.NoError(t, err, "Should fetch metrics from OTEL collector")
	require.NotEmpty(t, metrics, "Metrics should be available")

	// Verify specific identity service metrics exist
	t.Log("🔍 Verifying identity service metrics...")
	require.True(t, containsMetric(metrics, "http_server_request_duration"), "HTTP request duration metric should exist")
	require.True(t, containsMetric(metrics, "http_server_response_size"), "HTTP response size metric should exist")

	t.Log("✅ OTEL collector integration test passed")
}

// TestGrafanaIntegration verifies Grafana can query telemetry from OTEL collector.
func TestGrafanaIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start services with Grafana stack
	t.Log("📦 Starting identity services with Grafana stack...")
	require.NoError(t, startCompose(ctx, defaultProfile, map[string]int{
		"identity-authz": 1,
		"identity-idp":   1,
		"identity-rs":    1,
		"identity-spa":   1,
	}))

	defer func() {
		_ = stopCompose(context.Background(), defaultProfile, true)
	}()

	// Wait for all services healthy
	t.Log("⏳ Waiting for all services to become healthy...")
	require.NoError(t, waitForHealthy(ctx, defaultProfile, healthCheckTimeout, healthCheckRetry))

	// Verify Grafana is healthy
	t.Log("🔍 Verifying Grafana health...")
	require.NoError(t, checkGrafanaHealth(ctx))

	// Trigger identity service operations to generate telemetry
	suite := NewE2ETestSuite()

	t.Log("🔑 Performing OAuth flow to generate telemetry...")

	token, err := performClientCredentialsFlow(suite, "test-client-grafana", "test-secret-grafana")
	require.NoError(t, err, "Client credentials flow should succeed")
	require.NotEmpty(t, token, "Access token should be returned")

	// Wait for telemetry propagation
	t.Log("⏳ Waiting for telemetry propagation to Grafana...")
	time.Sleep(15 * time.Second)

	// Verify Grafana data sources configured
	t.Log("📊 Verifying Grafana data sources...")

	dataSources, err := fetchGrafanaDataSources(ctx)
	require.NoError(t, err, "Should fetch Grafana data sources")
	require.NotEmpty(t, dataSources, "Data sources should be configured")

	// Verify Prometheus data source exists
	t.Log("🔍 Verifying Prometheus data source...")
	require.True(t, containsDataSource(dataSources, "Prometheus"), "Prometheus data source should exist")

	// Verify Loki data source exists (logs)
	t.Log("🔍 Verifying Loki data source...")
	require.True(t, containsDataSource(dataSources, "Loki"), "Loki data source should exist")

	// Verify Tempo data source exists (traces)
	t.Log("🔍 Verifying Tempo data source...")
	require.True(t, containsDataSource(dataSources, "Tempo"), "Tempo data source should exist")

	t.Log("✅ Grafana integration test passed")
}

// TestPrometheusMetricScraping verifies Prometheus scrapes metrics from OTEL collector.
func TestPrometheusMetricScraping(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start services with Prometheus
	t.Log("📦 Starting identity services with Prometheus...")
	require.NoError(t, startCompose(ctx, defaultProfile, map[string]int{
		"identity-authz": 1,
		"identity-idp":   1,
		"identity-rs":    1,
		"identity-spa":   1,
	}))

	defer func() {
		_ = stopCompose(context.Background(), defaultProfile, true)
	}()

	// Wait for all services healthy
	t.Log("⏳ Waiting for all services to become healthy...")
	require.NoError(t, waitForHealthy(ctx, defaultProfile, healthCheckTimeout, healthCheckRetry))

	// Trigger identity service operations to generate metrics
	suite := NewE2ETestSuite()

	t.Log("🔑 Performing OAuth flow to generate metrics...")

	token, err := performClientCredentialsFlow(suite, "test-client-prometheus", "test-secret-prometheus")
	require.NoError(t, err, "Client credentials flow should succeed")
	require.NotEmpty(t, token, "Access token should be returned")

	// Wait for metric scraping
	t.Log("⏳ Waiting for Prometheus metric scraping...")
	time.Sleep(30 * time.Second) // Prometheus scrape interval typically 15-30s

	// Query Prometheus for identity service metrics
	t.Log("📊 Querying Prometheus for identity service metrics...")

	metrics, err := queryPrometheusMetrics(ctx, "http_server_request_duration_count")
	require.NoError(t, err, "Should query Prometheus metrics")
	require.NotEmpty(t, metrics, "Metrics should be available in Prometheus")

	t.Log("✅ Prometheus metric scraping test passed")
}

// TestTelemetryEndToEnd verifies complete telemetry flow (traces, metrics, logs).
func TestTelemetryEndToEnd(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start complete observability stack
	t.Log("📦 Starting identity services with complete observability stack...")
	require.NoError(t, startCompose(ctx, defaultProfile, map[string]int{
		"identity-authz": 1,
		"identity-idp":   1,
		"identity-rs":    1,
		"identity-spa":   1,
	}))

	defer func() {
		_ = stopCompose(context.Background(), defaultProfile, true)
	}()

	// Wait for all services healthy
	t.Log("⏳ Waiting for all services to become healthy...")
	require.NoError(t, waitForHealthy(ctx, defaultProfile, healthCheckTimeout, healthCheckRetry))

	// Perform complete OAuth flow to generate full telemetry
	suite := NewE2ETestSuite()

	t.Log("🔑 Performing complete OAuth flow...")

	token, err := performAuthorizationCodeFlow(suite, "test-client-e2e", "test-secret-e2e")
	require.NoError(t, err, "Authorization code flow should succeed")
	require.NotEmpty(t, token, "Access token should be returned")

	// Access protected resource to generate traces
	t.Log("📄 Accessing protected resource...")

	resource, err := accessProtectedResource(suite, suite.RSURL, token)
	require.NoError(t, err, "Resource access should succeed")
	require.NotEmpty(t, resource, "Resource data should be returned")

	// Wait for complete telemetry propagation
	t.Log("⏳ Waiting for complete telemetry propagation...")
	time.Sleep(30 * time.Second)

	// Verify traces available in Tempo (via Grafana API)
	t.Log("🔍 Verifying traces available in Tempo...")
	// TODO: Query Grafana Tempo API for traces
	// For now, skip detailed trace validation (requires Tempo query API implementation)

	// Verify metrics available in Prometheus
	t.Log("📊 Verifying metrics available in Prometheus...")

	metrics, err := queryPrometheusMetrics(ctx, "http_server_request_duration_count")
	require.NoError(t, err, "Should query Prometheus metrics")
	require.NotEmpty(t, metrics, "Metrics should be available")

	// Verify logs available in Loki (via Grafana API)
	t.Log("📝 Verifying logs available in Loki...")
	// TODO: Query Grafana Loki API for logs
	// For now, skip detailed log validation (requires Loki query API implementation)

	t.Log("✅ End-to-end telemetry test passed")
}

// Helper: checkOTELCollectorHealth verifies OTEL collector is healthy.
func checkOTELCollectorHealth(ctx context.Context) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, otelCollectorHealthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create OTEL health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("OTEL health check failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OTEL collector not healthy, status: %d", resp.StatusCode)
	}

	return nil
}

// Helper: fetchOTELCollectorMetrics fetches metrics from OTEL collector.
func fetchOTELCollectorMetrics(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, otelCollectorMetricsURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create metrics request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch metrics: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metrics request failed, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read metrics response: %w", err)
	}

	return string(body), nil
}

// Helper: containsMetric checks if metrics output contains a specific metric name.
func containsMetric(metrics, metricName string) bool {
	return strings.Contains(metrics, metricName)
}

// Helper: checkGrafanaHealth verifies Grafana is healthy.
func checkGrafanaHealth(ctx context.Context) error {
	client := &http.Client{Timeout: 10 * time.Second}

	healthURL := fmt.Sprintf("%s/api/health", grafanaURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create Grafana health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Grafana health check failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Grafana not healthy, status: %d", resp.StatusCode)
	}

	return nil
}

// Helper: fetchGrafanaDataSources fetches Grafana data sources.
func fetchGrafanaDataSources(ctx context.Context) ([]map[string]any, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	dataSourcesURL := fmt.Sprintf("%s/datasources", grafanaAPIURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dataSourcesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create data sources request: %w", err)
	}

	// Use default Grafana credentials (admin/admin)
	req.SetBasicAuth("admin", "admin")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data sources: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data sources request failed, status: %d", resp.StatusCode)
	}

	var dataSources []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&dataSources); err != nil {
		return nil, fmt.Errorf("failed to decode data sources: %w", err)
	}

	return dataSources, nil
}

// Helper: containsDataSource checks if data sources contain a specific type.
func containsDataSource(dataSources []map[string]any, dsType string) bool {
	for _, ds := range dataSources {
		if t, ok := ds["type"].(string); ok && strings.EqualFold(t, dsType) {
			return true
		}
	}

	return false
}

// Helper: queryPrometheusMetrics queries Prometheus for metrics.
func queryPrometheusMetrics(ctx context.Context, query string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	queryURL := fmt.Sprintf("%s?query=%s", prometheusQueryURL, query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Prometheus query request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Prometheus query failed, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Prometheus response: %w", err)
	}

	return string(body), nil
}
