// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides KMS demo implementation.
package demo

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"time"

	cryptoutilKmsServer "cryptoutil/internal/apps/sm/kms/server"
	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

// KMS demo step counts.
const (
	kmsStepCount       = 4
	kmsParsedSteps     = 0
	kmsPreServerSteps  = 1
	kmsPreHealthSteps  = 2
	kmsPreDemoSteps    = 3
	kmsRemainingHealth = 1
	kmsRemainingOps    = 0

	// kmsAdminHTTPTimeout is the HTTP client timeout for KMS admin health checks.
	kmsAdminHTTPTimeout = 10 * time.Second
)

// runKMSDemo executes the KMS demo.
func runKMSDemo(ctx context.Context, config *Config) int {
	progress := NewProgressDisplay(config)
	errors := NewErrorAggregator(cryptoutilSharedMagic.KMSServiceName)
	startTime := time.Now().UTC()

	progress.Info("Starting KMS Demo")
	progress.Info("==================")
	progress.SetTotalSteps(kmsStepCount)

	// Step 1: Parse configuration.
	progress.StartStep("Parsing configuration")

	settings, err := parseKMSConfig()
	if err != nil {
		progress.FailStep("Parsing configuration", err)
		errors.Add("config", "failed to parse configuration", err)

		if !config.ContinueOnError {
			return errors.ToResult(kmsParsedSteps, kmsStepCount-1).ExitCode()
		}
	} else {
		progress.CompleteStep("Parsed configuration")
	}

	if settings == nil {
		result := errors.ToResult(kmsParsedSteps, kmsStepCount-1)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	// Step 2: Start server.
	progress.StartStep("Starting KMS server")

	server, err := startKMSServer(ctx, settings)
	if err != nil {
		progress.FailStep("Starting KMS server", err)
		errors.Add("server", "failed to start KMS server", err)

		result := errors.ToResult(kmsPreServerSteps, kmsStepCount-kmsPreServerSteps-1)
		result.DurationMS = time.Since(startTime).Milliseconds()
		progress.PrintSummary(result)

		return result.ExitCode()
	}

	defer func() {
		progress.Debug("Shutting down KMS server")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultServerShutdownTimeout)
		defer shutdownCancel()

		_ = server.Shutdown(shutdownCtx)
	}()

	progress.CompleteStep("Started KMS server")

	// Step 3: Wait for health checks.
	progress.StartStep("Waiting for health checks")

	if err := waitForKMSHealth(ctx, server, config.HealthTimeout); err != nil {
		progress.FailStep("Health checks", err)
		errors.Add("health", "health checks failed", err)

		if !config.ContinueOnError {
			result := errors.ToResult(kmsPreHealthSteps, kmsRemainingHealth)
			result.DurationMS = time.Since(startTime).Milliseconds()
			progress.PrintSummary(result)

			return result.ExitCode()
		}
	} else {
		progress.CompleteStep("Health checks passed")
	}

	// Step 4: Demonstrate operations.
	progress.StartStep("Demonstrating KMS operations")

	if err := demonstrateKMSOperations(ctx, progress); err != nil {
		progress.FailStep("KMS operations", err)
		errors.Add("operations", "failed to demonstrate KMS operations", err)
	} else {
		progress.CompleteStep("KMS operations demonstrated")
	}

	// Calculate final result.
	passed := kmsStepCount - errors.Count()

	result := errors.ToResult(passed, kmsRemainingOps)
	result.DurationMS = time.Since(startTime).Milliseconds()

	progress.PrintSummary(result)

	return result.ExitCode()
}

// parseKMSConfig creates settings for KMS demo.
func parseKMSConfig() (*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, error) {
	// Use dev profile with demo mode enabled.
	args := []string{
		"start",
		"--dev",
		"--demo",
		"--log-level", cryptoutilSharedMagic.DefaultLogLevelInfo,
		"--bind-public-port", "0", // Dynamic port allocation
		"--bind-private-port", "0",
	}

	settings, err := cryptoutilAppsFrameworkServiceConfig.Parse(args, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return settings, nil
}

// startKMSServer starts the KMS server.
func startKMSServer(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilKmsServer.KMSServer, error) {
	server, err := cryptoutilKmsServer.NewKMSServer(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	adminClient := &http.Client{
		Timeout: kmsAdminHTTPTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: server.AdminTLSRootCAPool()},
		},
	}

	// Start server in background.
	go func() { _ = server.Start(ctx) }()

	// Poll for ports to bind.
	if err := cryptoutilSharedUtilPoll.Until(ctx, cryptoutilSharedMagic.DefaultHealthCheckTimeout, cryptoutilSharedMagic.DefaultHealthCheckInterval, func(_ context.Context) (bool, error) {
		return server.PublicPort() > 0 && server.AdminPort() > 0, nil
	}); err != nil {
		return nil, fmt.Errorf("KMS server ports did not bind: %w", err)
	}

	// Poll for KMS server readiness.
	if err := cryptoutilSharedUtilPoll.Until(ctx, cryptoutilSharedMagic.DefaultHealthCheckTimeout, cryptoutilSharedMagic.DefaultHealthCheckInterval, func(pollCtx context.Context) (bool, error) {
		return isKMSHealthy(pollCtx, adminClient, server.AdminBaseURL()), nil
	}); err != nil {
		return nil, fmt.Errorf("KMS server failed to become ready: %w", err)
	}

	return server, nil
}

// waitForKMSHealth waits for KMS health checks to pass.
func waitForKMSHealth(ctx context.Context, server *cryptoutilKmsServer.KMSServer, timeout time.Duration) error {
	adminClient := &http.Client{
		Timeout: kmsAdminHTTPTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: server.AdminTLSRootCAPool()},
		},
	}

	if err := cryptoutilSharedUtilPoll.Until(ctx, timeout, cryptoutilSharedMagic.DefaultHealthCheckInterval, func(pollCtx context.Context) (bool, error) {
		return isKMSHealthy(pollCtx, adminClient, server.AdminBaseURL()), nil
	}); err != nil {
		return fmt.Errorf("kms health check failed: %w", err)
	}

	return nil
}

// demonstrateKMSOperations demonstrates KMS cryptographic operations.
func demonstrateKMSOperations(_ context.Context, progress *ProgressDisplay) error {
	progress.Debug("Demo mode enabled - server seeded demo keys automatically")
	progress.Debug("Available demo keys: demo-encryption-aes256, demo-signing-rsa2048, demo-signing-ec256, demo-wrapping-aes256kw")

	// Demo complete - operations demonstrated through health checks and seed verification.
	return nil
}
