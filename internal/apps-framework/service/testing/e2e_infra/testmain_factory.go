// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package e2e_infra provides reusable helpers for E2E testing with docker compose.
package e2e_infra

import (
	"context"
	"fmt"
	http "net/http"
	"testing"
	"time"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// E2ETestConfig holds the PS-ID-specific configuration for E2E TestMain setup.
// Every field is PS-ID-specific; shared behavior lives in the factory functions.
type E2ETestConfig struct {
	// ComposeFile is the path to the PS-ID's docker-compose file.
	ComposeFile string

	// Profiles are Docker Compose profiles to activate (e.g., DefaultOTLPEnvironmentDefault, DockerServicePostgres).
	Profiles []string

	// HealthChecks maps compose service names to their health endpoint URLs.
	// Used by WaitForMultipleServices to wait for all services to be ready.
	HealthChecks map[string]string

	// HealthTimeout is the maximum time to wait for all services to pass health checks.
	HealthTimeout time.Duration

	// CACertPath is the path to the issuing CA cert written by pki-init.
	// Used to build the CA-validated HTTP client AFTER services are healthy.
	CACertPath string

	// ServiceLogName is the PS-ID display name used in log messages (e.g., "sm-kms").
	ServiceLogName string
}

// E2ETestEnv holds the shared resources initialized by SetupE2ETestMain.
// Assign the fields to package-level vars in the onReady callback.
type E2ETestEnv struct {
	// ComposeManager manages the docker compose lifecycle.
	ComposeManager *ComposeManager

	// InsecureClient uses InsecureSkipVerify=true and is used for compose readiness health checks.
	InsecureClient *http.Client

	// SecureClient validates TLS against the CA cert written by pki-init.
	// This client is created AFTER services pass health checks (pki-init has run by then).
	SecureClient *http.Client
}

// Cleanup stops the Docker Compose stack. Call this after m.Run() completes.
func (env *E2ETestEnv) Cleanup(ctx context.Context) {
	_ = env.ComposeManager.Stop(ctx)
}

// testmainFactoryDeps holds injectable factory dependencies for unit testing.
type testmainFactoryDeps struct {
	newComposeManagerFn func(composeFile string, profiles ...string) *ComposeManager
	newInsecureClientFn func() *http.Client
	newSecureClientFn   func(caCertPath string) *http.Client
	syncCertsFn         func(ctx context.Context, cm *ComposeManager, hostCACertPath string) error
	startFn             func(ctx context.Context, cm *ComposeManager) error
	waitForServicesFn   func(cm *ComposeManager, services map[string]string, timeout time.Duration) error
	stopFn              func(ctx context.Context, cm *ComposeManager) error
}

// defaultTestmainFactoryDeps returns production (non-stub) factory dependencies.
func defaultTestmainFactoryDeps() testmainFactoryDeps {
	return testmainFactoryDeps{
		newComposeManagerFn: NewComposeManager,
		newInsecureClientFn: cryptoutilSharedCryptoTls.NewClientForTest,
		newSecureClientFn:   cryptoutilSharedCryptoTls.NewClientForTestWithCA,
		syncCertsFn: func(ctx context.Context, cm *ComposeManager, hostCACertPath string) error {
			return cm.SyncCertOutputDirFromPkiInit(ctx, hostCACertPath)
		},
		startFn: func(ctx context.Context, cm *ComposeManager) error {
			return cm.Start(ctx)
		},
		waitForServicesFn: func(cm *ComposeManager, services map[string]string, timeout time.Duration) error {
			return cm.WaitForMultipleServices(services, timeout)
		},
		stopFn: func(ctx context.Context, cm *ComposeManager) error {
			return cm.Stop(ctx)
		},
	}
}

// newE2ETestEnvWithDeps sets up the E2ETestEnv using injected dependencies.
// It starts the compose stack, waits for health, and builds both HTTP clients.
// On failure, partial setup is cleaned up before returning the error.
func newE2ETestEnvWithDeps(ctx context.Context, cfg E2ETestConfig, deps testmainFactoryDeps) (*E2ETestEnv, error) {
	env := &E2ETestEnv{
		ComposeManager: deps.newComposeManagerFn(cfg.ComposeFile, cfg.Profiles...),
		InsecureClient: deps.newInsecureClientFn(),
	}

	if err := deps.startFn(ctx, env.ComposeManager); err != nil {
		_ = deps.stopFn(ctx, env.ComposeManager)

		return nil, fmt.Errorf("failed to start docker compose: %w", err)
	}

	fmt.Printf("Waiting for all %s instances to be healthy...\n", cfg.ServiceLogName)

	if err := deps.waitForServicesFn(env.ComposeManager, cfg.HealthChecks, cfg.HealthTimeout); err != nil {
		_ = deps.stopFn(ctx, env.ComposeManager)

		return nil, fmt.Errorf("service health checks failed: %w", err)
	}

	if err := deps.syncCertsFn(ctx, env.ComposeManager, cfg.CACertPath); err != nil {
		_ = deps.stopFn(ctx, env.ComposeManager)

		return nil, fmt.Errorf("failed to sync cert tree: %w", err)
	}

	env.SecureClient = deps.newSecureClientFn(cfg.CACertPath)

	fmt.Printf("All %s services healthy.\n", cfg.ServiceLogName)

	return env, nil
}

// NewE2ETestEnv sets up the full E2E test environment using production dependencies.
// Starts docker compose, waits for health, and builds both HTTP clients.
// Call env.Cleanup(ctx) after tests run to stop the compose stack.
func NewE2ETestEnv(ctx context.Context, cfg E2ETestConfig) (*E2ETestEnv, error) {
	return newE2ETestEnvWithDeps(ctx, cfg, defaultTestmainFactoryDeps())
}

// setupE2ETestMainWithDeps runs the full TestMain lifecycle with injected dependencies.
// Calls onReady(env) after env is fully initialized (before tests run).
// Calls runFn() to execute the test suite.
// Returns the test exit code (0 = all pass, non-zero = failures).
func setupE2ETestMainWithDeps(runFn func() int, cfg E2ETestConfig, onReady func(*E2ETestEnv), deps testmainFactoryDeps) int {
	ctx := context.Background()

	env, err := newE2ETestEnvWithDeps(ctx, cfg, deps)
	if err != nil {
		fmt.Printf("E2E setup failed: %v\n", err)

		return 1
	}

	onReady(env)

	exitCode := runFn()

	env.Cleanup(ctx)

	return exitCode
}

// SetupE2ETestMain is the top-level TestMain helper. It:
//  1. Sets up E2ETestEnv (start compose, wait for health, build HTTP clients)
//  2. Calls onReady(env) so the caller can populate package-level test vars
//  3. Runs m.Run()
//  4. Cleans up the compose stack
//  5. Returns the test exit code (pass to os.Exit in the caller)
//
// Usage in TestMain:
//
//	func TestMain(m *testing.M) {
//	    os.Exit(e2e_infra.SetupE2ETestMain(m, e2e_infra.E2ETestConfig{...}, func(env *e2e_infra.E2ETestEnv) {
//	        sharedHTTPClient    = env.InsecureClient
//	        sharedHTTPClientWithCA = env.SecureClient
//	        composeManager      = env.ComposeManager
//	    }))
//	}
func SetupE2ETestMain(m *testing.M, cfg E2ETestConfig, onReady func(*E2ETestEnv)) int {
	// Use a lazy closure so m.Run is only called when setup succeeds.
	// This prevents a nil-pointer dereference when m is nil in tests where setup always fails.
	return setupE2ETestMainWithDeps(func() int { return m.Run() }, cfg, onReady, defaultTestmainFactoryDeps())
}
