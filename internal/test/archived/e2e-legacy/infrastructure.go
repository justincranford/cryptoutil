// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// InfrastructureManager handles Docker Compose operations and service management.
type InfrastructureManager struct {
	startTime time.Time
	logger    *Logger
}

// NewInfrastructureManager creates a new infrastructure manager.
func NewInfrastructureManager(startTime time.Time, logFile *os.File) *InfrastructureManager {
	return &InfrastructureManager{
		startTime: startTime,
		logger:    NewLogger(startTime, logFile),
	}
}

// StopServices stops Docker Compose services.
func (im *InfrastructureManager) StopServices(ctx context.Context) error {
	output, err := runDockerComposeCommand(ctx, im.logger, dockerComposeDescStopServices, dockerComposeArgsStopServices)
	if err != nil {
		return fmt.Errorf("failed to stop services: %w, output: %s", err, string(output))
	}

	return nil
}

// StartServices starts Docker Compose services.
func (im *InfrastructureManager) StartServices(ctx context.Context) error {
	output, err := runDockerComposeCommand(ctx, im.logger, dockerComposeDescStartServices, dockerComposeArgsStartServices)
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w, output: %s", err, string(output))
	}

	return nil
}

// WaitForDockerServicesHealthy waits for Docker services to report healthy status.
func (im *InfrastructureManager) WaitForDockerServicesHealthy(ctx context.Context) error {
	giveUpTime := time.Now().UTC().Add(cryptoutilSharedMagic.TestTimeoutDockerHealth)
	checkCount := 0

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for Docker services")
		default:
		}

		if time.Now().UTC().After(giveUpTime) {
			// Before giving up, show container logs for failed services
			Log(im.logger, "❌ Docker services not healthy after %v, showing recent container logs...", cryptoutilSharedMagic.TestTimeoutDockerHealth)

			if err := im.showFailedContainerLogs(ctx); err != nil {
				Log(im.logger, "⚠️ Failed to show container logs: %v", err)
			}

			return fmt.Errorf("docker services not healthy after %v", cryptoutilSharedMagic.TestTimeoutDockerHealth)
		}

		checkCount++
		Log(im.logger, "🔍 Health check #%d: Checking %d services...", checkCount, len(dockerComposeServicesForHealthCheck))

		healthStatus := im.areDockerServicesHealthy(ctx, dockerComposeServicesForHealthCheck)

		unhealthyServices := logServiceHealthStatus(im.startTime, healthStatus)

		allHealthy := len(unhealthyServices) == 0
		if allHealthy {
			Log(im.logger, "✅ All %d Docker services are healthy after %d checks", len(dockerComposeServicesForHealthCheck), checkCount)

			return nil
		}

		Log(im.logger, "⏳ Waiting %v before next health check... (%d unhealthy: %v)",
			cryptoutilSharedMagic.TestTimeoutServiceRetry, len(unhealthyServices), unhealthyServices)
		time.Sleep(cryptoutilSharedMagic.TestTimeoutServiceRetry)
	}
}

// areDockerServicesHealthy checks all Docker services health status.
func (im *InfrastructureManager) areDockerServicesHealthy(ctx context.Context, services []ServiceAndJob) map[string]bool {
	output, err := runDockerComposeCommand(ctx, im.logger, dockerComposeDescBatchHealth, dockerComposeArgsPsServices)
	if err != nil {
		Log(im.logger, "❌ Failed to check services health: %v", err)

		healthStatus := make(map[string]bool)

		for _, service := range services {
			// Determine the key to use for this service/job
			if service.Service != "" {
				healthStatus[service.Service] = false
			} else {
				healthStatus[service.Job] = false
			}
		}

		return healthStatus
	}

	serviceMap, err := parseDockerComposePsOutput(output)
	if err != nil {
		Log(im.logger, "❌ Failed to parse Docker compose output: %v", err)

		healthStatus := make(map[string]bool)

		for _, service := range services {
			// Determine the key to use for this service/job
			if service.Service != "" {
				healthStatus[service.Service] = false
			} else {
				healthStatus[service.Job] = false
			}
		}

		return healthStatus
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	// Add logging for all use cases
	for _, service := range services {
		if service.Service == "" && service.Job != "" {
			// Use case 1: Job-only (standalone job)
			jobName := service.Job
			jobData, exists := serviceMap[jobName]

			if exists {
				if state, ok := jobData["State"].(string); ok && state == cryptoutilSharedMagic.DockerServiceStateExited {
					Log(im.logger, "✅ Standalone job %s is in exited state", jobName)

					var exitCode int
					if exitCodeFloat, ok := jobData["ExitCode"].(float64); ok {
						exitCode = int(exitCodeFloat)
						Log(im.logger, "✅ Standalone job %s ExitCode (float64): %d", jobName, exitCode)
					} else if exitCodeInt, ok := jobData["ExitCode"].(int); ok {
						exitCode = exitCodeInt
						Log(im.logger, "✅ Standalone job %s ExitCode (int): %d", jobName, exitCode)
					} else {
						Log(im.logger, "❌ Standalone job %s ExitCode field not found or wrong type", jobName)

						continue
					}

					if exitCode == 0 {
						Log(im.logger, "✅ Standalone job %s exited successfully with code 0", jobName)
					} else {
						Log(im.logger, "❌ Standalone job %s exited with non-zero code: %d", jobName, exitCode)
					}
				} else {
					if state == cryptoutilSharedMagic.DockerServiceStateRunning {
						Log(im.logger, "❌ Standalone job %s should not be running continuously", jobName)
					} else {
						Log(im.logger, "❌ Standalone job %s in unexpected state: %s", jobName, state)
					}
				}
			} else {
				Log(im.logger, "🔍 Standalone job %s not found (completed successfully)", jobName)
			}
		} else if service.Service != "" && service.Job != "" {
			// Use case 3: Service with healthcheck job
			jobName := service.Job
			jobData, exists := serviceMap[jobName]

			if exists {
				if state, ok := jobData["State"].(string); ok && state == cryptoutilSharedMagic.DockerServiceStateExited {
					Log(im.logger, "✅ Healthcheck job %s for service %s is in exited state", jobName, service.Service)

					var exitCode int
					if exitCodeFloat, ok := jobData["ExitCode"].(float64); ok {
						exitCode = int(exitCodeFloat)
						Log(im.logger, "✅ Healthcheck job %s ExitCode (float64): %d", jobName, exitCode)
					} else if exitCodeInt, ok := jobData["ExitCode"].(int); ok {
						exitCode = exitCodeInt
						Log(im.logger, "✅ Healthcheck job %s ExitCode (int): %d", jobName, exitCode)
					} else {
						Log(im.logger, "❌ Healthcheck job %s ExitCode field not found or wrong type", jobName)

						continue
					}

					if exitCode == 0 {
						Log(im.logger, "✅ Healthcheck job %s exited successfully with code 0", jobName)
					} else {
						Log(im.logger, "❌ Healthcheck job %s exited with non-zero code: %d", jobName, exitCode)
					}
				} else {
					if state == cryptoutilSharedMagic.DockerServiceStateRunning {
						Log(im.logger, "❌ Healthcheck job %s should not be running continuously", jobName)
					} else {
						Log(im.logger, "❌ Healthcheck job %s in unexpected state: %s", jobName, state)
					}
				}
			} else {
				Log(im.logger, "🔍 Healthcheck job %s for service %s not found (completed successfully)", jobName, service.Service)
			}
		} else if service.Service != "" && service.Job == "" {
			// Use case 2: Service-only (native health checks)
			if !healthStatus[service.Service] {
				if _, exists := serviceMap[service.Service]; !exists {
					Log(im.logger, "❌ Service %s not found in docker compose output", service.Service)
				}
			}
		}
	}

	return healthStatus
}

// WaitForServicesReachable waits for services to be reachable via HTTP.
func (im *InfrastructureManager) WaitForServicesReachable(ctx context.Context) error {
	// Verify cryptoutil ports are accessible
	if err := im.verifyCryptoutilPortsReachable(ctx); err != nil {
		return err
	}

	// Grafana excluded - requires --profile with-grafana (optional for E2E tests)

	// Wait for OTEL collector
	if err := im.waitForHTTPReady(ctx, cryptoutilSharedMagic.URLPrefixLocalhostHTTP+
		fmt.Sprintf("%d", cryptoutilSharedMagic.DefaultPublicPortOtelCollectorHealth)+"/",
		cryptoutilSharedMagic.TestTimeoutCryptoutilReady); err != nil {
		return fmt.Errorf("otel collector not ready: %w", err)
	}

	return nil
}

// verifyCryptoutilPortsReachable verifies HTTPS ports 8000, 8001, 8002 are accessible.
func (im *InfrastructureManager) verifyCryptoutilPortsReachable(ctx context.Context) error {
	// Check public API ports for basic connectivity (not health endpoints)
	publicPorts := []int{cryptoutilSharedMagic.KMSServicePort, cryptoutilSharedMagic.KMSE2EPostgreSQL1PublicPort, cryptoutilSharedMagic.KMSE2EPostgreSQL2PublicPort}
	for _, port := range publicPorts {
		url := fmt.Sprintf("https://localhost:%d/ui/swagger", port)
		Log(im.logger, "🔍 Checking public port %d at %s...", port, url)

		client := CreateInsecureHTTPClient()

		req, err := CreateHTTPGetRequest(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to create request for public port %d: %w", port, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("public port %d not reachable: %w", port, err)
		}

		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("public port %d returned status %d", port, resp.StatusCode)
		}

		Log(im.logger, "✅ Public port %d is reachable", port)
	}

	return nil
}

// waitForHTTPReady waits for an HTTP endpoint to return 200.
func (im *InfrastructureManager) waitForHTTPReady(ctx context.Context, url string, timeout time.Duration) error {
	giveUpTime := time.Now().UTC().Add(timeout)
	client := CreateHTTPClient()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s", url)
		default:
		}

		if time.Now().UTC().After(giveUpTime) {
			return fmt.Errorf("service not ready after %v: %s", timeout, url)
		}

		req, err := CreateHTTPGetRequest(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to create request to %s: %w", url, err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()

			return nil
		}

		if resp != nil {
			_ = resp.Body.Close()
		}

		time.Sleep(cryptoutilSharedMagic.TestTimeoutHTTPRetryInterval)
	}
}

// logServiceHealthStatus logs the health status of services and returns unhealthy services.
func logServiceHealthStatus(startTime time.Time, healthStatus map[string]bool) []string {
	var unhealthyServices []string

	for service, healthy := range healthStatus {
		status := "❌ UNHEALTHY"
		if healthy {
			status = "✅ HEALTHY"
		}

		fmt.Printf("[%s] [%v]    %s: %s\n",
			time.Now().UTC().Format("15:04:05"),
			time.Since(startTime).Round(time.Second),
			service, status)

		if !healthy {
			unhealthyServices = append(unhealthyServices, service)
		}
	}

	return unhealthyServices
}

// showFailedContainerLogs captures and displays recent logs from containers that have failed to start.
func (im *InfrastructureManager) showFailedContainerLogs(ctx context.Context) error {
	// Get list of containers that match our compose project
	containers, err := getDockerContainers(ctx, im.logger)
	if err != nil {
		return fmt.Errorf("failed to get Docker containers: %w", err)
	}

	if len(containers) == 0 {
		Log(im.logger, "⚠️ No containers found to show logs for")

		return nil
	}

	Log(im.logger, "📋 Showing recent logs from %d containers...", len(containers))

	for _, container := range containers {
		Log(im.logger, "📄 Recent logs from container: %s", container.Name)

		// Get last 50 lines of logs with timestamps
		cmd := exec.CommandContext(ctx, "docker", "logs", "--tail", "50", "--timestamps", container.Name)

		output, err := cmd.CombinedOutput()
		if err != nil {
			Log(im.logger, "⚠️ Failed to get logs for container %s: %v", container.Name, err)

			continue
		}

		// Split output into lines and log each line
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				Log(im.logger, "  %s: %s", container.Name, line)
			}
		}

		Log(im.logger, "📄 End of logs for container: %s", container.Name)
	}

	return nil
}

// Helper methods.
