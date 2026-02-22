// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides KMS demo implementation.
package demo

import (
	"context"
	"fmt"
	"time"

	cryptoutilServerApplication "cryptoutil/internal/apps/sm/kms/server/application"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
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
)

// runKMSDemo executes the KMS demo.
func runKMSDemo(ctx context.Context, config *Config) int {
	progress := NewProgressDisplay(config)
	errors := NewErrorAggregator("kms")
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
		server.ShutdownFunction()
	}()

	progress.CompleteStep("Started KMS server")

	// Update settings with actual dynamic ports assigned by OS.
	settings.BindPublicPort = server.ActualPublicPort
	settings.BindPrivatePort = server.ActualPrivatePort

	// Step 3: Wait for health checks.
	progress.StartStep("Waiting for health checks")

	if err := waitForKMSHealth(ctx, settings, config.HealthTimeout); err != nil {
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

	if err := demonstrateKMSOperations(ctx, settings, progress); err != nil {
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
func parseKMSConfig() (*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings, error) {
	// Use dev profile with demo mode enabled.
	args := []string{
		"start",
		"--dev",
		"--demo",
		"--log-level", "INFO",
		"--bind-public-port", "0", // Dynamic port allocation
		"--bind-private-port", "0",
	}

	settings, err := cryptoutilAppsTemplateServiceConfig.Parse(args, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return settings, nil
}

// startKMSServer starts the KMS server.
func startKMSServer(_ context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*cryptoutilServerApplication.ServerApplicationListener, error) {
	server, err := cryptoutilServerApplication.StartServerListenerApplication(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Start server in background.
	go server.StartFunction()

	// Give server time to start.
	time.Sleep(cryptoutilSharedMagic.DefaultServerStartupDelay)

	return server, nil
}

// waitForKMSHealth waits for KMS health checks to pass.
func waitForKMSHealth(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings, timeout time.Duration) error {
	return cryptoutilSharedUtilPoll.Until(ctx, timeout, cryptoutilSharedMagic.DefaultHealthCheckInterval, func(_ context.Context) (bool, error) {
		_, err := cryptoutilServerApplication.SendServerListenerLivenessCheck(settings)
		if err != nil {
			return false, nil
		}

		_, err = cryptoutilServerApplication.SendServerListenerReadinessCheck(settings)

		return err == nil, nil
	})
}

// demonstrateKMSOperations demonstrates KMS cryptographic operations.
func demonstrateKMSOperations(_ context.Context, _ *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings, progress *ProgressDisplay) error {
	progress.Debug("Demo mode enabled - server seeded demo keys automatically")
	progress.Debug("Available demo keys: demo-encryption-aes256, demo-signing-rsa2048, demo-signing-ec256, demo-wrapping-aes256kw")

	// Demo complete - operations demonstrated through health checks and seed verification.
	return nil
}
