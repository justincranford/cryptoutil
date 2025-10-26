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
	dockerComposeArgsStartServices = []string{"up", "-d", "--force-recreate"}
	dockerComposeArgsPsServices    = []string{"ps", "-a", "--format", "json"}
)

// Docker compose command description constants.
const (
	dockerComposeDescStopServices  = "Stop services"
	dockerComposeDescStartServices = "Start services"
	dockerComposeDescBatchHealth   = "Batch health check"
)

// Docker compose service names for batch health checking.
var (
	dockerComposeServicesForHealthCheck = []string{
		cryptoutilMagic.DockerServiceCryptoutilSqlite,
		cryptoutilMagic.DockerServiceCryptoutilPostgres1,
		cryptoutilMagic.DockerServiceCryptoutilPostgres2,
		cryptoutilMagic.DockerServicePostgres,
		cryptoutilMagic.DockerServiceGrafanaOtelLgtm,
		cryptoutilMagic.DockerServiceOtelCollectorHealthcheck,
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
	services := dockerComposeServicesForHealthCheck

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
		im.log("🔍 Health check #%d: Checking %d services...", checkCount, len(services))

		healthStatus := im.areDockerServicesHealthy(ctx, services)

		unhealthyServices := []string{}

		for service, healthy := range healthStatus {
			status := "❌ UNHEALTHY"
			if healthy {
				status = "✅ HEALTHY"
			}

			fmt.Printf("[%s] [%v]    %s: %s\n",
				time.Now().Format("15:04:05"),
				time.Since(im.startTime).Round(time.Second),
				service, status)

			if !healthy {
				unhealthyServices = append(unhealthyServices, service)
			}
		}

		allHealthy := len(unhealthyServices) == 0
		if allHealthy {
			im.log("✅ All %d Docker services are healthy after %d checks", len(services), checkCount)

			return nil
		}

		im.log("⏳ Waiting %v before next health check... (%d unhealthy: %v)",
			cryptoutilMagic.TestTimeoutServiceRetry, len(unhealthyServices), unhealthyServices)
		time.Sleep(cryptoutilMagic.TestTimeoutServiceRetry)
	}
}

// areDockerServicesHealthy checks all Docker services health status.
func (im *InfrastructureManager) areDockerServicesHealthy(ctx context.Context, services []string) map[string]bool {
	healthStatus := make(map[string]bool)

	output, err := im.runDockerComposeCommand(ctx, dockerComposeDescBatchHealth, dockerComposeArgsPsServices)
	if err != nil {
		im.log("❌ Failed to check services health: %v", err)

		for _, service := range services {
			healthStatus[service] = false
		}

		return healthStatus
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	serviceList := make([]map[string]any, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var service map[string]any
		if err := json.Unmarshal([]byte(line), &service); err != nil {
			continue
		}

		serviceList = append(serviceList, service)
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

	for _, serviceName := range services {
		service, exists := serviceMap[serviceName]
		if !exists {
			im.log("❌ Service %s not found in docker compose output", serviceName)

			healthStatus[serviceName] = false

			continue
		}

		if health, ok := service["Health"].(string); ok && health != "" {
			healthStatus[serviceName] = health == "healthy"
		} else if serviceName == cryptoutilMagic.DockerServiceOtelCollectorHealthcheck {
			im.log("🔍 Checking healthcheck service: %s", serviceName)
			if state, ok := service["State"].(string); ok && state == "exited" {
				im.log("✅ Healthcheck service is in exited state")
				// Handle both int and float64 types for ExitCode
				var exitCode int
				if exitCodeFloat, ok := service["ExitCode"].(float64); ok {
					exitCode = int(exitCodeFloat)
					im.log("✅ Healthcheck service ExitCode (float64): %d", exitCode)
				} else if exitCodeInt, ok := service["ExitCode"].(int); ok {
					exitCode = exitCodeInt
					im.log("✅ Healthcheck service ExitCode (int): %d", exitCode)
				} else {
					im.log("❌ Healthcheck service ExitCode field not found or wrong type")
					healthStatus[serviceName] = false

					continue
				}

				healthStatus[serviceName] = false
				if exitCode == 0 {
					im.log("✅ Healthcheck service exited successfully with code 0")
					healthStatus[serviceName] = true
				} else {
					im.log("❌ Healthcheck service exited with non-zero code: %d", exitCode)
				}
			} else {
				healthStatus[serviceName] = false

				if state == "running" {
					im.log("❌ Healthcheck service should not be running continuously")
				} else {
					im.log("❌ Healthcheck service in unexpected state: %s", state)
				}
			}
		} else {
			if state, ok := service["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateRunning {
				healthStatus[serviceName] = true
			} else {
				healthStatus[serviceName] = false
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
		fmt.Sprintf("%d", cryptoutilMagic.DefaultPublicPortInternalMetrics)+"/metrics",
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
