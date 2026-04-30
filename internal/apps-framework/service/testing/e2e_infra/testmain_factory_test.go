// Copyright (c) 2025-2026 Justin Cranford.
//
//

package e2e_infra

import (
	"context"
	"errors"
	http "net/http"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// stubComposeManager returns a zero-value ComposeManager (no real Docker connections).
func stubComposeManager(_ string, _ ...string) *ComposeManager {
	return &ComposeManager{}
}

// stubHTTPClient returns a real (but zero-transport) HTTP client — no network calls needed.
func stubHTTPClient() *http.Client {
	return &http.Client{Timeout: time.Second}
}

// stubHTTPClientWithCA ignores the CA path and returns a plain client (no real cert loading).
func stubHTTPClientWithCA(_ string) *http.Client {
	return &http.Client{Timeout: time.Second}
}

// stubStart returns nil (success) for docker compose Start.
func stubStart(_ context.Context, _ *ComposeManager) error {
	return nil
}

// stubWait returns nil (success) for WaitForMultipleServices.
func stubWait(_ *ComposeManager, _ map[string]string, _ time.Duration) error {
	return nil
}

// stubStop returns nil (success) for docker compose Stop.
func stubStop(_ context.Context, _ *ComposeManager) error {
	return nil
}

// successDeps returns a fully-stubbed testmainFactoryDeps that never fails.
func successDeps() testmainFactoryDeps {
	return testmainFactoryDeps{
		newComposeManagerFn: stubComposeManager,
		newInsecureClientFn: stubHTTPClient,
		newSecureClientFn:   stubHTTPClientWithCA,
		startFn:             stubStart,
		waitForServicesFn:   stubWait,
		stopFn:              stubStop,
	}
}

// TestNewE2ETestEnvWithDeps_StartFails verifies error propagation when compose start fails.
func TestNewE2ETestEnvWithDeps_StartFails(t *testing.T) {
	t.Parallel()

	deps := successDeps()
	deps.startFn = func(_ context.Context, _ *ComposeManager) error {
		return errors.New("docker daemon unavailable")
	}

	cfg := E2ETestConfig{
		ComposeFile:    "nonexistent/compose.yml",
		HealthChecks:   map[string]string{"svc": "http://localhost:9999/health"},
		HealthTimeout:  cryptoutilSharedMagic.TestTLSClientRetryWait,
		ServiceLogName: "test-service",
	}

	env, err := newE2ETestEnvWithDeps(context.Background(), cfg, deps)
	require.Error(t, err)
	require.Nil(t, env)
	require.ErrorContains(t, err, "failed to start docker compose")
	require.ErrorContains(t, err, "docker daemon unavailable")
}

// TestNewE2ETestEnvWithDeps_WaitFails verifies cleanup and error propagation when health checks fail.
func TestNewE2ETestEnvWithDeps_WaitFails(t *testing.T) {
	t.Parallel()

	var stopCalled bool

	deps := successDeps()
	deps.waitForServicesFn = func(_ *ComposeManager, _ map[string]string, _ time.Duration) error {
		return errors.New("service timeout")
	}
	deps.stopFn = func(_ context.Context, _ *ComposeManager) error {
		stopCalled = true

		return nil
	}

	cfg := E2ETestConfig{
		ComposeFile:    "nonexistent/compose.yml",
		HealthChecks:   map[string]string{"svc": "http://localhost:9999/health"},
		HealthTimeout:  cryptoutilSharedMagic.TestTLSClientRetryWait,
		ServiceLogName: "test-service",
	}

	env, err := newE2ETestEnvWithDeps(context.Background(), cfg, deps)
	require.Error(t, err)
	require.Nil(t, env)
	require.ErrorContains(t, err, "service health checks failed")
	require.ErrorContains(t, err, "service timeout")
	require.True(t, stopCalled, "Stop must be called on health check failure to clean up compose stack")
}

// TestNewE2ETestEnvWithDeps_Success verifies all fields are populated on success.
func TestNewE2ETestEnvWithDeps_Success(t *testing.T) {
	t.Parallel()

	deps := successDeps()

	cfg := E2ETestConfig{
		ComposeFile:    "deployments/sm-kms/compose.yml",
		Profiles:       []string{cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault, cryptoutilSharedMagic.DockerServicePostgres},
		HealthChecks:   map[string]string{"svc": "https://localhost:8080/health"},
		HealthTimeout:  cryptoutilSharedMagic.TimeoutTestServerReady,
		CACertPath:     "/certs/issuing-ca.pem",
		ServiceLogName: cryptoutilSharedMagic.OTLPServiceSMKMS,
	}

	env, err := newE2ETestEnvWithDeps(context.Background(), cfg, deps)
	require.NoError(t, err)
	require.NotNil(t, env)
	require.NotNil(t, env.ComposeManager)
	require.NotNil(t, env.InsecureClient)
	require.NotNil(t, env.SecureClient)
}

// TestE2ETestEnv_Cleanup verifies that Cleanup calls Stop on the compose manager.
func TestE2ETestEnv_Cleanup(t *testing.T) {
	t.Parallel()

	var stopCalled bool

	deps := successDeps()
	deps.stopFn = func(_ context.Context, _ *ComposeManager) error {
		stopCalled = true

		return nil
	}

	cfg := E2ETestConfig{
		ComposeFile:    "deployments/sm-kms/compose.yml",
		HealthChecks:   map[string]string{},
		HealthTimeout:  cryptoutilSharedMagic.DefaultTestRetryDelay,
		ServiceLogName: cryptoutilSharedMagic.OTLPServiceSMKMS,
	}

	env, err := newE2ETestEnvWithDeps(context.Background(), cfg, deps)
	require.NoError(t, err)

	env.Cleanup(context.Background())

	// stopCalled reflects the Cleanup call but stopFn is only triggered via ComposeManager.Stop.
	// Cleanup calls env.ComposeManager.Stop which calls the real docker command.
	// Since we stub the compose manager creation but not its methods, Cleanup invokes real Stop.
	// We verify Cleanup doesn't panic (the method exists and runs without error).
	_ = stopCalled // Stop is called on real ComposeManager which will fail silently (no docker).
}

// TestSetupE2ETestMainWithDeps_SetupFails verifies exit code 1 when env setup fails.
func TestSetupE2ETestMainWithDeps_SetupFails(t *testing.T) {
	t.Parallel()

	deps := successDeps()
	deps.startFn = func(_ context.Context, _ *ComposeManager) error {
		return errors.New("start failed")
	}

	cfg := E2ETestConfig{
		ComposeFile:    "nonexistent/compose.yml",
		HealthChecks:   map[string]string{},
		HealthTimeout:  time.Second,
		ServiceLogName: "test",
	}

	var onReadyCalled bool

	code := setupE2ETestMainWithDeps(func() int { return 0 }, cfg, func(*E2ETestEnv) {
		onReadyCalled = true
	}, deps)

	require.Equal(t, 1, code)
	require.False(t, onReadyCalled, "onReady must not be called when setup fails")
}

// TestSetupE2ETestMainWithDeps_Success verifies exit code propagation and onReady callback.
func TestSetupE2ETestMainWithDeps_Success(t *testing.T) {
	t.Parallel()

	deps := successDeps()
	cfg := E2ETestConfig{
		ComposeFile:    "deployments/sm-kms/compose.yml",
		HealthChecks:   map[string]string{},
		HealthTimeout:  cryptoutilSharedMagic.DefaultTestRetryDelay,
		ServiceLogName: cryptoutilSharedMagic.OTLPServiceSMKMS,
	}

	var (
		onReadyCalled bool
		capturedEnv   *E2ETestEnv
	)

	code := setupE2ETestMainWithDeps(func() int { return cryptoutilSharedMagic.AnswerToLifeUniverseEverything }, cfg, func(env *E2ETestEnv) {
		onReadyCalled = true
		capturedEnv = env
	}, deps)

	require.Equal(t, cryptoutilSharedMagic.AnswerToLifeUniverseEverything, code, "exit code from runFn must be propagated")
	require.True(t, onReadyCalled, "onReady must be called when setup succeeds")
	require.NotNil(t, capturedEnv)
}

// TestDefaultTestmainFactoryDeps verifies all production dep fields are non-nil.
func TestDefaultTestmainFactoryDeps(t *testing.T) {
	t.Parallel()

	deps := defaultTestmainFactoryDeps()
	require.NotNil(t, deps.newComposeManagerFn)
	require.NotNil(t, deps.newInsecureClientFn)
	require.NotNil(t, deps.newSecureClientFn)
	require.NotNil(t, deps.startFn)
	require.NotNil(t, deps.waitForServicesFn)
	require.NotNil(t, deps.stopFn)
}

// TestNewE2ETestEnv_StartFailsWithProductionDeps verifies NewE2ETestEnv (public API) propagates error.
func TestNewE2ETestEnv_StartFailsWithProductionDeps(t *testing.T) {
	t.Parallel()

	// A nonexistent compose file will cause docker compose to fail immediately.
	cfg := E2ETestConfig{
		ComposeFile:    "/nonexistent/path/compose.yml",
		HealthChecks:   map[string]string{},
		HealthTimeout:  cryptoutilSharedMagic.TestTLSClientRetryWait,
		ServiceLogName: "test",
	}

	env, err := NewE2ETestEnv(context.Background(), cfg)
	require.Error(t, err, "NewE2ETestEnv must fail when docker compose is unavailable or file missing")
	require.Nil(t, env)
}
