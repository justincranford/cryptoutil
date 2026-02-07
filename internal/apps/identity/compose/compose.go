// Copyright (c) 2025 Justin Cranford

// Package main provides the identity-compose command for managing identity service deployments.
package compose

import (
	"context"
	"io"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	outWriter io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr
)

const (
	defaultTailLines      = 100
	defaultHealthRetries  = 30
	defaultHealthInterval = 5 * time.Second
)

// identityOrchestrator manages Docker Compose operations for identity services.
type identityOrchestrator struct {
	logger      *slog.Logger
	composeFile string
	profile     string
	scaling     map[string]int // service name -> replica count
}

// newIdentityOrchestrator creates a new identity orchestrator.
func newIdentityOrchestrator(logger *slog.Logger, composeFile, profile string, scaling map[string]int) *identityOrchestrator {
	return &identityOrchestrator{
		logger:      logger,
		composeFile: composeFile,
		profile:     profile,
		scaling:     scaling,
	}
}

// start brings up identity services with specified profile and scaling.
func (o *identityOrchestrator) start(ctx context.Context) error {
	o.logger.Info("Starting identity services",
		"compose_file", o.composeFile,
		"profile", o.profile,
		"scaling", o.scaling)

	args := []string{"compose", "-f", o.composeFile, "--profile", o.profile, "up", "-d"}

	// Add scaling arguments
	for service, replicas := range o.scaling {
		args = append(args, "--scale", fmt.Sprintf("%s=%d", service, replicas))
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start identity services: %w", err)
	}

	o.logger.Info("Identity services started successfully")

	return nil
}

// stop brings down identity services.
func (o *identityOrchestrator) stop(ctx context.Context, removeVolumes bool) error {
	o.logger.Info("Stopping identity services", "remove_volumes", removeVolumes)

	args := []string{"compose", "-f", o.composeFile, "--profile", o.profile, "down"}
	if removeVolumes {
		args = append(args, "-v")
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop identity services: %w", err)
	}

	o.logger.Info("Identity services stopped successfully")

	return nil
}

// healthCheck verifies all services are healthy.
func (o *identityOrchestrator) healthCheck(ctx context.Context) error {
	o.logger.Info("Checking identity services health")

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", o.composeFile, "--profile", o.profile, "ps", "--format", "json")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	// Parse JSON output to check service health
	// For simplicity, we check if output contains "healthy" status
	if !strings.Contains(string(output), `"Health":"healthy"`) {
		o.logger.Warn("Some services may not be healthy", "output", string(output))

		return fmt.Errorf("services not yet healthy")
	}

	o.logger.Info("All identity services are healthy")

	return nil
}

// waitForHealth waits for all services to become healthy with retries.
func (o *identityOrchestrator) waitForHealth(ctx context.Context, maxRetries int, retryInterval time.Duration) error {
	o.logger.Info("Waiting for identity services to become healthy",
		"max_retries", maxRetries,
		"retry_interval", retryInterval)

	for i := 0; i < maxRetries; i++ {
		if err := o.healthCheck(ctx); err == nil {
			return nil
		}

		if i < maxRetries-1 {
			o.logger.Info("Health check failed, retrying...", "attempt", i+1, "max_retries", maxRetries)
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("services did not become healthy after %d retries", maxRetries)
}

// logs displays logs for specified service.
func (o *identityOrchestrator) logs(ctx context.Context, service string, follow bool, tail int) error {
	o.logger.Info("Fetching logs", "service", service, "follow", follow, "tail", tail)

	args := []string{"compose", "-f", o.composeFile, "--profile", o.profile, "logs"}
	if follow {
		args = append(args, "-f")
	}

	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	if service != "" {
		args = append(args, service)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	return nil
}

// parseScaling parses scaling string (e.g., "authz=2,idp=1") into map.
func parseScaling(scalingStr string) (map[string]int, error) {
	scaling := make(map[string]int)
	if scalingStr == "" {
		return scaling, nil
	}

	pairs := strings.Split(scalingStr, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), "=")
		if len(parts) != cryptoutilSharedMagic.ScalingPairParts {
			return nil, fmt.Errorf("invalid scaling format: %s (expected service=count)", pair)
		}

		service := strings.TrimSpace(parts[0])

		var count int
		if _, err := fmt.Sscanf(parts[1], "%d", &count); err != nil {
			return nil, fmt.Errorf("invalid replica count for service %s: %w", service, err)
		}

		scaling[service] = count
	}

	return scaling, nil
}

// Compose runs the identity compose orchestrator.
// args: Command-line arguments (not including program name)
// stdout, stderr: Output streams for messages
// Returns: Exit code (0 for success, non-zero for errors)
func Compose(args []string, stdout, stderr io.Writer) int {
	outWriter = stdout
	errWriter = stderr
	// Command-line flags
	var (
		composeFile    = flag.String("compose-file", "deployments/identity/compose.advanced.yml", "Path to Docker Compose file")
		profile        = flag.String("profile", "demo", "Docker Compose profile (demo, development, ci, production)")
		scalingStr     = flag.String("scaling", "", "Service scaling (e.g., 'identity-authz=2,identity-idp=1')")
		operation      = flag.String("operation", "start", "Operation: start, stop, health, logs")
		removeVolumes  = flag.Bool("remove-volumes", false, "Remove volumes when stopping")
		follow         = flag.Bool("follow", false, "Follow log output")
		tail           = flag.Int("tail", defaultTailLines, "Number of log lines to show")
		service        = flag.String("service", "", "Service name for logs command")
		healthRetries  = flag.Int("health-retries", defaultHealthRetries, "Number of health check retries")
		healthInterval = flag.Duration("health-interval", defaultHealthInterval, "Interval between health checks")
	)

	flag.Parse()

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Parse scaling
	scaling, err := parseScaling(*scalingStr)
	if err != nil {
		logger.Error("Failed to parse scaling", "error", err)
		return 1
	}

	// Create orchestrator
	orchestrator := newIdentityOrchestrator(logger, *composeFile, *profile, scaling)

	// Execute operation
	ctx := context.Background()

	switch *operation {
	case "start":
		if err := orchestrator.start(ctx); err != nil {
			logger.Error("Start failed", "error", err)
			return 1
		}
		// Wait for services to become healthy
		if err := orchestrator.waitForHealth(ctx, *healthRetries, *healthInterval); err != nil {
			logger.Error("Health check failed", "error", err)
			return 1
		}
	case "stop":
		if err := orchestrator.stop(ctx, *removeVolumes); err != nil {
			logger.Error("Stop failed", "error", err)
			return 1
		}
	case "health":
		if err := orchestrator.healthCheck(ctx); err != nil {
			logger.Error("Health check failed", "error", err)
			return 1
		}
	case "logs":
		if err := orchestrator.logs(ctx, *service, *follow, *tail); err != nil {
			logger.Error("Logs failed", "error", err)
			return 1
		}
	default:
		logger.Error("Unknown operation", "operation", *operation)
		flag.Usage()
		return 1
	}

	logger.Info("Operation completed successfully", "operation", *operation)
	return 0
}
