//go:build e2e

package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
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
	giveUpTime := time.Now().Add(cryptoutilMagic.TestTimeoutDockerHealth)
	checkCount := 0

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for Docker services")
		default:
		}

		if time.Now().After(giveUpTime) {
			return fmt.Errorf("docker services not healthy after %v", cryptoutilMagic.TestTimeoutDockerHealth)
		}

		checkCount++
		im.logger.Log("🔍 Health check #%d: Checking %d services...", checkCount, len(dockerComposeServicesForHealthCheck))

		healthStatus := im.areDockerServicesHealthy(ctx, dockerComposeServicesForHealthCheck)

		unhealthyServices := logServiceHealthStatus(im.startTime, healthStatus)

		allHealthy := len(unhealthyServices) == 0
		if allHealthy {
			im.logger.Log("✅ All %d Docker services are healthy after %d checks", len(dockerComposeServicesForHealthCheck), checkCount)

			return nil
		}

		im.logger.Log("⏳ Waiting %v before next health check... (%d unhealthy: %v)",
			cryptoutilMagic.TestTimeoutServiceRetry, len(unhealthyServices), unhealthyServices)
		time.Sleep(cryptoutilMagic.TestTimeoutServiceRetry)
	}
}

// areDockerServicesHealthy checks all Docker services health status.
func (im *InfrastructureManager) areDockerServicesHealthy(ctx context.Context, services []ServiceNameAndJob) map[string]bool {
	output, err := runDockerComposeCommand(ctx, im.logger, dockerComposeDescBatchHealth, dockerComposeArgsPsServices)
	if err != nil {
		im.logger.Log("❌ Failed to check services health: %v", err)

		healthStatus := make(map[string]bool)
		for _, service := range services {
			healthStatus[service.Name] = false
		}

		return healthStatus
	}

	serviceMap, err := parseDockerComposePsOutput(output)
	if err != nil {
		im.logger.Log("❌ Failed to parse Docker compose output: %v", err)

		healthStatus := make(map[string]bool)
		for _, service := range services {
			healthStatus[service.Name] = false
		}

		return healthStatus
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	// Add logging for healthcheck service special cases
	for _, service := range services {
		if service.HealthcheckJob != "" {
			// This service uses a healthcheck job
			jobName := service.HealthcheckJob
			jobData, exists := serviceMap[jobName]

			if exists {
				if state, ok := jobData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateExited {
					im.logger.Log("✅ Healthcheck job %s is in exited state", jobName)

					// Handle both int and float64 types for ExitCode
					var exitCode int
					if exitCodeFloat, ok := jobData["ExitCode"].(float64); ok {
						exitCode = int(exitCodeFloat)
						im.logger.Log("✅ Healthcheck job %s ExitCode (float64): %d", jobName, exitCode)
					} else if exitCodeInt, ok := jobData["ExitCode"].(int); ok {
						exitCode = exitCodeInt
						im.logger.Log("✅ Healthcheck job %s ExitCode (int): %d", jobName, exitCode)
					} else {
						im.logger.Log("❌ Healthcheck job %s ExitCode field not found or wrong type", jobName)

						continue
					}

					if exitCode == 0 {
						im.logger.Log("✅ Healthcheck job %s exited successfully with code 0", jobName)
					} else {
						im.logger.Log("❌ Healthcheck job %s exited with non-zero code: %d", jobName, exitCode)
					}
				} else {
					if state == cryptoutilMagic.DockerServiceStateRunning {
						im.logger.Log("❌ Healthcheck job %s should not be running continuously", jobName)
					} else {
						im.logger.Log("❌ Healthcheck job %s in unexpected state: %s", jobName, state)
					}
				}
			} else {
				im.logger.Log("🔍 Healthcheck job %s not found (completed successfully)", jobName)
			}
		} else if !healthStatus[service.Name] {
			// Log when regular services are not found
			if _, exists := serviceMap[service.Name]; !exists {
				im.logger.Log("❌ Service %s not found in docker compose output", service.Name)
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

	// Wait for Grafana
	if err := im.waitForHTTPReady(ctx, cryptoutilMagic.URLPrefixLocalhostHTTP+
		fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortGrafana)+"/api/health",
		cryptoutilMagic.TestTimeoutCryptoutilReady); err != nil {
		return fmt.Errorf("grafana not ready: %w", err)
	}

	// Wait for OTEL collector
	if err := im.waitForHTTPReady(ctx, cryptoutilMagic.URLPrefixLocalhostHTTP+
		fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortOtelCollectorHealth)+"/",
		cryptoutilMagic.TestTimeoutCryptoutilReady); err != nil {
		return fmt.Errorf("otel collector not ready: %w", err)
	}

	return nil
}

// verifyCryptoutilPortsReachable verifies HTTPS ports 8080, 8081, 8082 are accessible.
func (im *InfrastructureManager) verifyCryptoutilPortsReachable(ctx context.Context) error {
	// Check public API ports for basic connectivity (not health endpoints)
	publicPorts := []int{8080, 8081, 8082}
	for _, port := range publicPorts {
		url := fmt.Sprintf("https://localhost:%d/ui/swagger", port)
		im.logger.Log("🔍 Checking public port %d at %s...", port, url)

		client := CreateInsecureHTTPClient()

		req, err := CreateHTTPGetRequest(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to create request for public port %d: %w", port, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("public port %d not reachable: %w", port, err)
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("public port %d returned status %d", port, resp.StatusCode)
		}

		im.logger.Log("✅ Public port %d is reachable", port)
	}

	return nil
}

// waitForHTTPReady waits for an HTTP endpoint to return 200.
func (im *InfrastructureManager) waitForHTTPReady(ctx context.Context, url string, timeout time.Duration) error {
	giveUpTime := time.Now().Add(timeout)
	client := CreateHTTPClient()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s", url)
		default:
		}

		if time.Now().After(giveUpTime) {
			return fmt.Errorf("service not ready after %v: %s", timeout, url)
		}

		req, err := CreateHTTPGetRequest(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to create request to %s: %w", url, err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			im.logger.Log("✅ Service ready at %s", url)

			return nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(cryptoutilMagic.TestTimeoutHTTPRetryInterval)
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
			time.Now().Format("15:04:05"),
			time.Since(startTime).Round(time.Second),
			service, status)

		if !healthy {
			unhealthyServices = append(unhealthyServices, service)
		}
	}

	return unhealthyServices
}

// Helper methods.
