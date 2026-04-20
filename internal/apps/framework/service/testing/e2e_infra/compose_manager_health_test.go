// Copyright (c) 2025 Justin Cranford
//
//

package e2e_infra

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// --- WaitForHealth tests ---

// TestWaitForHealth_Timeout verifies that WaitForHealth returns error when timeout fires
// before the ticker and no server is reachable. The 1ms timeout guarantees the timeout
// channel fires first (ticker is set to TestUnitPollIntervalMs via SetDefaultHealthPollInterval).
//
// Sequential: modifies package-level defaultHealthPollInterval state.
func TestWaitForHealth_Timeout(t *testing.T) {
	restore := SetDefaultHealthPollInterval(cryptoutilSharedMagic.TestUnitPollIntervalMs)
	defer restore()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  &http.Client{Timeout: time.Second},
	}

	// 1ms timeout fires before the poll tick.
	err := cm.WaitForHealth("http://127.0.0.1:19999/health", 1*time.Millisecond)
	require.Error(t, err)
	require.ErrorContains(t, err, "health check timeout after")
}

// TestWaitForHealth_ConnError verifies the connection-error retry path (ticker fires, but
// server is not reachable). Uses a port that has no listener.
//
// Sequential: modifies package-level defaultHealthPollInterval state.
func TestWaitForHealth_ConnError(t *testing.T) {
	restore := SetDefaultHealthPollInterval(cryptoutilSharedMagic.TestUnitPollIntervalMs)
	defer restore()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  &http.Client{Timeout: cryptoutilSharedMagic.TestUnitHTTPClientTimeoutMs},
	}

	// Timeout after TestUnitShortTimeoutMs, ticker at TestUnitPollIntervalMs → several retries.
	err := cm.WaitForHealth("http://127.0.0.1:19999/health", cryptoutilSharedMagic.TestUnitShortTimeoutMs)
	require.Error(t, err)
	require.ErrorContains(t, err, "health check timeout after")
	// Error message must contain a non-negative attempts count: "(N attempts)" where N >= 0.
	// This kills the attempts++ → attempts-- mutation: negative attempts would produce "(-N attempts)".
	require.NotContains(t, err.Error(), "(-", "attempts counter must not be negative")
}

// TestWaitForHealth_NonOKStatus verifies the non-200 retry path: server exists but
// returns 503. Eventually times out.
//
// Sequential: modifies package-level defaultHealthPollInterval state.
func TestWaitForHealth_NonOKStatus(t *testing.T) {
	restore := SetDefaultHealthPollInterval(cryptoutilSharedMagic.TestUnitPollIntervalMs)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  srv.Client(),
	}

	err := cm.WaitForHealth(srv.URL+"/health", cryptoutilSharedMagic.TestUnitMediumTimeoutMs)
	require.Error(t, err)
	require.ErrorContains(t, err, "health check timeout after")
}

// TestWaitForHealth_Success verifies WaitForHealth returns nil when server responds 200.
//
// Sequential: modifies package-level defaultHealthPollInterval state.
func TestWaitForHealth_Success(t *testing.T) {
	restore := SetDefaultHealthPollInterval(cryptoutilSharedMagic.TestUnitPollIntervalMs)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  srv.Client(),
	}

	err := cm.WaitForHealth(srv.URL+"/health", cryptoutilSharedMagic.TimeoutHTTPHealthRequest)
	require.NoError(t, err)
}

// TestWaitForHealth_InvalidURL verifies the request-creation-error retry path.
// "\x00" in the URL causes http.NewRequestWithContext to return an error on each tick,
// exercising the "request creation error" continue branch.
//
// Sequential: modifies package-level defaultHealthPollInterval state.
func TestWaitForHealth_InvalidURL(t *testing.T) {
	restore := SetDefaultHealthPollInterval(cryptoutilSharedMagic.TestUnitPollIntervalMs)
	defer restore()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  &http.Client{Timeout: time.Second},
	}

	// Control characters in host cause NewRequestWithContext to return an error.
	err := cm.WaitForHealth("http://invalid\x00host/health", cryptoutilSharedMagic.TestUnitShortTimeoutMs)
	require.Error(t, err)
}

// --- WaitForMultipleServices tests ---

// TestWaitForMultipleServices_Empty verifies that an empty services map returns nil immediately.
func TestWaitForMultipleServices_Empty(t *testing.T) {
	t.Parallel()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  &http.Client{Timeout: time.Second},
	}

	err := cm.WaitForMultipleServices(map[string]string{}, time.Second)
	require.NoError(t, err)
}

// TestWaitForMultipleServices_Timeout verifies that a non-reachable service causes error.
//
// Sequential: modifies package-level defaultHealthPollInterval state.
func TestWaitForMultipleServices_Timeout(t *testing.T) {
	restore := SetDefaultHealthPollInterval(cryptoutilSharedMagic.TestUnitPollIntervalMs)
	defer restore()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  &http.Client{Timeout: cryptoutilSharedMagic.TestUnitHTTPClientTimeoutMs},
	}

	services := map[string]string{
		"test-svc": "http://127.0.0.1:19999/health",
	}

	err := cm.WaitForMultipleServices(services, cryptoutilSharedMagic.TestUnitShortTimeoutMs)
	require.Error(t, err)
	require.ErrorContains(t, err, "test-svc")
}

// TestWaitForMultipleServices_Success verifies all healthy services return nil.
//
// Sequential: modifies package-level defaultHealthPollInterval state.
func TestWaitForMultipleServices_Success(t *testing.T) {
	restore := SetDefaultHealthPollInterval(cryptoutilSharedMagic.TestUnitPollIntervalMs)
	defer restore()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cm := &ComposeManager{
		ComposeFile: "nonexistent/compose.yml",
		HTTPClient:  srv.Client(),
	}

	services := map[string]string{
		"svc-a": srv.URL + "/health",
		"svc-b": srv.URL + "/health",
	}

	err := cm.WaitForMultipleServices(services, cryptoutilSharedMagic.TimeoutHTTPHealthRequest)
	require.NoError(t, err)
}

// --- WaitForServicesHealthy tests ---

// TestWaitForServicesHealthy_DockerFails verifies error when docker command fails
// (e.g. nonexistent compose file path).
func TestWaitForServicesHealthy_DockerFails(t *testing.T) {
	t.Parallel()

	cm := NewComposeManager("/nonexistent/compose.yml")
	services := []ServiceAndJob{{Service: "test-svc"}}
	err := cm.WaitForServicesHealthy(context.Background(), services)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get docker compose ps output")
}

// TestWaitForServicesHealthy_ParseFails verifies error when docker ps output is invalid JSON.
func TestWaitForServicesHealthy_ParseFails(t *testing.T) {
	t.Parallel()

	cm := NewComposeManager("/nonexistent/compose.yml")
	cm.psOutputFn = func(_ context.Context) ([]byte, error) {
		return []byte("not valid json"), nil
	}

	services := []ServiceAndJob{{Service: "test-svc"}}
	err := cm.WaitForServicesHealthy(context.Background(), services)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to parse docker compose ps output")
}

// TestWaitForServicesHealthy_UnhealthyService verifies error when a service is not healthy.
func TestWaitForServicesHealthy_UnhealthyService(t *testing.T) {
	t.Parallel()

	cm := NewComposeManager("/nonexistent/compose.yml")
	// Service is running (not exited), no health check → healthy. But the test-svc is not in the map.
	cm.psOutputFn = func(_ context.Context) ([]byte, error) {
		// Returns output with a different service only — test-svc is not present.
		return []byte(`{"Name":"compose-other-svc-1","State":"running","Health":""}`), nil
	}

	services := []ServiceAndJob{{Service: "test-svc"}}
	err := cm.WaitForServicesHealthy(context.Background(), services)
	require.Error(t, err)
	require.ErrorContains(t, err, "unhealthy services")
}

// TestWaitForServicesHealthy_AllHealthy verifies nil return when all services are healthy.
func TestWaitForServicesHealthy_AllHealthy(t *testing.T) {
	t.Parallel()

	cm := NewComposeManager("/nonexistent/compose.yml")
	cm.psOutputFn = func(_ context.Context) ([]byte, error) {
		// Omit the Health field so determineServiceHealthStatus falls through to state check.
		// A service with State="running" and no Health field is considered healthy.
		return []byte(`{"Name":"compose-test-svc-1","State":"running"}`), nil
	}

	services := []ServiceAndJob{{Service: "test-svc"}}
	err := cm.WaitForServicesHealthy(context.Background(), services)
	require.NoError(t, err)
}

// --- defaultTestmainFactoryDeps closure body tests ---

// TestDefaultTestmainFactoryDeps_ClosureBodies verifies each closure body executes.
// startFn and stopFn will fail because compose file doesn't exist, but the closure
// body IS executed (covering those statements).
// waitForServicesFn with empty services returns nil immediately.
// newSecureClientFn panics on bad CA paths — that panic is expected and caught.
func TestDefaultTestmainFactoryDeps_ClosureBodies(t *testing.T) {
	t.Parallel()

	deps := defaultTestmainFactoryDeps()
	ctx := context.Background()

	cm := deps.newComposeManagerFn("/nonexistent/compose.yml")
	require.NotNil(t, cm)

	// insecureClient closure body.
	insecureClient := deps.newInsecureClientFn()
	require.NotNil(t, insecureClient)

	// secureClient closure body: NewClientForTestWithCA panics on missing CA file.
	// Cover the closure body by catching the expected panic.
	require.Panics(t, func() {
		_ = deps.newSecureClientFn("/nonexistent/ca.pem")
	})

	// startFn body → calls cm.Start → docker command fails → error.
	err := deps.startFn(ctx, cm)
	require.Error(t, err, "startFn must fail with nonexistent compose file")

	// waitForServicesFn body → calls cm.WaitForMultipleServices with empty map → nil.
	err = deps.waitForServicesFn(cm, map[string]string{}, cryptoutilSharedMagic.TestTLSClientRetryWait)
	require.NoError(t, err, "waitForServicesFn with empty services must return nil")

	// stopFn body → calls cm.Stop → docker command fails → error.
	err = deps.stopFn(ctx, cm)
	require.Error(t, err, "stopFn must fail with nonexistent compose file")
}

// --- SetupE2ETestMain tests ---

// TestSetupE2ETestMain_SetupFails verifies that SetupE2ETestMain returns 1 when setup fails
// and that nil m is safe when the setup always fails (m.Run is never invoked).
func TestSetupE2ETestMain_SetupFails(t *testing.T) {
	t.Parallel()

	cfg := E2ETestConfig{
		ComposeFile:    "/nonexistent/compose.yml",
		HealthChecks:   map[string]string{},
		HealthTimeout:  cryptoutilSharedMagic.TestTLSClientRetryWait,
		ServiceLogName: "test",
	}

	// nil m is safe because the lazy closure `func() int { return m.Run() }` is only
	// invoked when setup succeeds — which it won't here.
	code := SetupE2ETestMain(nil, cfg, func(*E2ETestEnv) {})
	require.Equal(t, 1, code)
}
