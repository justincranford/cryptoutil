//go:build e2e

package test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// Docker compose command arguments constants.
var (
	dockerComposeArgsStopServices  = []string{"down", "-v", "--remove-orphans"}
	dockerComposeArgsStartServices = []string{"up", "-d", "--force-recreate", "--build"}
	dockerComposeArgsPsServices    = []string{"ps", "-a", "--format", "json"}
)

// Docker compose command description constants.
const (
	dockerComposeDescStopServices  = "Stop services"
	dockerComposeDescStartServices = "Start services"
	dockerComposeDescBatchHealth   = "Batch health check"
)

// ServiceNameAndJob represents a service and its optional healthcheck job.
type ServiceNameAndJob struct {
	Name           string // The service name to check
	HealthcheckJob string // Optional healthcheck job name (empty if no healthcheck job)
}

// Docker compose service names for batch health checking.
var (
	dockerComposeServicesForHealthCheck = []ServiceNameAndJob{
		{Name: cryptoutilMagic.DockerServiceCryptoutilSqlite},
		{Name: cryptoutilMagic.DockerServiceCryptoutilPostgres1},
		{Name: cryptoutilMagic.DockerServiceCryptoutilPostgres2},
		{Name: cryptoutilMagic.DockerServicePostgres},
		{Name: cryptoutilMagic.DockerServiceGrafanaOtelLgtm},
		{Name: cryptoutilMagic.DockerServiceOtelCollector, HealthcheckJob: cryptoutilMagic.DockerJobOtelCollectorHealthcheck},
	}
)

// InfrastructureManager handles Docker Compose operations and service management.
type InfrastructureManager struct {
	startTime time.Time
	logFile   *os.File
}

// NewInfrastructureManager creates a new infrastructure manager.
func NewInfrastructureManager(startTime time.Time, logFile *os.File) *InfrastructureManager {
	return &InfrastructureManager{
		startTime: startTime,
		logFile:   logFile,
	}
}

// StopServices stops Docker Compose services.
func (im *InfrastructureManager) StopServices(ctx context.Context) error {
	output, err := im.runDockerComposeCommand(ctx, dockerComposeDescStopServices, dockerComposeArgsStopServices)
	if err != nil {
		return fmt.Errorf("failed to stop services: %w, output: %s", err, string(output))
	}

	return nil
}

// StartServices starts Docker Compose services.
func (im *InfrastructureManager) StartServices(ctx context.Context) error {
	output, err := im.runDockerComposeCommand(ctx, dockerComposeDescStartServices, dockerComposeArgsStartServices)
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
		im.log("🔍 Health check #%d: Checking %d services...", checkCount, len(dockerComposeServicesForHealthCheck))

		healthStatus := im.areDockerServicesHealthy(ctx, dockerComposeServicesForHealthCheck)

		unhealthyServices := logServiceHealthStatus(im.startTime, healthStatus)

		allHealthy := len(unhealthyServices) == 0
		if allHealthy {
			im.log("✅ All %d Docker services are healthy after %d checks", len(dockerComposeServicesForHealthCheck), checkCount)

			return nil
		}

		im.log("⏳ Waiting %v before next health check... (%d unhealthy: %v)",
			cryptoutilMagic.TestTimeoutServiceRetry, len(unhealthyServices), unhealthyServices)
		time.Sleep(cryptoutilMagic.TestTimeoutServiceRetry)
	}
}

// areDockerServicesHealthy checks all Docker services health status.
func (im *InfrastructureManager) areDockerServicesHealthy(ctx context.Context, services []ServiceNameAndJob) map[string]bool {
	output, err := im.runDockerComposeCommand(ctx, dockerComposeDescBatchHealth, dockerComposeArgsPsServices)
	if err != nil {
		im.log("❌ Failed to check services health: %v", err)

		healthStatus := make(map[string]bool)
		for _, service := range services {
			healthStatus[service.Name] = false
		}

		return healthStatus
	}

	serviceMap, err := parseDockerComposePsOutput(output)
	if err != nil {
		im.log("❌ Failed to parse Docker compose output: %v", err)

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
					im.log("✅ Healthcheck job %s is in exited state", jobName)

					// Handle both int and float64 types for ExitCode
					var exitCode int
					if exitCodeFloat, ok := jobData["ExitCode"].(float64); ok {
						exitCode = int(exitCodeFloat)
						im.log("✅ Healthcheck job %s ExitCode (float64): %d", jobName, exitCode)
					} else if exitCodeInt, ok := jobData["ExitCode"].(int); ok {
						exitCode = exitCodeInt
						im.log("✅ Healthcheck job %s ExitCode (int): %d", jobName, exitCode)
					} else {
						im.log("❌ Healthcheck job %s ExitCode field not found or wrong type", jobName)

						continue
					}

					if exitCode == 0 {
						im.log("✅ Healthcheck job %s exited successfully with code 0", jobName)
					} else {
						im.log("❌ Healthcheck job %s exited with non-zero code: %d", jobName, exitCode)
					}
				} else {
					if state == cryptoutilMagic.DockerServiceStateRunning {
						im.log("❌ Healthcheck job %s should not be running continuously", jobName)
					} else {
						im.log("❌ Healthcheck job %s in unexpected state: %s", jobName, state)
					}
				}
			} else {
				im.log("🔍 Healthcheck job %s not found (completed successfully)", jobName)
			}
		} else if !healthStatus[service.Name] {
			// Log when regular services are not found
			if _, exists := serviceMap[service.Name]; !exists {
				im.log("❌ Service %s not found in docker compose output", service.Name)
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
		im.log("🔍 Checking public port %d at %s...", port, url)

		client := &http.Client{
			Timeout: cryptoutilMagic.DockerHTTPClientTimeoutSeconds * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

		im.log("✅ Public port %d is reachable", port)
	}

	return nil
}

// waitForHTTPReady waits for an HTTP endpoint to return 200.
func (im *InfrastructureManager) waitForHTTPReady(ctx context.Context, url string, timeout time.Duration) error {
	giveUpTime := time.Now().Add(timeout)
	client := &http.Client{Timeout: cryptoutilMagic.TestTimeoutHTTPClient}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s", url)
		default:
		}

		if time.Now().After(giveUpTime) {
			return fmt.Errorf("service not ready after %v: %s", timeout, url)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request to %s: %w", url, err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			im.log("✅ Service ready at %s", url)

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
func (im *InfrastructureManager) log(format string, args ...any) {
	message := fmt.Sprintf("[%s] [%v] %s\n",
		time.Now().Format("15:04:05"),
		time.Since(im.startTime).Round(time.Second),
		fmt.Sprintf(format, args...))

	// Write to console
	fmt.Print(message)

	// Write to log file if available
	if im.logFile != nil {
		if _, err := im.logFile.WriteString(message); err != nil {
			// If we can't write to the log file, at least write to console
			fmt.Printf("⚠️ Failed to write to log file: %v\n", err)
		}
	}
}

func (im *InfrastructureManager) logCommand(description, command, output string) {
	im.log("📋 [%s] %s", description, command)

	if output != "" {
		im.log("📋 [%s] Output: %s", description, strings.TrimSpace(output))
	}
}

// getComposeFilePath returns the compose file path appropriate for the current OS.
// Since E2E tests run from test/e2e/ directory, we need to navigate up to project root.
func (im *InfrastructureManager) getComposeFilePath() string {
	// Navigate up from test/e2e/ to project root, then to deployments/compose/compose.yml
	projectRoot := filepath.Join("..", "..")
	composePath := filepath.Join(projectRoot, "deployments", "compose", "compose.yml")

	// Convert to absolute path to ensure it works regardless of working directory
	absPath, err := filepath.Abs(composePath)
	if err != nil {
		// Fallback to relative path if absolute path fails
		if runtime.GOOS == "windows" {
			return cryptoutilMagic.DockerComposeRelativeFilePathWindows
		}

		return cryptoutilMagic.DockerComposeRelativeFilePathLinux
	}

	return absPath
}

// parseDockerComposePsOutput parses docker compose ps --format json output into a service name to service data map.
func parseDockerComposePsOutput(output []byte) (map[string]map[string]any, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	serviceList := make([]map[string]any, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var service map[string]any
		if err := json.Unmarshal([]byte(line), &service); err != nil {
			return nil, fmt.Errorf("failed to parse Docker service JSON: %w", err)
		}

		serviceList = append(serviceList, service)
	}

	if len(serviceList) == 0 {
		return nil, fmt.Errorf("no services found in Docker compose output")
	}

	serviceMap := make(map[string]map[string]any)

	for _, service := range serviceList {
		if name, ok := service["Name"].(string); ok {
			if strings.Contains(name, "compose-") {
				parts := strings.Split(name, "-")
				if len(parts) >= cryptoutilMagic.DockerServiceNamePartsMin {
					serviceName := strings.Join(parts[1:len(parts)-1], "-")
					serviceMap[serviceName] = service
				}
			}
		}
	}

	return serviceMap, nil
}

// determineServiceHealthStatus determines the health status of services from a service map.
// This function contains the shared logic for determining if services are healthy.
// It handles services with and without healthcheck jobs.
func determineServiceHealthStatus(serviceMap map[string]map[string]any, services []ServiceNameAndJob) map[string]bool {
	healthStatus := make(map[string]bool)

	for _, service := range services {
		var serviceNameToCheck string
		if service.HealthcheckJob != "" {
			// Use the healthcheck job to determine service health
			serviceNameToCheck = service.HealthcheckJob
		} else {
			// Check the service directly
			serviceNameToCheck = service.Name
		}

		serviceData, exists := serviceMap[serviceNameToCheck]
		if !exists {
			// Service/job not found
			if service.HealthcheckJob != "" {
				// For healthcheck jobs: not found means it completed successfully and was cleaned up
				healthStatus[service.Name] = true
			} else {
				// For regular services: not found means unhealthy
				healthStatus[service.Name] = false
			}

			continue
		}

		if service.HealthcheckJob != "" {
			// This is a healthcheck job - check if it exited successfully
			if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateExited {
				// Handle both int and float64 types for ExitCode
				var exitCode int
				if exitCodeFloat, ok := serviceData["ExitCode"].(float64); ok {
					exitCode = int(exitCodeFloat)
				} else if exitCodeInt, ok := serviceData["ExitCode"].(int); ok {
					exitCode = exitCodeInt
				} else {
					// ExitCode field not found or wrong type
					healthStatus[service.Name] = false

					continue
				}

				healthStatus[service.Name] = exitCode == 0
			} else {
				// Not in exited state
				healthStatus[service.Name] = false
			}
		} else {
			// Regular service - check health or running state
			if health, ok := serviceData["Health"].(string); ok && health != "" {
				// Services with health checks
				healthStatus[service.Name] = health == cryptoutilMagic.DockerServiceHealthHealthy
			} else {
				// Services without health checks: check if running
				if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateRunning {
					healthStatus[service.Name] = true
				} else {
					healthStatus[service.Name] = false
				}
			}
		}
	}

	return healthStatus
}

// runDockerComposeCommand executes a docker compose command with the given arguments.
//
//	Windows: docker compose -f .\deployments\compose\compose.yml <command> <args>
//	Linux:   docker compose -f ./deployments/compose/compose.yml <command> <args>
//
// Always use relative path for cross-platform compatibility in
// in GitHub Actions (Ubuntu runners) and Windows (`act` runner).
func (im *InfrastructureManager) runDockerComposeCommand(ctx context.Context, description string, args []string) ([]byte, error) {
	// Log start message based on description
	switch description {
	case dockerComposeDescStopServices:
		im.log("🧹 Stopping Docker Compose services")
	case dockerComposeDescStartServices:
		im.log("🚀 Starting Docker Compose services")
	}

	composeFile := im.getComposeFilePath()
	allArgs := append([]string{"docker", "compose", "-f", composeFile}, args...)
	cmd := exec.CommandContext(ctx, allArgs[0], allArgs[1:]...)
	output, err := cmd.CombinedOutput()
	im.logCommand(description, cmd.String(), string(output))

	if err != nil {
		return output, fmt.Errorf("docker compose command failed: %w", err)
	}

	// Log success message based on description
	switch description {
	case dockerComposeDescStopServices:
		im.log("✅ Existing services stopped successfully")
	case dockerComposeDescStartServices:
		im.log("✅ Docker Compose services started successfully")
	}

	return output, nil
}
